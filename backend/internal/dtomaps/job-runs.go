package dtomaps

import (
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	temporalfailure "go.temporal.io/api/failure/v1"
	"go.temporal.io/api/history/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// including pending activities
func ToJobRunDto(
	logger *slog.Logger,
	input *workflowservice.DescribeWorkflowExecutionResponse,
) *mgmtv1alpha1.JobRun {
	jr := ToJobRunDtoFromWorkflowExecutionInfo(input.GetWorkflowExecutionInfo(), logger)
	jr.PendingActivities = toPendingActivitiesDto(input.GetPendingActivities())
	return jr
}

// returns a job run without any pending activities
func ToJobRunDtoFromWorkflowExecutionInfo(workflow *workflowpb.WorkflowExecutionInfo, logger *slog.Logger) *mgmtv1alpha1.JobRun {
	var completedTime *timestamppb.Timestamp
	if workflow.GetCloseTime() != nil {
		completedTime = workflow.GetCloseTime()
	}
	return &mgmtv1alpha1.JobRun{
		Id:          workflow.GetExecution().GetWorkflowId(),
		JobId:       GetJobIdFromWorkflow(logger, workflow.GetSearchAttributes()),
		Name:        workflow.GetType().GetName(),
		Status:      toWorfklowStatus(workflow.GetStatus()),
		StartedAt:   workflow.GetStartTime(),
		CompletedAt: completedTime,
	}
}

func GetJobIdFromWorkflow(logger *slog.Logger, searchAttributes *commonpb.SearchAttributes) string {
	scheduledByIDPayload := searchAttributes.GetIndexedFields()["TemporalScheduledById"]
	var scheduledByID string
	err := converter.GetDefaultDataConverter().FromPayload(scheduledByIDPayload, &scheduledByID)
	if err != nil {
		// not returning an error here so that the runs don't break on the frontend due to temporal archiving old workflows
		// should probably revisit this at some point or if we get bit by trying to do something with an empty job id
		logger.Error(fmt.Errorf("unable to get job id from workflow: %w", err).Error())
	}
	return scheduledByID
}

func ToJobRunEventTaskDto(event *history.HistoryEvent, taskError *mgmtv1alpha1.JobRunEventTaskError) *mgmtv1alpha1.JobRunEventTask {
	return &mgmtv1alpha1.JobRunEventTask{
		Id:        event.GetEventId(),
		Type:      event.GetEventType().String(),
		EventTime: event.GetEventTime(),
		Error:     taskError,
	}
}

func ToJobRunEventTaskErrorDto(failure *temporalfailure.Failure, retryState enums.RetryState) *mgmtv1alpha1.JobRunEventTaskError {
	msg := failure.Message
	if failure.GetCause() != nil {
		msg = fmt.Sprintf("%s: %s", failure.GetMessage(), failure.GetCause().GetMessage())
	}
	return &mgmtv1alpha1.JobRunEventTaskError{
		Message:    msg,
		RetryState: retryState.String(),
	}
}

func toPendingActivitiesDto(activities []*workflowpb.PendingActivityInfo) []*mgmtv1alpha1.PendingActivity {
	dtos := []*mgmtv1alpha1.PendingActivity{}
	for _, activity := range activities {
		var lastFailure *mgmtv1alpha1.ActivityFailure
		if activity.LastFailure != nil {
			lastFailure = &mgmtv1alpha1.ActivityFailure{
				Message: activity.GetLastFailure().GetMessage(),
			}
		}
		dtos = append(dtos, &mgmtv1alpha1.PendingActivity{
			Status:       toActivityStatus(activity.State),
			ActivityName: activity.ActivityType.Name,
			LastFailure:  lastFailure,
		})
	}
	return dtos
}

func toActivityStatus(state enums.PendingActivityState) mgmtv1alpha1.ActivityStatus {
	switch state {
	case enums.PENDING_ACTIVITY_STATE_STARTED:
		return mgmtv1alpha1.ActivityStatus_ACTIVITY_STATUS_STARTED
	case enums.PENDING_ACTIVITY_STATE_SCHEDULED:
		return mgmtv1alpha1.ActivityStatus_ACTIVITY_STATUS_SCHEDULED
	case enums.PENDING_ACTIVITY_STATE_CANCEL_REQUESTED:
		return mgmtv1alpha1.ActivityStatus_ACTIVITY_STATUS_CANCELED
	default:
		return mgmtv1alpha1.ActivityStatus_ACTIVITY_STATUS_UNSPECIFIED
	}
}

func toWorfklowStatus(input enums.WorkflowExecutionStatus) mgmtv1alpha1.JobRunStatus {
	switch input {
	case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_COMPLETE
	case enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_RUNNING
	case enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_RUNNING
	case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_FAILED
	case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_TIMED_OUT
	case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_CANCELED
	case enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_TERMINATED
	default:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_UNSPECIFIED
	}
}
