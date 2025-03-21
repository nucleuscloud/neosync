package datasync_workflow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	benthosbuilder "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder"
	benthosbuilder_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	runconfigs "github.com/nucleuscloud/neosync/internal/runconfigs"
	"github.com/nucleuscloud/neosync/internal/testutil"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	accountstatus_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/account-status"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	jobhooks_by_timing_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/jobhooks-by-timing"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	accounthook_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/workflow"
	tablesync_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/workflow"
	"go.uber.org/atomic"

	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func Test_Workflow_BenthosConfigsFails(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	var activityOpts *syncactivityopts_activity.Activity
	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
			AccountId: uuid.NewString(),
		}, nil)
	var accStatsActivity *accountstatus_activity.Activity
	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)

	var genact *genbenthosconfigs_activity.Activity
	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).Return(nil, errors.New("TestFailure"))

	env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
		Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil).Twice()

	datasyncWorkflow := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
	env.ExecuteWorkflow(datasyncWorkflow.Workflow, &WorkflowRequest{})

	assert.True(t, env.IsWorkflowCompleted())

	err := env.GetWorkflowError()
	require.Error(t, err)
	var applicationErr *temporal.ApplicationError
	require.True(t, errors.As(err, &applicationErr))
	assert.Equal(t, "TestFailure", applicationErr.Error())

	env.AssertExpectations(t)
}

func Test_Workflow_Succeeds_Zero_BenthosConfigs(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	var activityOpts *syncactivityopts_activity.Activity
	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
			AccountId: uuid.NewString(),
		}, nil)
	var accStatsActivity *accountstatus_activity.Activity
	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)
	env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
		Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil)

	var genact *genbenthosconfigs_activity.Activity
	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosbuilder.BenthosConfigResponse{}}, nil)

	datasyncWorkflow := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
	env.ExecuteWorkflow(datasyncWorkflow.Workflow, &WorkflowRequest{})

	assert.True(t, env.IsWorkflowCompleted())

	err := env.GetWorkflowError()
	assert.Nil(t, err)

	result := &WorkflowResponse{}
	err = env.GetWorkflowResult(result)
	assert.Nil(t, err)
	assert.Equal(t, result, &WorkflowResponse{})

	env.AssertExpectations(t)
}

func Test_Workflow_Succeeds_SingleSync(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	var activityOpts *syncactivityopts_activity.Activity
	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
			AccountId: uuid.NewString(),
		}, nil)
	var accStatsActivity *accountstatus_activity.Activity
	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)
	env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
		Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil).Twice()

	var genact *genbenthosconfigs_activity.Activity
	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosbuilder.BenthosConfigResponse{
			{
				Name:      "public.users",
				DependsOn: []*runconfigs.DependsOn{},
				Config:    &neosync_benthos.BenthosConfig{},
			},
		}}, nil)

	var jobHookTimingActivity *jobhooks_by_timing_activity.Activity
	env.OnActivity(jobHookTimingActivity.RunJobHooksByTiming, mock.Anything, mock.Anything).
		Return(&jobhooks_by_timing_activity.RunJobHooksByTimingResponse{ExecCount: 1}, nil)

	syncWorkflow := tablesync_workflow.New(10)
	env.OnWorkflow(syncWorkflow.TableSync, mock.Anything, mock.Anything).
		Return(&tablesync_workflow.TableSyncResponse{}, nil)

	datasyncWorkflow := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
	env.ExecuteWorkflow(datasyncWorkflow.Workflow, &WorkflowRequest{})

	assert.True(t, env.IsWorkflowCompleted(), "Workflow did not complete as expected")

	err := env.GetWorkflowError()
	assert.Nil(t, err, "Expected no error during workflow execution, but got one: %v", err)

	result := &WorkflowResponse{}
	err = env.GetWorkflowResult(result)
	assert.Nil(t, err, "Failed to retrieve workflow result: %v", err)
	assert.Equal(t, result, &WorkflowResponse{}, "Error: Workflow result does not match the expected value")

	env.AssertExpectations(t)
}

func Test_Workflow_Follows_Synchronous_DependentFlow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	var accStatsActivity *accountstatus_activity.Activity
	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)
	env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
		Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil)

	var genact *genbenthosconfigs_activity.Activity
	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosbuilder.BenthosConfigResponse{
			{
				Name:      "public.users",
				DependsOn: []*runconfigs.DependsOn{},
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
				TableSchema: "public",
				TableName:   "users",
				Columns:     []string{"id"},
			},
			{
				Name:      "public.foo",
				DependsOn: []*runconfigs.DependsOn{{Table: "public.users", Columns: []string{"id"}}},
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
				TableSchema: "public",
				TableName:   "foo",
				Columns:     []string{"id"},
			},
		}}, nil)
	var activityOpts *syncactivityopts_activity.Activity
	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
		}, nil)

	var jobHookTimingActivity *jobhooks_by_timing_activity.Activity
	env.OnActivity(jobHookTimingActivity.RunJobHooksByTiming, mock.Anything, mock.Anything).
		Return(&jobhooks_by_timing_activity.RunJobHooksByTimingResponse{ExecCount: 1}, nil)

	count := 0
	var tableSyncWorkflow tablesync_workflow.Workflow
	env.OnWorkflow(tableSyncWorkflow.TableSync, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, req *tablesync_workflow.TableSyncRequest) (*tablesync_workflow.TableSyncResponse, error) {
			if req.TableSchema == "public" && req.TableName == "users" {
				// This is the root sync.
				assert.Equal(t, 0, count, "Expected 'users' sync to be called first")
				count++
			} else if req.TableSchema == "public" && req.TableName == "foo" {
				// The dependent sync should occur after 'users' completes.
				assert.Equal(t, 1, count, "Expected 'foo' sync to be called after 'users'")
				count++
			}
			return &tablesync_workflow.TableSyncResponse{}, nil
		})

	datasyncWorkflow := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
	env.ExecuteWorkflow(datasyncWorkflow.Workflow, &WorkflowRequest{})

	assert.True(t, env.IsWorkflowCompleted())
	assert.Equal(t, count, 2)

	err := env.GetWorkflowError()
	assert.Nil(t, err)

	result := &WorkflowResponse{}
	err = env.GetWorkflowResult(result)
	assert.Nil(t, err)
	assert.Equal(t, result, &WorkflowResponse{})

	env.AssertExpectations(t)
}

func Test_Workflow_Follows_Multiple_Dependents(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	var accStatsActivity *accountstatus_activity.Activity
	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)
	env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
		Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil)

	var genact *genbenthosconfigs_activity.Activity
	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosbuilder.BenthosConfigResponse{
			{
				Name:        "public.users",
				DependsOn:   []*runconfigs.DependsOn{},
				TableSchema: "public",
				TableName:   "users",
				Columns:     []string{"id"},
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
			},
			{
				Name:        "public.accounts",
				DependsOn:   []*runconfigs.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "accounts",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
			},
			{
				Name:        "public.foo",
				DependsOn:   []*runconfigs.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "foo",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
			},
		}}, nil)
	var activityOpts *syncactivityopts_activity.Activity
	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
		}, nil)

	var jobHookTimingActivity *jobhooks_by_timing_activity.Activity
	env.OnActivity(jobHookTimingActivity.RunJobHooksByTiming, mock.Anything, mock.Anything).
		Return(&jobhooks_by_timing_activity.RunJobHooksByTimingResponse{ExecCount: 1}, nil)

	counter := atomic.NewInt32(0)
	var tableSyncWorkflow tablesync_workflow.Workflow
	env.OnWorkflow(tableSyncWorkflow.TableSync, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, req *tablesync_workflow.TableSyncRequest) (*tablesync_workflow.TableSyncResponse, error) {
			switch req.TableName {
			case "users", "accounts":
				// Both are root syncs – increment the counter.
				counter.Add(1)
			case "foo":
				// Dependent sync should run only after both roots are complete.
				assert.Equal(t, int32(2), counter.Load(), "Expected both 'users' and 'accounts' to finish before 'foo' runs")
				counter.Add(1)
			default:
				t.Errorf("unexpected table name: %s", req.TableName)
			}
			return &tablesync_workflow.TableSyncResponse{}, nil
		})

	datasyncWorkflow := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
	env.ExecuteWorkflow(datasyncWorkflow.Workflow, &WorkflowRequest{})

	assert.True(t, env.IsWorkflowCompleted())
	assert.Equal(t, counter.Load(), int32(3))

	err := env.GetWorkflowError()
	assert.Nil(t, err)

	result := &WorkflowResponse{}
	err = env.GetWorkflowResult(result)
	assert.Nil(t, err)
	assert.Equal(t, result, &WorkflowResponse{})

	env.AssertExpectations(t)
}

func Test_Workflow_Follows_Multiple_Dependent_Redis_Cleanup(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	var accStatsActivity *accountstatus_activity.Activity
	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)
	env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
		Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil)
	var genact *genbenthosconfigs_activity.Activity
	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosbuilder.BenthosConfigResponse{
			{
				Name:        "public.users",
				DependsOn:   []*runconfigs.DependsOn{},
				TableSchema: "public",
				TableName:   "users",
				Columns:     []string{"id"},
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
				RedisConfig: []*benthosbuilder_shared.BenthosRedisConfig{
					{
						Key:    "fake-redis-key",
						Table:  "public.users",
						Column: "id",
					},
				},
			},
			{
				Name:        "public.accounts",
				DependsOn:   []*runconfigs.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "accounts",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
				RedisConfig: []*benthosbuilder_shared.BenthosRedisConfig{
					{
						Key:    "fake-redis-key2",
						Table:  "public.accounts",
						Column: "id",
					},
				},
			},
			{
				Name:        "public.foo",
				DependsOn:   []*runconfigs.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "foo",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
			},
		}}, nil)
	var activityOpts *syncactivityopts_activity.Activity
	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
		}, nil)

	var jobHookTimingActivity *jobhooks_by_timing_activity.Activity
	env.OnActivity(jobHookTimingActivity.RunJobHooksByTiming, mock.Anything, mock.Anything).
		Return(&jobhooks_by_timing_activity.RunJobHooksByTimingResponse{ExecCount: 1}, nil)

	counter := atomic.NewInt32(0)
	var tableSyncWorkflow tablesync_workflow.Workflow
	env.OnWorkflow(tableSyncWorkflow.TableSync, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, req *tablesync_workflow.TableSyncRequest) (*tablesync_workflow.TableSyncResponse, error) {
			switch req.TableName {
			case "users", "accounts":
				// Both are root syncs – increment the counter.
				counter.Add(1)
			case "foo":
				// Dependent sync should run only after both roots are complete.
				assert.Equal(t, int32(2), counter.Load(), "Expected both 'users' and 'accounts' to finish before 'foo' runs")
				counter.Add(1)
			default:
				t.Errorf("unexpected table name: %s", req.TableName)
			}
			return &tablesync_workflow.TableSyncResponse{}, nil
		})

	redisCleanupCount := atomic.NewInt32(0)
	var syncRedisCleanupActivity *syncrediscleanup_activity.Activity
	env.OnActivity(syncRedisCleanupActivity.DeleteRedisHash, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, req *syncrediscleanup_activity.DeleteRedisHashRequest) (*syncrediscleanup_activity.DeleteRedisHashResponse, error) {
			redisCleanupCount.Add(1)
			return &syncrediscleanup_activity.DeleteRedisHashResponse{}, nil
		})

	datasyncWorkflow := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
	env.ExecuteWorkflow(datasyncWorkflow.Workflow, &WorkflowRequest{})

	assert.True(t, env.IsWorkflowCompleted())
	assert.Equal(t, counter.Load(), int32(3))
	assert.Equal(t, int32(2), redisCleanupCount.Load(), "Expected two redis cleanup calls (one for each RedisConfig)")

	err := env.GetWorkflowError()
	assert.Nil(t, err)

	result := &WorkflowResponse{}
	err = env.GetWorkflowResult(result)
	assert.Nil(t, err)
	assert.Equal(t, result, &WorkflowResponse{})

	env.AssertExpectations(t)
}

func Test_Workflow_Halts_Activities_OnError(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	var genact *genbenthosconfigs_activity.Activity
	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosbuilder.BenthosConfigResponse{
			{
				Name:        "public.users",
				DependsOn:   []*runconfigs.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "users",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
			},
			{
				Name:        "public.accounts",
				DependsOn:   []*runconfigs.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "accounts",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
			},
			{
				Name:        "public.foo",
				DependsOn:   []*runconfigs.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "foo",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
			},
		}}, nil)
	var activityOpts *syncactivityopts_activity.Activity
	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
		}, nil)
	var accStatsActivity *accountstatus_activity.Activity
	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)
	env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
		Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil)

	var jobHookTimingActivity *jobhooks_by_timing_activity.Activity
	env.OnActivity(jobHookTimingActivity.RunJobHooksByTiming, mock.Anything, mock.Anything).
		Return(&jobhooks_by_timing_activity.RunJobHooksByTimingResponse{ExecCount: 1}, nil)

	var tableSyncWorkflow tablesync_workflow.Workflow
	env.OnWorkflow(tableSyncWorkflow.TableSync, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, req *tablesync_workflow.TableSyncRequest) (*tablesync_workflow.TableSyncResponse, error) {
			if req.TableSchema == "public" && req.TableName == "accounts" {
				return nil, errors.New("TestFailure")
			}
			return &tablesync_workflow.TableSyncResponse{}, nil
		})

	datasyncWorkflow := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
	env.ExecuteWorkflow(datasyncWorkflow.Workflow, &WorkflowRequest{})

	require.True(t, env.IsWorkflowCompleted())

	err := env.GetWorkflowError()
	require.Error(t, err)
	var applicationErr *temporal.ApplicationError
	require.True(t, errors.As(err, &applicationErr))
	require.Equal(t, "TestFailure", applicationErr.Error())

	env.AssertExpectations(t)
}

func Test_Workflow_Halts_Activities_On_InvalidAccountStatus(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	var genact *genbenthosconfigs_activity.Activity
	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosbuilder.BenthosConfigResponse{
			{
				Name:        "public.users",
				DependsOn:   []*runconfigs.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "users",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
			},
			{
				Name:        "public.accounts",
				DependsOn:   []*runconfigs.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "accounts",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
			},
			{
				Name:        "public.foo",
				DependsOn:   []*runconfigs.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "foo",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
			},
		}}, nil)
	var activityOpts *syncactivityopts_activity.Activity
	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
		}, nil)

	var accStatsActivity *accountstatus_activity.Activity
	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true, ShouldPoll: true}, nil).Once()
	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: false}, nil).Once()

	env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
		Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil)

	var jobHookTimingActivity *jobhooks_by_timing_activity.Activity
	env.OnActivity(jobHookTimingActivity.RunJobHooksByTiming, mock.Anything, mock.Anything).
		Return(&jobhooks_by_timing_activity.RunJobHooksByTimingResponse{ExecCount: 1}, nil)

	var tableSyncWorkflow tablesync_workflow.Workflow
	env.OnWorkflow(tableSyncWorkflow.TableSync, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, req *tablesync_workflow.TableSyncRequest) (*tablesync_workflow.TableSyncResponse, error) {
			workflow.Sleep(ctx, 3*time.Second)
			return &tablesync_workflow.TableSyncResponse{}, nil
		})

	datasyncWorkflow := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
	env.ExecuteWorkflow(datasyncWorkflow.Workflow, &WorkflowRequest{})

	require.True(t, env.IsWorkflowCompleted())

	err := env.GetWorkflowError()
	require.Error(t, err)
	var applicationErr *temporal.ApplicationError
	require.True(t, errors.As(err, &applicationErr))
	require.ErrorContains(t, applicationErr, invalidAccountStatusError.Error())

	env.AssertExpectations(t)
}

func Test_Workflow_Cleans_Up_Redis_OnError(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	var genact *genbenthosconfigs_activity.Activity
	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosbuilder.BenthosConfigResponse{
			{
				Name:        "public.users",
				DependsOn:   []*runconfigs.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "users",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
				RedisConfig: []*benthosbuilder_shared.BenthosRedisConfig{
					{
						Key:    "fake-redis-key",
						Table:  "public.users",
						Column: "id",
					},
				},
			},
			{
				Name:        "public.accounts",
				DependsOn:   []*runconfigs.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "accounts",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
			},
			{
				Name:        "public.foo",
				DependsOn:   []*runconfigs.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "foo",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
									OrderByColumns: []string{"id"},
								},
							},
						},
					},
				},
			},
		}}, nil)
	var activityOpts *syncactivityopts_activity.Activity
	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
		}, nil)
	var accStatsActivity *accountstatus_activity.Activity
	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)

	env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
		Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil)

	var jobHookTimingActivity *jobhooks_by_timing_activity.Activity
	env.OnActivity(jobHookTimingActivity.RunJobHooksByTiming, mock.Anything, mock.Anything).
		Return(&jobhooks_by_timing_activity.RunJobHooksByTimingResponse{ExecCount: 1}, nil)

	var tableSyncWorkflow tablesync_workflow.Workflow
	env.OnWorkflow(tableSyncWorkflow.TableSync, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, req *tablesync_workflow.TableSyncRequest) (*tablesync_workflow.TableSyncResponse, error) {
			if req.TableName == "users" {
				return nil, errors.New("TestFailure")
			}
			return &tablesync_workflow.TableSyncResponse{}, nil
		})

	redisCleanupCount := atomic.NewInt32(0)
	var syncRedisCleanupActivity *syncrediscleanup_activity.Activity
	env.OnActivity(syncRedisCleanupActivity.DeleteRedisHash, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, req *syncrediscleanup_activity.DeleteRedisHashRequest) (*syncrediscleanup_activity.DeleteRedisHashResponse, error) {
			redisCleanupCount.Add(1)
			return &syncrediscleanup_activity.DeleteRedisHashResponse{}, nil
		})

	datasyncWorkflow := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
	env.ExecuteWorkflow(datasyncWorkflow.Workflow, &WorkflowRequest{})

	assert.True(t, env.IsWorkflowCompleted())
	assert.Equal(t, int32(1), redisCleanupCount.Load(), "Expected one redis cleanup call")

	err := env.GetWorkflowError()
	assert.Error(t, err)
	var applicationErr *temporal.ApplicationError
	assert.True(t, errors.As(err, &applicationErr))
	assert.Equal(t, "TestFailure", applicationErr.Error())

	env.AssertExpectations(t)
}

func Test_Workflow_Max_InFlight(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Atomic counters to track the current and maximum in-flight child workflows.
	currentInFlight := atomic.NewInt32(0)
	maxObserved := atomic.NewInt32(0)

	var activityOpts *syncactivityopts_activity.Activity
	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
			AccountId: "test-account",
		}, nil)

	var accStatsActivity *accountstatus_activity.Activity
	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)

	env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
		Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil)

	// Return several root configurations so that all can be started concurrently.
	var genact *genbenthosconfigs_activity.Activity
	numConfigs := 5
	configs := make([]*benthosbuilder.BenthosConfigResponse, numConfigs)
	for i := 0; i < numConfigs; i++ {
		configs[i] = &benthosbuilder.BenthosConfigResponse{
			Name:        fmt.Sprintf("config-%d", i),
			DependsOn:   []*runconfigs.DependsOn{},
			TableSchema: "public",
			TableName:   fmt.Sprintf("table%d", i),
			Columns:     []string{"id"},
			Config:      &neosync_benthos.BenthosConfig{},
		}
	}
	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{
			BenthosConfigs: configs,
		}, nil)

	var jobHookTimingActivity *jobhooks_by_timing_activity.Activity
	env.OnActivity(jobHookTimingActivity.RunJobHooksByTiming, mock.Anything, mock.Anything).
		Return(&jobhooks_by_timing_activity.RunJobHooksByTimingResponse{ExecCount: 1}, nil)

	var tableSyncWorkflow tablesync_workflow.Workflow
	env.OnWorkflow(tableSyncWorkflow.TableSync, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, req *tablesync_workflow.TableSyncRequest) (*tablesync_workflow.TableSyncResponse, error) {
			// Increment the in-flight counter.
			curr := currentInFlight.Inc()

			// Update the maximum observed value if needed.
			for {
				prev := maxObserved.Load()
				if curr > prev {
					if maxObserved.CompareAndSwap(prev, curr) {
						break
					}
				} else {
					break
				}
			}

			// Simulate some work (using workflow.Sleep which is simulated in tests).
			workflow.Sleep(ctx, time.Second)

			// Decrement the counter when done.
			currentInFlight.Dec()
			return &tablesync_workflow.TableSyncResponse{}, nil
		})

	datasyncWorkflow := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
	env.ExecuteWorkflow(datasyncWorkflow.Workflow, &WorkflowRequest{JobId: "test-job"})

	require.True(t, env.IsWorkflowCompleted())
	err := env.GetWorkflowError()
	require.NoError(t, err)

	// Assert that the maximum number of concurrently running syncs never exceeded maxConcurrency.
	assert.LessOrEqual(t, maxObserved.Load(), int32(3), "Max in-flight child workflows exceeded limit")

	result := &WorkflowResponse{}
	err = env.GetWorkflowResult(result)
	require.NoError(t, err)
	assert.Equal(t, &WorkflowResponse{}, result)

	env.AssertExpectations(t)
}

func Test_isConfigReady(t *testing.T) {
	isReady, err := isConfigReady(nil, nil)
	assert.Error(t, err)

	completed := sync.Map{}
	isReady, err = isConfigReady(nil, &completed)
	assert.NoError(t, err)
	assert.False(t, isReady, "config is nil")

	completed = sync.Map{}
	isReady, err = isConfigReady(&benthosbuilder.BenthosConfigResponse{
		Name:      "foo",
		DependsOn: []*runconfigs.DependsOn{},
	},
		&completed)
	assert.NoError(t, err)
	assert.True(
		t,
		isReady,
		"has no dependencies",
	)

	completed = sync.Map{}
	completed.Store("bar", []string{"id"})
	isReady, err = isConfigReady(&benthosbuilder.BenthosConfigResponse{
		Name:      "foo",
		DependsOn: []*runconfigs.DependsOn{{Table: "bar", Columns: []string{"id"}}, {Table: "baz", Columns: []string{"id"}}},
	},
		&completed)
	assert.NoError(t, err)
	assert.False(
		t,
		isReady,
		"not all dependencies are finished",
	)

	completed = sync.Map{}
	completed.Store("bar", []string{"id"})
	completed.Store("baz", []string{"id"})
	isReady, err = isConfigReady(&benthosbuilder.BenthosConfigResponse{
		Name:      "foo",
		DependsOn: []*runconfigs.DependsOn{{Table: "bar", Columns: []string{"id"}}, {Table: "baz", Columns: []string{"id"}}},
	}, &completed)
	assert.NoError(t, err)
	assert.True(
		t,
		isReady,
		"all dependencies are finished",
	)

	completed = sync.Map{}
	completed.Store("bar", []string{"id"})
	isReady, err = isConfigReady(&benthosbuilder.BenthosConfigResponse{
		Name:      "foo",
		DependsOn: []*runconfigs.DependsOn{{Table: "bar", Columns: []string{"id", "f_id"}}},
	},
		&completed)
	assert.NoError(t, err)
	assert.False(
		t,
		isReady,
		"not all dependencies columns are finished",
	)
}

func Test_updateCompletedMap(t *testing.T) {
	completedMap := sync.Map{}
	table := "public.users"
	cols := []string{"id"}
	err := updateCompletedMap(table, &completedMap, cols)
	assert.NoError(t, err)
	val, loaded := completedMap.Load(table)
	assert.True(t, loaded)
	assert.Equal(t, cols, val)

	completedMap = sync.Map{}
	table = "public.users"
	completedMap.Store(table, []string{"name"})
	err = updateCompletedMap(table, &completedMap, []string{"id"})
	assert.NoError(t, err)
	val, loaded = completedMap.Load(table)
	assert.True(t, loaded)
	assert.Equal(t, []string{"name", "id"}, val)
}

func Test_isReadyForCleanUp(t *testing.T) {
	assert.True(t, isReadyForCleanUp("", "", nil), "no dependencies")

	assert.False(
		t,
		isReadyForCleanUp(
			"table",
			"col",
			map[string]map[string][]string{
				"other_table": {"table": []string{"col"}},
			},
		),
		"has dependency",
	)

	assert.True(
		t,
		isReadyForCleanUp(
			"table",
			"col",
			map[string]map[string][]string{
				"other_table": {"table": []string{"col1"}},
			},
		),
		"no dependency",
	)
}

func Test_Workflow_Initial_AccountStatus(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	var activityOptsActivity *syncactivityopts_activity.Activity
	env.OnActivity(activityOptsActivity.RetrieveActivityOptions, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			AccountId:            uuid.NewString(),
			RequestedRecordCount: shared.Ptr(uint64(4)),
		}, nil)

	var checkStatusActivity *accountstatus_activity.Activity
	env.OnActivity(checkStatusActivity.CheckAccountStatus, mock.Anything, mock.Anything).
		Return(&accountstatus_activity.CheckAccountStatusResponse{
			IsValid: false,
			Reason:  shared.Ptr("test failure"),
		}, nil)

	env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
		Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil)

	datasyncWorkflow := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
	env.ExecuteWorkflow(datasyncWorkflow.Workflow, &WorkflowRequest{})

	assert.True(t, env.IsWorkflowCompleted())

	err := env.GetWorkflowError()
	assert.Error(t, err)
	var applicationErr *temporal.ApplicationError
	assert.True(t, errors.As(err, &applicationErr))
	assert.ErrorContains(t, applicationErr, invalidAccountStatusError.Error())

	env.AssertExpectations(t)
}
