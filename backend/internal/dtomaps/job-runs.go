package dtomaps

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	k8s_utils "github.com/nucleuscloud/neosync/backend/internal/utils/k8s"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToJobRunDto(
	input *neosyncdevv1alpha1.JobRun,
	jobId *string,
) *mgmtv1alpha1.JobRun {

	return &mgmtv1alpha1.JobRun{
		Id:        input.Labels[k8s_utils.NeosyncUuidLabel],
		JobId:     *jobId,
		Name:      input.Name,
		Status:    mgmtv1alpha1.JobRunStatus(0), // TODO @alisha implement
		CreatedAt: timestamppb.New(input.CreationTimestamp.Time),
	}
}
