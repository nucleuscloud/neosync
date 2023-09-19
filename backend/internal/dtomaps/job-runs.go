package dtomaps

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToJobRunDto(
	input *workflowpb.WorkflowExecutionInfo,
) *mgmtv1alpha1.JobRun {

	attributes := input.GetSearchAttributes().IndexedFields

	return &mgmtv1alpha1.JobRun{
		Id:        input.Execution.WorkflowId,
		JobId:     attributes["TemporalScheduledById"].String(),
		Name:      input.Type.Name,
		Status:    mgmtv1alpha1.JobRunStatus(0), // TODO @alisha implement
		CreatedAt: timestamppb.New(*input.StartTime),
	}
}

func toJobStatusDto(input *neosyncdevv1alpha1.JobRun) *mgmtv1alpha1.JobRunStatus {
	var status *mgmtv1alpha1.JobRunStatusType
	if len(input.Status.Conditions) > 0 {
		s := getStatus(input.Status.Conditions[0].Type)
		status = &s
	}

	return &mgmtv1alpha1.JobRunStatus{
		Status:         *status,
		StartTime:      timestamppb.New(input.Status.StartTime.Time),
		CompletionTime: timestamppb.New(input.Status.CompletionTime.Time),
	}
}

func getStatus(status string) mgmtv1alpha1.JobRunStatusType {
	switch status {
	case "Succeeded":
		return mgmtv1alpha1.JobRunStatusType_JOB_RUN_STATUS_COMPLETE
	default:
		return mgmtv1alpha1.JobRunStatusType_JOB_RUN_STATUS_UNSPECIFIED
	}
}
