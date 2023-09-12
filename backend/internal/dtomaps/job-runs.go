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
		Status:    toJobStatusDto(input),
		CreatedAt: timestamppb.New(input.CreationTimestamp.Time),
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
