package datasync_workflow

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	benthosbuilder "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder"
	benthosbuilder_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	accountstatus_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/account-status"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	jobhooks_by_timing_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/jobhooks-by-timing"
	posttablesync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/post-table-sync"
	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
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

var (
	invalidAccountStatusError = errors.New("exiting workflow due to invalid account status")
)

func withGenerateBenthosConfigsActivityOptions(ctx workflow.Context) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
		HeartbeatTimeout: 1 * time.Minute,
	})
}

func withCheckAccountStatusActivityOptions(ctx workflow.Context) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
		},
		HeartbeatTimeout: 1 * time.Minute,
	})
}

func withJobHookTimingActivityOptions(ctx workflow.Context) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
		HeartbeatTimeout: 1 * time.Minute,
	})
}

func Workflow(wfctx workflow.Context, req *WorkflowRequest) (*WorkflowResponse, error) {
	ctx, cancelHandler := workflow.WithCancel(wfctx)
	logger := workflow.GetLogger(ctx)

	logger = log.With(logger, "jobId", req.JobId)
	logger.Info("data sync workflow starting")

	actOptResp, err := retrieveActivityOptions(ctx, req.JobId, logger)
	if err != nil {
		return nil, err
	}
	logger = log.With(
		logger,
		"accountId", actOptResp.AccountId,
	)

	if actOptResp.RequestedRecordCount != nil && *actOptResp.RequestedRecordCount > 0 {
		logger.Info(fmt.Sprintf("requested record count of %d", *actOptResp.RequestedRecordCount))
	}
	var initialCheckAccountStatusResponse *accountstatus_activity.CheckAccountStatusResponse
	var a *accountstatus_activity.Activity
	err = workflow.ExecuteActivity(
		withCheckAccountStatusActivityOptions(ctx),
		a.CheckAccountStatus,
		&accountstatus_activity.CheckAccountStatusRequest{AccountId: actOptResp.AccountId, RequestedRecordCount: actOptResp.RequestedRecordCount}).
		Get(ctx, &initialCheckAccountStatusResponse)
	if err != nil {
		logger.Error("encountered error while checking account status", "error", err)
		cancelHandler()
		return nil, fmt.Errorf("unable to continue workflow due to error when checking account status: %w", err)
	}
	if !initialCheckAccountStatusResponse.IsValid {
		logger.Warn("account is no longer is valid state")
		cancelHandler()
		reason := "no reason provided"
		if initialCheckAccountStatusResponse.Reason != nil {
			reason = *initialCheckAccountStatusResponse.Reason
		}
		return nil, fmt.Errorf("halting job run due to account in invalid state. Reason: %q: %w", reason, invalidAccountStatusError)
	}

	var bcResp *genbenthosconfigs_activity.GenerateBenthosConfigsResponse
	logger.Info("scheduling GenerateBenthosConfigs for execution.")
	var genbenthosactivity *genbenthosconfigs_activity.Activity
	err = workflow.ExecuteActivity(
		withGenerateBenthosConfigsActivityOptions(ctx),
		genbenthosactivity.GenerateBenthosConfigs,
		&genbenthosconfigs_activity.GenerateBenthosConfigsRequest{
			JobId: req.JobId,
		}).
		Get(ctx, &bcResp)
	if err != nil {
		return nil, err
	}

	if len(bcResp.BenthosConfigs) == 0 {
		logger.Info("found 0 benthos configs, ending workflow.")
		return &WorkflowResponse{}, nil
	}

	err = execRunJobHooksByTiming(ctx, &jobhooks_by_timing_activity.RunJobHooksByTimingRequest{JobId: req.JobId, Timing: mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_PRESYNC}, logger)
	if err != nil {
		return nil, err
	}

	logger.Info("scheduling RunSqlInitTableStatements for execution.")
	var resp *runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse
	var runSqlInitTableStatements *runsqlinittablestmts_activity.Activity
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, *actOptResp.SyncActivityOptions),
		runSqlInitTableStatements.RunSqlInitTableStatements,
		&runsqlinittablestmts_activity.RunSqlInitTableStatementsRequest{
			JobId: req.JobId,
		}).
		Get(ctx, &resp)
	if err != nil {
		return nil, err
	}
	logger.Info("completed RunSqlInitTableStatements.")

	redisDependsOn := map[string]map[string][]string{} // schema.table -> dependson
	redisConfigs := map[string]*benthosbuilder_shared.BenthosRedisConfig{}
	for _, cfg := range bcResp.BenthosConfigs {
		for _, redisCfg := range cfg.RedisConfig {
			redisConfigs[redisCfg.Key] = redisCfg
		}
		redisDependsOn[cfg.Name] = cfg.RedisDependsOn
	}

	// spawn account status checker in loop
	stopChan := workflow.NewNamedChannel(ctx, "account-status")
	if initialCheckAccountStatusResponse.ShouldPoll {
		accountStatusTimerDuration := getAccountStatusTimerDuration()
		workflow.GoNamed(
			ctx,
			"account-status-check",
			func(ctx workflow.Context) {
				shouldStop := false
				for {
					selector := workflow.NewNamedSelector(ctx, "account-status-select")
					timer := workflow.NewTimer(ctx, accountStatusTimerDuration)
					selector.AddFuture(timer, func(f workflow.Future) {
						err := f.Get(ctx, nil)
						if err != nil {
							logger.Error("time receive failed", "error", err)
							return
						}

						var result *accountstatus_activity.CheckAccountStatusResponse
						var a *accountstatus_activity.Activity
						err = workflow.ExecuteActivity(
							withCheckAccountStatusActivityOptions(ctx),
							a.CheckAccountStatus,
							&accountstatus_activity.CheckAccountStatusRequest{AccountId: actOptResp.AccountId}).
							Get(ctx, &result)
						if err != nil {
							logger.Error("encountered error while checking account status", "error", err)
							stopChan.Send(ctx, true)
							shouldStop = true
							cancelHandler()
							return
						}
						if !result.IsValid {
							logger.Warn("account is no longer is valid state")
							stopChan.Send(ctx, true)
							shouldStop = true
							cancelHandler()
							return
						}
					})

					selector.Select(ctx)

					if shouldStop {
						logger.Warn("exiting account status check")
						return
					}
					if ctx.Err() != nil {
						logger.Warn("workflow canceled due to error or stop signal", "error", ctx.Err())
						return
					}
				}
			})
	}

	workselector := workflow.NewSelector(ctx)
	var activityErr error

	workselector.AddReceive(stopChan, func(c workflow.ReceiveChannel, more bool) {
		// Stop signal received, exit the routing
		logger.Warn("received signal to stop workflow based on account status")
		activityErr = invalidAccountStatusError
		cancelHandler()
	})

	splitConfigs := splitBenthosConfigs(bcResp.BenthosConfigs)
	if len(splitConfigs.Root) == 0 && len(splitConfigs.Dependents) > 0 {
		return nil, fmt.Errorf("root config not found. unable to process configs")
	}

	started := sync.Map{}
	completed := sync.Map{}

	executeSyncActivity := func(bc *benthosbuilder.BenthosConfigResponse, logger log.Logger) {
		future := invokeSync(bc, ctx, &started, &completed, logger, &bcResp.AccountId, actOptResp.SyncActivityOptions)
		workselector.AddFuture(future, func(f workflow.Future) {
			var result sync_activity.SyncResponse
			err := f.Get(ctx, &result)
			if err != nil {
				logger.Error("activity did not complete", "err", err)
				activityErr = err
				cancelHandler()

				// empty depends on map will clean up all redis inserts
				redisErr := runRedisCleanUpActivity(ctx, logger, map[string]map[string][]string{}, req.JobId, redisConfigs)
				if redisErr != nil {
					logger.Error("redis clean up activity did not complete")
				}
				return
			}
			logger.Info("config sync completed", "name", bc.Name)
			err = runPostTableSyncActivity(ctx, logger, actOptResp, bc.Name)
			if err != nil {
				logger.Error(fmt.Sprintf("post table sync activity did not complete: %s", err.Error()), "schema", bc.TableSchema, "table", bc.TableName)
			}
			delete(redisDependsOn, bc.Name)
			err = runRedisCleanUpActivity(ctx, logger, redisDependsOn, req.JobId, redisConfigs)
			if err != nil {
				logger.Error(fmt.Sprintf("redis clean up activity did not complete: %s", err))
			}
		})
	}

	for _, bc := range splitConfigs.Root {
		logger := log.With(logger, withBenthosConfigResponseLoggerTags(bc)...)
		executeSyncActivity(bc, logger)

		if ctx.Err() != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil, fmt.Errorf("workflow canceled due to error or stop signal: %w", ctx.Err())
			}
			return nil, ctx.Err()
		}
	}

	logger.Info("all root tables spawned, moving on to children")
	for i := 0; i < len(bcResp.BenthosConfigs); i++ {
		logger.Debug("*** blocking select ***", "i", i)
		workselector.Select(ctx)
		if activityErr != nil {
			return nil, activityErr
		}
		logger.Debug("*** post select ***", "i", i)

		if ctx.Err() != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil, fmt.Errorf("workflow canceled due to error or stop signal: %w", ctx.Err())
			}
			return nil, fmt.Errorf("exiting workflow in root sync due to err: %w", ctx.Err())
		}

		// todo: deadlock detection
		for _, bc := range splitConfigs.Dependents {
			if ctx.Err() != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return nil, fmt.Errorf("workflow canceled due to error or stop signal: %w", ctx.Err())
				}
				return nil, fmt.Errorf("exiting workflow in dependent sync due err: %w", ctx.Err())
			}
			bc := bc
			if _, configStarted := started.Load(bc.Name); configStarted {
				continue
			}
			isReady, err := isConfigReady(bc, &completed)
			if err != nil {
				return nil, err
			}

			if !isReady {
				continue
			}

			executeSyncActivity(bc, log.With(logger, withBenthosConfigResponseLoggerTags(bc)...))
		}
	}

	logger.Info("data syncs completed")

	err = execRunJobHooksByTiming(ctx, &jobhooks_by_timing_activity.RunJobHooksByTimingRequest{JobId: req.JobId, Timing: mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_POSTSYNC}, logger)
	if err != nil {
		return nil, err
	}

	logger.Info("data sync workflow completed")
	return &WorkflowResponse{}, nil
}

func execRunJobHooksByTiming(ctx workflow.Context, req *jobhooks_by_timing_activity.RunJobHooksByTimingRequest, logger log.Logger) error {
	logger.Info(fmt.Sprintf("scheduling %q RunJobHooksByTiming for execution", req.Timing))
	var resp *jobhooks_by_timing_activity.RunJobHooksByTimingResponse
	var timingActivity *jobhooks_by_timing_activity.Activity
	err := workflow.ExecuteActivity(
		withJobHookTimingActivityOptions(ctx),
		timingActivity.RunJobHooksByTiming,
		req,
	).Get(ctx, &resp)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("completed %d %q RunJobHooksByTiming", resp.ExecCount, req.Timing))
	return nil
}

func retrieveActivityOptions(
	ctx workflow.Context,
	jobId string,
	logger log.Logger,
) (*syncactivityopts_activity.RetrieveActivityOptionsResponse, error) {
	logger.Info("scheduling RetrieveActivityOptions for execution.")

	var actOptResp *syncactivityopts_activity.RetrieveActivityOptionsResponse
	var activityOptsActivity *syncactivityopts_activity.Activity
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 1 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 2,
			},
			HeartbeatTimeout: 1 * time.Minute,
		}),
		activityOptsActivity.RetrieveActivityOptions,
		&syncactivityopts_activity.RetrieveActivityOptionsRequest{
			JobId: jobId,
		}).
		Get(ctx, &actOptResp)
	if err != nil {
		return nil, err
	}
	logger.Info("completed RetrieveActivityOptions.")
	return actOptResp, nil
}

func runPostTableSyncActivity(
	ctx workflow.Context,
	logger log.Logger,
	actOptResp *syncactivityopts_activity.RetrieveActivityOptionsResponse,
	name string,
) error {
	logger.Debug("executing post table sync activity")
	var resp *posttablesync_activity.RunPostTableSyncResponse
	var postTableSyncActivity *posttablesync_activity.Activity
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 2 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 2,
			},
			HeartbeatTimeout: 1 * time.Minute,
		}),
		postTableSyncActivity.RunPostTableSync,
		&posttablesync_activity.RunPostTableSyncRequest{
			AccountId: actOptResp.AccountId,
			Name:      name,
		}).Get(ctx, &resp)
	if err != nil {
		return err
	}
	return nil
}

func runRedisCleanUpActivity(
	ctx workflow.Context,
	logger log.Logger,
	dependsOnMap map[string]map[string][]string,
	jobId string,
	redisConfigs map[string]*benthosbuilder_shared.BenthosRedisConfig,
) error {
	if len(redisConfigs) > 0 {
		for k, cfg := range redisConfigs {
			if !isReadyForCleanUp(cfg.Table, cfg.Column, dependsOnMap) {
				continue
			}
			logger.Debug("executing redis clean up activity")
			var resp *syncrediscleanup_activity.DeleteRedisHashResponse
			var redisCleanUpActivity *syncrediscleanup_activity.Activity
			err := workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
					StartToCloseTimeout: 2 * time.Minute,
					RetryPolicy: &temporal.RetryPolicy{
						MaximumAttempts: 2,
					},
					HeartbeatTimeout: 1 * time.Minute,
				}),
				redisCleanUpActivity.DeleteRedisHash,
				&syncrediscleanup_activity.DeleteRedisHashRequest{
					JobId:   jobId,
					HashKey: cfg.Key,
				}).Get(ctx, &resp)
			if err != nil {
				return err
			}
			delete(redisConfigs, k)
		}
	}
	return nil
}

func isReadyForCleanUp(table, col string, dependsOnMap map[string]map[string][]string) bool {
	for _, dependsOn := range dependsOnMap {
		for t, cols := range dependsOn {
			if t == table && slices.Contains(cols, col) {
				return false
			}
		}
	}
	return true
}

func withBenthosConfigResponseLoggerTags(bc *benthosbuilder.BenthosConfigResponse) []any {
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

func getSyncMetadata(config *benthosbuilder.BenthosConfigResponse) *sync_activity.SyncMetadata {
	return &sync_activity.SyncMetadata{Schema: config.TableSchema, Table: config.TableName}
}

func invokeSync(
	config *benthosbuilder.BenthosConfigResponse,
	ctx workflow.Context,
	started, completed *sync.Map,
	logger log.Logger,
	accountId *string,
	syncActivityOptions *workflow.ActivityOptions,
) workflow.Future {
	metadata := getSyncMetadata(config)
	future, settable := workflow.NewFuture(ctx)
	logger.Debug("triggering config sync")
	started.Store(config.Name, struct{}{})
	workflow.GoNamed(ctx, config.Name, func(ctx workflow.Context) {
		var benthosConfig string
		var accId string
		if accountId != nil && *accountId != "" {
			accId = *accountId
		} else if config.Config != nil {
			configbits, err := yaml.Marshal(config.Config)
			if err != nil {
				logger.Error("unable to marshal benthos config", "err", err)
				settable.SetError(fmt.Errorf("unable to marshal benthos config: %w", err))
				return
			}
			benthosConfig = string(configbits)
		}

		logger.Info("scheduling Sync for execution.")

		var result sync_activity.SyncResponse
		activity := sync_activity.Activity{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, *syncActivityOptions),
			activity.Sync,
			&sync_activity.SyncRequest{BenthosConfig: benthosConfig, AccountId: accId, Name: config.Name, BenthosDsns: config.BenthosDsns}, metadata).Get(ctx, &result)
		if err == nil {
			tn := neosync_benthos.BuildBenthosTable(config.TableSchema, config.TableName)
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
			return fmt.Errorf("unable to retrieve completed columns from completed map. Expected []string, received: %T", val)
		}
		currCols = append(currCols, columns...)
		completed.Store(tableName, currCols)
	} else {
		completed.Store(tableName, columns)
	}
	return nil
}

func isConfigReady(config *benthosbuilder.BenthosConfigResponse, completed *sync.Map) (bool, error) {
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
				return false, fmt.Errorf("unable to retrieve completed columns from completed map. Expected []string, received: %T", val)
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
	Root       []*benthosbuilder.BenthosConfigResponse
	Dependents []*benthosbuilder.BenthosConfigResponse
}

func splitBenthosConfigs(configs []*benthosbuilder.BenthosConfigResponse) *SplitConfigs {
	out := &SplitConfigs{
		Root:       []*benthosbuilder.BenthosConfigResponse{},
		Dependents: []*benthosbuilder.BenthosConfigResponse{},
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

func getAccountStatusTimerDuration() time.Duration {
	envtime := viper.GetInt("CHECK_ACCOUNT_TIMER_SECONDS")
	if envtime == 0 {
		return 5 * time.Second
	}
	return time.Duration(envtime) * time.Second
}
