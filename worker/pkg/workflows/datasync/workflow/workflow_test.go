package datasync_workflow

import (
	"context"
	"errors"
	"testing"
	"time"

	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/atomic"
)

func Test_Workflow_BenthosConfigsFails(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.OnActivity(genbenthosconfigs_activity.GenerateBenthosConfigs, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("TestFailure"))

	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

	assert.True(t, env.IsWorkflowCompleted())
	assert.True(t, env.IsWorkflowCompleted())

	err := env.GetWorkflowError()
	assert.Error(t, err)
	var applicationErr *temporal.ApplicationError
	assert.True(t, errors.As(err, &applicationErr))
	assert.Equal(t, "TestFailure", applicationErr.Error())

	env.AssertExpectations(t)
}

func Test_Workflow_Succeeds_Zero_BenthosConfigs(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.OnActivity(genbenthosconfigs_activity.GenerateBenthosConfigs, mock.Anything, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{}}, nil)

	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

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

	env.OnActivity(genbenthosconfigs_activity.GenerateBenthosConfigs, mock.Anything, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
			{
				Name:      "public.users",
				DependsOn: []*tabledependency.DependsOn{},
				Config:    &neosync_benthos.BenthosConfig{},
			},
		}}, nil)
	env.OnActivity(syncactivityopts_activity.RetrieveActivityOptions, mock.Anything, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
		}, nil)
	env.OnActivity(runsqlinittablestmts_activity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
	env.OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&sync_activity.SyncResponse{}, nil)

	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

	assert.True(t, env.IsWorkflowCompleted())

	err := env.GetWorkflowError()
	assert.Nil(t, err)

	result := &WorkflowResponse{}
	err = env.GetWorkflowResult(result)
	assert.Nil(t, err)
	assert.Equal(t, result, &WorkflowResponse{})

	env.AssertExpectations(t)
}

func Test_Workflow_Follows_Synchronous_DependentFlow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.OnActivity(genbenthosconfigs_activity.GenerateBenthosConfigs, mock.Anything, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
			{
				Name:      "public.users",
				DependsOn: []*tabledependency.DependsOn{},
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
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
				DependsOn: []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}},
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
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
	env.OnActivity(syncactivityopts_activity.RetrieveActivityOptions, mock.Anything, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
		}, nil)
	env.OnActivity(runsqlinittablestmts_activity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
	count := 0
	env.
		OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "users"}, mock.Anything).
		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata, workflowMetadata *shared.WorkflowMetadata) (*sync_activity.SyncResponse, error) {
			assert.Equal(t, count, 0)
			count += 1
			return &sync_activity.SyncResponse{}, nil
		})
	env.
		OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "foo"}, mock.Anything).
		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata, workflowMetadata *shared.WorkflowMetadata) (*sync_activity.SyncResponse, error) {
			assert.Equal(t, count, 1)
			count += 1
			return &sync_activity.SyncResponse{}, nil
		})

	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

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

	env.OnActivity(genbenthosconfigs_activity.GenerateBenthosConfigs, mock.Anything, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
			{
				Name:        "public.users",
				DependsOn:   []*tabledependency.DependsOn{},
				TableSchema: "public",
				TableName:   "users",
				Columns:     []string{"id"},
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
								},
							},
						},
					},
				},
			},
			{
				Name:        "public.accounts",
				DependsOn:   []*tabledependency.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "accounts",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
								},
							},
						},
					},
				},
			},
			{
				Name:        "public.foo",
				DependsOn:   []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "foo",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
								},
							},
						},
					},
				},
			},
		}}, nil)
	env.OnActivity(syncactivityopts_activity.RetrieveActivityOptions, mock.Anything, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
		}, nil)
	env.OnActivity(runsqlinittablestmts_activity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
	counter := atomic.NewInt32(0)
	env.
		OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "users"}, mock.Anything).
		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata, workflowMetadata *shared.WorkflowMetadata) (*sync_activity.SyncResponse, error) {
			counter.Add(1)
			return &sync_activity.SyncResponse{}, nil
		})
	env.
		OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "accounts"}, mock.Anything).
		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata, workflowMetadata *shared.WorkflowMetadata) (*sync_activity.SyncResponse, error) {
			counter.Add(1)
			return &sync_activity.SyncResponse{}, nil
		})
	env.
		OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "foo"}, mock.Anything).
		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata, workflowMetadata *shared.WorkflowMetadata) (*sync_activity.SyncResponse, error) {
			assert.Equal(t, counter.Load(), int32(2))
			counter.Add(1)
			return &sync_activity.SyncResponse{}, nil
		})

	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

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

	env.OnActivity(genbenthosconfigs_activity.GenerateBenthosConfigs, mock.Anything, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
			{
				Name:        "public.users",
				DependsOn:   []*tabledependency.DependsOn{},
				TableSchema: "public",
				TableName:   "users",
				Columns:     []string{"id"},
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
								},
							},
						},
					},
				},
				RedisConfig: []*genbenthosconfigs_activity.BenthosRedisConfig{
					{
						Key:    "fake-redis-key",
						Table:  "public.users",
						Column: "id",
					},
				},
			},
			{
				Name:        "public.accounts",
				DependsOn:   []*tabledependency.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "accounts",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
								},
							},
						},
					},
				},
				RedisConfig: []*genbenthosconfigs_activity.BenthosRedisConfig{
					{
						Key:    "fake-redis-key2",
						Table:  "public.accounts",
						Column: "id",
					},
				},
			},
			{
				Name:        "public.foo",
				DependsOn:   []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "foo",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
								},
							},
						},
					},
				},
			},
		}}, nil)
	env.OnActivity(syncactivityopts_activity.RetrieveActivityOptions, mock.Anything, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
		}, nil)
	env.OnActivity(runsqlinittablestmts_activity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
	counter := atomic.NewInt32(0)
	env.
		OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "users"}, mock.Anything).
		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata, workflowMetadata *shared.WorkflowMetadata) (*sync_activity.SyncResponse, error) {
			counter.Add(1)
			return &sync_activity.SyncResponse{}, nil
		})
	env.
		OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "accounts"}, mock.Anything).
		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata, workflowMetadata *shared.WorkflowMetadata) (*sync_activity.SyncResponse, error) {
			counter.Add(1)
			return &sync_activity.SyncResponse{}, nil
		})
	env.
		OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "foo"}, mock.Anything).
		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata, workflowMetadata *shared.WorkflowMetadata) (*sync_activity.SyncResponse, error) {
			assert.Equal(t, counter.Load(), int32(2))
			counter.Add(1)
			return &sync_activity.SyncResponse{}, nil
		})

	env.OnActivity(syncrediscleanup_activity.DeleteRedisHash, mock.Anything, mock.Anything).
		Return(&syncrediscleanup_activity.DeleteRedisHashResponse{}, nil)
	env.OnActivity(syncrediscleanup_activity.DeleteRedisHash, mock.Anything, mock.Anything).
		Return(&syncrediscleanup_activity.DeleteRedisHashResponse{}, nil)

	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

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

func Test_Workflow_Halts_Activities_OnError(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.OnActivity(genbenthosconfigs_activity.GenerateBenthosConfigs, mock.Anything, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
			{
				Name:        "public.users",
				DependsOn:   []*tabledependency.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "users",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
								},
							},
						},
					},
				},
			},
			{
				Name:        "public.accounts",
				DependsOn:   []*tabledependency.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "accounts",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
								},
							},
						},
					},
				},
			},
			{
				Name:        "public.foo",
				DependsOn:   []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "foo",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
								},
							},
						},
					},
				},
			},
		}}, nil)
	env.OnActivity(runsqlinittablestmts_activity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
	env.OnActivity(syncactivityopts_activity.RetrieveActivityOptions, mock.Anything, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
		}, nil)

	env.
		OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "users"}, mock.Anything).
		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata, workflowMetadata *shared.WorkflowMetadata) (*sync_activity.SyncResponse, error) {
			return &sync_activity.SyncResponse{}, nil
		})
	env.
		OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "accounts"}, mock.Anything).
		Return(nil, errors.New("TestFailure"))

	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

	assert.True(t, env.IsWorkflowCompleted())

	err := env.GetWorkflowError()
	assert.Error(t, err)
	var applicationErr *temporal.ApplicationError
	assert.True(t, errors.As(err, &applicationErr))
	assert.Equal(t, "TestFailure", applicationErr.Error())

	env.AssertExpectations(t)
}

func Test_Workflow_Cleans_Up_Redis_OnError(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.OnActivity(genbenthosconfigs_activity.GenerateBenthosConfigs, mock.Anything, mock.Anything, mock.Anything).
		Return(&genbenthosconfigs_activity.GenerateBenthosConfigsResponse{BenthosConfigs: []*genbenthosconfigs_activity.BenthosConfigResponse{
			{
				Name:        "public.users",
				DependsOn:   []*tabledependency.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "users",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
								},
							},
						},
					},
				},
				RedisConfig: []*genbenthosconfigs_activity.BenthosRedisConfig{
					{
						Key:    "fake-redis-key",
						Table:  "public.users",
						Column: "id",
					},
				},
			},
			{
				Name:        "public.accounts",
				DependsOn:   []*tabledependency.DependsOn{},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "accounts",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
								},
							},
						},
					},
				},
			},
			{
				Name:        "public.foo",
				DependsOn:   []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.accounts", Columns: []string{"id"}}},
				Columns:     []string{"id"},
				TableSchema: "public",
				TableName:   "foo",
				Config: &neosync_benthos.BenthosConfig{
					StreamConfig: neosync_benthos.StreamConfig{
						Input: &neosync_benthos.InputConfig{
							Inputs: neosync_benthos.Inputs{
								SqlSelect: &neosync_benthos.SqlSelect{
									Columns: []string{"id"},
								},
							},
						},
					},
				},
			},
		}}, nil)
	env.OnActivity(runsqlinittablestmts_activity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
		Return(&runsqlinittablestmts_activity.RunSqlInitTableStatementsResponse{}, nil)
	env.OnActivity(syncactivityopts_activity.RetrieveActivityOptions, mock.Anything, mock.Anything, mock.Anything).
		Return(&syncactivityopts_activity.RetrieveActivityOptionsResponse{
			SyncActivityOptions: &workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			},
		}, nil)

	env.
		OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "users"}, mock.Anything).
		Return(func(ctx context.Context, req *sync_activity.SyncRequest, metadata *sync_activity.SyncMetadata, workflowMetadata *shared.WorkflowMetadata) (*sync_activity.SyncResponse, error) {
			return &sync_activity.SyncResponse{}, nil
		})
	env.
		OnActivity(sync_activity.Sync, mock.Anything, mock.Anything, &sync_activity.SyncMetadata{Schema: "public", Table: "accounts"}, mock.Anything).
		Return(nil, errors.New("TestFailure"))

	env.OnActivity(syncrediscleanup_activity.DeleteRedisHash, mock.Anything, mock.Anything).
		Return(&syncrediscleanup_activity.DeleteRedisHashResponse{}, nil)

	env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

	assert.True(t, env.IsWorkflowCompleted())

	err := env.GetWorkflowError()
	assert.Error(t, err)
	var applicationErr *temporal.ApplicationError
	assert.True(t, errors.As(err, &applicationErr))
	assert.Equal(t, "TestFailure", applicationErr.Error())

	env.AssertExpectations(t)
}
func Test_isConfigReady(t *testing.T) {
	assert.False(t, isConfigReady(nil, nil), "config is nil")
	assert.True(
		t,
		isConfigReady(
			&genbenthosconfigs_activity.BenthosConfigResponse{
				Name:      "foo",
				DependsOn: []*tabledependency.DependsOn{},
			},
			nil,
		),
		"has no dependencies",
	)

	assert.False(
		t,
		isConfigReady(
			&genbenthosconfigs_activity.BenthosConfigResponse{
				Name:      "foo",
				DependsOn: []*tabledependency.DependsOn{{Table: "bar", Columns: []string{"id"}}, {Table: "baz", Columns: []string{"id"}}},
			},
			map[string][]string{
				"bar": {"id"},
			},
		),
		"not all dependencies are finished",
	)

	assert.True(
		t,
		isConfigReady(
			&genbenthosconfigs_activity.BenthosConfigResponse{
				Name:      "foo",
				DependsOn: []*tabledependency.DependsOn{{Table: "bar", Columns: []string{"id"}}, {Table: "baz", Columns: []string{"id"}}},
			},
			map[string][]string{
				"bar": {"id"},
				"baz": {"id"},
			},
		),
		"all dependencies are finished",
	)

	assert.False(
		t,
		isConfigReady(
			&genbenthosconfigs_activity.BenthosConfigResponse{
				Name:      "foo",
				DependsOn: []*tabledependency.DependsOn{{Table: "bar", Columns: []string{"id", "f_id"}}},
			},
			map[string][]string{
				"bar": {"id"},
			},
		),
		"not all dependencies columns are finished",
	)
}

func Test_isReadyForCleanUp(t *testing.T) {
	assert.True(t, isReadyForCleanUp("", "", nil), "no dependencies")

	assert.False(
		t,
		isReadyForCleanUp(
			"table",
			"col",
			map[string][]*tabledependency.DependsOn{
				"config": {{
					Table:   "table",
					Columns: []string{"col"},
				}},
			},
		),
		"has dependency",
	)

	assert.True(
		t,
		isReadyForCleanUp(
			"table",
			"col",
			map[string][]*tabledependency.DependsOn{
				"config": {{
					Table:   "table",
					Columns: []string{"col1"},
				}},
			},
		),
		"has dependency",
	)
}
