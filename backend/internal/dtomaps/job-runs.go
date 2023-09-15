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
