package datasync_workflow

// import (
// 	"context"
// 	"errors"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
// 	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
// 	accountstatus_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/account-status"
// 	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
// 	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
// 	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
// 	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
// 	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
// 	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/require"
// 	"go.temporal.io/sdk/temporal"
// 	"go.temporal.io/sdk/testsuite"
// 	"go.temporal.io/sdk/workflow"
// 	"go.uber.org/atomic"
// )

// func Test_Workflow_BenthosConfigsFails(t *testing.T) {
// 	testSuite := &testsuite.WorkflowTestSuite{}
// 	env := testSuite.NewTestWorkflowEnvironment()

// 	var activityOpts *syncactivityopts_activity.Activity
// 	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
// 		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
// 			SyncActivityOptions: &workflow.ActivityOptions{
// 				StartToCloseTimeout: time.Minute,
// 			},
// 			AccountId: uuid.NewString(),
// 		}, nil)
// 	var accStatsActivity *accountstatus_activity.Activity
// 	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
// 		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)

// 	var genact *genbenthosconfigs_activity.Activity
// 	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).Return(nil, errors.New("TestFailure"))

// 	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

// 	assert.True(t, env.IsWorkflowCompleted())
// 	assert.True(t, env.IsWorkflowCompleted())

// 	err := env.GetWorkflowError()
// 	assert.Error(t, err)
// 	var applicationErr *temporal.ApplicationError
// 	assert.True(t, errors.As(err, &applicationErr))
// 	assert.Equal(t, "TestFailure", applicationErr.Error())

// 	env.AssertExpectations(t)
// }

// func Test_Workflow_Succeeds_Zero_BenthosConfigs(t *testing.T) {
// 	testSuite := &testsuite.WorkflowTestSuite{}
// 	env := testSuite.NewTestWorkflowEnvironment()

// 	var activityOpts *syncactivityopts_activity.Activity
// 	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
// 		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
// 			SyncActivityOptions: &workflow.ActivityOptions{
// 				StartToCloseTimeout: time.Minute,
// 			},
// 			AccountId: uuid.NewString(),
// 		}, nil)
// 	var accStatsActivity *accountstatus_activity.Activity
// 	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
// 		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)

// 	var genact *genbenthosconfigs_activity.Activity
// 	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
// 		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{}}, nil)

// 	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

// 	assert.True(t, env.IsWorkflowCompleted())

// 	err := env.GetWorkflowError()
// 	assert.Nil(t, err)

// 	result := &WorkflowResponse{}
// 	err = env.GetWorkflowResult(result)
// 	assert.Nil(t, err)
// 	assert.Equal(t, result, &WorkflowResponse{})

// 	env.AssertExpectations(t)
// }

// func Test_Workflow_Succeeds_SingleSync(t *testing.T) {
// 	testSuite := &testsuite.WorkflowTestSuite{}
// 	env := testSuite.NewTestWorkflowEnvironment()

// 	var activityOpts *syncactivityopts_activity.Activity
// 	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
// 		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
// 			SyncActivityOptions: &workflow.ActivityOptions{
// 				StartToCloseTimeout: time.Minute,
// 			},
// 			AccountId: uuid.NewString(),
// 		}, nil)
// 	var accStatsActivity *accountstatus_activity.Activity
// 	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
// 		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)

// 	var genact *genbenthosconfigs_activity.Activity
// 	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
// 		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
// 			{
// 				Name:      "public.users",
// 				DependsOn: []*tabledependency.DependsOn{},
// 				Config:    &neosync_benthos.BenthosConfig{},
// 			},
// 		}}, nil)
// 	var sqlInitActivity *runsqlinittablestmts_activity.Activity
// 	env.OnActivity(sqlInitActivity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
// 		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
// 	syncActivity := sync_activity.Activity{}
// 	env.OnActivity(syncActivity.Sync, mock.Anything, mock.Anything, mock.Anything).Return(&sync_activity.SyncResponse{}, nil)

// 	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

// 	assert.True(t, env.IsWorkflowCompleted())

// 	err := env.GetWorkflowError()
// 	assert.Nil(t, err)

// 	result := &WorkflowResponse{}
// 	err = env.GetWorkflowResult(result)
// 	assert.Nil(t, err)
// 	assert.Equal(t, result, &WorkflowResponse{})

// 	env.AssertExpectations(t)
// }

// func Test_Workflow_Follows_Synchronous_DependentFlow(t *testing.T) {
// 	testSuite := &testsuite.WorkflowTestSuite{}
// 	env := testSuite.NewTestWorkflowEnvironment()

// 	var accStatsActivity *accountstatus_activity.Activity
// 	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
// 		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)

// 	var genact *genbenthosconfigs_activity.Activity
// 	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
// 		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
// 			{
// 				Name:      "public.users",
// 				DependsOn: []*tabledependency.DependsOn{},
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 				TableSchema: "public",
// 				TableName:   "users",
// 				Columns:     []string{"id"},
// 			},
// 			{
// 				Name:      "public.foo",
// 				DependsOn: []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}},
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 				TableSchema: "public",
// 				TableName:   "foo",
// 				Columns:     []string{"id"},
// 			},
// 		}}, nil)
// 	var activityOpts *syncactivityopts_activity.Activity
// 	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
// 		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
// 			SyncActivityOptions: &workflow.ActivityOptions{
// 				StartToCloseTimeout: time.Minute,
// 			},
// 		}, nil)
// 	var sqlInitActivity *runsqlinittablestmts_activity.Activity
// 	env.OnActivity(sqlInitActivity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
// 		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
// 	count := 0
// 	syncActivity := sync_activity.Activity{}
// 	env.
// 		OnActivity(syncActivity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "users"}).
// 		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata) (*sync_activity.SyncResponse, error) {
// 			assert.Equal(t, count, 0)
// 			count += 1
// 			return &sync_activity.SyncResponse{}, nil
// 		})
// 	env.
// 		OnActivity(syncActivity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "foo"}).
// 		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata) (*sync_activity.SyncResponse, error) {
// 			assert.Equal(t, count, 1)
// 			count += 1
// 			return &sync_activity.SyncResponse{}, nil
// 		})

// 	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

// 	assert.True(t, env.IsWorkflowCompleted())
// 	assert.Equal(t, count, 2)

// 	err := env.GetWorkflowError()
// 	assert.Nil(t, err)

// 	result := &WorkflowResponse{}
// 	err = env.GetWorkflowResult(result)
// 	assert.Nil(t, err)
// 	assert.Equal(t, result, &WorkflowResponse{})

// 	env.AssertExpectations(t)
// }

// func Test_Workflow_Follows_Multiple_Dependents(t *testing.T) {
// 	testSuite := &testsuite.WorkflowTestSuite{}
// 	env := testSuite.NewTestWorkflowEnvironment()

// 	var accStatsActivity *accountstatus_activity.Activity
// 	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
// 		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)
// 	var genact *genbenthosconfigs_activity.Activity
// 	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
// 		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
// 			{
// 				Name:        "public.users",
// 				DependsOn:   []*tabledependency.DependsOn{},
// 				TableSchema: "public",
// 				TableName:   "users",
// 				Columns:     []string{"id"},
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			{
// 				Name:        "public.accounts",
// 				DependsOn:   []*tabledependency.DependsOn{},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "accounts",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			{
// 				Name:        "public.foo",
// 				DependsOn:   []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "foo",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		}}, nil)
// 	var activityOpts *syncactivityopts_activity.Activity
// 	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
// 		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
// 			SyncActivityOptions: &workflow.ActivityOptions{
// 				StartToCloseTimeout: time.Minute,
// 			},
// 		}, nil)
// 	var sqlInitActivity *runsqlinittablestmts_activity.Activity
// 	env.OnActivity(sqlInitActivity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
// 		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
// 	counter := atomic.NewInt32(0)
// 	syncActivity := sync_activity.Activity{}
// 	env.
// 		OnActivity(syncActivity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "users"}).
// 		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata) (*sync_activity.SyncResponse, error) {
// 			counter.Add(1)
// 			return &sync_activity.SyncResponse{}, nil
// 		})
// 	env.
// 		OnActivity(syncActivity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "accounts"}).
// 		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata) (*sync_activity.SyncResponse, error) {
// 			counter.Add(1)
// 			return &sync_activity.SyncResponse{}, nil
// 		})
// 	env.
// 		OnActivity(syncActivity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "foo"}).
// 		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata) (*sync_activity.SyncResponse, error) {
// 			assert.Equal(t, counter.Load(), int32(2))
// 			counter.Add(1)
// 			return &sync_activity.SyncResponse{}, nil
// 		})

// 	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

// 	assert.True(t, env.IsWorkflowCompleted())
// 	assert.Equal(t, counter.Load(), int32(3))

// 	err := env.GetWorkflowError()
// 	assert.Nil(t, err)

// 	result := &WorkflowResponse{}
// 	err = env.GetWorkflowResult(result)
// 	assert.Nil(t, err)
// 	assert.Equal(t, result, &WorkflowResponse{})

// 	env.AssertExpectations(t)
// }

// func Test_Workflow_Follows_Multiple_Dependent_Redis_Cleanup(t *testing.T) {
// 	testSuite := &testsuite.WorkflowTestSuite{}
// 	env := testSuite.NewTestWorkflowEnvironment()

// 	var accStatsActivity *accountstatus_activity.Activity
// 	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
// 		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)
// 	var genact *genbenthosconfigs_activity.Activity
// 	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
// 		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
// 			{
// 				Name:        "public.users",
// 				DependsOn:   []*tabledependency.DependsOn{},
// 				TableSchema: "public",
// 				TableName:   "users",
// 				Columns:     []string{"id"},
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 				RedisConfig: []*genbenthosconfigs_activity.BenthosRedisConfig{
// 					{
// 						Key:    "fake-redis-key",
// 						Table:  "public.users",
// 						Column: "id",
// 					},
// 				},
// 			},
// 			{
// 				Name:        "public.accounts",
// 				DependsOn:   []*tabledependency.DependsOn{},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "accounts",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 				RedisConfig: []*genbenthosconfigs_activity.BenthosRedisConfig{
// 					{
// 						Key:    "fake-redis-key2",
// 						Table:  "public.accounts",
// 						Column: "id",
// 					},
// 				},
// 			},
// 			{
// 				Name:        "public.foo",
// 				DependsOn:   []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "foo",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		}}, nil)
// 	var activityOpts *syncactivityopts_activity.Activity
// 	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
// 		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
// 			SyncActivityOptions: &workflow.ActivityOptions{
// 				StartToCloseTimeout: time.Minute,
// 			},
// 		}, nil)
// 	var sqlInitActivity *runsqlinittablestmts_activity.Activity
// 	env.OnActivity(sqlInitActivity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
// 		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
// 	counter := atomic.NewInt32(0)
// 	syncActivities := &sync_activity.Activity{}
// 	env.
// 		OnActivity(syncActivities.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "users"}).
// 		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata) (*sync_activity.SyncResponse, error) {
// 			counter.Add(1)
// 			return &sync_activity.SyncResponse{}, nil
// 		})
// 	env.
// 		OnActivity(syncActivities.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "accounts"}).
// 		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata) (*sync_activity.SyncResponse, error) {
// 			counter.Add(1)
// 			return &sync_activity.SyncResponse{}, nil
// 		})
// 	env.
// 		OnActivity(syncActivities.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "foo"}).
// 		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata) (*sync_activity.SyncResponse, error) {
// 			assert.Equal(t, counter.Load(), int32(2))
// 			counter.Add(1)
// 			return &sync_activity.SyncResponse{}, nil
// 		})

// 	env.OnActivity(syncrediscleanup_activity.DeleteRedisHash, mock.Anything, mock.Anything).
// 		Return(&syncrediscleanup_activity.DeleteRedisHashResponse{}, nil)
// 	env.OnActivity(syncrediscleanup_activity.DeleteRedisHash, mock.Anything, mock.Anything).
// 		Return(&syncrediscleanup_activity.DeleteRedisHashResponse{}, nil)

// 	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

// 	assert.True(t, env.IsWorkflowCompleted())
// 	assert.Equal(t, counter.Load(), int32(3))

// 	err := env.GetWorkflowError()
// 	assert.Nil(t, err)

// 	result := &WorkflowResponse{}
// 	err = env.GetWorkflowResult(result)
// 	assert.Nil(t, err)
// 	assert.Equal(t, result, &WorkflowResponse{})

// 	env.AssertExpectations(t)
// }

// func Test_Workflow_Halts_Activities_OnError(t *testing.T) {
// 	testSuite := &testsuite.WorkflowTestSuite{}
// 	env := testSuite.NewTestWorkflowEnvironment()

// 	var genact *genbenthosconfigs_activity.Activity
// 	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
// 		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
// 			{
// 				Name:        "public.users",
// 				DependsOn:   []*tabledependency.DependsOn{},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "users",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			{
// 				Name:        "public.accounts",
// 				DependsOn:   []*tabledependency.DependsOn{},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "accounts",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			{
// 				Name:        "public.foo",
// 				DependsOn:   []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "foo",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		}}, nil)
// 	var sqlInitActivity *runsqlinittablestmts_activity.Activity
// 	env.OnActivity(sqlInitActivity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
// 		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
// 	var activityOpts *syncactivityopts_activity.Activity
// 	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
// 		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
// 			SyncActivityOptions: &workflow.ActivityOptions{
// 				StartToCloseTimeout: time.Minute,
// 			},
// 		}, nil)
// 	var accStatsActivity *accountstatus_activity.Activity
// 	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
// 		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)

// 	syncActivity := sync_activity.Activity{}
// 	env.
// 		OnActivity(syncActivity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "users"}).
// 		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata) (*sync_activity.SyncResponse, error) {
// 			return &sync_activity.SyncResponse{}, nil
// 		})
// 	env.
// 		OnActivity(syncActivity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "accounts"}).
// 		Return(nil, errors.New("TestFailure"))

// 	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

// 	require.True(t, env.IsWorkflowCompleted())

// 	err := env.GetWorkflowError()
// 	require.Error(t, err)
// 	var applicationErr *temporal.ApplicationError
// 	require.True(t, errors.As(err, &applicationErr))
// 	require.Equal(t, "TestFailure", applicationErr.Error())

// 	env.AssertExpectations(t)
// }

// func Test_Workflow_Halts_Activities_On_InvalidAccountStatus(t *testing.T) {
// 	testSuite := &testsuite.WorkflowTestSuite{}
// 	env := testSuite.NewTestWorkflowEnvironment()

// 	var genact *genbenthosconfigs_activity.Activity
// 	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
// 		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
// 			{
// 				Name:        "public.users",
// 				DependsOn:   []*tabledependency.DependsOn{},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "users",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			{
// 				Name:        "public.accounts",
// 				DependsOn:   []*tabledependency.DependsOn{},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "accounts",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			{
// 				Name:        "public.foo",
// 				DependsOn:   []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "foo",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		}}, nil)
// 	var sqlInitActivity *runsqlinittablestmts_activity.Activity
// 	env.OnActivity(sqlInitActivity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
// 		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
// 	var activityOpts *syncactivityopts_activity.Activity
// 	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
// 		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
// 			SyncActivityOptions: &workflow.ActivityOptions{
// 				StartToCloseTimeout: time.Minute,
// 			},
// 		}, nil)

// 	var accStatsActivity *accountstatus_activity.Activity
// 	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
// 		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true, ShouldPoll: true}, nil).Once()
// 	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
// 		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: false}, nil).Once()

// 	syncActivity := sync_activity.Activity{}
// 	env.
// 		OnActivity(syncActivity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "users"}).
// 		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata) (*sync_activity.SyncResponse, error) {
// 			return &sync_activity.SyncResponse{}, nil
// 		})
// 	env.
// 		OnActivity(syncActivity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "accounts"}).
// 		Return(nil, errors.New("AccountTestFailure"))

// 	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

// 	require.True(t, env.IsWorkflowCompleted())

// 	err := env.GetWorkflowError()
// 	require.Error(t, err)
// 	var applicationErr *temporal.ApplicationError
// 	require.True(t, errors.As(err, &applicationErr))
// 	require.ErrorContains(t, applicationErr, invalidAccountStatusError.Error())

// 	env.AssertExpectations(t)
// }

// func Test_Workflow_Cleans_Up_Redis_OnError(t *testing.T) {
// 	testSuite := &testsuite.WorkflowTestSuite{}
// 	env := testSuite.NewTestWorkflowEnvironment()

// 	var genact *genbenthosconfigs_activity.Activity
// 	env.OnActivity(genact.GenerateBenthosConfigs, mock.Anything, mock.Anything).
// 		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
// 			{
// 				Name:        "public.users",
// 				DependsOn:   []*tabledependency.DependsOn{},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "users",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 				RedisConfig: []*genbenthosconfigs_activity.BenthosRedisConfig{
// 					{
// 						Key:    "fake-redis-key",
// 						Table:  "public.users",
// 						Column: "id",
// 					},
// 				},
// 			},
// 			{
// 				Name:        "public.accounts",
// 				DependsOn:   []*tabledependency.DependsOn{},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "accounts",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			{
// 				Name:        "public.foo",
// 				DependsOn:   []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
// 				Columns:     []string{"id"},
// 				TableSchema: "public",
// 				TableName:   "foo",
// 				Config: &neosync_benthos.BenthosConfig{
// 					StreamConfig: neosync_benthos.StreamConfig{
// 						Input: &neosync_benthos.InputConfig{
// 							Inputs: neosync_benthos.Inputs{
// 								SqlSelect: &neosync_benthos.SqlSelect{
// 									Columns: []string{"id"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		}}, nil)
// 	var sqlInitActivity *runsqlinittablestmts_activity.Activity
// 	env.OnActivity(sqlInitActivity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
// 		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
// 	var activityOpts *syncactivityopts_activity.Activity
// 	env.OnActivity(activityOpts.RetrieveActivityOptions, mock.Anything, mock.Anything).
// 		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
// 			SyncActivityOptions: &workflow.ActivityOptions{
// 				StartToCloseTimeout: time.Minute,
// 			},
// 		}, nil)
// 	var accStatsActivity *accountstatus_activity.Activity
// 	env.OnActivity(accStatsActivity.CheckAccountStatus, mock.Anything, mock.Anything).
// 		Return(&accountstatus_activity.CheckAccountStatusResponse{IsValid: true}, nil)

// 	syncActivities := &sync_activity.Activity{}
// 	env.
// 		OnActivity(syncActivities.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "users"}).
// 		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata) (*sync_activity.SyncResponse, error) {
// 			return &sync_activity.SyncResponse{}, nil
// 		})
// 	env.
// 		OnActivity(syncActivities.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "accounts"}, mock.Anything).
// 		Return(nil, errors.New("TestFailure"))

// 	env.OnActivity(syncrediscleanup_activity.DeleteRedisHash, mock.Anything, mock.Anything).
// 		Return(&syncrediscleanup_activity.DeleteRedisHashResponse{}, nil)

// 	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

// 	assert.True(t, env.IsWorkflowCompleted())

// 	err := env.GetWorkflowError()
// 	assert.Error(t, err)
// 	var applicationErr *temporal.ApplicationError
// 	assert.True(t, errors.As(err, &applicationErr))
// 	assert.Equal(t, "TestFailure", applicationErr.Error())

// 	env.AssertExpectations(t)
// }
// func Test_isConfigReady(t *testing.T) {
// 	isReady, err := isConfigReady(nil, nil)
// 	assert.NoError(t, err)
// 	assert.False(t, isReady, "config is nil")

// 	isReady, err = isConfigReady(&genbenthosconfigs_activity.BenthosConfigResponse{
// 		Name:      "foo",
// 		DependsOn: []*tabledependency.DependsOn{},
// 	},
// 		nil)
// 	assert.NoError(t, err)
// 	assert.True(
// 		t,
// 		isReady,
// 		"has no dependencies",
// 	)

// 	completed := sync.Map{}
// 	completed.Store("bar", []string{"id"})
// 	isReady, err = isConfigReady(&genbenthosconfigs_activity.BenthosConfigResponse{
// 		Name:      "foo",
// 		DependsOn: []*tabledependency.DependsOn{{Table: "bar", Columns: []string{"id"}}, {Table: "baz", Columns: []string{"id"}}},
// 	},
// 		&completed)
// 	assert.NoError(t, err)
// 	assert.False(
// 		t,
// 		isReady,
// 		"not all dependencies are finished",
// 	)

// 	completed = sync.Map{}
// 	completed.Store("bar", []string{"id"})
// 	completed.Store("baz", []string{"id"})
// 	isReady, err = isConfigReady(&genbenthosconfigs_activity.BenthosConfigResponse{
// 		Name:      "foo",
// 		DependsOn: []*tabledependency.DependsOn{{Table: "bar", Columns: []string{"id"}}, {Table: "baz", Columns: []string{"id"}}},
// 	}, &completed)
// 	assert.NoError(t, err)
// 	assert.True(
// 		t,
// 		isReady,
// 		"all dependencies are finished",
// 	)

// 	completed = sync.Map{}
// 	completed.Store("bar", []string{"id"})
// 	isReady, err = isConfigReady(&genbenthosconfigs_activity.BenthosConfigResponse{
// 		Name:      "foo",
// 		DependsOn: []*tabledependency.DependsOn{{Table: "bar", Columns: []string{"id", "f_id"}}},
// 	},
// 		&completed)
// 	assert.NoError(t, err)
// 	assert.False(
// 		t,
// 		isReady,
// 		"not all dependencies columns are finished",
// 	)
// }

// func Test_updateCompletedMap(t *testing.T) {
// 	completedMap := sync.Map{}
// 	table := "public.users"
// 	cols := []string{"id"}
// 	err := updateCompletedMap(table, &completedMap, cols)
// 	assert.NoError(t, err)
// 	val, loaded := completedMap.Load(table)
// 	assert.True(t, loaded)
// 	assert.Equal(t, cols, val)

// 	completedMap = sync.Map{}
// 	table = "public.users"
// 	completedMap.Store(table, []string{"name"})
// 	err = updateCompletedMap(table, &completedMap, []string{"id"})
// 	assert.NoError(t, err)
// 	val, loaded = completedMap.Load(table)
// 	assert.True(t, loaded)
// 	assert.Equal(t, []string{"name", "id"}, val)
// }

// func Test_isReadyForCleanUp(t *testing.T) {
// 	assert.True(t, isReadyForCleanUp("", "", nil), "no dependencies")

// 	assert.False(
// 		t,
// 		isReadyForCleanUp(
// 			"table",
// 			"col",
// 			map[string]map[string][]string{
// 				"other_table": {"table": []string{"col"}},
// 			},
// 		),
// 		"has dependency",
// 	)

// 	assert.True(
// 		t,
// 		isReadyForCleanUp(
// 			"table",
// 			"col",
// 			map[string]map[string][]string{
// 				"other_table": {"table": []string{"col1"}},
// 			},
// 		),
// 		"no dependency",
// 	)
// }

// func Test_Workflow_Initial_AccountStatus(t *testing.T) {
// 	testSuite := &testsuite.WorkflowTestSuite{}
// 	env := testSuite.NewTestWorkflowEnvironment()

// 	var activityOptsActivity *syncactivityopts_activity.Activity
// 	env.OnActivity(activityOptsActivity.RetrieveActivityOptions, mock.Anything, mock.Anything).
// 		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
// 			AccountId:            uuid.NewString(),
// 			RequestedRecordCount: shared.Ptr(uint64(4)),
// 		}, nil)

// 	var checkStatusActivity *accountstatus_activity.Activity
// 	env.OnActivity(checkStatusActivity.CheckAccountStatus, mock.Anything, mock.Anything).
// 		Return(&accountstatus_activity.CheckAccountStatusResponse{
// 			IsValid: false,
// 			Reason:  shared.Ptr("test failure"),
// 		}, nil)

// 	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

// 	assert.True(t, env.IsWorkflowCompleted())

// 	err := env.GetWorkflowError()
// 	assert.Error(t, err)
// 	var applicationErr *temporal.ApplicationError
// 	assert.True(t, errors.As(err, &applicationErr))
// 	assert.ErrorContains(t, applicationErr, invalidAccountStatusError.Error())

// 	env.AssertExpectations(t)
// }
