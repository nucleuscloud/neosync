package accounthook_workflow

import (
	"errors"
	"testing"

	accounthook_events "github.com/nucleuscloud/neosync/internal/ee/events"
	execute_hook_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/activities/execute"
	hooks_by_event_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/activities/hooks-by-event"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_ProcessAccountHook(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	var hooksByEventActivity *hooks_by_event_activity.Activity

	env.OnActivity(hooksByEventActivity.GetAccountHooksByEvent, mock.Anything, mock.Anything).
		Return(&hooks_by_event_activity.RunHooksByEventResponse{
			HookIds: []string{"hook1", "hook2"},
		}, nil).Once()

	var executeHookActivity *execute_hook_activity.Activity
	env.OnActivity(executeHookActivity.ExecuteAccountHook, mock.Anything, mock.Anything).
		Return(&execute_hook_activity.ExecuteHookResponse{}, nil).Twice()

	env.RegisterWorkflow(ProcessAccountHook)

	env.ExecuteWorkflow(ProcessAccountHook, &ProcessAccountHookRequest{
		Event: accounthook_events.NewEvent_JobRunCreated("123", "456", "789"),
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result *ProcessAccountHookResponse
	err := env.GetWorkflowResult(&result)
	require.NoError(t, err)
	require.Equal(t, &ProcessAccountHookResponse{}, result)

	env.AssertExpectations(t)
}

func Test_ProcessAccountHook_Error(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	var hooksByEventActivity *hooks_by_event_activity.Activity
	env.OnActivity(hooksByEventActivity.GetAccountHooksByEvent, mock.Anything, mock.Anything).
		Return(&hooks_by_event_activity.RunHooksByEventResponse{
			HookIds: []string{"hook1", "hook2"},
		}, nil).Once()

	var executeHookActivity *execute_hook_activity.Activity
	env.OnActivity(executeHookActivity.ExecuteAccountHook, mock.Anything, mock.Anything).
		Return(nil, errors.New("error"))

	env.RegisterWorkflow(ProcessAccountHook)

	env.ExecuteWorkflow(ProcessAccountHook, &ProcessAccountHookRequest{
		Event: accounthook_events.NewEvent_JobRunCreated("123", "456", "789"),
	})

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}
