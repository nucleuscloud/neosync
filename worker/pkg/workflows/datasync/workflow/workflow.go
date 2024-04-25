package datasync_workflow

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
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
		HeartbeatTimeout: 1 * time.Minute,
	})
	logger := workflow.GetLogger(ctx)

	logger = log.With(logger, "jobId", req.JobId)
	logger.Info("data sync workflow starting")

	workflowMetadata := &shared.WorkflowMetadata{
		WorkflowId: wfinfo.WorkflowExecution.ID,
		RunId:      wfinfo.WorkflowExecution.RunID,
	}

	var bcResp *genbenthosconfigs_activity.GenerateBenthosConfigsResponse
	logger.Info("scheduling GenerateBenthosConfigs for execution.")
	var genbenthosactivity *genbenthosconfigs_activity.Activity
	err := workflow.ExecuteActivity(ctx, genbenthosactivity.GenerateBenthosConfigs, &genbenthosconfigs_activity.GenerateBenthosConfigsRequest{
		JobId: req.JobId,
	}).Get(ctx, &bcResp)
	if err != nil {
		return nil, err
	}

	if len(bcResp.BenthosConfigs) == 0 {
		logger.Info("found 0 benthos configs, ending workflow.")
		return &WorkflowResponse{}, nil
	}

	var actOptResp *syncactivityopts_activity.RetrieveActivityOptionsResponse
	ctx = workflow.WithActivityOptions(wfctx, workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
		},
		HeartbeatTimeout: 1 * time.Minute,
	})
	logger.Info("scheduling RetrieveActivityOptions for execution.")
	err = workflow.ExecuteActivity(ctx, syncactivityopts_activity.RetrieveActivityOptions, &syncactivityopts_activity.RetrieveActivityOptionsRequest{
		JobId: req.JobId,
	}, workflowMetadata).Get(ctx, &actOptResp)
	if err != nil {
		return nil, err
	}
	logger.Info("completed RetrieveActivityOptions.")

	ctx = workflow.WithActivityOptions(wfctx, *actOptResp.SyncActivityOptions)
	logger.Info("scheduling RunSqlInitTableStatements for execution.")
	var resp *runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse
	err = workflow.ExecuteActivity(ctx, runsqlinittablestmts_activity.RunSqlInitTableStatements, &runsqlinittablestmts_activity.RunSqlInitTableStatementsRequest{
		JobId:      req.JobId,
		WorkflowId: wfinfo.WorkflowExecution.ID,
	}).Get(ctx, &resp)
	if err != nil {
		return nil, err
	}
	logger.Info("completed RunSqlInitTableStatements.")

	started := sync.Map{}
	completed := sync.Map{}

	allDependsOn := map[string][]*tabledependency.DependsOn{} // configName -> dependson
	redisConfigs := map[string]*genbenthosconfigs_activity.BenthosRedisConfig{}
	for _, cfg := range bcResp.BenthosConfigs {
		for _, redisCfg := range cfg.RedisConfig {
			redisConfigs[redisCfg.Key] = redisCfg
		}
		allDependsOn[cfg.Name] = cfg.DependsOn
	}

	workselector := workflow.NewSelector(ctx)

	splitConfigs := splitBenthosConfigs(bcResp.BenthosConfigs)
	var activityErr error
	childctx, cancelHandler := workflow.WithCancel(ctx)

	if len(splitConfigs.Root) == 0 && len(splitConfigs.Dependents) > 0 {
		return nil, fmt.Errorf("root config not found. unable to process configs")
	}
	for _, bc := range splitConfigs.Root {
		bc := bc
		logger := log.With(logger, withBenthosConfigResponseLoggerTags(bc)...)
		future := invokeSync(bc, childctx, &started, &completed, logger)
		workselector.AddFuture(future, func(f workflow.Future) {
			logger.Info("config sync completed")
			var result sync_activity.SyncResponse
			err := f.Get(childctx, &result)
			if err != nil {
				logger.Error("activity did not complete", "err", err)
				redisErr := runRedisCleanUpActivity(wfctx, logger, actOptResp, map[string][]*tabledependency.DependsOn{}, req.JobId, wfinfo.WorkflowExecution.ID, redisConfigs)
				if redisErr != nil {
					logger.Error("redis clean up activity did not complete")
				}
				logger.Error("sync activity did not complete", "err", err)
				cancelHandler()
				activityErr = err
			}
			delete(allDependsOn, bc.Name)
			// clean up redis
			err = runRedisCleanUpActivity(wfctx, logger, actOptResp, allDependsOn, req.JobId, wfinfo.WorkflowExecution.ID, redisConfigs)
			if err != nil {
				logger.Error("redis clean up activity did not complete")
			}
		})
	}

	logger.Info("all root tables spawned, moving on to children")
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
			_, configStarted := started.Load(bc.Name)
			if configStarted {
				continue
			}
			isReady, err := isConfigReady(bc, &completed)
			if err != nil {
				return nil, err
			}

			if !isReady {
				continue
			}
			logger := log.With(logger, withBenthosConfigResponseLoggerTags(bc)...)
			future := invokeSync(bc, childctx, &started, &completed, logger)
			workselector.AddFuture(future, func(f workflow.Future) {
				logger.Info("config sync completed", "name", bc.Name)
				var result sync_activity.SyncResponse
				err := f.Get(childctx, &result)
				if err != nil {
					logger.Error("activity did not complete", "err", err)
					redisErr := runRedisCleanUpActivity(wfctx, logger, actOptResp, map[string][]*tabledependency.DependsOn{}, req.JobId, wfinfo.WorkflowExecution.ID, redisConfigs)
					if redisErr != nil {
						logger.Error("redis clean up activity did not complete")
					}
					cancelHandler()
					activityErr = err
				}
				delete(allDependsOn, bc.Name)
				// clean up redis
				err = runRedisCleanUpActivity(wfctx, logger, actOptResp, allDependsOn, req.JobId, wfinfo.WorkflowExecution.ID, redisConfigs)
				if err != nil {
					logger.Error("redis clean up activity did not complete")
				}
			})
		}
	}
	logger.Info("data sync workflow completed")
	return &WorkflowResponse{}, nil
}

func runRedisCleanUpActivity(
	wfctx workflow.Context,
	logger log.Logger,
	actOptResp *syncactivityopts_activity.RetrieveActivityOptionsResponse,
	dependsOnMap map[string][]*tabledependency.DependsOn,
	jobId, workflowId string,
	redisConfigs map[string]*genbenthosconfigs_activity.BenthosRedisConfig,
) error {
	if len(redisConfigs) > 0 {
		for k, cfg := range redisConfigs {
			if !isReadyForCleanUp(cfg.Table, cfg.Column, dependsOnMap) {
				continue
			}
			ctx := workflow.WithActivityOptions(wfctx, *actOptResp.SyncActivityOptions)
			logger.Debug("executing redis clean up activity")
			var resp *syncrediscleanup_activity.DeleteRedisHashResponse
			err := workflow.ExecuteActivity(ctx, syncrediscleanup_activity.DeleteRedisHash, &syncrediscleanup_activity.DeleteRedisHashRequest{
				JobId:      jobId,
				WorkflowId: workflowId,
				HashKey:    cfg.Key,
			}).Get(ctx, &resp)
			if err != nil {
				return err
			}
			delete(redisConfigs, k)
		}
	}
	return nil
}

func isReadyForCleanUp(table, col string, dependsOnMap map[string][]*tabledependency.DependsOn) bool {
	for _, dependsOn := range dependsOnMap {
		for _, d := range dependsOn {
			if d.Table == table && slices.Contains(d.Columns, col) {
				return false
			}
		}
	}
	return true
}

func withBenthosConfigResponseLoggerTags(bc *genbenthosconfigs_activity.BenthosConfigResponse) []any {
	keyvals := []any{}

	if bc.Name != "" {
		keyvals = append(keyvals, "name", bc.Name)
	}
	if bc.TableSchema != "" {
		keyvals = append(keyvals, "schema", bc.TableSchema)
	}
	if bc.TableName != "" {
		keyvals = append(keyvals, "table", bc.TableName)
	}

	return keyvals
}

func getSyncMetadata(config *genbenthosconfigs_activity.BenthosConfigResponse) *sync_activity.SyncMetadata {
	names := strings.Split(config.Name, ".")
	schema, table := names[0], names[1]
	return &sync_activity.SyncMetadata{Schema: schema, Table: table}
}

func invokeSync(
	config *genbenthosconfigs_activity.BenthosConfigResponse,
	ctx workflow.Context,
	started, completed *sync.Map,
	logger log.Logger,
) workflow.Future {
	metadata := getSyncMetadata(config)
	future, settable := workflow.NewFuture(ctx)
	logger.Debug("triggering config sync")
	started.Store(config.Name, struct{}{})
	workflow.GoNamed(ctx, config.Name, func(ctx workflow.Context) {
		configbits, err := yaml.Marshal(config.Config)
		if err != nil {
			logger.Error("unable to marshal benthos config", "err", err)
			settable.SetError(fmt.Errorf("unable to marshal benthos config: %w", err))
			return
		}
		logger.Info("scheduling Sync for execution.")

		var result sync_activity.SyncResponse
		activity := sync_activity.Activity{}
		err = workflow.ExecuteActivity(
			ctx,
			activity.Sync,
			&sync_activity.SyncRequest{BenthosConfig: string(configbits), BenthosDsns: config.BenthosDsns}, metadata).Get(ctx, &result)
		if err == nil {
			tn := fmt.Sprintf("%s.%s", config.TableSchema, config.TableName)
			err = updateCompletedMap(tn, completed, config.Columns)
			if err != nil {
				settable.Set(result, err)
			}
		}
		settable.Set(result, err)
	})
	return future
}

func updateCompletedMap(tableName string, completed *sync.Map, columns []string) error {
	val, loaded := completed.Load(tableName)
	if loaded {
		currCols, ok := val.([]string)
		if !ok {
			return fmt.Errorf("unable to retrieve completed colums from completed map. Expected []string, received: %T", val)
		}
		currCols = append(currCols, columns...)
		completed.Store(tableName, currCols)
	} else {
		completed.Store(tableName, columns)
	}
	return nil
}

func isConfigReady(config *genbenthosconfigs_activity.BenthosConfigResponse, completed *sync.Map) (bool, error) {
	if config == nil {
		return false, nil
	}

	if len(config.DependsOn) == 0 {
		return true, nil
	}
	// check that all columns in dependency has been completed
	for _, dep := range config.DependsOn {
		val, loaded := completed.Load(dep.Table)
		if loaded {
			completedCols, ok := val.([]string)
			if !ok {
				return false, fmt.Errorf("unable to retrieve completed colums from completed map. Expected []string, received: %T", val)
			}
			for _, dc := range dep.Columns {
				if !slices.Contains(completedCols, dc) {
					return false, nil
				}
			}
		} else {
			return false, nil
		}
	}
	return true, nil
}

type SplitConfigs struct {
	Root       []*genbenthosconfigs_activity.BenthosConfigResponse
	Dependents []*genbenthosconfigs_activity.BenthosConfigResponse
}

func splitBenthosConfigs(configs []*genbenthosconfigs_activity.BenthosConfigResponse) *SplitConfigs {
	out := &SplitConfigs{
		Root:       []*genbenthosconfigs_activity.BenthosConfigResponse{},
		Dependents: []*genbenthosconfigs_activity.BenthosConfigResponse{},
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
