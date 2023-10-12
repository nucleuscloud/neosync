package datasync

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"

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
		StartToCloseTimeout: 2 * time.Minute, // this will need to be drastically increased and probably settable via the UI
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
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

	started := map[string]struct{}{}
	completed := map[string]struct{}{}

	workselector := workflow.NewSelector(ctx)

	splitConfigs := splitBenthosConfigs(bcResp.BenthosConfigs)
	var activityErr error
	childctx, cancelHandler := workflow.WithCancel(ctx)

	for _, bc := range splitConfigs.Root {
		bc := bc
		future := invokeSync(bc, childctx, started, completed, logger)
		workselector.AddFuture(future, func(f workflow.Future) {
			logger.Info("config sync completed", "name", bc.Name)
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
		logger.Debug("*** blocking select ***", "i", i)
		workselector.Select(ctx)
		if activityErr != nil {
			return nil, activityErr
		}
		logger.Debug("*** post select ***", "i", i)

		// todo: deadlock detection
		for _, bc := range splitConfigs.Dependents {
			bc := bc
			if _, ok := started[bc.Name]; ok {
				continue
			}
			if !isConfigReady(bc, completed) {
				continue
			}
			future := invokeSync(bc, childctx, started, completed, logger)
			workselector.AddFuture(future, func(f workflow.Future) {
				logger.Info("config sync completed", "name", bc.Name)
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

	logger.Info("workflow completed")

	return &WorkflowResponse{}, nil
}

func getSyncMetadata(config *benthosConfigResponse) *SyncMetadata {
	names := strings.Split(config.Name, ".")
	schema, table := names[0], names[1]
	return &SyncMetadata{Schema: schema, Table: table}
}

func invokeSync(
	config *benthosConfigResponse,
	ctx workflow.Context,
	started map[string]struct{},
	completed map[string]struct{},
	logger log.Logger,
) workflow.Future {
	metadata := getSyncMetadata(config)
	future, settable := workflow.NewFuture(ctx)
	logger.Info("triggering config sync", "name", config.Name, "metadata", metadata)
	started[config.Name] = struct{}{}
	var wfActivites *Activities
	workflow.GoNamed(ctx, config.Name, func(ctx workflow.Context) {
		configbits, err := yaml.Marshal(config.Config)
		if err != nil {
			logger.Error("unable to marshal benthos config", "err", err)
			settable.SetError(fmt.Errorf("unable to marshal benthos config: %w", err))
			return
		}
		var result SyncResponse
		err = workflow.ExecuteActivity(
			ctx,
			wfActivites.Sync,
			&SyncRequest{BenthosConfig: string(configbits)}, metadata).Get(ctx, &result)
		completed[config.Name] = struct{}{}
		settable.Set(result, err)
	})
	return future
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
