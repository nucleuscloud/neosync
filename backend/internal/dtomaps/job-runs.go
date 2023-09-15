package dtomaps

import (
	"fmt"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToJobRunDto(
	input *workflowpb.WorkflowExecutionInfo,
) *mgmtv1alpha1.JobRun {

	attributes := input.GetSearchAttributes().IndexedFields
	fmt.Println(string(attributes["TemporalScheduledById"].Data))

	return &mgmtv1alpha1.JobRun{
		Id:        input.Execution.WorkflowId,
		JobId:     strings.Trim(string(attributes["TemporalScheduledById"].Data), "\""),
		Name:      input.Type.Name,
		Status:    mgmtv1alpha1.JobRunStatus(0), // TODO @alisha implement
		CreatedAt: timestamppb.New(*input.StartTime),
	}
}
