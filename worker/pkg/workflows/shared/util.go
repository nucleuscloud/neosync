package workflow_shared

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	accounthook_events "github.com/nucleuscloud/neosync/internal/ee/events"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	accounthook_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/workflow"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

// Utility function that handles spawning job run lifecycle hooks: created, success, failed
// Should only be used by root workflows that are responsible for handling the lifecycle of a job run
func HandleWorkflowEventLifecycle[T any](
	ctx workflow.Context,
	eelicense license.EEInterface,
	jobId,
	runId string, // typically the temporal workflow execution id
	logger log.Logger,
	getAccountId func() (string, error),
	fn func(ctx workflow.Context, logger log.Logger) (*T, error),
) (*T, error) {
	if !eelicense.IsValid() {
		logger.Debug("ee license is not valid, skipping event lifecycle")
		return fn(ctx, logger)
	}

	accountId, err := getAccountId()
	if err != nil {
		return nil, err
	}

	createdFuture := workflow.ExecuteChildWorkflow(
		workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
			WorkflowID:        getAccountHookChildWorkflowId(runId, "job-run-created", workflow.Now(ctx)),
			StaticSummary:     "Account Hook: Job Run Created",
		}),
		accounthook_workflow.ProcessAccountHook,
		&accounthook_workflow.ProcessAccountHookRequest{
			Event: accounthook_events.NewEvent_JobRunCreated(accountId, jobId, runId),
		},
	)
	if err := ensureChildSpawned(ctx, createdFuture, logger); err != nil {
		return nil, err
	}

	resp, err := fn(ctx, logger)
	if err != nil {
		failedFuture := workflow.ExecuteChildWorkflow(
			workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
				ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
				WorkflowID:        getAccountHookChildWorkflowId(runId, "job-run-failed", workflow.Now(ctx)),
				StaticSummary:     "Account Hook: Job Run Failed",
			}),
			accounthook_workflow.ProcessAccountHook,
			&accounthook_workflow.ProcessAccountHookRequest{
				Event: accounthook_events.NewEvent_JobRunFailed(accountId, jobId, runId),
			},
		)
		if err := ensureChildSpawned(ctx, failedFuture, logger); err != nil {
			return nil, err
		}
		return nil, err
	}

	completedFuture := workflow.ExecuteChildWorkflow(
		workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
			WorkflowID:        getAccountHookChildWorkflowId(runId, "job-run-succeeded", workflow.Now(ctx)),
			StaticSummary:     "Account Hook: Job Run Succeeded",
		}),
		accounthook_workflow.ProcessAccountHook,
		&accounthook_workflow.ProcessAccountHookRequest{
			Event: accounthook_events.NewEvent_JobRunSucceeded(accountId, jobId, runId),
		},
	)
	if err := ensureChildSpawned(ctx, completedFuture, logger); err != nil {
		return nil, err
	}

	return resp, nil
}

func ensureChildSpawned(ctx workflow.Context, future workflow.ChildWorkflowFuture, logger log.Logger) error {
	var childWE workflow.Execution
	if waitErr := future.GetChildWorkflowExecution().Get(ctx, &childWE); waitErr != nil {
		return waitErr
	}
	logger.Debug(fmt.Sprintf("child wf event spawned: %s", childWE.ID))
	return nil
}

func getAccountHookChildWorkflowId(parentJobRunId, eventName string, now time.Time) string {
	return BuildChildWorkflowId(parentJobRunId, "hook-"+eventName, now)
}

// Builds a child workflow id that is unique for the given parent execution. Sanitizes the name and cuts to the max allowed limit
func BuildChildWorkflowId(parentExecutionId, name string, ts time.Time) string {
	id := fmt.Sprintf("%s-%s-%d", parentExecutionId, SanitizeWorkflowID(strings.ToLower(name)), ts.UnixNano())
	if len(id) > 1000 {
		id = id[:1000]
	}
	return id
}

var invalidWorkflowIDChars = regexp.MustCompile(`[^a-zA-Z0-9_\-]`)

func SanitizeWorkflowID(id string) string {
	return invalidWorkflowIDChars.ReplaceAllString(id, "_")
}
