package dtomaps

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	k8s_utils "github.com/nucleuscloud/neosync/backend/internal/utils/k8s"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToJobDto(
	getTransformerName func(value string) (string, error),
	inputJob *neosyncdevv1alpha1.JobConfig,
	inputSourceConnId string,
	inputDestConnIds []string,
) *mgmtv1alpha1.Job {
	mappings := []*mgmtv1alpha1.JobMapping{}
	for _, table := range inputJob.Spec.Source.Sql.Schemas {
		for _, column := range table.Columns {
			transformerName := "passthrough"
			if column.Transformer != nil {
				name, err := getTransformerName(column.Transformer.Name)
				if err != nil {
					transformerName = "invalid"
				} else {
					transformerName = name
				}
			}
			mappings = append(mappings, &mgmtv1alpha1.JobMapping{
				Schema:      table.Schema,
				Table:       table.Table,
				Column:      column.Name,
				Exclude:     *column.Exclude,
				Transformer: transformerName,
			})
		}
	}

	return &mgmtv1alpha1.Job{
		Id:                       inputJob.Labels[k8s_utils.NeosyncUuidLabel],
		Name:                     inputJob.Name,
		CreatedAt:                timestamppb.New(inputJob.CreationTimestamp.Time),
		UpdatedAt:                timestamppb.New(inputJob.CreationTimestamp.Time), // TODO @alisha implement
		Status:                   mgmtv1alpha1.JobStatus(0),                        // TODO @alisha implement
		ConnectionSourceId:       inputSourceConnId,
		CronSchedule:             inputJob.Spec.CronSchedule,
		ConnectionDestinationIds: inputDestConnIds,
		Mappings:                 mappings,
		SourceOptions: &mgmtv1alpha1.JobSourceOptions{
			HaltOnNewColumnAddition: *inputJob.Spec.Source.Sql.HaltOnSchemaChange,
		},
	}
}
