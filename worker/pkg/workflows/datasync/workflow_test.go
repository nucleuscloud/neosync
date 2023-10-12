package datasync

import (
	"errors"
	"testing"

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
