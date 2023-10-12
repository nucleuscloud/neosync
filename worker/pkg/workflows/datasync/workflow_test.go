package datasync

import (
	"context"
	"errors"
	"testing"

	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_Workflow_BenthosConfigsFails() {
	activities := &Activities{}
	s.env.OnActivity(activities.GenerateBenthosConfigs, mock.Anything, mock.Anything).Return(nil, errors.New("TestFailure"))

	s.env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

	s.True(s.env.IsWorkflowCompleted())

	err := s.env.GetWorkflowError()
	s.Error(err)
	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("TestFailure", applicationErr.Error())
}

func (s *UnitTestSuite) Test_Workflow_Succeeds_Zero_BenthosConfigs() {
	activities := &Activities{}
	s.env.OnActivity(activities.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosConfigResponse{}}, nil)

	s.env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

	s.True(s.env.IsWorkflowCompleted())

	err := s.env.GetWorkflowError()
	s.Nil(err)

	result := &WorkflowResponse{}
	err = s.env.GetWorkflowResult(result)
	s.Nil(err)
	s.Equal(result, &WorkflowResponse{})
}

func (s *UnitTestSuite) Test_Workflow_Succeeds_SingleSync() {
	activities := &Activities{}
	s.env.OnActivity(activities.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosConfigResponse{
			{
				Name:      "public.users",
				DependsOn: []string{},
				Config:    &neosync_benthos.BenthosConfig{},
			},
		}}, nil)
	s.env.OnActivity(activities.Sync, mock.Anything, mock.Anything, mock.Anything).Return(&SyncResponse{}, nil)

	s.env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

	s.True(s.env.IsWorkflowCompleted())

	err := s.env.GetWorkflowError()
	s.Nil(err)

	result := &WorkflowResponse{}
	err = s.env.GetWorkflowResult(result)
	s.Nil(err)
	s.Equal(result, &WorkflowResponse{})
}

func (s *UnitTestSuite) Test_Workflow_Follows_Synchronous_DependentFlow() {
	activities := &Activities{}
	s.env.OnActivity(activities.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosConfigResponse{
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
	s.env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &SyncMetadata{Schema: "public", Table: "users"}).
		Return(func(ctx context.Context, req *SyncRequest, metadata *SyncMetadata) (*SyncResponse, error) {
			s.Equal(count, 0)
			count += 1
			return &SyncResponse{}, nil
		})
	s.env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &SyncMetadata{Schema: "public", Table: "foo"}).
		Return(func(ctx context.Context, req *SyncRequest, metadata *SyncMetadata) (*SyncResponse, error) {
			s.Equal(count, 1)
			count += 1
			return &SyncResponse{}, nil
		})

	s.env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

	s.True(s.env.IsWorkflowCompleted())
	s.Equal(count, 2)

	err := s.env.GetWorkflowError()
	s.Nil(err)

	result := &WorkflowResponse{}
	err = s.env.GetWorkflowResult(result)
	s.Nil(err)
	s.Equal(result, &WorkflowResponse{})
}

func (s *UnitTestSuite) Test_Workflow_Follows_Multiple_Dependents() {
	activities := &Activities{}
	s.env.OnActivity(activities.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosConfigResponse{
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
	count := 0
	s.env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &SyncMetadata{Schema: "public", Table: "users"}).
		Return(func(ctx context.Context, req *SyncRequest, metadata *SyncMetadata) (*SyncResponse, error) {
			count += 1
			return &SyncResponse{}, nil
		})
	s.env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &SyncMetadata{Schema: "public", Table: "accounts"}).
		Return(func(ctx context.Context, req *SyncRequest, metadata *SyncMetadata) (*SyncResponse, error) {
			count += 1
			return &SyncResponse{}, nil
		})
	s.env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &SyncMetadata{Schema: "public", Table: "foo"}).
		Return(func(ctx context.Context, req *SyncRequest, metadata *SyncMetadata) (*SyncResponse, error) {
			s.Equal(count, 2)
			count += 1
			return &SyncResponse{}, nil
		})

	s.env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

	s.True(s.env.IsWorkflowCompleted())
	s.Equal(count, 3)

	err := s.env.GetWorkflowError()
	s.Nil(err)

	result := &WorkflowResponse{}
	err = s.env.GetWorkflowResult(result)
	s.Nil(err)
	s.Equal(result, &WorkflowResponse{})
}

func (s *UnitTestSuite) Test_Workflow_Halts_Activities_OnError() {
	activities := &Activities{}
	s.env.OnActivity(activities.GenerateBenthosConfigs, mock.Anything, mock.Anything).
		Return(&GenerateBenthosConfigsResponse{BenthosConfigs: []*benthosConfigResponse{
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

	s.env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &SyncMetadata{Schema: "public", Table: "users"}).
		Return(func(ctx context.Context, req *SyncRequest, metadata *SyncMetadata) (*SyncResponse, error) {
			return &SyncResponse{}, nil
		})
	s.env.
		OnActivity(activities.Sync, mock.Anything, mock.Anything, &SyncMetadata{Schema: "public", Table: "accounts"}).
		Return(nil, errors.New("TestFailure"))

	s.env.ExecuteWorkflow(Workflow, &WorkflowRequest{})

	s.True(s.env.IsWorkflowCompleted())

	err := s.env.GetWorkflowError()
	s.Error(err)
	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("TestFailure", applicationErr.Error())
}
