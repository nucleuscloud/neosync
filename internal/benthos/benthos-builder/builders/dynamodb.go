package benthosbuilder_builders

import (
	"context"
	"errors"
	"fmt"
	"slices"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	neosync_redis "github.com/nucleuscloud/neosync/internal/redis"
	"github.com/nucleuscloud/neosync/internal/runconfigs"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

type dyanmodbSyncBuilder struct {
	transformerclient mgmtv1alpha1connect.TransformersServiceClient
}

func NewDynamoDbSyncBuilder(
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
) bb_internal.BenthosBuilder {
	return &dyanmodbSyncBuilder{
		transformerclient: transformerclient,
	}
}

func (b *dyanmodbSyncBuilder) BuildSourceConfigs(ctx context.Context, params *bb_internal.SourceParams) ([]*bb_internal.BenthosSourceConfig, error) {
	sourceConnection := params.SourceConnection
	job := params.Job

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

	benthosConfigs := []*bb_internal.BenthosSourceConfig{}
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
			runconfigs.NewRunConfig(tableMapping.Table, runconfigs.RunTypeInsert, []string{}, nil, columns, columns, nil, nil, splitColumnPaths),
			map[string][]*bb_internal.ReferenceKey{},
			map[string][]*bb_internal.ReferenceKey{},
			params.Job.Id,
			params.JobRunId,
			&neosync_redis.RedisConfig{},
			tableMapping.Mappings,
			map[string]*sqlmanager_shared.DatabaseSchemaRow{},
			job.GetSource().GetOptions(),
			mappedKeys,
		)
		if err != nil {
			return nil, err
		}
		for _, pc := range processorConfigs {
			bc.StreamConfig.Pipeline.Processors = append(bc.StreamConfig.Pipeline.Processors, *pc)
		}

		benthosConfigs = append(benthosConfigs, &bb_internal.BenthosSourceConfig{
			Config:      bc,
			Name:        fmt.Sprintf("%s.%s", tableMapping.Schema, tableMapping.Table), // todo
			TableSchema: tableMapping.Schema,
			TableName:   tableMapping.Table,
			RunType:     runconfigs.RunTypeInsert,
			DependsOn:   []*runconfigs.DependsOn{},
			Columns:     columns,

			Metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, tableMapping.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, tableMapping.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "sync"),
			},
		})
	}

	return benthosConfigs, nil
}

func (b *dyanmodbSyncBuilder) BuildDestinationConfig(ctx context.Context, params *bb_internal.DestinationParams) (*bb_internal.BenthosDestinationConfig, error) {
	config := &bb_internal.BenthosDestinationConfig{}

	benthosConfig := params.SourceConfig
	destinationOpts := params.DestinationOpts

	dynamoConfig := params.DestConnection.GetConnectionConfig().GetDynamodbConfig()
	if dynamoConfig == nil {
		return nil, errors.New("destination must have configured dyanmodb config")
	}
	dynamoDestinationOpts := destinationOpts.GetDynamodbOptions()
	if dynamoDestinationOpts == nil {
		return nil, errors.New("destination must have configured dyanmodb options")
	}
	tableMap := map[string]string{}
	for _, tm := range dynamoDestinationOpts.GetTableMappings() {
		tableMap[tm.GetSourceTable()] = tm.GetDestinationTable()
	}
	mappedTable, ok := tableMap[benthosConfig.TableName]
	if !ok {
		return nil, fmt.Errorf("did not find table map for %q when building dynamodb destination config", benthosConfig.TableName)
	}
	config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
		AwsDynamoDB: &neosync_benthos.OutputAwsDynamoDB{
			Table: mappedTable,
			JsonMapColumns: map[string]string{
				"": ".",
			},

			Batching: &neosync_benthos.Batching{
				// https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_BatchWriteItem.html
				// A single call to BatchWriteItem can transmit up to 16MB of data over the network, consisting of up to 25 item put or delete operations
				// Specifying the count here may not be enough if the overall data is above 16MB.
				// Benthos will fall back on error to single writes however
				Period: "5s",
				Count:  25,
			},

			Region:      dynamoConfig.GetRegion(),
			Endpoint:    dynamoConfig.GetEndpoint(),
			Credentials: buildBenthosS3Credentials(dynamoConfig.GetCredentials()),
		},
	})

	return config, nil
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

func buildBenthosS3Credentials(mgmtCreds *mgmtv1alpha1.AwsS3Credentials) *neosync_benthos.AwsCredentials {
	if mgmtCreds == nil {
		return nil
	}
	creds := &neosync_benthos.AwsCredentials{}
	if mgmtCreds.Profile != nil {
		creds.Profile = *mgmtCreds.Profile
	}
	if mgmtCreds.AccessKeyId != nil {
		creds.Id = *mgmtCreds.AccessKeyId
	}
	if mgmtCreds.SecretAccessKey != nil {
		creds.Secret = *mgmtCreds.SecretAccessKey
	}
	if mgmtCreds.SessionToken != nil {
		creds.Token = *mgmtCreds.SessionToken
	}
	if mgmtCreds.FromEc2Role != nil {
		creds.FromEc2Role = *mgmtCreds.FromEc2Role
	}
	if mgmtCreds.RoleArn != nil {
		creds.Role = *mgmtCreds.RoleArn
	}
	if mgmtCreds.RoleExternalId != nil {
		creds.RoleExternalId = *mgmtCreds.RoleExternalId
	}

	return creds
}
