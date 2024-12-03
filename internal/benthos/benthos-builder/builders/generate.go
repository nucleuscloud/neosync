package benthosbuilder_builders

import (
	"context"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type generateBuilder struct {
	transformerclient mgmtv1alpha1connect.TransformersServiceClient
	sqlmanagerclient  sqlmanager.SqlManagerClient
	connectionclient  mgmtv1alpha1connect.ConnectionServiceClient
}

func NewGenerateBuilder(
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	connectionclient mgmtv1alpha1connect.ConnectionServiceClient,
) bb_internal.BenthosBuilder {
	return &generateBuilder{
		transformerclient: transformerclient,
		sqlmanagerclient:  sqlmanagerclient,
		connectionclient:  connectionclient,
	}
}

func (b *generateBuilder) BuildSourceConfigs(ctx context.Context, params *bb_internal.SourceParams) ([]*bb_internal.BenthosSourceConfig, error) {
	logger := params.Logger
	job := params.Job
	configs := []*bb_internal.BenthosSourceConfig{}

	jobSource := job.GetSource()
	sourceOptions := jobSource.GetOptions().GetGenerate()
	if sourceOptions == nil {
		return nil, fmt.Errorf("job does not have Generate source options, has: %T", jobSource.GetOptions().Config)
	}
	sourceConnection, err := shared.GetJobSourceConnection(ctx, jobSource, b.connectionclient)
	if err != nil {
		return nil, fmt.Errorf("unable to get connection by id: %w", err)
	}

	db, err := b.sqlmanagerclient.NewSqlConnection(ctx, connectionmanager.NewUniqueSession(connectionmanager.WithSessionGroup(params.WorkflowId)), sourceConnection, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	defer db.Db().Close()

	groupedMappings := groupMappingsByTable(job.Mappings)
	groupedTableMapping := getTableMappingsMap(groupedMappings)
	colTransformerMap := getColumnTransformerMap(groupedTableMapping)
	sourceTableOpts := groupGenerateSourceOptionsByTable(sourceOptions.Schemas)
	groupedSchemas, err := db.Db().GetSchemaColumnMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get database schema for connection: %w", err)
	}

	for _, tableMapping := range groupedMappings {
		tableName := neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table)
		var count = 0
		tableOpt := sourceTableOpts[tableName]
		if tableOpt != nil {
			count = tableOpt.Count
		}

		tableColInfo, ok := groupedSchemas[tableName]
		if !ok {
			return nil, fmt.Errorf("missing table column info")
		}
		tableColTransformers, ok := colTransformerMap[tableName]
		if !ok {
			return nil, fmt.Errorf("missing table column transformers mapping")
		}

		jsCode, err := extractJsFunctionsAndOutputs(ctx, b.transformerclient, tableMapping.Mappings)
		if err != nil {
			return nil, err
		}

		mutations, err := buildMutationConfigs(ctx, b.transformerclient, tableMapping.Mappings, tableColInfo, false)
		if err != nil {
			return nil, err
		}
		var processors []*neosync_benthos.ProcessorConfig
		// for the generate input, benthos requires a mapping, so falling back to a
		// generic empty object if the mutations are empty
		if mutations == "" {
			mutations = "root = {}"
		}
		processors = append(processors, &neosync_benthos.ProcessorConfig{Mutation: &mutations})

		if jsCode != "" {
			processors = append(processors, &neosync_benthos.ProcessorConfig{NeosyncJavascript: &neosync_benthos.NeosyncJavascriptConfig{Code: jsCode}})
		}
		if len(processors) > 0 {
			// add catch and error processor
			processors = append(processors, &neosync_benthos.ProcessorConfig{Catch: []*neosync_benthos.ProcessorConfig{
				{Error: &neosync_benthos.ErrorProcessorConfig{
					ErrorMsg: `${! error()}`,
				}},
			}})
		}

		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						Generate: &neosync_benthos.Generate{
							Interval: "",
							Count:    count,
							Mapping:  "root = {}",
						},
					},
				},
				Pipeline: &neosync_benthos.PipelineConfig{
					Threads:    -1,
					Processors: []neosync_benthos.ProcessorConfig{}, // leave empty. processors should be on output
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

		columns := buildPlainColumns(tableMapping.Mappings)
		columnDefaultProperties, err := getColumnDefaultProperties(logger, db.Driver(), columns, tableColInfo, tableColTransformers)
		if err != nil {
			return nil, err
		}

		configs = append(configs, &bb_internal.BenthosSourceConfig{
			Name:      neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table), // todo: may need to expand on this
			Config:    bc,
			DependsOn: []*tabledependency.DependsOn{},

			TableSchema:             tableMapping.Schema,
			TableName:               tableMapping.Table,
			Columns:                 columns,
			ColumnDefaultProperties: columnDefaultProperties,

			Processors: processors,

			Metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, tableMapping.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, tableMapping.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "generate"),
			},
		})
	}

	return configs, nil
}

func (b *generateBuilder) BuildDestinationConfig(ctx context.Context, params *bb_internal.DestinationParams) (*bb_internal.BenthosDestinationConfig, error) {
	config := &bb_internal.BenthosDestinationConfig{}

	benthosConfig := params.SourceConfig
	destOpts, err := getDestinationOptions(params.DestinationOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to parse destination options: %w", err)
	}

	processorConfigs := []neosync_benthos.ProcessorConfig{}
	for _, pc := range benthosConfig.Processors {
		processorConfigs = append(processorConfigs, *pc)
	}

	config.BenthosDsns = append(config.BenthosDsns, &bb_shared.BenthosDsn{ConnectionId: params.DestConnection.Id})
	config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
		Fallback: []neosync_benthos.Outputs{
			{
				// retry processor and output several times
				Retry: &neosync_benthos.RetryConfig{
					InlineRetryConfig: neosync_benthos.InlineRetryConfig{
						MaxRetries: 10,
					},
					Output: neosync_benthos.OutputConfig{
						Outputs: neosync_benthos.Outputs{
							PooledSqlInsert: &neosync_benthos.PooledSqlInsert{
								ConnectionId: params.DestConnection.GetId(),

								Schema:                  benthosConfig.TableSchema,
								Table:                   benthosConfig.TableName,
								Columns:                 benthosConfig.Columns,
								ColumnDefaultProperties: benthosConfig.ColumnDefaultProperties,
								OnConflictDoNothing:     destOpts.OnConflictDoNothing,
								TruncateOnRetry:         destOpts.Truncate,

								ArgsMapping: buildPlainInsertArgs(benthosConfig.Columns),

								Batching: &neosync_benthos.Batching{
									Period: destOpts.BatchPeriod,
									Count:  destOpts.BatchCount,
								},
								MaxInFlight: int(destOpts.MaxInFlight),
							},
						},
						Processors: processorConfigs,
					},
				},
			},
			// kills activity depending on error
			{Error: &neosync_benthos.ErrorOutputConfig{
				ErrorMsg: `${! meta("fallback_error")}`,
				Batching: &neosync_benthos.Batching{
					Period: destOpts.BatchPeriod,
					Count:  destOpts.BatchCount,
				},
			}},
		},
	})

	return config, nil
}

type generateSourceTableOptions struct {
	Count int
}

func groupGenerateSourceOptionsByTable(
	schemaOptions []*mgmtv1alpha1.GenerateSourceSchemaOption,
) map[string]*generateSourceTableOptions {
	groupedMappings := map[string]*generateSourceTableOptions{}

	for idx := range schemaOptions {
		schemaOpt := schemaOptions[idx]
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &generateSourceTableOptions{
				Count: int(tableOpt.RowCount), // todo: probably need to update rowcount int64 to int32
			}
		}
	}

	return groupedMappings
}
