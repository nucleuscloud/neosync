package sync_cmd

import (
	"fmt"

	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

func toJob(
	cmd *cmdConfig,
	sourceConnection *mgmtv1alpha1.Connection,
	destinationConnection *mgmtv1alpha1.Connection,
	sourceSchema []*mgmtv1alpha1.DatabaseColumn,
) (*mgmtv1alpha1.Job, error) {
	sourceConnOpts, err := toJobSourceOption(sourceConnection)
	if err != nil {
		return nil, err
	}
	jobId := uuid.NewString()
	if cmd.Source.ConnectionOpts != nil && cmd.Source.ConnectionOpts.JobId != nil &&
		*cmd.Source.ConnectionOpts.JobId != "" {
		jobId = *cmd.Source.ConnectionOpts.JobId
	}
	tables := map[string]string{}
	for _, m := range sourceSchema {
		tables[m.Table] = m.Table
	}
	return &mgmtv1alpha1.Job{
		Id:        jobId,
		Name:      "cli-sync",
		AccountId: *cmd.AccountId,
		Source: &mgmtv1alpha1.JobSource{
			Options: sourceConnOpts,
		},
		Destinations: []*mgmtv1alpha1.JobDestination{
			toJobDestination(cmd, destinationConnection, tables),
		},
		Mappings: toJobMappings(sourceSchema),
	}, nil
}

func toJobDestination(
	cmd *cmdConfig,
	destinationConnection *mgmtv1alpha1.Connection,
	tables map[string]string,
) *mgmtv1alpha1.JobDestination {
	return &mgmtv1alpha1.JobDestination{
		ConnectionId: destinationConnection.Id,
		Id:           uuid.NewString(),
		Options:      cmdConfigToDestinationConnectionOptions(cmd, tables),
	}
}

func toJobSourceOption(
	sourceConnection *mgmtv1alpha1.Connection,
) (*mgmtv1alpha1.JobSourceOptions, error) {
	switch sourceConnection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return &mgmtv1alpha1.JobSourceOptions{
			Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
				Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
					ConnectionId: sourceConnection.Id,
				},
			},
		}, nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return &mgmtv1alpha1.JobSourceOptions{
			Config: &mgmtv1alpha1.JobSourceOptions_Mysql{
				Mysql: &mgmtv1alpha1.MysqlSourceConnectionOptions{
					ConnectionId: sourceConnection.Id,
				},
			},
		}, nil
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		return &mgmtv1alpha1.JobSourceOptions{
			Config: &mgmtv1alpha1.JobSourceOptions_AwsS3{
				AwsS3: &mgmtv1alpha1.AwsS3SourceConnectionOptions{
					ConnectionId: sourceConnection.Id,
				},
			},
		}, nil
	case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
		return &mgmtv1alpha1.JobSourceOptions{
			Config: &mgmtv1alpha1.JobSourceOptions_Dynamodb{
				Dynamodb: &mgmtv1alpha1.DynamoDBSourceConnectionOptions{
					ConnectionId: sourceConnection.Id,
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported connection type")
	}
}

// if is generated and not idenity then set to generate default
func toJobMappings(sourceSchema []*mgmtv1alpha1.DatabaseColumn) []*mgmtv1alpha1.JobMapping {
	mappings := []*mgmtv1alpha1.JobMapping{}

	for _, colInfo := range sourceSchema {
		mappings = append(mappings, &mgmtv1alpha1.JobMapping{
			Schema:      colInfo.Schema,
			Table:       colInfo.Table,
			Column:      colInfo.Column,
			Transformer: toTransformer(colInfo),
		})
	}

	return mappings
}

func toTransformer(colInfo *mgmtv1alpha1.DatabaseColumn) *mgmtv1alpha1.JobMappingTransformer {
	if colInfo.GeneratedType != nil && colInfo.GetGeneratedType() != "" {
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
					GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
				},
			},
		}
	}
	return &mgmtv1alpha1.JobMappingTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
				PassthroughConfig: &mgmtv1alpha1.Passthrough{},
			},
		},
	}
}
