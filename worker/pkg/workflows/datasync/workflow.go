package datasync

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.temporal.io/sdk/workflow"
	"gopkg.in/yaml.v3"
)

type WorkflowRequest struct {
	JobId string
}

type WorkflowResponse struct{}

func Workflow(wfctx workflow.Context, req *WorkflowRequest) (*WorkflowResponse, error) {
	neosyncUrl := viper.GetString("NEOSYNC_URL")
	if neosyncUrl == "" {
		neosyncUrl = "localhost:8080"
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	wfinfo := workflow.GetInfo(wfctx)

	ctx := workflow.WithActivityOptions(wfctx, ao)
	logger := workflow.GetLogger(ctx)
	_ = logger

	var wfActivites *Activities
	var bcResp *GenerateBenthosConfigsResponse
	err := workflow.ExecuteActivity(ctx, wfActivites.GenerateBenthosConfigs, &GenerateBenthosConfigsRequest{
		JobId:      req.JobId,
		BackendUrl: neosyncUrl,
		WorkflowId: wfinfo.WorkflowExecution.ID,
	}).Get(ctx, &bcResp)
	if err != nil {
		return nil, err
	}

	// benthos jobs that have been completed
	completed := map[string]struct{}{}
	started := map[string]struct{}{}
	// syncChan := workflow.NewChannel(wfctx)
	workselector := workflow.NewSelector(ctx)

	wg := workflow.NewWaitGroup(ctx)
	wg.Add(len(bcResp.BenthosConfigs))

	// workflow.GoNamed(ctx, "ConfigSync", func(ctx workflow.Context) {
	// 	var bcName string
	// 	syncChan.Receive(ctx, &bcName)

	// 	if bcName == "" {
	// 		wg.Done()
	// 	}

	// 	logger.Info("benthos config has completed", "name", bcName)
	// 	completed[bcName] = struct{}{}

	// 	logger.Info(fmt.Sprintf("all completed configs: %s", strings.Join(getMapKeys(completed), ",")))

	// 	for len(completed) != len(bcResp.BenthosConfigs) {
	// 		for _, bc := range bcResp.BenthosConfigs {
	// 			bc := bc
	// 			if _, ok := completed[bc.Name]; ok {
	// 				continue
	// 			}
	// 			isready := true
	// 			for _, dep := range bc.DependsOn {
	// 				if _, ok := completed[dep]; !ok {
	// 					isready = false
	// 					break
	// 				}
	// 			}
	// 			if !isready {
	// 				continue
	// 			}
	// 			configbits, _ := yaml.Marshal(bc.Config)
	// 			if err != nil {
	// 				logger.Error("unable to marshal benthos config", "err", err)
	// 				// return nil, fmt.Errorf("unable to marshal benthos config: %w", err)
	// 			}
	// 			future, settable := workflow.NewFuture(ctx)
	// 			workflow.Go(ctx, func(ctx workflow.Context) {
	// 				var result SyncResponse
	// 				err := workflow.ExecuteActivity(ctx, wfActivites.Sync, &SyncRequest{BenthosConfig: string(configbits)}).Get(ctx, &result)
	// 				if err != nil {
	// 					settable.SetError(err)
	// 				} else {
	// 					settable.SetValue(result)
	// 					syncChan.Send(ctx, bc.Name)
	// 				}
	// 			})
	// 			workselector.AddFuture(future, func(f workflow.Future) {
	// 				var result SyncResponse
	// 				err := f.Get(ctx, &result)
	// 				if err != nil {
	// 					logger.Error("activity did not complete", "err", err)
	// 					// todo: cancel workflow
	// 				}
	// 			})
	// 		}
	// 	}
	// 	wg.Done()
	// })

	splitConfigs := splitBenthosConfigs(bcResp.BenthosConfigs)
	var activityErr error
	childctx, cancelHandler := workflow.WithCancel(ctx)

	for _, bc := range splitConfigs.Root {
		bc := bc
		configbits, err := yaml.Marshal(bc.Config)
		if err != nil {
			logger.Error("unable to marshal benthos config", "err", err)
			return nil, fmt.Errorf("unable to marshal benthos config: %w", err)
		}
		future, settable := workflow.NewFuture(childctx)
		logger.Info("triggering config sync", "name", bc.Name)
		started[bc.Name] = struct{}{}
		workflow.GoNamed(childctx, bc.Name, func(ctx workflow.Context) {
			defer wg.Done()
			var result SyncResponse
			err := workflow.ExecuteActivity(ctx, wfActivites.Sync, &SyncRequest{BenthosConfig: string(configbits)}).Get(ctx, &result)
			completed[bc.Name] = struct{}{}
			settable.Set(result, err)
		})
		workselector.AddFuture(future, func(f workflow.Future) {
			logger.Info("config sync completed (future)", "name", bc.Name)
			var result SyncResponse
			err := f.Get(childctx, &result)
			if err != nil {
				logger.Error("activity did not complete", "err", err)
				cancelHandler()
				activityErr = err
			}
		})
	}

	for i := 0; i < len(bcResp.BenthosConfigs); i++ {
		logger.Info("*** blocking select ***", "i", i)
		workselector.Select(ctx)
		if activityErr != nil {
			return nil, fmt.Errorf("activity failed: %w", activityErr)
		}
		logger.Info("*** post select ***", "i", i)

		for _, bc := range splitConfigs.Dependents {
			bc := bc
			if _, ok := started[bc.Name]; ok {
				continue
			}
			if !isConfigReady(bc, completed) {
				continue
			}
			started[bc.Name] = struct{}{}

			configbits, err := yaml.Marshal(bc.Config)
			if err != nil {
				logger.Error("unable to marshal benthos config", "err", err)
				return nil, fmt.Errorf("unable to marshal benthos config: %w", err)
			}
			future, settable := workflow.NewFuture(childctx)
			logger.Info("triggering config sync", "name", bc.Name)
			workflow.GoNamed(childctx, bc.Name, func(ctx workflow.Context) {
				defer wg.Done()
				completed[bc.Name] = struct{}{}
				var result SyncResponse
				err := workflow.ExecuteActivity(ctx, wfActivites.Sync, &SyncRequest{BenthosConfig: string(configbits)}).Get(ctx, &result)
				settable.Set(result, err)
			})
			workselector.AddFuture(future, func(f workflow.Future) {
				logger.Info("config sync completed (future)", "name", bc.Name)
				var result SyncResponse
				err := f.Get(childctx, &result)
				if err != nil {
					logger.Error("activity did not complete", "err", err)
					cancelHandler()
					activityErr = err
				}
			})
		}
	}

	logger.Info("waiting for wg to finish")
	wg.Wait(ctx)

	logger.Info("workflow completed")

	return &WorkflowResponse{}, nil
}

func isConfigReady(config *benthosConfigResponse, completed map[string]struct{}) bool {
	if config == nil {
		return false
	}

	if len(config.DependsOn) == 0 {
		return true
	}
	for _, dep := range config.DependsOn {
		if _, ok := completed[dep]; !ok {
			return false
		}
	}
	return true
}

// func getMapKeys[T any](val map[string]T) []string {
// 	keys := make([]string, 0, len(val))
// 	for key := range val {
// 		keys = append(keys, key)
// 	}
// 	return keys
// }

type SplitConfigs struct {
	Root       []*benthosConfigResponse
	Dependents []*benthosConfigResponse
}

func splitBenthosConfigs(configs []*benthosConfigResponse) *SplitConfigs {
	out := &SplitConfigs{
		Root:       []*benthosConfigResponse{},
		Dependents: []*benthosConfigResponse{},
	}
	for _, cfg := range configs {
		if len(cfg.DependsOn) == 0 {
			out.Root = append(out.Root, cfg)
		} else {
			out.Dependents = append(out.Dependents, cfg)
		}
	}

	return out
}
