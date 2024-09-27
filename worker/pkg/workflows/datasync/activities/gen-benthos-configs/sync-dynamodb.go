package genbenthosconfigs_activity

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type dynamoSyncResp struct {
	BenthosConfigs []*BenthosConfigResponse
}

func (b *benthosBuilder) getDynamoDbSyncBenthosConfigResponses(
	ctx context.Context,
	job *mgmtv1alpha1.Job,
	slogger *slog.Logger,
) (*dynamoSyncResp, error) {
	sourceConnection, err := shared.GetJobSourceConnection(ctx, job.GetSource(), b.connclient)
	if err != nil {
		return nil, fmt.Errorf("unable to get source connection by id: %w", err)
	}
	sourceConnectionType := shared.GetConnectionType(sourceConnection)
	slogger = slogger.With(
		"sourceConnectionType", sourceConnectionType,
	)
	_ = slogger

	dynamoSourceConfig := sourceConnection.GetConnectionConfig().GetDynamodbConfig()
	if dynamoSourceConfig == nil {
		return nil, fmt.Errorf("source connection was not dynamodb. Got %T", sourceConnection.GetConnectionConfig().Config)
	}
	awsManager := awsmanager.New()
	dynamoClient, err := awsManager.NewDynamoDbClient(ctx, dynamoSourceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create DynamoDB client: %w", err)
	}

	dynamoJobSourceOpts := job.GetSource().GetOptions().GetDynamodb()
	tableOptsMap := toDynamoDbSourceTableOptionMap(dynamoJobSourceOpts.GetTables())

	groupedMappings := groupMappingsByTable(job.GetMappings())

	benthosConfigs := []*BenthosConfigResponse{}
	// todo: may need to filter here based on the destination config mappings if there is no source->destination table map
	for _, tableMapping := range groupedMappings {
		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						AwsDynamoDB: &neosync_benthos.InputAwsDynamoDB{
							Table:          tableMapping.Table,
							Where:          getWhereFromSourceTableOption(tableOptsMap[tableMapping.Table]),
							ConsistentRead: dynamoJobSourceOpts.GetEnableConsistentRead(),
							Region:         dynamoSourceConfig.GetRegion(),
							Endpoint:       dynamoSourceConfig.GetEndpoint(),
							Credentials:    buildBenthosS3Credentials(dynamoSourceConfig.GetCredentials()),
						},
					},
				},
				Pipeline: &neosync_benthos.PipelineConfig{
					Threads:    -1,
					Processors: []neosync_benthos.ProcessorConfig{},
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

		columns := []string{}
		for _, jm := range tableMapping.Mappings {
			columns = append(columns, jm.Column)
		}

		tableKey, err := dynamoClient.GetTableKey(ctx, tableMapping.Table)
		if err != nil {
			return nil, fmt.Errorf("failed to describe table %s: %w", tableMapping.Table, err)
		}

		tableKeyList := []string{tableKey.HashKey}
		if tableKey.RangeKey != "" {
			tableKeyList = append(tableKeyList, tableKey.RangeKey)
		}

		mappedKeys := slices.Concat(columns, tableKeyList)
		splitColumnPaths := true
		processorConfigs, err := buildProcessorConfigsByRunType(
			ctx,
			b.transformerclient,
			tabledependency.NewRunConfig(tableMapping.Table, tabledependency.RunTypeInsert, []string{}, nil, columns, columns, nil, splitColumnPaths),
			map[string][]*referenceKey{},
			map[string][]*referenceKey{},
			b.jobId,
			b.runId,
			&shared.RedisConfig{},
			tableMapping.Mappings,
			map[string]*sqlmanager_shared.ColumnInfo{},
			job.GetSource().GetOptions(),
			mappedKeys,
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

			SourceConnectionType: sourceConnectionType,
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
