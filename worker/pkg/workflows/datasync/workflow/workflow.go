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
	"github.com/nucleuscloud/neosync/internal/ee/license"
	"github.com/nucleuscloud/neosync/internal/runconfigs"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	accountstatus_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/account-status"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	jobhooks_by_timing_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/jobhooks-by-timing"
	posttablesync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/post-table-sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	schemainit_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/schemainit/workflow"
	workflow_shared "github.com/nucleuscloud/neosync/worker/pkg/workflows/shared"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/activities/sync"
	tablesync_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/workflow"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"

	"go.temporal.io/sdk/workflow"
)

type WorkflowRequest struct {
	JobId string
}

type WorkflowResponse struct{}

type Workflow struct {
	eelicense license.EEInterface
}

func New(eelicense license.EEInterface) *Workflow {
	return &Workflow{
		eelicense: eelicense,
	}
}

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

func (w *Workflow) Workflow(ctx workflow.Context, req *WorkflowRequest) (*WorkflowResponse, error) {
	logger := workflow.GetLogger(ctx)
	getAccountId := func() (string, error) {
		actOptResp, err := retrieveActivityOptions(ctx, req.JobId, logger)
		if err != nil {
			return "", err
		}
		return actOptResp.AccountId, nil
	}
	runWorkflow := func(ctx workflow.Context, logger log.Logger) (*WorkflowResponse, error) {
		return executeWorkflow(ctx, req)
	}
	wfinfo := workflow.GetInfo(ctx)
	return workflow_shared.HandleWorkflowEventLifecycle(
		ctx,
		w.eelicense,
		req.JobId,
		wfinfo.WorkflowExecution.ID,
		logger,
		getAccountId,
		runWorkflow,
	)
}

func executeWorkflow(wfctx workflow.Context, req *WorkflowRequest) (*WorkflowResponse, error) {
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

	info := workflow.GetInfo(ctx)
	var bcResp *genbenthosconfigs_activity.GenerateBenthosConfigsResponse
	logger.Info("scheduling GenerateBenthosConfigs for execution.")
	var genbenthosactivity *genbenthosconfigs_activity.Activity
	err = workflow.ExecuteActivity(
		withGenerateBenthosConfigsActivityOptions(ctx),
		genbenthosactivity.GenerateBenthosConfigs,
		&genbenthosconfigs_activity.GenerateBenthosConfigsRequest{
			JobId:    req.JobId,
			JobRunId: info.WorkflowExecution.ID,
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

	err = runSchemaInitWorkflowByDestination(ctx, logger, actOptResp.AccountId, req.JobId, info.WorkflowExecution.ID, actOptResp.Destinations)
	if err != nil {
		return nil, err
	}

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

	maxConcurrency := getTableSyncMaxConcurrency()
	inFlight := 0
	completedCount := 0
	started := sync.Map{}
	completed := sync.Map{}

	executeSyncActivity := func(bc *benthosbuilder.BenthosConfigResponse, logger log.Logger) {
		future := invokeSync(bc, ctx, &started, &completed, logger, &bcResp.AccountId, actOptResp.SyncActivityOptions)
		inFlight++
		workselector.AddFuture(future, func(f workflow.Future) {
			var wfResult tablesync_workflow.TableSyncResponse
			err := f.Get(ctx, &wfResult)
			inFlight--
			completedCount++
			if err != nil {
				logger.Error("activity did not complete", "err", err)
				activityErr = err
				cancelHandler()

				// empty depends on map will clean up all redis inserts
				detachedCtx, _ := workflow.NewDisconnectedContext(ctx)
				redisErr := runRedisCleanUpActivity(detachedCtx, logger, map[string]map[string][]string{}, req.JobId, redisConfigs)
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
		// Ensures concurrency limits are respected.
		for inFlight >= maxConcurrency {
			logger.Debug("max concurrency reached; blocking until one sync finishes")
			workselector.Select(ctx)
			if activityErr != nil {
				return nil, activityErr
			}
			if ctx.Err() != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return nil, fmt.Errorf("workflow canceled due to error/stop: %w", ctx.Err())
				}
				return nil, ctx.Err()
			}
		}
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
		// Ensures that the select statement below does not block indefinitely
		if len(bcResp.BenthosConfigs) == completedCount {
			break
		}
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

			// Ensures concurrency limits are respected.
			if inFlight >= maxConcurrency {
				logger.Debug("max concurrency reached; blocking until one sync finishes for a dependent")
				workselector.Select(ctx)
				if activityErr != nil {
					return nil, activityErr
				}
				if ctx.Err() != nil {
					if errors.Is(ctx.Err(), context.Canceled) {
						return nil, fmt.Errorf("workflow canceled due to error or stop signal: %w", ctx.Err())
					}
					return nil, fmt.Errorf("exiting workflow in dependent sync due to err: %w", ctx.Err())
				}
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

func runSchemaInitWorkflowByDestination(
	ctx workflow.Context,
	logger log.Logger,
	accountId, jobId, jobRunId string,
	destinations []*mgmtv1alpha1.JobDestination,
) error {
	initSchemaActivityOptions := &workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
		HeartbeatTimeout: 1 * time.Minute,
	}
	for _, destination := range destinations {
		// right now only mysql supports schema drift
		schemaDrift := destination.GetOptions().GetMysqlOptions() != nil
		logger.Info("scheduling Schema Initialization workflow for execution.", "destinationId", destination.GetId())
		siWf := &schemainit_workflow.Workflow{}
		var wfResult schemainit_workflow.SchemaInitResponse
		id := fmt.Sprintf("init-schema-%s", destination.GetId())
		err := workflow.ExecuteChildWorkflow(workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID:    workflow_shared.BuildChildWorkflowId(jobRunId, id, workflow.Now(ctx)),
			StaticSummary: fmt.Sprintf("Initializing Schema for %s", destination.GetId()),
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 1,
			},
		}), siWf.SchemaInit, &schemainit_workflow.SchemaInitRequest{
			AccountId:                 accountId,
			JobId:                     jobId,
			SchemaInitActivityOptions: initSchemaActivityOptions,
			JobRunId:                  jobRunId,
			DestinationId:             destination.GetId(),
			UseSchemaDrift:            schemaDrift,
		}).Get(ctx, &wfResult)
		if err != nil {
			return err
		}
		logger.Info("completed Schema Initialization workflow.", "destinationId", destination.GetId())
	}
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
	info := workflow.GetInfo(ctx)
	metadata := getSyncMetadata(config)
	_ = metadata
	future, settable := workflow.NewFuture(ctx)
	logger.Debug("triggering config sync")
	started.Store(config.Name, struct{}{})
	workflow.GoNamed(ctx, config.Name, func(ctx workflow.Context) {
		var accId string
		if accountId != nil && *accountId != "" {
			accId = *accountId
		}
		logger.Info("scheduling Sync for execution.")

		tsWf := &tablesync_workflow.Workflow{}
		var wfResult tablesync_workflow.TableSyncResponse

		err := workflow.ExecuteChildWorkflow(workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID:    workflow_shared.BuildChildWorkflowId(info.WorkflowExecution.ID, config.Name, workflow.Now(ctx)),
			StaticSummary: fmt.Sprintf("Syncing %s.%s", config.TableSchema, config.TableName),
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 1,
			},
		}), tsWf.TableSync, &tablesync_workflow.TableSyncRequest{
			AccountId:           accId,
			Id:                  config.Name,
			SyncActivityOptions: syncActivityOptions,
			ContinuationToken:   nil,
			JobRunId:            info.WorkflowExecution.ID,
			TableSchema:         config.TableSchema,
			TableName:           config.TableName,
		}).Get(ctx, &wfResult)
		if err == nil {
			tn := neosync_benthos.BuildBenthosTable(config.TableSchema, config.TableName)
			err = updateCompletedMap(tn, completed, config.Columns)
			if err != nil {
				settable.Set(wfResult, err)
			}
		}
		settable.Set(wfResult, err)
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

func toStringSliceMap(m *sync.Map) (map[string][]string, error) {
	result := make(map[string][]string)
	var typeErr error

	m.Range(func(k, v any) bool {
		key, okKey := k.(string)
		val, okVal := v.([]string)
		if !okKey || !okVal {
			typeErr = fmt.Errorf("failed type assertion for key=%T and value=%T", k, v)
			return false
		}
		result[key] = val
		return true
	})

	return result, typeErr
}

func isConfigReady(config *benthosbuilder.BenthosConfigResponse, completed *sync.Map) (bool, error) {
	if completed == nil {
		return false, fmt.Errorf("completed map is nil: cannot determine if config is ready")
	}
	if config == nil {
		return false, nil
	}
	completedMap, err := toStringSliceMap(completed)
	if err != nil {
		return false, err
	}
	return runconfigs.AreConfigDependenciesSatisfied(config.DependsOn, completedMap), nil
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

func getTableSyncMaxConcurrency() int {
	maxConcurrency := viper.GetInt("TABLESYNC_MAX_CONCURRENCY")
	if maxConcurrency <= 0 {
		return 3 // default max concurrency
	}
	return maxConcurrency
}
