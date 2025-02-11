package accounthook_workflow

import (
	"fmt"
	"time"

	execute_hook_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/activities/execute"
	hooks_by_event_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/activities/hooks-by-event"
	accounthook_events "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/events"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type ProcessAccountHookRequest struct {
	Event *accounthook_events.Event
}

type ProcessAccountHookResponse struct{}

func ProcessAccountHook(wfctx workflow.Context, req *ProcessAccountHookRequest) (*ProcessAccountHookResponse, error) {
	var hooksByEventActivity *hooks_by_event_activity.Activity
	var resp *hooks_by_event_activity.RunHooksByEventResponse
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(wfctx, workflow.ActivityOptions{
			StartToCloseTimeout: 1 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 1,
			},
			HeartbeatTimeout: 1 * time.Minute,
			Summary:          "Retrieves the configured account hooks for the given event",
		}),
		hooksByEventActivity.GetAccountHooksByEvent,
		&hooks_by_event_activity.RunHooksByEventRequest{
			AccountId: req.Event.AccountId,
			EventName: req.Event.Name,
		}).
		Get(wfctx, &resp)
	if err != nil {
		return nil, err
	}

	futures := make([]workflow.Future, len(resp.HookIds))
	var executeHookActivity *execute_hook_activity.Activity

	for i, hookId := range resp.HookIds {
		futures[i] = workflow.ExecuteActivity(
			workflow.WithActivityOptions(wfctx, workflow.ActivityOptions{
				StartToCloseTimeout: 5 * time.Minute,
				RetryPolicy: &temporal.RetryPolicy{
					MaximumAttempts: 3,
				},
				Summary:          "Runs the configured account hook",
				HeartbeatTimeout: 1 * time.Minute,
			}),
			executeHookActivity.ExecuteAccountHook,
			&execute_hook_activity.ExecuteHookRequest{
				HookId: hookId,
				Event:  req.Event,
			},
		)
	}

	for _, future := range futures {
		var execResp *execute_hook_activity.ExecuteHookResponse
		if err := future.Get(wfctx, &execResp); err != nil {
			return nil, fmt.Errorf("error executing hook: %w", err)
		}
	}

	return &ProcessAccountHookResponse{}, nil
}
