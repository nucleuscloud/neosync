package dtomaps

import (
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	k8s_utils "github.com/nucleuscloud/neosync/backend/internal/utils/k8s"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToJobDto(
	logger *slog.Logger,
	getTransformerName func(value string) (string, error),
	inputJob *neosyncdevv1alpha1.JobConfig,
	inputSource *mgmtv1alpha1.Connection,
	inputDestinations []*mgmtv1alpha1.Connection,
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
		Id:           inputJob.Labels[k8s_utils.NeosyncUuidLabel],
		Name:         inputJob.Name,
		CreatedAt:    timestamppb.New(inputJob.CreationTimestamp.Time),
		UpdatedAt:    timestamppb.New(inputJob.CreationTimestamp.Time), // TODO @alisha implement
		Status:       mgmtv1alpha1.JobStatus(0),                        // TODO @alisha implement
		CronSchedule: inputJob.Spec.CronSchedule,
		Mappings:     mappings,
		Source:       toSourceConnectionDto(inputSource, inputJob),
		Destinations: toDestinationConnectionDto(logger, inputDestinations, inputJob),
	}
}

func toSourceConnectionDto(connection *mgmtv1alpha1.Connection, job *neosyncdevv1alpha1.JobConfig) *mgmtv1alpha1.JobSource {
	switch connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return &mgmtv1alpha1.JobSource{
			ConnectionId: connection.Id,
			Options: &mgmtv1alpha1.JobSourceOptions{
				Config: &mgmtv1alpha1.JobSourceOptions_SqlOptions{
					SqlOptions: &mgmtv1alpha1.SqlSourceConnectionOptions{
						HaltOnNewColumnAddition: job.Spec.Source.Sql.HaltOnSchemaChange,
					},
				},
			},
		}
	default:
		return &mgmtv1alpha1.JobSource{
			ConnectionId: connection.Id,
			Options:      &mgmtv1alpha1.JobSourceOptions{},
		}
	}
}

func toDestinationConnectionDto(
	logger *slog.Logger,
	connections []*mgmtv1alpha1.Connection,
	job *neosyncdevv1alpha1.JobConfig,
) []*mgmtv1alpha1.JobDestination {
	connectionIdMap := map[string]string{}
	for _, connection := range connections {
		connectionIdMap[connection.Name] = connection.Id
	}
	destinations := []*mgmtv1alpha1.JobDestination{}
	for _, dest := range job.Spec.Destinations {
		var jobDestination *mgmtv1alpha1.JobDestination
		if dest.Sql != nil {
			connId, ok := connectionIdMap[dest.Sql.ConnectionRef.Name]
			if ok {
				jobDestination = &mgmtv1alpha1.JobDestination{
					ConnectionId: connId,
					Options: &mgmtv1alpha1.JobDestinationOptions{
						Config: &mgmtv1alpha1.JobDestinationOptions_SqlOptions{
							SqlOptions: &mgmtv1alpha1.SqlDestinationConnectionOptions{
								TruncateBeforeInsert: dest.Sql.TruncateBeforeInsert,
								InitDbSchema:         dest.Sql.InitDbSchema,
							},
						},
					},
				}
			} else {
				logger.Error("unable to locate connection id")
			}
		} else {
			logger.Error("this job config destination type unsupported")
			jobDestination = &mgmtv1alpha1.JobDestination{
				ConnectionId: "",
				Options:      &mgmtv1alpha1.JobDestinationOptions{},
			}
		}
		destinations = append(destinations, jobDestination)
	}
	return destinations
}
