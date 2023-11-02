package datasync_workflow

import (
	"context"
	"errors"
	"testing"

	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	datasync_activities "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.uber.org/atomic"
)

func Test_Workflow_BenthosConfigsFails(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	activities := &datasync_activities.Activities{}
	env.OnActivity(activities.GenerateBenthosConfigs, mock.Anything, mock.Anything).Return(nil, errors.New("TestFailure"))

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

	activities := &datasync_activities.Activities{}
	env.OnActivity(activities.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&datasync_activities.GenerateBenthosConfigsResponse{BenthosConfigs: []*datasync_activities.BenthosConfigResponse{}}, nil)

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

	activities := &datasync_activities.Activities{}
	env.OnActivity(activities.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&datasync_activities.GenerateBenthosConfigsResponse{BenthosConfigs: []*datasync_activities.BenthosConfigResponse{
			{
				Name:      "public.users",
				DependsOn: []string{},
				Config:    &neosync_benthos.BenthosConfig{},
			},
		}}, nil)
	env.OnActivity(activities.Sync, mock.Anything, mock.Anything, mock.Anything).Return(&datasync_activities.SyncResponse{}, nil)

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

	activities := &datasync_activities.Activities{}
	env.OnActivity(activities.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&datasync_activities.GenerateBenthosConfigsResponse{BenthosConfigs: []*datasync_activities.BenthosConfigResponse{
			{
				Name:      "public.users",
				DependsOn: []string{},
				Config:    &neosync_benthos.BenthosConfig{},
			},
			{
				Name:      "public.foo",
				DependsOn: []string{"public.users"},
			},
		}}, nil)
	count := 0
	env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &datasync_activities.SyncMetadata{Schema: "public", Table: "users"}).
		Return(func(ctx context.Context, req *datasync_activities.SyncRequest, metadata *datasync_activities.SyncMetadata) (*datasync_activities.SyncResponse, error) {
			assert.Equal(t, count, 0)
			count += 1
			return &datasync_activities.SyncResponse{}, nil
		})
	env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &datasync_activities.SyncMetadata{Schema: "public", Table: "foo"}).
		Return(func(ctx context.Context, req *datasync_activities.SyncRequest, metadata *datasync_activities.SyncMetadata) (*datasync_activities.SyncResponse, error) {
			assert.Equal(t, count, 1)
			count += 1
			return &datasync_activities.SyncResponse{}, nil
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

	activities := &datasync_activities.Activities{}
	env.OnActivity(activities.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&datasync_activities.GenerateBenthosConfigsResponse{BenthosConfigs: []*datasync_activities.BenthosConfigResponse{
			{
				Name:      "public.users",
				DependsOn: []string{},
				Config:    &neosync_benthos.BenthosConfig{},
			},
			{
				Name:      "public.accounts",
				DependsOn: []string{},
				Config:    &neosync_benthos.BenthosConfig{},
			},
			{
				Name:      "public.foo",
				DependsOn: []string{"public.users", "public.accounts"},
			},
		}}, nil)
	counter := atomic.NewInt32(0)
	env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &datasync_activities.SyncMetadata{Schema: "public", Table: "users"}).
		Return(func(ctx context.Context, req *datasync_activities.SyncRequest, metadata *datasync_activities.SyncMetadata) (*datasync_activities.SyncResponse, error) {
			counter.Add(1)
			return &datasync_activities.SyncResponse{}, nil
		})
	env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &datasync_activities.SyncMetadata{Schema: "public", Table: "accounts"}).
		Return(func(ctx context.Context, req *datasync_activities.SyncRequest, metadata *datasync_activities.SyncMetadata) (*datasync_activities.SyncResponse, error) {
			counter.Add(1)
			return &datasync_activities.SyncResponse{}, nil
		})
	env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &datasync_activities.SyncMetadata{Schema: "public", Table: "foo"}).
		Return(func(ctx context.Context, req *datasync_activities.SyncRequest, metadata *datasync_activities.SyncMetadata) (*datasync_activities.SyncResponse, error) {
			assert.Equal(t, counter.Load(), int32(2))
			counter.Add(1)
			return &datasync_activities.SyncResponse{}, nil
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

func Test_Workflow_Halts_Activities_OnError(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	activities := &datasync_activities.Activities{}
	env.OnActivity(activities.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&datasync_activities.GenerateBenthosConfigsResponse{BenthosConfigs: []*datasync_activities.BenthosConfigResponse{
			{
				Name:      "public.users",
				DependsOn: []string{},
				Config:    &neosync_benthos.BenthosConfig{},
			},
			{
				Name:      "public.accounts",
				DependsOn: []string{},
				Config:    &neosync_benthos.BenthosConfig{},
			},
			{
				Name:      "public.foo",
				DependsOn: []string{"public.users", "public.accounts"},
			},
		}}, nil)

	env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &datasync_activities.SyncMetadata{Schema: "public", Table: "users"}).
		Return(func(ctx context.Context, req *datasync_activities.SyncRequest, metadata *datasync_activities.SyncMetadata) (*datasync_activities.SyncResponse, error) {
			return &datasync_activities.SyncResponse{}, nil
		})
	env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &datasync_activities.SyncMetadata{Schema: "public", Table: "accounts"}).
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
