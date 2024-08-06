package genbenthosconfigs_activity

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"google.golang.org/protobuf/encoding/protojson"
)

type dynamoSyncResp struct {
	BenthosConfigs []*BenthosConfigResponse
}

func (b *benthosBuilder) getDynamoDbSyncBenthosConfigResponses(
	ctx context.Context,
	job *mgmtv1alpha1.Job,
	slogger *slog.Logger,
) (*dynamoSyncResp, error) {
	_ = slogger
	sourceConnection, err := shared.GetJobSourceConnection(ctx, job.GetSource(), b.connclient)
	if err != nil {
		return nil, fmt.Errorf("unable to get source connection by id: %w", err)
	}

	dynamoSourceConfig := sourceConnection.GetConnectionConfig().GetDynamodbConfig()
	if dynamoSourceConfig == nil {
		return nil, fmt.Errorf("source connection was not dynamodb. Got %T", sourceConnection.GetConnectionConfig().Config)
	}

	dynamoJobSourceOpts := job.GetSource().GetOptions().GetDynamodb()
	tableOptsMap := toDynamoDbSourceTableOptionMap(dynamoJobSourceOpts.GetTables())

	groupedMappings := groupMappingsByTable(job.GetMappings())

	sourceOptBits, err := protojson.Marshal(job.GetSource().GetOptions())
	if err != nil {
		return nil, err
	}

	benthosConfigs := []*BenthosConfigResponse{}
	// todo: may need to filter here based on the destination config mappings if there is no source->destination table map
	for _, tableMapping := range groupedMappings {
		columns := []string{}
		for _, jm := range tableMapping.Mappings {
			columns = append(columns, jm.Column)
		}

		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						AwsDynamoDB: &neosync_benthos.InputAwsDynamoDB{
							Table: tableMapping.Table,
							Where: getWhereFromSourceTableOption(tableOptsMap[tableMapping.Table]),

							Region:      dynamoSourceConfig.GetRegion(),
							Endpoint:    dynamoSourceConfig.GetEndpoint(),
							Credentials: buildBenthosS3Credentials(dynamoSourceConfig.GetCredentials()),
						},
					},
				},
				Pipeline: &neosync_benthos.PipelineConfig{
					Threads: -1,
					Processors: []neosync_benthos.ProcessorConfig{
						{
							NeosyncDefaultMapping: &neosync_benthos.NeosyncDefaultMappingConfig{
								JobSourceOptionsString: string(sourceOptBits),
								MappedKeys:             columns,
							},
						},
					},
				},
				Output: &neosync_benthos.OutputConfig{
					Outputs: neosync_benthos.Outputs{
						Broker: &neosync_benthos.OutputBrokerConfig{
							Pattern: "fan_out",
							Outputs: []neosync_benthos.Outputs{},
						},
					},
				},
			},
		}

		processorConfigs, err := buildProcessorConfigsByRunType(
			ctx,
			b.transformerclient,
			&tabledependency.RunConfig{RunType: tabledependency.RunTypeInsert, Table: tableMapping.Table, SelectColumns: columns, InsertColumns: columns, SplitColumnPaths: true},
			map[string][]*referenceKey{},
			map[string][]*referenceKey{},
			b.jobId,
			b.runId,
			&shared.RedisConfig{},
			tableMapping.Mappings,
			map[string]*sqlmanager_shared.ColumnInfo{},
		)
		if err != nil {
			return nil, err
		}
		for _, pc := range processorConfigs {
			bc.StreamConfig.Pipeline.Processors = append(bc.StreamConfig.Pipeline.Processors, *pc)
		}

		benthosConfigs = append(benthosConfigs, &BenthosConfigResponse{
			Config:      bc,
			Name:        fmt.Sprintf("%s.%s", tableMapping.Schema, tableMapping.Table), // todo
			TableSchema: tableMapping.Schema,
			TableName:   tableMapping.Table,
			RunType:     tabledependency.RunTypeInsert,
			DependsOn:   []*tabledependency.DependsOn{},
			Columns:     columns,

			metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, tableMapping.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, tableMapping.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "sync"),
			},
		})
	}

	return &dynamoSyncResp{
		BenthosConfigs: benthosConfigs,
	}, nil
}

func getWhereFromSourceTableOption(opt *mgmtv1alpha1.DynamoDBSourceTableOption) *string {
	if opt == nil {
		return nil
	}
	return opt.WhereClause
}

func toDynamoDbSourceTableOptionMap(tableOpts []*mgmtv1alpha1.DynamoDBSourceTableOption) map[string]*mgmtv1alpha1.DynamoDBSourceTableOption {
	output := map[string]*mgmtv1alpha1.DynamoDBSourceTableOption{}
	for _, opt := range tableOpts {
		output[opt.Table] = opt
	}
	return output
}
