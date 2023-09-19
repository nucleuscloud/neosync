package dtomaps

import (
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToJobRunDto(
	logger *slog.Logger,
	input *workflowservice.DescribeWorkflowExecutionResponse,
) *mgmtv1alpha1.JobRun {
	executionInfo := input.GetWorkflowExecutionInfo()

	closeTime := executionInfo.GetCloseTime()
	var completedTime *timestamppb.Timestamp
	if closeTime != nil {
		completedTime = timestamppb.New(*executionInfo.GetCloseTime())
	}

	return &mgmtv1alpha1.JobRun{
		Id:                executionInfo.Execution.WorkflowId,
		JobId:             GetJobIdFromWorkflow(logger, executionInfo.GetSearchAttributes()),
		Name:              executionInfo.Type.Name,
		Status:            toWorfklowStatus(input),
		StartedAt:         timestamppb.New(*executionInfo.StartTime),
		CompletedAt:       completedTime,
		PendingActivities: toPendingActivitiesDto(input.GetPendingActivities()),
	}
}

func GetJobIdFromWorkflow(logger *slog.Logger, searchAttributes *common.SearchAttributes) string {
	scheduledByIDPayload := searchAttributes.IndexedFields["TemporalScheduledById"]
	var scheduledByID string
	err := converter.GetDefaultDataConverter().FromPayload(scheduledByIDPayload, &scheduledByID)
	if err != nil {
		logger.Error(fmt.Errorf("unable to get job id from workflow: %w", err).Error())
	}
	return scheduledByID
}

func ToJobRunEventDto(event *history.HistoryEvent, name, eventType string) *mgmtv1alpha1.JobRunEvent {
	return &mgmtv1alpha1.JobRunEvent{
		Id:        event.EventId,
		Name:      name,
		CreatedAt: timestamppb.New(*event.EventTime),
		Type:      eventType,
	}
}

func toPendingActivitiesDto(activities []*workflow.PendingActivityInfo) []*mgmtv1alpha1.PendingActivity {
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

func toWorfklowStatus(input *workflowservice.DescribeWorkflowExecutionResponse) mgmtv1alpha1.JobRunStatus {
	if input.PendingActivities != nil {
		for _, activity := range input.PendingActivities {
			if activity.LastFailure != nil {
				return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_ERROR
			}
		}
	}

	switch input.GetWorkflowExecutionInfo().Status {
	case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_COMPLETE
	case enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_RUNNING
	case enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_RUNNING
	case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_FAILED
	case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_FAILED
	case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_CANCELED
	case enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_TERMINATED
	default:
		return mgmtv1alpha1.JobRunStatus_JOB_RUN_STATUS_UNSPECIFIED
	}
}
