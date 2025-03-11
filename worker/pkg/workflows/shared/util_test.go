package workflow_shared

import (
	"errors"
	"testing"

	"github.com/nucleuscloud/neosync/internal/testutil"
	accounthook_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func Test_SanitizeWorkflowID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "replaces special characters",
			input:    "public.users@123",
			expected: "public_users_123",
		},
		{
			name:     "keeps valid characters",
			input:    "public-users-123",
			expected: "public-users-123",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeWorkflowID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_HandleWorkflowEventLifecycle(t *testing.T) {
	t.Run("executes function successfully with valid license", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		// Register account hook workflow
		env.RegisterWorkflow(accounthook_workflow.ProcessAccountHook)

		// Mock the account hook workflow calls
		env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
			Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil).Times(2)

		// Setup test data
		jobId := "job-123"
		runId := "run-456"
		accountId := "acc-789"
		expectedResult := "success"

		// Create a mock function that returns our test result
		testFn := func(ctx workflow.Context, logger log.Logger) (*string, error) {
			return &expectedResult, nil
		}

		// Create a mock getAccountId function
		getAccountId := func() (string, error) {
			return accountId, nil
		}

		// Execute the workflow
		var result *string
		env.ExecuteWorkflow(func(ctx workflow.Context) (*string, error) {
			return HandleWorkflowEventLifecycle(
				ctx,
				testutil.NewFakeEELicense(testutil.WithIsValid()),
				jobId,
				runId,
				workflow.GetLogger(ctx),
				getAccountId,
				testFn,
			)
		})

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		require.NoError(t, env.GetWorkflowResult(&result))
		assert.Equal(t, expectedResult, *result)
	})

	t.Run("skips events when license is invalid", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		// Register account hook workflow
		env.RegisterWorkflow(accounthook_workflow.ProcessAccountHook)

		// Mock the account hook workflow - should never be called
		env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
			Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil).Never()

		// Setup test data
		jobId := "job-123"
		runId := "run-456"
		expectedResult := "success"

		// Create a mock function that returns our test result
		testFn := func(ctx workflow.Context, logger log.Logger) (*string, error) {
			return &expectedResult, nil
		}

		// Create a mock getAccountId function - should never be called
		getAccountId := func() (string, error) {
			t.Fatal("getAccountId should not be called when license is invalid")
			return "", nil
		}

		// Execute the workflow
		var result *string
		env.ExecuteWorkflow(func(ctx workflow.Context) (*string, error) {
			return HandleWorkflowEventLifecycle(
				ctx,
				testutil.NewFakeEELicense(), // invalid license
				jobId,
				runId,
				workflow.GetLogger(ctx),
				getAccountId,
				testFn,
			)
		})

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		require.NoError(t, env.GetWorkflowResult(&result))
		assert.Equal(t, expectedResult, *result)
	})

	t.Run("handles getAccountId error", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		// Register account hook workflow
		env.RegisterWorkflow(accounthook_workflow.ProcessAccountHook)

		// Mock the account hook workflow - should never be called since getAccountId fails
		env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
			Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil).Never()

		// Setup test data
		jobId := "job-123"
		runId := "run-456"
		expectedError := errors.New("failed to get account id")

		// Create a mock function that should never be called
		testFn := func(ctx workflow.Context, logger log.Logger) (*string, error) {
			t.Fatal("testFn should not be called when getAccountId fails")
			return nil, nil
		}

		// Create a mock getAccountId function that returns an error
		getAccountId := func() (string, error) {
			return "", expectedError
		}

		// Execute the workflow
		env.ExecuteWorkflow(func(ctx workflow.Context) (*string, error) {
			return HandleWorkflowEventLifecycle(
				ctx,
				testutil.NewFakeEELicense(testutil.WithIsValid()),
				jobId,
				runId,
				workflow.GetLogger(ctx),
				getAccountId,
				testFn,
			)
		})

		require.True(t, env.IsWorkflowCompleted())
		workflowErr := env.GetWorkflowError()
		require.Error(t, workflowErr)
		assert.Contains(t, workflowErr.Error(), expectedError.Error())
	})

	t.Run("handles function error and triggers failed event", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		// Register account hook workflow
		env.RegisterWorkflow(accounthook_workflow.ProcessAccountHook)

		// Mock the account hook workflow calls - expect created and failed events
		env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
			Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil).Times(2)

		// Setup test data
		jobId := "job-123"
		runId := "run-456"
		accountId := "acc-789"
		expectedError := errors.New("function failed")

		// Create a mock function that returns an error
		testFn := func(ctx workflow.Context, logger log.Logger) (*string, error) {
			return nil, expectedError
		}

		// Create a mock getAccountId function
		getAccountId := func() (string, error) {
			return accountId, nil
		}

		// Execute the workflow
		env.ExecuteWorkflow(func(ctx workflow.Context) (*string, error) {
			return HandleWorkflowEventLifecycle(
				ctx,
				testutil.NewFakeEELicense(testutil.WithIsValid()),
				jobId,
				runId,
				workflow.GetLogger(ctx),
				getAccountId,
				testFn,
			)
		})

		require.True(t, env.IsWorkflowCompleted())
		workflowErr := env.GetWorkflowError()
		require.Error(t, workflowErr)
		assert.Contains(t, workflowErr.Error(), expectedError.Error())
	})
}
