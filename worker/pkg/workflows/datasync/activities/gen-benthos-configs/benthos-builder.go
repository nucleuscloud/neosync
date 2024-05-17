package genbenthosconfigs_activity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

const (
	generateDefault            = "generate_default"
	passthrough                = "passthrough"
	dbDefault                  = "DEFAULT"
	jobmappingSubsetErrMsg     = "job mappings are not equal to or a subset of the database schema found in the source connection"
	haltOnSchemaAdditionErrMsg = "job mappings does not contain a column mapping for all " +
		"columns found in the source connection for the selected schemas and tables"
)

type benthosBuilder struct {
	sqlmanager sql_manager.SqlManagerClient

	jobclient         mgmtv1alpha1connect.JobServiceClient
	connclient        mgmtv1alpha1connect.ConnectionServiceClient
	transformerclient mgmtv1alpha1connect.TransformersServiceClient

	jobId string
	runId string

	redisConfig *shared.RedisConfig

	metricsEnabled bool
}

func newBenthosBuilder(
	sqlmanager sql_manager.SqlManagerClient,

	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,

	jobId, runId string,

	redisConfig *shared.RedisConfig,

	metricsEnabled bool,
) *benthosBuilder {
	return &benthosBuilder{
		sqlmanager:        sqlmanager,
		jobclient:         jobclient,
		connclient:        connclient,
		transformerclient: transformerclient,
		jobId:             jobId,
		runId:             runId,
		redisConfig:       redisConfig,
		metricsEnabled:    metricsEnabled,
	}
}

func (b *benthosBuilder) GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
	slogger *slog.Logger,
) (*GenerateBenthosConfigsResponse, error) {
	job, err := b.getJobById(ctx, req.JobId)
	if err != nil {
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}
	responses := []*BenthosConfigResponse{}

	groupedMappings := groupMappingsByTable(job.Mappings)
	groupedTableMapping := map[string]*tableMapping{}
	for _, tm := range groupedMappings {
		groupedTableMapping[neosync_benthos.BuildBenthosTable(tm.Schema, tm.Table)] = tm
	}
	uniqueSchemas := shared.GetUniqueSchemasFromMappings(job.Mappings)

	colTransformerMap := map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{} // schema.table ->  column -> transformer
	for table, mapping := range groupedTableMapping {
		colTransformerMap[table] = map[string]*mgmtv1alpha1.JobMappingTransformer{}
		for _, m := range mapping.Mappings {
			colTransformerMap[table][m.Column] = m.Transformer
		}
	}

	// reverse of table dependency
	// map of foreign key to source table + column
	var tableConstraintsSource map[string]map[string]*sql_manager.ForeignKey // schema.table -> column -> ForeignKey
	var aiGenerateMappings []*aiGenerateMappings

	switch jobSourceConfig := job.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_AiGenerate:
		sourceResponses, aimappings, err := b.getAiGenerateBenthosConfigResponses(ctx, job, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos AI Generate source config responses: %w", err)
		}
		aiGenerateMappings = aimappings
		responses = append(responses, sourceResponses...)
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		sourceTableOpts := groupGenerateSourceOptionsByTable(jobSourceConfig.Generate.Schemas)
		// TODO this needs to be updated to get db schema
		sourceResponses, err := buildBenthosGenerateSourceConfigResponses(ctx, b.transformerclient, groupedMappings, sourceTableOpts, map[string]*sql_manager.ColumnInfo{})
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos generate source config responses: %w", err)
		}
		responses = append(responses, sourceResponses...)

	case *mgmtv1alpha1.JobSourceOptions_Postgres, *mgmtv1alpha1.JobSourceOptions_Mysql:
		sourceConnection, err := shared.GetJobSourceConnection(ctx, job.GetSource(), b.connclient)
		if err != nil {
			return nil, fmt.Errorf("unable to get connection by id: %w", err)
		}

		sqlSourceOpts, err := getSqlJobSourceOpts(job.Source)
		if err != nil {
			return nil, err
		}
		var sourceTableOpts map[string]*sqlSourceTableOptions
		if sqlSourceOpts != nil {
			sourceTableOpts = groupJobSourceOptionsByTable(sqlSourceOpts)
		}

		db, err := b.sqlmanager.NewPooledSqlDb(ctx, slogger, sourceConnection)
		if err != nil {
			return nil, fmt.Errorf("unable to create new sql db: %w", err)
		}
		defer db.Db.Close()

		groupedSchemas, err := db.Db.GetSchemaColumnMap(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get database schema for connection: %w", err)
		}
		if !areMappingsSubsetOfSchemas(groupedSchemas, job.Mappings) {
			return nil, errors.New(jobmappingSubsetErrMsg)
		}
		if sqlSourceOpts != nil && sqlSourceOpts.HaltOnNewColumnAddition &&
			shouldHaltOnSchemaAddition(groupedSchemas, job.Mappings) {
			return nil, errors.New(haltOnSchemaAdditionErrMsg)
		}

		// todo should use GetForeignKeyReferencesMap instead
		tableDependencyMap, err := db.Db.GetForeignKeyConstraintsMap(ctx, uniqueSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve database foreign key constraints: %w", err)
		}
		slogger.Info(fmt.Sprintf("found %d foreign key constraints for database", getMapValuesCount(tableDependencyMap)))

		primaryKeyMap, err := db.Db.GetPrimaryKeyConstraintsMap(ctx, uniqueSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to get all primary key constraints: %w", err)
		}
		slogger.Info(fmt.Sprintf("found %d primary key constraints for database", getMapValuesCount(primaryKeyMap)))

		tables := []string{}
		tableSubsetMap := buildTableSubsetMap(sourceTableOpts)
		tableColMap := map[string][]string{}
		for _, m := range groupedMappings {
			cols := []string{}
			for _, c := range m.Mappings {
				cols = append(cols, c.Column)
			}
			tn := sql_manager.BuildTable(m.Schema, m.Table)
			tableColMap[tn] = cols
			tables = append(tables, tn)
		}
		runConfigs, err := tabledependency.GetRunConfigs(tableDependencyMap, tables, tableSubsetMap, primaryKeyMap, tableColMap)
		if err != nil {
			return nil, err
		}

		// reverse of table dependency
		// map of foreign key to source table + column
		fkReferenceMap, err := db.Db.GetForeignKeyReferencesMap(ctx, uniqueSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve database foreign key constraints: %w", err)
		}
		tableQueryMap, err := buildSelectQueryMap(db.Driver, groupedTableMapping, sourceTableOpts, fkReferenceMap, runConfigs, sqlSourceOpts.SubsetByForeignKeyConstraints)
		if err != nil {
			return nil, fmt.Errorf("unable to build select queries: %w", err)
		}

		sourceResponses, err := buildBenthosSqlSourceConfigResponses(ctx, b.transformerclient, groupedTableMapping, runConfigs, sourceConnection.Id, db.Driver, tableQueryMap, groupedSchemas, tableDependencyMap, colTransformerMap, b.jobId, b.runId, b.redisConfig, tableConstraintsSource)
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos sql source config responses: %w", err)
		}
		responses = append(responses, sourceResponses...)

	default:
		return nil, errors.New("unsupported job source")
	}

	// builds a map of table key to columns for AI Generated schemas as they are calculated lazily instead of via job mappings
	aiGroupedTableCols := map[string][]string{}
	for _, agm := range aiGenerateMappings {
		key := neosync_benthos.BuildBenthosTable(agm.Schema, agm.Table)
		for _, col := range agm.Columns {
			aiGroupedTableCols[key] = append(aiGroupedTableCols[key], col.Column)
		}
	}

	for destIdx, destination := range job.Destinations {
		destinationConnection, err := shared.GetConnectionById(ctx, b.connclient, destination.ConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get destination connection (%s) by id: %w", destination.ConnectionId, err)
		}

		for _, resp := range responses {
			tableKey := neosync_benthos.BuildBenthosTable(resp.TableSchema, resp.TableName)
			dstEnvVarKey := fmt.Sprintf("DESTINATION_%d_CONNECTION_DSN", destIdx)
			dsn := fmt.Sprintf("${%s}", dstEnvVarKey)

			switch connection := destinationConnection.ConnectionConfig.Config.(type) {
			case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
				driver, err := getSqlDriverFromConnection(destinationConnection)
				if err != nil {
					return nil, err
				}
				destOpts := getDestinationOptions(destination)
				resp.BenthosDsns = append(resp.BenthosDsns, &shared.BenthosDsn{EnvVarKey: dstEnvVarKey, ConnectionId: destinationConnection.Id})

				if resp.Config.Input.SqlSelect != nil || resp.Config.Input.PooledSqlRaw != nil {
					if resp.RunType == tabledependency.RunTypeUpdate {
						args := resp.Columns
						args = append(args, resp.primaryKeys...)
						resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
							Fallback: []neosync_benthos.Outputs{
								{
									PooledSqlUpdate: &neosync_benthos.PooledSqlUpdate{
										Driver: driver,
										Dsn:    dsn,

										Schema:       resp.TableSchema,
										Table:        resp.TableName,
										Columns:      resp.Columns,
										WhereColumns: resp.primaryKeys,
										ArgsMapping:  buildPlainInsertArgs(args),

										Batching: &neosync_benthos.Batching{
											Period: "5s",
											Count:  100,
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
					} else {
						// adds redis hash output for transformed primary keys
						constraints := tableConstraintsSource[tableKey]
						for col := range constraints {
							transformer := colTransformerMap[tableKey][col]
							if shouldProcessStrict(transformer) {
								if b.redisConfig == nil {
									return nil, fmt.Errorf("missing redis config. this operation requires redis")
								}
								hashedKey := neosync_benthos.HashBenthosCacheKey(b.jobId, b.runId, tableKey, col)
								resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
									RedisHashOutput: &neosync_benthos.RedisHashOutputConfig{
										Url:            b.redisConfig.Url,
										Key:            hashedKey,
										FieldsMapping:  fmt.Sprintf(`root = {meta("neosync_%s_%s_%s"): json(%q)}`, resp.TableSchema, resp.TableName, col, col), // map of original value to transformed value
										WalkMetadata:   false,
										WalkJsonObject: false,
										Kind:           &b.redisConfig.Kind,
										Master:         b.redisConfig.Master,
										Tls:            shared.BuildBenthosRedisTlsConfig(b.redisConfig),
									},
								})
								resp.RedisConfig = append(resp.RedisConfig, &BenthosRedisConfig{
									Key:    hashedKey,
									Table:  tableKey,
									Column: col,
								})
							}
						}
						resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
							Fallback: []neosync_benthos.Outputs{
								{
									PooledSqlInsert: &neosync_benthos.PooledSqlInsert{
										Driver: driver,
										Dsn:    dsn,

										Schema:              resp.TableSchema,
										Table:               resp.TableName,
										Columns:             resp.Columns,
										OnConflictDoNothing: destOpts.OnConflictDoNothing,
										TruncateOnRetry:     destOpts.Truncate,
										ArgsMapping:         buildPlainInsertArgs(resp.Columns),

										Batching: &neosync_benthos.Batching{
											Period: "5s",
											Count:  100,
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
					}
				} else if resp.Config.Input.Generate != nil {
					tm := groupedTableMapping[tableKey]
					if tm == nil {
						return nil, fmt.Errorf("unable to find table mapping for key (%s) when building destination connection", tableKey)
					}
					cols := buildPlainColumns(tm.Mappings)
					processorConfigs := []neosync_benthos.ProcessorConfig{}
					for _, pc := range resp.Processors {
						processorConfigs = append(processorConfigs, *pc)
					}

					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
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

												Schema:              resp.TableSchema,
												Table:               resp.TableName,
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
				} else if resp.Config.Input.OpenAiGenerate != nil {
					cols, ok := aiGroupedTableCols[tableKey]
					if !ok {
						return nil, fmt.Errorf("unable to find table columns for key (%s) when building destination connection", tableKey)
					}

					processorConfigs := []neosync_benthos.ProcessorConfig{}
					for _, pc := range resp.Processors {
						processorConfigs = append(processorConfigs, *pc)
					}

					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
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

												Schema:              resp.TableSchema,
												Table:               resp.TableName,
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
				} else {
					return nil, errors.New("unable to build destination connection due to unsupported source connection")
				}
			case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
				if resp.RunType == tabledependency.RunTypeUpdate {
					continue
				}
				s3pathpieces := []string{}
				if connection.AwsS3Config.PathPrefix != nil && *connection.AwsS3Config.PathPrefix != "" {
					s3pathpieces = append(s3pathpieces, strings.Trim(*connection.AwsS3Config.PathPrefix, "/"))
				}

				s3pathpieces = append(
					s3pathpieces,
					"workflows",
					req.WorkflowId,
					"activities",
					neosync_benthos.BuildBenthosTable(resp.TableSchema, resp.TableName),
					"data",
					`${!count("files")}.txt.gz`,
				)

				resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
					Fallback: []neosync_benthos.Outputs{
						{
							AwsS3: &neosync_benthos.AwsS3Insert{
								Bucket:      connection.AwsS3Config.Bucket,
								MaxInFlight: 64,
								Path:        fmt.Sprintf("/%s", strings.Join(s3pathpieces, "/")),
								Batching: &neosync_benthos.Batching{
									Count:  100,
									Period: "5s",
									Processors: []*neosync_benthos.BatchProcessor{
										{Archive: &neosync_benthos.ArchiveProcessor{Format: "lines"}},
										{Compress: &neosync_benthos.CompressProcessor{Algorithm: "gzip"}},
									},
								},
								Credentials: buildBenthosS3Credentials(connection.AwsS3Config.Credentials),
								Region:      connection.AwsS3Config.GetRegion(),
								Endpoint:    connection.AwsS3Config.GetEndpoint(),
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
			default:
				return nil, fmt.Errorf("unsupported destination connection config")
			}
		}
	}

	if b.metricsEnabled {
		labels := metrics.MetricLabels{
			metrics.NewEqLabel(metrics.AccountIdLabel, job.AccountId),
			metrics.NewEqLabel(metrics.JobIdLabel, job.Id),
			metrics.NewEqLabel(metrics.TemporalWorkflowId, "${TEMPORAL_WORKFLOW_ID}"),
			metrics.NewEqLabel(metrics.TemporalRunId, "${TEMPORAL_RUN_ID}"),
		}
		for _, resp := range responses {
			joinedLabels := append(labels, resp.metriclabels...) //nolint:gocritic
			resp.Config.Metrics = &neosync_benthos.Metrics{
				OtelCollector: &neosync_benthos.MetricsOtelCollector{},
				Mapping:       joinedLabels.ToBenthosMeta(),
			}
		}
	}

	slogger.Info(fmt.Sprintf("successfully built %d benthos configs", len(responses)))
	return &GenerateBenthosConfigsResponse{
		BenthosConfigs: responses,
	}, nil
}

// func getForeignKeyToSourceMap(tableDependencies map[string][]*sql_manager.ColumnConstraint) map[string]map[string]*sql_manager.ForeignKey {
// 	tc := map[string]map[string]*sql_manager.ForeignKey{} // schema.table -> column -> ForeignKey
// 	for table, constraints := range tableDependencies {
// 		for _, c := range constraints {
// 			_, ok := tc[c.ForeignKey.Table]
// 			if !ok {
// 				tc[c.ForeignKey.Table] = map[string]*sql_manager.ForeignKey{}
// 			}
// 			tc[c.ForeignKey.Table][c.ForeignKey.Column] = &sql_manager.ForeignKey{
// 				Table:  table,
// 				Column: c.Column,
// 			}
// 		}
// 	}
// 	return tc
// }

func buildTableSubsetMap(tableOpts map[string]*sqlSourceTableOptions) map[string]string {
	tableSubsetMap := map[string]string{}
	for table, opts := range tableOpts {
		if opts != nil && opts.WhereClause != nil && *opts.WhereClause != "" {
			tableSubsetMap[table] = *opts.WhereClause
		}
	}
	return tableSubsetMap
}

type generateSourceTableOptions struct {
	Count int
}

func buildBenthosGenerateSourceConfigResponses(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	mappings []*tableMapping,
	sourceTableOpts map[string]*generateSourceTableOptions,
	columnInfo map[string]*sql_manager.ColumnInfo,
) ([]*BenthosConfigResponse, error) {
	responses := []*BenthosConfigResponse{}

	for _, tableMapping := range mappings {
		var count = 0
		tableOpt := sourceTableOpts[neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table)]
		if tableOpt != nil {
			count = tableOpt.Count
		}

		jsCode, err := extractJsFunctionsAndOutputs(ctx, transformerclient, tableMapping.Mappings)
		if err != nil {
			return nil, err
		}

		mutations, err := buildMutationConfigs(ctx, transformerclient, tableMapping.Mappings, columnInfo)
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
			processors = append(processors, &neosync_benthos.ProcessorConfig{Javascript: &neosync_benthos.JavascriptConfig{Code: jsCode}})
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

		responses = append(responses, &BenthosConfigResponse{
			Name:      neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table), // todo: may need to expand on this
			Config:    bc,
			DependsOn: []*tabledependency.DependsOn{},

			TableSchema: tableMapping.Schema,
			TableName:   tableMapping.Table,

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

func (b *benthosBuilder) getJobById(
	ctx context.Context,
	jobId string,
) (*mgmtv1alpha1.Job, error) {
	getjobResp, err := b.jobclient.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: jobId,
	}))
	if err != nil {
		return nil, err
	}

	return getjobResp.Msg.Job, nil
}

func hasTransformer(t mgmtv1alpha1.TransformerSource) bool {
	return t != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED && t != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH
}

type sqlSourceTableOptions struct {
	WhereClause *string
}

func buildBenthosSqlSourceConfigResponses(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	groupedTableMapping map[string]*tableMapping,
	runconfigs []*tabledependency.RunConfig,
	dsnConnectionId string,
	driver string,
	selectQueryMap map[string]string,
	groupedColumnInfo map[string]map[string]*sql_manager.ColumnInfo,
	tableDependencies map[string][]*sql_manager.ForeignConstraint,
	colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
	tableConstraintsSource map[string]map[string]*sql_manager.ForeignKey,
) ([]*BenthosConfigResponse, error) {
	responses := []*BenthosConfigResponse{}

	fkSourceMap := buildForeignKeySourceMap(tableDependencies)
	// filter this list by table constraints that has transformer
	tableConstraints := map[string]map[string]*sql_manager.ForeignKey{} // schema.table -> column -> foreignKey
	for table, constraints := range fkSourceMap {
		_, ok := tableConstraints[table]
		if !ok {
			tableConstraints[table] = map[string]*sql_manager.ForeignKey{}
		}
		for col, tc := range constraints {
			// only add constraint if foreign key has transformer
			transformer, transformerOk := colTransformerMap[tc.Table][tc.Column]
			if transformerOk && shouldProcessStrict(transformer) {
				tableConstraints[table][col] = tc
			}
		}
	}

	for _, config := range runconfigs {
		mappings, ok := groupedTableMapping[config.Table]
		if !ok {
			return nil, fmt.Errorf("missing column mappings for table: %s", config.Table)
		}
		query, ok := selectQueryMap[config.Table]
		if !ok {
			return nil, fmt.Errorf("select query not found for table: %s", config.Table)
		}
		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
							Driver: driver,
							Dsn:    "${SOURCE_CONNECTION_DSN}",
							Query:  query,
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

		if config.RunType == tabledependency.RunTypeUpdate {
			columnConstraints, ok := tableConstraintsSource[config.Table]
			if !ok {
				columnConstraints = map[string]*sql_manager.ForeignKey{}
			}
			// sql update processor configs
			processorConfigs, err := buildSqlUpdateProcessorConfigs(config, redisConfig, jobId, runId, mappings.Mappings, columnConstraints)
			if err != nil {
				return nil, err
			}

			for _, pc := range processorConfigs {
				bc.StreamConfig.Pipeline.Processors = append(bc.StreamConfig.Pipeline.Processors, *pc)
			}
		} else {
			columnConstraints, ok := tableConstraints[config.Table]
			if !ok {
				columnConstraints = map[string]*sql_manager.ForeignKey{}
			}
			// sql insert processor configs
			processorConfigs, err := buildProcessorConfigs(ctx, transformerclient, mappings.Mappings, groupedColumnInfo[config.Table], columnConstraints, config.PrimaryKeys, jobId, runId, redisConfig)
			if err != nil {
				return nil, err
			}

			for _, pc := range processorConfigs {
				bc.StreamConfig.Pipeline.Processors = append(bc.StreamConfig.Pipeline.Processors, *pc)
			}
		}

		redisDependsOnMap := map[string][]string{}
		for _, fk := range tableConstraints[config.Table] {
			if _, exists := redisDependsOnMap[fk.Table]; !exists {
				redisDependsOnMap[fk.Table] = []string{}
			}
			redisDependsOnMap[fk.Table] = append(redisDependsOnMap[fk.Table], fk.Column)
		}

		responses = append(responses, &BenthosConfigResponse{
			Name:           fmt.Sprintf("%s.%s", config.Table, config.RunType),
			Config:         bc,
			DependsOn:      config.DependsOn,
			RedisDependsOn: redisDependsOnMap,
			RunType:        config.RunType,

			BenthosDsns: []*shared.BenthosDsn{{ConnectionId: dsnConnectionId, EnvVarKey: "SOURCE_CONNECTION_DSN"}},

			TableSchema: mappings.Schema,
			TableName:   mappings.Table,
			Columns:     config.Columns,
			primaryKeys: config.PrimaryKeys,

			metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, mappings.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, mappings.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "sync"),
			},
		})
	}

	return responses, nil
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

func areMappingsSubsetOfSchemas(
	groupedSchemas map[string]map[string]*sql_manager.ColumnInfo,
	mappings []*mgmtv1alpha1.JobMapping,
) bool {
	tableColMappings := getUniqueColMappingsMap(mappings)

	for key := range groupedSchemas {
		// For this method, we only care about the schemas+tables that we currently have mappings for
		if _, ok := tableColMappings[key]; !ok {
			delete(groupedSchemas, key)
		}
	}

	if len(tableColMappings) != len(groupedSchemas) {
		return false
	}

	// tests to make sure that every column in the col mappings is present in the db schema
	for table, cols := range tableColMappings {
		schemaCols, ok := groupedSchemas[table]
		if !ok {
			return false
		}
		// job mappings has more columns than the schema
		if len(cols) > len(schemaCols) {
			return false
		}
		for col := range cols {
			if _, ok := schemaCols[col]; !ok {
				return false
			}
		}
	}
	return true
}

func getUniqueColMappingsMap(
	mappings []*mgmtv1alpha1.JobMapping,
) map[string]map[string]struct{} {
	tableColMappings := map[string]map[string]struct{}{}
	for _, mapping := range mappings {
		key := neosync_benthos.BuildBenthosTable(mapping.Schema, mapping.Table)
		if _, ok := tableColMappings[key]; ok {
			tableColMappings[key][mapping.Column] = struct{}{}
		} else {
			tableColMappings[key] = map[string]struct{}{
				mapping.Column: {},
			}
		}
	}
	return tableColMappings
}

func shouldHaltOnSchemaAddition(
	groupedSchemas map[string]map[string]*sql_manager.ColumnInfo,
	mappings []*mgmtv1alpha1.JobMapping,
) bool {
	tableColMappings := getUniqueColMappingsMap(mappings)

	if len(tableColMappings) != len(groupedSchemas) {
		return true
	}

	for table, cols := range groupedSchemas {
		mappingCols, ok := tableColMappings[table]
		if !ok {
			return true
		}
		if len(cols) > len(mappingCols) {
			return true
		}
		for col := range cols {
			if _, ok := mappingCols[col]; !ok {
				return true
			}
		}
	}
	return false
}

func buildPlainColumns(mappings []*mgmtv1alpha1.JobMapping) []string {
	columns := make([]string, len(mappings))
	for idx := range mappings {
		columns[idx] = mappings[idx].Column
	}
	return columns
}

func splitTableKey(key string) (schema, table string) {
	pieces := strings.Split(key, ".")
	if len(pieces) == 1 {
		return "public", pieces[0]
	}
	return pieces[0], pieces[1]
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

func getSqlDriverFromConnection(conn *mgmtv1alpha1.Connection) (string, error) {
	switch conn.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return sql_manager.PostgresDriver, nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return sql_manager.MysqlDriver, nil
	default:
		return "", fmt.Errorf("unsupported sql connection config")
	}
}

type destinationOptions struct {
	OnConflictDoNothing bool
	Truncate            bool
	TruncateCascade     bool
}

func getDestinationOptions(dest *mgmtv1alpha1.JobDestination) *destinationOptions {
	if dest == nil || dest.Options == nil || dest.Options.Config == nil {
		return &destinationOptions{}
	}
	switch config := dest.Options.Config.(type) {
	case *mgmtv1alpha1.JobDestinationOptions_PostgresOptions:
		return &destinationOptions{
			OnConflictDoNothing: config.PostgresOptions.GetOnConflict().GetDoNothing(),
			Truncate:            config.PostgresOptions.GetTruncateTable().GetTruncateBeforeInsert(),
			TruncateCascade:     config.PostgresOptions.GetTruncateTable().GetCascade(),
		}
	case *mgmtv1alpha1.JobDestinationOptions_MysqlOptions:
		return &destinationOptions{
			OnConflictDoNothing: config.MysqlOptions.GetOnConflict().GetDoNothing(),
			Truncate:            config.MysqlOptions.GetTruncateTable().GetTruncateBeforeInsert(),
		}
	default:
		return &destinationOptions{}
	}
}

func groupJobSourceOptionsByTable(
	sqlSourceOpts *sqlJobSourceOpts,
) map[string]*sqlSourceTableOptions {
	groupedMappings := map[string]*sqlSourceTableOptions{}
	for _, schemaOpt := range sqlSourceOpts.SchemaOpt {
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &sqlSourceTableOptions{
				WhereClause: tableOpt.WhereClause,
			}
		}
	}
	return groupedMappings
}

type sqlJobSourceOpts struct {
	HaltOnNewColumnAddition       bool
	SubsetByForeignKeyConstraints bool
	SchemaOpt                     []*schemaOptions
}
type schemaOptions struct {
	Schema string
	Tables []*tableOptions
}
type tableOptions struct {
	Table       string
	WhereClause *string
}

func getSqlJobSourceOpts(
	source *mgmtv1alpha1.JobSource,
) (*sqlJobSourceOpts, error) {
	switch jobSourceConfig := source.GetOptions().GetConfig().(type) {
	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		if jobSourceConfig.Postgres == nil {
			return nil, nil
		}
		schemaOpt := []*schemaOptions{}
		for _, opt := range jobSourceConfig.Postgres.Schemas {
			tableOpts := []*tableOptions{}
			for _, t := range opt.GetTables() {
				tableOpts = append(tableOpts, &tableOptions{
					Table:       t.Table,
					WhereClause: t.WhereClause,
				})
			}
			schemaOpt = append(schemaOpt, &schemaOptions{
				Schema: opt.GetSchema(),
				Tables: tableOpts,
			})
		}
		return &sqlJobSourceOpts{
			HaltOnNewColumnAddition:       jobSourceConfig.Postgres.HaltOnNewColumnAddition,
			SubsetByForeignKeyConstraints: jobSourceConfig.Postgres.SubsetByForeignKeyConstraints,
			SchemaOpt:                     schemaOpt,
		}, nil
	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		if jobSourceConfig.Mysql == nil {
			return nil, nil
		}
		schemaOpt := []*schemaOptions{}
		for _, opt := range jobSourceConfig.Mysql.Schemas {
			tableOpts := []*tableOptions{}
			for _, t := range opt.GetTables() {
				tableOpts = append(tableOpts, &tableOptions{
					Table:       t.Table,
					WhereClause: t.WhereClause,
				})
			}
			schemaOpt = append(schemaOpt, &schemaOptions{
				Schema: opt.GetSchema(),
				Tables: tableOpts,
			})
		}
		return &sqlJobSourceOpts{
			HaltOnNewColumnAddition:       jobSourceConfig.Mysql.HaltOnNewColumnAddition,
			SubsetByForeignKeyConstraints: jobSourceConfig.Mysql.SubsetByForeignKeyConstraints,
			SchemaOpt:                     schemaOpt,
		}, nil
	default:
		return nil, errors.New("unsupported job source options type")
	}
}

func groupMappingsByTable(
	mappings []*mgmtv1alpha1.JobMapping,
) []*tableMapping {
	groupedMappings := map[string][]*mgmtv1alpha1.JobMapping{}

	for _, mapping := range mappings {
		key := neosync_benthos.BuildBenthosTable(mapping.Schema, mapping.Table)
		groupedMappings[key] = append(groupedMappings[key], mapping)
	}

	output := make([]*tableMapping, 0, len(groupedMappings))
	for key, mappings := range groupedMappings {
		schema, table := splitTableKey(key)
		output = append(output, &tableMapping{
			Schema:   schema,
			Table:    table,
			Mappings: mappings,
		})
	}
	return output
}

type tableMapping struct {
	Schema   string
	Table    string
	Mappings []*mgmtv1alpha1.JobMapping
}

func shouldProcessColumn(t *mgmtv1alpha1.JobMappingTransformer) bool {
	return t != nil &&
		t.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED &&
		t.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH
}

func shouldProcessStrict(t *mgmtv1alpha1.JobMappingTransformer) bool {
	return t != nil &&
		t.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED &&
		t.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL &&
		t.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH &&
		t.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT
}

func buildPlainInsertArgs(cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	pieces := make([]string, len(cols))
	for idx := range cols {
		pieces[idx] = fmt.Sprintf("this.%q", cols[idx])
	}
	return fmt.Sprintf("root = [%s]", strings.Join(pieces, ", "))
}

func convertStringSliceToString(slc []string) (string, error) {
	var returnStr string

	if len(slc) == 0 {
		returnStr = "[]"
	} else {
		sliceBytes, err := json.Marshal(slc)
		if err != nil {
			return "", err
		}
		returnStr = string(sliceBytes)
	}
	return returnStr, nil
}

func getMapValuesCount[K comparable, V any](m map[K][]V) int {
	count := 0
	for _, v := range m {
		count += len(v)
	}
	return count
}

func findTopForeignKeySource(tableName, col string, tableDependencies map[string][]*sql_manager.ForeignConstraint) *sql_manager.ForeignKey {
	// Add the foreign key dependencies of the current table
	if foreignKeys, ok := tableDependencies[tableName]; ok {
		for _, fk := range foreignKeys {
			if fk.Column == col {
				// Recursively add dependent tables and their foreign keys
				return findTopForeignKeySource(fk.ForeignKey.Table, fk.ForeignKey.Column, tableDependencies)
			}
		}
	}
	return &sql_manager.ForeignKey{
		Table:  tableName,
		Column: col,
	}
}

// builds schema.table -> FK column ->  PK schema table column
// find top level primary key column if foreign keys are nested
func buildForeignKeySourceMap(tableDeps map[string][]*sql_manager.ForeignConstraint) map[string]map[string]*sql_manager.ForeignKey {
	outputMap := map[string]map[string]*sql_manager.ForeignKey{}
	for tableName, constraints := range tableDeps {
		if _, ok := outputMap[tableName]; !ok {
			outputMap[tableName] = map[string]*sql_manager.ForeignKey{}
		}
		for _, con := range constraints {
			fk := findTopForeignKeySource(tableName, con.Column, tableDeps)
			outputMap[tableName][con.Column] = fk
		}
	}
	return outputMap
}
