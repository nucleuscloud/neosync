package dtomaps

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	k8s_utils "github.com/nucleuscloud/neosync/backend/internal/utils/k8s"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToJobDto(
	inputJob *neosyncdevv1alpha1.JobConfig,
	inputSourceConnId string,
	inputDestConnIds []string,
) *mgmtv1alpha1.Job {
	mappings := []*mgmtv1alpha1.JobMapping{}
	for _, schema := range inputJob.Spec.Source.Sql.Schemas {
		for _, table := range schema.Schema
	}
	return &mgmtv1alpha1.Job{
		Id:                       inputJob.Labels[k8s_utils.NeosyncUuidLabel],
		Name:                     inputJob.Name,
		CreatedAt:                timestamppb.New(inputJob.CreationTimestamp.Time),
		Status:                   "",
		ConnectionSourceId:       inputSourceConnId,
		CronSchedule:             "",
		HaltOnNewColumnAddition:  *inputJob.Spec.Source.Sql.HaltOnSchemaChange,
		ConnectionDestinationIds: inputDestConnIds,
		Mappings:                 mappings,
	}
}
