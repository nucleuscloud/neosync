package genbenthosconfigs_activity

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

func (b *benthosBuilder) getGenerateBenthosConfigResponses(
	ctx context.Context,
	job *mgmtv1alpha1.Job,
	slogger *slog.Logger,
) ([]*BenthosConfigResponse, error) {
	jobSource := job.GetSource()
	sourceOptions := job.GetSource().GetOptions().GetGenerate()
	if sourceOptions == nil {
		return nil, fmt.Errorf("job does not have Generate source options, has: %T", jobSource.GetOptions().Config)
	}
	sourceConnection, err := shared.GetJobSourceConnection(ctx, job.GetSource(), b.connclient)
	if err != nil {
		return nil, fmt.Errorf("unable to get connection by id: %w", err)
	}

	db, err := b.sqlmanagerclient.NewPooledSqlDb(ctx, slogger, sourceConnection)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	defer db.Db.Close()

	groupedMappings := groupMappingsByTable(job.Mappings)
	groupedTableMapping := getTableMappingsMap(groupedMappings)
	colTransformerMap := getColumnTransformerMap(groupedTableMapping)
	sourceTableOpts := groupGenerateSourceOptionsByTable(sourceOptions.Schemas)
	groupedSchemas, err := db.Db.GetSchemaColumnMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get database schema for connection: %w", err)
	}

	sourceResponses, err := buildBenthosGenerateSourceConfigResponses(slogger, ctx, b.transformerclient, groupedMappings, sourceTableOpts, groupedSchemas, colTransformerMap, db.Driver)
	if err != nil {
		return nil, fmt.Errorf("unable to build benthos generate source config responses: %w", err)
	}

	return sourceResponses, nil
}

type generateSourceTableOptions struct {
	Count int
}

func buildBenthosGenerateSourceConfigResponses(
	slogger *slog.Logger,
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	mappings []*tableMapping,
	sourceTableOpts map[string]*generateSourceTableOptions,
	groupedcolumnInfo map[string]map[string]*sqlmanager_shared.ColumnInfo,
	groupedColTransformers map[string]map[string]*mgmtv1alpha1.JobMappingTransformer,
	driver string,
) ([]*BenthosConfigResponse, error) {
	responses := []*BenthosConfigResponse{}

	for _, tableMapping := range mappings {
		tableName := neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table)
		var count = 0
		tableOpt := sourceTableOpts[tableName]
		if tableOpt != nil {
			count = tableOpt.Count
		}

		tableColInfo, ok := groupedcolumnInfo[tableName]
		if !ok {
			return nil, fmt.Errorf("missing table column info")
		}
		tableColTransformers, ok := groupedColTransformers[tableName]
		if !ok {
			return nil, fmt.Errorf("missing table column transformers mapping")
		}

		jsCode, err := extractJsFunctionsAndOutputs(ctx, transformerclient, tableMapping.Mappings)
		if err != nil {
			return nil, err
		}

		mutations, err := buildMutationConfigs(ctx, transformerclient, tableMapping.Mappings, tableColInfo, false)
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
		columnDefaultProperties, err := getColumnDefaultProperties(slogger, driver, columns, tableColInfo, tableColTransformers)
		if err != nil {
			return nil, err
		}

		responses = append(responses, &BenthosConfigResponse{
			Name:      neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table), // todo: may need to expand on this
			Config:    bc,
			DependsOn: []*tabledependency.DependsOn{},

			TableSchema:             tableMapping.Schema,
			TableName:               tableMapping.Table,
			Columns:                 columns,
			ColumnDefaultProperties: columnDefaultProperties,

			Processors: processors,

			metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, tableMapping.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, tableMapping.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "generate"),
			},
		})
	}

	return responses, nil
}

func (b *benthosBuilder) getSqlGenerateOutput(
	driver string,
	benthosConfig *BenthosConfigResponse,
	destination *mgmtv1alpha1.JobDestination,
	dsn string,
) []neosync_benthos.Outputs {
	outputs := []neosync_benthos.Outputs{}
	destOpts := getDestinationOptions(destination)

	processorConfigs := []neosync_benthos.ProcessorConfig{}
	for _, pc := range benthosConfig.Processors {
		processorConfigs = append(processorConfigs, *pc)
	}

	outputs = append(outputs, neosync_benthos.Outputs{
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
								Driver: driver,
								Dsn:    dsn,

								Schema:                  benthosConfig.TableSchema,
								Table:                   benthosConfig.TableName,
								Columns:                 benthosConfig.Columns,
								ColumnDefaultProperties: benthosConfig.ColumnDefaultProperties,
								OnConflictDoNothing:     destOpts.OnConflictDoNothing,
								TruncateOnRetry:         destOpts.Truncate,

								ArgsMapping: buildPlainInsertArgs(benthosConfig.Columns),

								Batching: &neosync_benthos.Batching{
									Period: "5s",
									Count:  100,
								},
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
					Period: "5s",
					Count:  100,
				},
			}},
		},
	})

	return outputs
}

func (b *benthosBuilder) getSqlAiGenerateOutput(
	driver string,
	benthosConfig *BenthosConfigResponse,
	destination *mgmtv1alpha1.JobDestination,
	dsn string,
	aiGroupedTableCols map[string][]string,
) ([]neosync_benthos.Outputs, error) {
	outputs := []neosync_benthos.Outputs{}
	destOpts := getDestinationOptions(destination)
	tableKey := neosync_benthos.BuildBenthosTable(benthosConfig.TableSchema, benthosConfig.TableName)

	cols, ok := aiGroupedTableCols[tableKey]
	if !ok {
		return nil, fmt.Errorf("unable to find table columns for key (%s) when building destination connection", tableKey)
	}

	processorConfigs := []neosync_benthos.ProcessorConfig{}
	for _, pc := range benthosConfig.Processors {
		processorConfigs = append(processorConfigs, *pc)
	}

	outputs = append(outputs, neosync_benthos.Outputs{
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
								Driver: driver,
								Dsn:    dsn,

								Schema:              benthosConfig.TableSchema,
								Table:               benthosConfig.TableName,
								Columns:             cols,
								OnConflictDoNothing: destOpts.OnConflictDoNothing,
								TruncateOnRetry:     destOpts.Truncate,

								ArgsMapping: buildPlainInsertArgs(cols),

								Batching: &neosync_benthos.Batching{
									Period: "5s",
									Count:  100,
								},
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
					Period: "5s",
					Count:  100,
				},
			}},
		},
	})

	return outputs, nil
}
