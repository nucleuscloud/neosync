package datasync_workflow

import (
	"fmt"
	"slices"
	"strings"
	"time"

	datasync_activities "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
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
	wfinfo := workflow.GetInfo(wfctx)

	ctx := workflow.WithActivityOptions(wfctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	logger := workflow.GetLogger(ctx)
	_ = logger

	workflowMetadata := &datasync_activities.WorkflowMetadata{
		WorkflowId: wfinfo.WorkflowExecution.ID,
		RunId:      wfinfo.WorkflowExecution.RunID,
	}

	var wfActivites *datasync_activities.Activities
	var bcResp *datasync_activities.GenerateBenthosConfigsResponse
	logger.Info("executing generate benthos configs activity")
	err := workflow.ExecuteActivity(ctx, wfActivites.GenerateBenthosConfigs, &datasync_activities.GenerateBenthosConfigsRequest{
		JobId: req.JobId,
	}, workflowMetadata).Get(ctx, &bcResp)
	if err != nil {
		return nil, err
	}

	if len(bcResp.BenthosConfigs) == 0 {
		return &WorkflowResponse{}, nil
	}

	var actOptResp *datasync_activities.RetrieveActivityOptionsResponse
	logger.Info("executing retrieval of activity options activity")
	ctx = workflow.WithActivityOptions(wfctx, workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
		},
	})
	err = workflow.ExecuteActivity(ctx, wfActivites.RetrieveActivityOptions, &datasync_activities.RetrieveActivityOptionsRequest{
		JobId: req.JobId,
	}, workflowMetadata).Get(ctx, &actOptResp)
	if err != nil {
		return nil, err
	}

	ctx = workflow.WithActivityOptions(wfctx, *actOptResp.SyncActivityOptions)
	logger.Info("executing running init statements in job destinations activity")
	var resp *datasync_activities.RunSqlInitTableStatementsResponse
	err = workflow.ExecuteActivity(ctx, wfActivites.RunSqlInitTableStatements, &datasync_activities.RunSqlInitTableStatementsRequest{
		JobId:      req.JobId,
		WorkflowId: wfinfo.WorkflowExecution.ID,
	}).Get(ctx, &resp)
	if err != nil {
		return nil, err
	}

	started := map[string]struct{}{}
	completed := map[string][]string{}

	workselector := workflow.NewSelector(ctx)

	splitConfigs := splitBenthosConfigs(bcResp.BenthosConfigs)
	var activityErr error
	childctx, cancelHandler := workflow.WithCancel(ctx)

	if len(splitConfigs.Root) == 0 && len(splitConfigs.Dependents) > 0 {
		return nil, fmt.Errorf("root config not found. unable to process configs")
	}
	for _, bc := range splitConfigs.Root {
		bc := bc
		future := invokeSync(bc, childctx, started, completed, workflowMetadata, logger)
		workselector.AddFuture(future, func(f workflow.Future) {
			logger.Info("config sync completed", "name", bc.Name)
			var result sync_activity.SyncResponse
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

			future := invokeSync(bc, childctx, started, completed, workflowMetadata, logger)
			workselector.AddFuture(future, func(f workflow.Future) {
				logger.Info("config sync completed", "name", bc.Name)
				var result sync_activity.SyncResponse
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

func getSyncMetadata(config *datasync_activities.BenthosConfigResponse) *datasync_activities.SyncMetadata {
	names := strings.Split(config.Name, ".")
	schema, table := names[0], names[1]
	return &datasync_activities.SyncMetadata{Schema: schema, Table: table}
}

func invokeSync(
	config *datasync_activities.BenthosConfigResponse,
	ctx workflow.Context,
	started map[string]struct{},
	completed map[string][]string,
	workflowMetadata *datasync_activities.WorkflowMetadata,
	logger log.Logger,
) workflow.Future {
	metadata := getSyncMetadata(config)
	future, settable := workflow.NewFuture(ctx)
	logger.Info("triggering config sync", "name", config.Name, "metadata", metadata)
	started[config.Name] = struct{}{}
	workflow.GoNamed(ctx, config.Name, func(ctx workflow.Context) {
		configbits, err := yaml.Marshal(config.Config)
		if err != nil {
			logger.Error("unable to marshal benthos config", "err", err)
			settable.SetError(fmt.Errorf("unable to marshal benthos config: %w", err))
			return
		}
		var result sync_activity.SyncResponse
		err = workflow.ExecuteActivity(
			ctx,
			sync_activity.Sync,
			&sync_activity.SyncRequest{BenthosConfig: string(configbits), BenthosDsns: config.BenthosDsns}, metadata, workflowMetadata).Get(ctx, &result)
		tn := fmt.Sprintf("%s.%s", config.TableSchema, config.TableName)
		_, ok := completed[tn]
		if ok {
			completed[tn] = append(completed[tn], config.Columns...)
		} else {
			completed[tn] = config.Columns
		}
		settable.Set(result, err)
	})
	return future
}

func isConfigReady(config *datasync_activities.BenthosConfigResponse, completed map[string][]string) bool {
	if config == nil {
		return false
	}

	if len(config.DependsOn) == 0 {
		return true
	}
	// check that all columns in dependency has been completed
	for _, dep := range config.DependsOn {
		if cols, ok := completed[dep.Table]; ok {
			for _, dc := range dep.Columns {
				if !slices.Contains(cols, dc) {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

type SplitConfigs struct {
	Root       []*datasync_activities.BenthosConfigResponse
	Dependents []*datasync_activities.BenthosConfigResponse
}

func splitBenthosConfigs(configs []*datasync_activities.BenthosConfigResponse) *SplitConfigs {
	out := &SplitConfigs{
		Root:       []*datasync_activities.BenthosConfigResponse{},
		Dependents: []*datasync_activities.BenthosConfigResponse{},
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
