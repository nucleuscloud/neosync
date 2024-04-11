package genbenthosconfigs_activity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
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
	sqladapter sql_manager.SqlManager

	jobclient         mgmtv1alpha1connect.JobServiceClient
	connclient        mgmtv1alpha1connect.ConnectionServiceClient
	transformerclient mgmtv1alpha1connect.TransformersServiceClient

	jobId string
	runId string

	redisConfig *shared.RedisConfig

	metricsEnabled bool
}

func newBenthosBuilder(
	sqladapter sql_manager.SqlManager,

	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,

	jobId, runId string,

	redisConfig *shared.RedisConfig,

	metricsEnabled bool,
) *benthosBuilder {
	return &benthosBuilder{
		sqladapter:        sqladapter,
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
	var tableConstraintsSource map[string]map[string]*dbschemas_utils.ForeignKey // schema.table -> column -> ForeignKey
	var groupedColInfoMap map[string]map[string]*dbschemas_utils.ColumnInfo

	switch jobSourceConfig := job.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		sourceTableOpts := groupGenerateSourceOptionsByTable(jobSourceConfig.Generate.Schemas)
		// TODO this needs to be updated to get db schema
		sourceResponses, err := buildBenthosGenerateSourceConfigResponses(ctx, b.transformerclient, groupedMappings, sourceTableOpts, map[string]*dbschemas_utils.ColumnInfo{})
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos generate source config responses: %w", err)
		}
		responses = append(responses, sourceResponses...)

	case *mgmtv1alpha1.JobSourceOptions_Postgres, *mgmtv1alpha1.JobSourceOptions_Mysql:
		sourceConnection, err := b.getJobSourceConnection(ctx, job.GetSource())
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

		db, err := b.sqladapter.NewSqlDb(ctx, slogger, sourceConnection)
		if err != nil {
			return nil, fmt.Errorf("unable to create new sql db: %w", err)
		}
		defer db.ClosePool()

		dbschemas, err := db.GetDatabaseSchema(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get database schema for connection: %w", err)
		}
		groupedSchemas := sql_manager.GetUniqueSchemaColMappings(dbschemas)
		if !areMappingsSubsetOfSchemas(groupedSchemas, job.Mappings) {
			return nil, errors.New(jobmappingSubsetErrMsg)
		}
		if sqlSourceOpts != nil && sqlSourceOpts.HaltOnNewColumnAddition &&
			shouldHaltOnSchemaAddition(groupedSchemas, job.Mappings) {
			return nil, errors.New(haltOnSchemaAdditionErrMsg)
		}

		groupedColInfoMap = groupedSchemas

		allConstraints, err := db.GetAllForeignKeyConstraints(ctx, uniqueSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve database foreign key constraints: %w", err)
		}
		slogger.Info(fmt.Sprintf("found %d foreign key constraints for database", len(allConstraints)))
		td := sql_manager.GetDbTableDependencies(allConstraints)

		primaryKeys, err := db.GetAllPrimaryKeyConstraints(ctx, uniqueSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to get all primary key constraints: %w", err)
		}
		primaryKeyMap := sql_manager.GetTablePrimaryKeysMap(primaryKeys)

		tables := filterNullTables(groupedMappings)
		tableSubsetMap := buildTableSubsetMap(sourceTableOpts)
		dependencyConfigs := tabledependency.GetRunConfigs(td, tables, tableSubsetMap)

		// reverse of table dependency
		// map of foreign key to source table + column
		tableConstraintsSource = getForeignKeyToSourceMap(td)
		tableQueryMap, err := buildSelectQueryMap(db.Driver, groupedTableMapping, sourceTableOpts, td, dependencyConfigs, sqlSourceOpts.SubsetByForeignKeyConstraints)
		if err != nil {
			return nil, fmt.Errorf("unable to build select queries: %w", err)
		}

		sourceResponses, err := buildBenthosSqlSourceConfigResponses(ctx, b.transformerclient, groupedMappings, sourceConnection.Id, db.Driver, tableQueryMap, groupedSchemas, td, colTransformerMap, primaryKeyMap, b.jobId, b.runId, b.redisConfig)
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos sql source config responses: %w", err)
		}
		responses = append(responses, sourceResponses...)

		dependencyMap := map[string][]*tabledependency.RunConfig{}
		for _, cfg := range dependencyConfigs {
			_, ok := dependencyMap[cfg.Table]
			if ok {
				dependencyMap[cfg.Table] = append(dependencyMap[cfg.Table], cfg)
			} else {
				dependencyMap[cfg.Table] = []*tabledependency.RunConfig{cfg}
			}
		}

		for _, resp := range responses {
			tableName := neosync_benthos.BuildBenthosTable(resp.TableSchema, resp.TableName)
			configs := dependencyMap[tableName]
			if len(configs) > 1 {
				// circular dependency
				for _, c := range configs {
					if c.Columns != nil && c.Columns.Exclude != nil && len(c.Columns.Exclude) > 0 {
						resp.excludeColumns = c.Columns.Exclude
						resp.DependsOn = c.DependsOn
					} else if c.Columns != nil && c.Columns.Include != nil && len(c.Columns.Include) > 0 {
						pks := primaryKeyMap[tableName]
						if len(pks) == 0 {
							return nil, fmt.Errorf("no primary keys found for table (%s). Unable to build update query", tableName)
						}

						// config for sql update
						resp.updateConfig = c
						resp.primaryKeys = pks
					}
				}
			} else if len(configs) == 1 {
				resp.DependsOn = configs[0].DependsOn
			} else {
				return nil, fmt.Errorf("unexpected number of dependency configs")
			}
		}

	default:
		return nil, errors.New("unsupported job source")
	}

	updateResponses := []*BenthosConfigResponse{} // update configs for circular dependecies
	for destIdx, destination := range job.Destinations {
		destinationConnection, err := b.getConnectionById(ctx, destination.ConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get destination connection (%s) by id: %w", destination.ConnectionId, err)
		}

		for _, resp := range responses {
			tableKey := neosync_benthos.BuildBenthosTable(resp.TableSchema, resp.TableName)
			tm := groupedTableMapping[tableKey]
			if tm == nil {
				return nil, fmt.Errorf("unable to find table mapping for key (%s) when building destination connection", tableKey)
			}
			dstEnvVarKey := fmt.Sprintf("DESTINATION_%d_CONNECTION_DSN", destIdx)
			dsn := fmt.Sprintf("${%s}", dstEnvVarKey)

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
							FieldsMapping:  fmt.Sprintf(`root = {meta("neosync_%s"): json(%q)}`, col, col), // map of original value to transformed value
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

			switch connection := destinationConnection.ConnectionConfig.Config.(type) {
			case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
				driver, err := getSqlDriverFromConnection(destinationConnection)
				if err != nil {
					return nil, err
				}
				onConflictDoNothing := getOnConflictDoNothing(destination)

				resp.BenthosDsns = append(resp.BenthosDsns, &shared.BenthosDsn{EnvVarKey: dstEnvVarKey, ConnectionId: destinationConnection.Id})

				if resp.Config.Input.SqlSelect != nil || resp.Config.Input.PooledSqlRaw != nil {
					out := buildOutputArgs(resp, tm)
					resp.Columns = out.Columns
					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
						Fallback: []neosync_benthos.Outputs{
							{
								PooledSqlInsert: &neosync_benthos.PooledSqlInsert{
									Driver: driver,
									Dsn:    dsn,

									Schema:              resp.TableSchema,
									Table:               resp.TableName,
									Columns:             out.Columns,
									OnConflictDoNothing: &onConflictDoNothing,
									ArgsMapping:         out.ArgsMapping,

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

					if resp.updateConfig != nil {
						// circular dependency -> create update benthos config
						updateResp, err := createSqlUpdateBenthosConfig(ctx, b.transformerclient, resp, dsn, resp.TableSchema, resp.TableName, tm, groupedColInfoMap, tableConstraintsSource[neosync_benthos.BuildBenthosTable(resp.TableSchema, resp.TableName)], b.jobId, b.runId, b.redisConfig)
						if err != nil {
							return nil, fmt.Errorf("unable to create sql update benthos config: %w", err)
						}
						updateResponses = append(updateResponses, updateResp)
					}
				} else if resp.Config.Input.Generate != nil {
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
												OnConflictDoNothing: &onConflictDoNothing,

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

				cols := buildPlainColumns(tm.Mappings)
				resp.Columns = cols
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

	responses = append(responses, updateResponses...)

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

func getForeignKeyToSourceMap(tableDependencies map[string]*dbschemas_utils.TableConstraints) map[string]map[string]*dbschemas_utils.ForeignKey {
	tc := map[string]map[string]*dbschemas_utils.ForeignKey{} // schema.table -> column -> ForeignKey
	for table, constraints := range tableDependencies {
		for _, c := range constraints.Constraints {
			_, ok := tc[c.ForeignKey.Table]
			if !ok {
				tc[c.ForeignKey.Table] = map[string]*dbschemas_utils.ForeignKey{}
			}
			tc[c.ForeignKey.Table][c.ForeignKey.Column] = &dbschemas_utils.ForeignKey{
				Table:  table,
				Column: c.Column,
			}
		}
	}
	return tc
}

func buildTableSubsetMap(tableOpts map[string]*sqlSourceTableOptions) map[string]string {
	tableSubsetMap := map[string]string{}
	for table, opts := range tableOpts {
		if opts != nil && opts.WhereClause != nil && *opts.WhereClause != "" {
			tableSubsetMap[table] = *opts.WhereClause
		}
	}
	return tableSubsetMap
}

type sqlUpdateOutput struct {
	Query       string
	ArgsMapping string
	Columns     []string
	WhereCols   []string
}

func buildOutputArgs(resp *BenthosConfigResponse, tm *tableMapping) *sqlUpdateOutput {
	if len(resp.excludeColumns) > 0 {
		// insert excluding specific columns
		filteredInsertMappings := []*mgmtv1alpha1.JobMapping{}
		for _, m := range tm.Mappings {
			if !slices.Contains(resp.excludeColumns, m.Column) {
				filteredInsertMappings = append(filteredInsertMappings, m)
			}
		}
		cols := buildPlainColumns(filteredInsertMappings)
		return &sqlUpdateOutput{
			ArgsMapping: buildPlainInsertArgs(cols),
			Columns:     cols,
		}
	} else if resp.updateConfig != nil && resp.updateConfig.Columns != nil && len(resp.updateConfig.Columns.Include) > 0 {
		// update specific columns
		filteredUpdateMappings := []*mgmtv1alpha1.JobMapping{}
		for _, m := range tm.Mappings {
			if slices.Contains(resp.updateConfig.Columns.Include, m.Column) {
				filteredUpdateMappings = append(filteredUpdateMappings, m)
			}
		}
		updateCols := buildPlainColumns(filteredUpdateMappings)
		updateArgsMapping := []string{}
		updateArgsMapping = append(updateArgsMapping, updateCols...)
		updateArgsMapping = append(updateArgsMapping, resp.primaryKeys...)

		return &sqlUpdateOutput{
			ArgsMapping: buildPlainInsertArgs(updateArgsMapping),
			Columns:     updateCols,
			WhereCols:   resp.primaryKeys,
		}
	}
	// insert all columns
	cols := buildPlainColumns(tm.Mappings)
	return &sqlUpdateOutput{
		ArgsMapping: buildPlainInsertArgs(cols),
		Columns:     cols,
	}
}

type generateSourceTableOptions struct {
	Count int
}

func buildBenthosGenerateSourceConfigResponses(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	mappings []*tableMapping,
	sourceTableOpts map[string]*generateSourceTableOptions,
	columnInfo map[string]*dbschemas_utils.ColumnInfo,
) ([]*BenthosConfigResponse, error) {
	responses := []*BenthosConfigResponse{}

	for _, tableMapping := range mappings {
		if shared.AreAllColsNull(tableMapping.Mappings) {
			// skiping table as no columns are mapped
			continue
		}

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
					ErrorMsg: `${! meta("fallback_error")}`,
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

func (b *benthosBuilder) getConnectionById(
	ctx context.Context,
	connectionId string,
) (*mgmtv1alpha1.Connection, error) {
	getConnResp, err := b.connclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: connectionId,
	}))
	if err != nil {
		return nil, err
	}
	return getConnResp.Msg.Connection, nil
}

// filters out tables where all cols are set to null
func filterNullTables(mappings []*tableMapping) []string {
	tables := []string{}
	for _, group := range mappings {
		if !shared.AreAllColsNull(group.Mappings) {
			tables = append(tables, dbschemas_utils.BuildTable(group.Schema, group.Table))
		}
	}
	return tables
}

func getDriverFromBenthosInput(input *neosync_benthos.Inputs) (string, error) {
	if input.SqlSelect != nil {
		return input.SqlSelect.Driver, nil
	} else if input.PooledSqlRaw != nil {
		return input.PooledSqlRaw.Driver, nil
	}
	return "", errors.New("invalid benthos input when trying to find database driver")
}

func createSqlUpdateBenthosConfig(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	insertConfig *BenthosConfigResponse,
	dsn,
	schema string,
	table string,
	tm *tableMapping,
	groupedColInfo map[string]map[string]*dbschemas_utils.ColumnInfo,
	fkMap map[string]*dbschemas_utils.ForeignKey,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
) (*BenthosConfigResponse, error) {
	driver, err := getDriverFromBenthosInput(&insertConfig.Config.Input.Inputs)
	if err != nil {
		return nil, err
	}

	colSourceMap := map[string]mgmtv1alpha1.TransformerSource{}
	for _, col := range tm.Mappings {
		colSourceMap[col.Column] = col.GetTransformer().Source
	}

	sourceResponses, err := buildBenthosSqlSourceConfigResponses(
		ctx,
		transformerclient,
		[]*tableMapping{tm},
		"", // does not matter what is here. gets overwritten with insert config
		driver,
		map[string]string{
			insertConfig.updateConfig.Table: "", // gets overwritten below
		},
		groupedColInfo,
		map[string]*dbschemas_utils.TableConstraints{},
		map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{},
		map[string][]string{},
		jobId,
		runId,
		redisConfig,
	)
	if err != nil {
		return nil, err
	}

	if len(sourceResponses) > 0 {
		newResp := sourceResponses[0]

		// create processor
		if insertConfig.updateConfig != nil && insertConfig.updateConfig.Columns != nil && insertConfig.updateConfig.Columns.Include != nil {
			processorConfigs := []neosync_benthos.ProcessorConfig{}
			for pkCol, fk := range fkMap {
				// only need redis processors if the primary key has a transformer
				if !hasTransformer(colSourceMap[pkCol]) || !slices.Contains(insertConfig.updateConfig.Columns.Include, fk.Column) {
					continue
				}

				// circular dependent foreign key
				hashedKey := neosync_benthos.HashBenthosCacheKey(jobId, runId, fk.Table, pkCol)
				requestMap := fmt.Sprintf(`root = if this.%q == null { deleted() } else { this }`, fk.Column)
				argsMapping := fmt.Sprintf(`root = [%q, json(%q)]`, hashedKey, fk.Column)
				resultMap := fmt.Sprintf("root.%q = this", fk.Column)
				fkBranch, err := buildRedisGetBranchConfig(resultMap, argsMapping, &requestMap, redisConfig)
				if err != nil {
					return nil, err
				}
				processorConfigs = append(processorConfigs, neosync_benthos.ProcessorConfig{Branch: fkBranch})

				// primary key
				pkRequestMap := fmt.Sprintf(`root = if this.%q == null { deleted() } else { this }`, pkCol)
				pkArgsMapping := fmt.Sprintf(`root = [%q, json(%q)]`, hashedKey, pkCol)
				pkResultMap := fmt.Sprintf("root.%q = this", pkCol)
				pkBranch, err := buildRedisGetBranchConfig(pkResultMap, pkArgsMapping, &pkRequestMap, redisConfig)
				if err != nil {
					return nil, err
				}
				processorConfigs = append(processorConfigs, neosync_benthos.ProcessorConfig{Branch: pkBranch})
			}
			if len(processorConfigs) > 0 {
				// add catch and error processor
				processorConfigs = append(processorConfigs, neosync_benthos.ProcessorConfig{Catch: []*neosync_benthos.ProcessorConfig{
					{Error: &neosync_benthos.ErrorProcessorConfig{
						ErrorMsg: `${! meta("fallback_error")}`,
					}},
				}})
			}
			newResp.Config.StreamConfig.Pipeline.Processors = processorConfigs
		}

		newResp.updateConfig = insertConfig.updateConfig
		newResp.DependsOn = insertConfig.updateConfig.DependsOn
		newResp.Name = fmt.Sprintf("%s.update", insertConfig.Name)
		newResp.primaryKeys = insertConfig.primaryKeys
		newResp.metriclabels = append(newResp.metriclabels, metrics.NewEqLabel(metrics.IsUpdateConfigLabel, "true"))
		output := buildOutputArgs(newResp, tm)
		newResp.Columns = output.Columns
		if newResp.Config.Input.SqlSelect != nil {
			newResp.Config.Input.SqlSelect.Where = insertConfig.Config.Input.SqlSelect.Where // keep the where clause the same as insert
		} else if newResp.Config.Input.PooledSqlRaw != nil {
			newResp.Config.Input.PooledSqlRaw.Query = insertConfig.Config.Input.PooledSqlRaw.Query // keep this the same for the insert
		}
		newResp.BenthosDsns = insertConfig.BenthosDsns
		newResp.Config.Output.Broker.Outputs = append(newResp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
			Fallback: []neosync_benthos.Outputs{
				{
					PooledSqlUpdate: &neosync_benthos.PooledSqlUpdate{
						Driver: driver,
						Dsn:    dsn,

						Schema:       schema,
						Table:        table,
						Columns:      output.Columns,
						WhereColumns: output.WhereCols,
						ArgsMapping:  output.ArgsMapping,

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
		return newResp, nil
	}
	return nil, errors.New("unable to build sql update benthos config")
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
	mappings []*tableMapping,
	dsnConnectionId string,
	driver string,
	selectQueryMap map[string]string,
	groupedColumnInfo map[string]map[string]*dbschemas_utils.ColumnInfo,
	tableDependencies map[string]*dbschemas_utils.TableConstraints,
	colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer,
	primaryKeys map[string][]string,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
) ([]*BenthosConfigResponse, error) {
	responses := []*BenthosConfigResponse{}

	// filter this list by table constraints that has transformer
	tableConstraints := map[string]map[string]*dbschemas_utils.ForeignKey{} // schema.table -> column -> foreignKey
	for table, constraints := range tableDependencies {
		_, ok := tableConstraints[table]
		if !ok {
			tableConstraints[table] = map[string]*dbschemas_utils.ForeignKey{}
		}
		for _, tc := range constraints.Constraints {
			// only add constraint if foreign key has transformer
			transformer, transformerOk := colTransformerMap[tc.ForeignKey.Table][tc.ForeignKey.Column]
			if transformerOk && shouldProcessStrict(transformer) {
				tableConstraints[table][tc.Column] = tc.ForeignKey
			}
		}
	}

	for i := range mappings {
		tableMapping := mappings[i]
		if shared.AreAllColsNull(tableMapping.Mappings) {
			// skipping table as no columns are mapped
			continue
		}

		table := neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table)
		query, ok := selectQueryMap[table]
		if !ok {
			return nil, fmt.Errorf("select query not found for table: %s", table)
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

		columnConstraints, ok := tableConstraints[table]
		if !ok {
			columnConstraints = map[string]*dbschemas_utils.ForeignKey{}
		}

		processorConfigs, err := buildProcessorConfigs(ctx, transformerclient, tableMapping.Mappings, groupedColumnInfo[table], columnConstraints, primaryKeys[table], jobId, runId, redisConfig)
		if err != nil {
			return nil, err
		}

		for _, pc := range processorConfigs {
			bc.StreamConfig.Pipeline.Processors = append(bc.StreamConfig.Pipeline.Processors, *pc)
		}

		responses = append(responses, &BenthosConfigResponse{
			Name:      neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table), // todo: may need to expand on this
			Config:    bc,
			DependsOn: []*tabledependency.DependsOn{},

			BenthosDsns: []*shared.BenthosDsn{{ConnectionId: dsnConnectionId, EnvVarKey: "SOURCE_CONNECTION_DSN"}},

			TableSchema: tableMapping.Schema,
			TableName:   tableMapping.Table,

			metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, tableMapping.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, tableMapping.Table),
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
	groupedSchemas map[string]map[string]*dbschemas_utils.ColumnInfo,
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
	groupedSchemas map[string]map[string]*dbschemas_utils.ColumnInfo,
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

func getOnConflictDoNothing(dest *mgmtv1alpha1.JobDestination) bool {
	if dest == nil || dest.Options == nil || dest.Options.Config == nil {
		return false
	}
	switch config := dest.Options.Config.(type) {
	case *mgmtv1alpha1.JobDestinationOptions_PostgresOptions:
		return config.PostgresOptions.GetOnConflict().GetDoNothing()
	case *mgmtv1alpha1.JobDestinationOptions_MysqlOptions:
		return config.MysqlOptions.GetOnConflict().GetDoNothing()
	default:
		return false
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

func (b *benthosBuilder) getJobSourceConnection(ctx context.Context, jobSource *mgmtv1alpha1.JobSource) (*mgmtv1alpha1.Connection, error) {
	var connectionId string
	switch jobSourceConfig := jobSource.GetOptions().GetConfig().(type) {
	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		connectionId = jobSourceConfig.Postgres.GetConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		connectionId = jobSourceConfig.Mysql.GetConnectionId()
	default:
		return nil, errors.New("unsupported job source options type")
	}
	sourceConnection, err := b.getConnectionById(ctx, connectionId)
	if err != nil {
		return nil, fmt.Errorf("unable to get connection by id (%s): %w", connectionId, err)
	}
	return sourceConnection, nil
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

func buildProcessorConfigs(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	cols []*mgmtv1alpha1.JobMapping,
	tableColumnInfo map[string]*dbschemas_utils.ColumnInfo,
	columnConstraints map[string]*dbschemas_utils.ForeignKey,
	primaryKeys []string,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
) ([]*neosync_benthos.ProcessorConfig, error) {
	jsCode, err := extractJsFunctionsAndOutputs(ctx, transformerclient, cols)
	if err != nil {
		return nil, err
	}

	mutations, err := buildMutationConfigs(ctx, transformerclient, cols, tableColumnInfo)
	if err != nil {
		return nil, err
	}

	cacheBranches, err := buildBranchCacheConfigs(cols, columnConstraints, jobId, runId, redisConfig)
	if err != nil {
		return nil, err
	}

	pkMapping := buildPrimaryKeyMappingConfigs(cols, primaryKeys)

	var processorConfigs []*neosync_benthos.ProcessorConfig
	if pkMapping != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Mapping: &pkMapping})
	}
	if mutations != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Mutation: &mutations})
	}
	if jsCode != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Javascript: &neosync_benthos.JavascriptConfig{Code: jsCode}})
	}
	if len(cacheBranches) > 0 {
		for _, config := range cacheBranches {
			processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Branch: config})
		}
	}

	if len(processorConfigs) > 0 {
		// add catch and error processor
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Catch: []*neosync_benthos.ProcessorConfig{
			{Error: &neosync_benthos.ErrorProcessorConfig{
				ErrorMsg: `${! meta("fallback_error")}`,
			}},
		}})
	}

	return processorConfigs, err
}

func extractJsFunctionsAndOutputs(ctx context.Context, transformerclient mgmtv1alpha1connect.TransformersServiceClient, cols []*mgmtv1alpha1.JobMapping) (string, error) {
	var benthosOutputs []string
	var jsFunctions []string

	for _, col := range cols {
		if shouldProcessStrict(col.Transformer) {
			if _, ok := col.Transformer.Config.Config.(*mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig); ok {
				val, err := convertUserDefinedFunctionConfig(ctx, transformerclient, col.Transformer)
				if err != nil {
					return "", errors.New("unable to look up user defined transformer config by id")
				}
				col.Transformer = val
			}
			switch col.Transformer.Source {
			case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT:
				code := col.Transformer.Config.GetTransformJavascriptConfig().Code
				if code != "" {
					jsFunctions = append(jsFunctions, constructJsFunction(code, col.Column, col.Transformer.Source))
					benthosOutputs = append(benthosOutputs, constructBenthosJavascriptObject(col.Column, col.Transformer.Source))
				}
			case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT:
				code := col.Transformer.Config.GetGenerateJavascriptConfig().Code
				if code != "" {
					jsFunctions = append(jsFunctions, constructJsFunction(code, col.Column, col.Transformer.Source))
					benthosOutputs = append(benthosOutputs, constructBenthosJavascriptObject(col.Column, col.Transformer.Source))
				}
			}
		}
	}

	if len(jsFunctions) > 0 {
		return constructBenthosJsProcessor(jsFunctions, benthosOutputs), nil
	} else {
		return "", nil
	}
}

func buildMutationConfigs(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	cols []*mgmtv1alpha1.JobMapping,
	tableColumnInfo map[string]*dbschemas_utils.ColumnInfo,
) (string, error) {
	mutations := []string{}

	for _, col := range cols {
		colInfo := tableColumnInfo[col.Column]
		if shouldProcessColumn(col.Transformer) {
			if _, ok := col.Transformer.Config.Config.(*mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig); ok {
				// handle user defined transformer -> get the user defined transformer configs using the id
				val, err := convertUserDefinedFunctionConfig(ctx, transformerclient, col.Transformer)
				if err != nil {
					return "", errors.New("unable to look up user defined transformer config by id")
				}
				col.Transformer = val
			}
			if col.Transformer.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT && col.Transformer.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT {
				mutation, err := computeMutationFunction(col, colInfo)
				if err != nil {
					return "", fmt.Errorf("%s is not a supported transformer: %w", col.Transformer, err)
				}
				mutations = append(mutations, fmt.Sprintf("root.%q = %s", col.Column, mutation))
			}
		}
	}

	return strings.Join(mutations, "\n"), nil
}

func buildPrimaryKeyMappingConfigs(cols []*mgmtv1alpha1.JobMapping, primaryKeys []string) string {
	mappings := []string{}
	for _, col := range cols {
		if shouldProcessColumn(col.Transformer) && slices.Contains(primaryKeys, col.Column) {
			mappings = append(mappings, fmt.Sprintf("meta neosync_%s = this.%q", col.Column, col.Column))
		}
	}
	return strings.Join(mappings, "\n")
}

func buildBranchCacheConfigs(
	cols []*mgmtv1alpha1.JobMapping,
	columnConstraints map[string]*dbschemas_utils.ForeignKey,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
) ([]*neosync_benthos.BranchConfig, error) {
	branchConfigs := []*neosync_benthos.BranchConfig{}
	for _, col := range cols {
		fk, ok := columnConstraints[col.Column]
		if ok {
			// skip self referencing cols
			if fk.Table == fmt.Sprintf("%s.%s", col.Schema, col.Table) {
				continue
			}

			hashedKey := neosync_benthos.HashBenthosCacheKey(jobId, runId, fk.Table, fk.Column)
			requestMap := fmt.Sprintf(`root = if this.%q == null { deleted() } else { this }`, col.Column)
			argsMapping := fmt.Sprintf(`root = [%q, json(%q)]`, hashedKey, col.Column)
			resultMap := fmt.Sprintf("root.%q = this", col.Column)
			br, err := buildRedisGetBranchConfig(resultMap, argsMapping, &requestMap, redisConfig)
			if err != nil {
				return nil, err
			}
			branchConfigs = append(branchConfigs, br)
		}
	}
	return branchConfigs, nil
}

func buildRedisGetBranchConfig(
	resultMap, argsMapping string,
	requestMap *string,
	redisConfig *shared.RedisConfig,
) (*neosync_benthos.BranchConfig, error) {
	if redisConfig == nil {
		return nil, fmt.Errorf("missing redis config. this operation requires redis")
	}
	return &neosync_benthos.BranchConfig{
		RequestMap: requestMap,
		Processors: []neosync_benthos.ProcessorConfig{
			{
				Redis: &neosync_benthos.RedisProcessorConfig{
					Url:         redisConfig.Url,
					Command:     "hget",
					ArgsMapping: argsMapping,
					Kind:        &redisConfig.Kind,
					Master:      redisConfig.Master,
					Tls:         shared.BuildBenthosRedisTlsConfig(redisConfig),
				},
			},
		},
		ResultMap: &resultMap,
	}, nil
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

func constructJsFunction(jsCode, col string, source mgmtv1alpha1.TransformerSource) string {
	switch source {
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT:
		return fmt.Sprintf(`
function fn_%s(value, input){
  %s
};
`, col, jsCode)
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT:
		return fmt.Sprintf(`
function fn_%s(){
  %s
};
`, col, jsCode)
	default:
		return ""
	}
}

func constructBenthosJsProcessor(jsFunctions, benthosOutputs []string) string {
	jsFunctionStrings := strings.Join(jsFunctions, "\n")

	benthosOutputString := strings.Join(benthosOutputs, "\n")

	jsCode := fmt.Sprintf(`
(() => {
%s
const input = benthos.v0_msg_as_structured();
const output = { ...input };
%s
benthos.v0_msg_set_structured(output);
})();`, jsFunctionStrings, benthosOutputString)
	return jsCode
}

func constructBenthosJavascriptObject(col string, source mgmtv1alpha1.TransformerSource) string {
	switch source {
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT:
		return fmt.Sprintf(`output["%[1]s"] = fn_%[1]s(input["%[1]s"], input);`, col)
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT:
		return fmt.Sprintf(`output["%[1]s"] = fn_%[1]s();`, col)
	default:
		return ""
	}
}

// takes in an user defined config with just an id field and return the right transformer config for that user defined function id
func convertUserDefinedFunctionConfig(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	t *mgmtv1alpha1.JobMappingTransformer,
) (*mgmtv1alpha1.JobMappingTransformer, error) {
	transformer, err := transformerclient.GetUserDefinedTransformerById(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{TransformerId: t.Config.GetUserDefinedTransformerConfig().Id}))
	if err != nil {
		return nil, err
	}

	return &mgmtv1alpha1.JobMappingTransformer{
		Source: transformer.Msg.Transformer.Source,
		Config: transformer.Msg.Transformer.Config,
	}, nil
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

/*
function transformers
root.{destination_col} = transformerfunction(args)
*/

func computeMutationFunction(col *mgmtv1alpha1.JobMapping, colInfo *dbschemas_utils.ColumnInfo) (string, error) {
	var maxLen int64 = 10000
	if colInfo != nil && colInfo.CharacterMaximumLength != nil && *colInfo.CharacterMaximumLength > 0 {
		maxLen = int64(*colInfo.CharacterMaximumLength)
	}

	switch col.Transformer.Source {
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CATEGORICAL:
		categories := col.Transformer.Config.GetGenerateCategoricalConfig().Categories
		return fmt.Sprintf(`generate_categorical(categories: %q)`, categories), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_EMAIL:
		return fmt.Sprintf(`generate_email(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL:
		pd := col.Transformer.Config.GetTransformEmailConfig().PreserveDomain
		pl := col.Transformer.Config.GetTransformEmailConfig().PreserveLength
		excludedDomains := col.Transformer.Config.GetTransformEmailConfig().ExcludedDomains

		excludedDomainsStr, err := convertStringSliceToString(excludedDomains)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("transform_email(email:this.%q,preserve_domain:%t,preserve_length:%t,excluded_domains:%v,max_length:%d)", col.Column, pd, pl, excludedDomainsStr, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL:
		return "generate_bool()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CARD_NUMBER:
		luhn := col.Transformer.Config.GetGenerateCardNumberConfig().ValidLuhn
		return fmt.Sprintf(`generate_card_number(valid_luhn:%t)`, luhn), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CITY:
		return fmt.Sprintf(`generate_city(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_E164_PHONE_NUMBER:
		min := col.Transformer.Config.GetGenerateE164PhoneNumberConfig().Min
		max := col.Transformer.Config.GetGenerateE164PhoneNumberConfig().Max
		return fmt.Sprintf(`generate_e164_phone_number(min:%d,max:%d)`, min, max), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FIRST_NAME:
		return fmt.Sprintf(`generate_first_name(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FLOAT64:
		randomSign := col.Transformer.Config.GetGenerateFloat64Config().RandomizeSign
		min := col.Transformer.Config.GetGenerateFloat64Config().Min
		max := col.Transformer.Config.GetGenerateFloat64Config().Max
		precision := col.Transformer.Config.GetGenerateFloat64Config().Precision
		return fmt.Sprintf(`generate_float64(randomize_sign:%t, min:%f, max:%f, precision:%d)`, randomSign, min, max, precision), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_ADDRESS:
		return fmt.Sprintf(`generate_full_address(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME:
		return fmt.Sprintf(`generate_full_name(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_GENDER:
		ab := col.Transformer.Config.GetGenerateGenderConfig().Abbreviate
		return fmt.Sprintf(`generate_gender(abbreviate:%t,max_length:%d)`, ab, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64_PHONE_NUMBER:
		return "generate_int64_phone_number()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64:
		sign := col.Transformer.Config.GetGenerateInt64Config().RandomizeSign
		min := col.Transformer.Config.GetGenerateInt64Config().Min
		max := col.Transformer.Config.GetGenerateInt64Config().Max
		return fmt.Sprintf(`generate_int64(randomize_sign:%t,min:%d, max:%d)`, sign, min, max), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_LAST_NAME:
		return fmt.Sprintf(`generate_last_name(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SHA256HASH:
		return `generate_sha256hash()`, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SSN:
		return "generate_ssn()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STATE:
		return "generate_state()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STREET_ADDRESS:
		return fmt.Sprintf(`generate_street_address(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STRING_PHONE_NUMBER:
		min := col.Transformer.Config.GetGenerateStringPhoneNumberConfig().Min
		max := col.Transformer.Config.GetGenerateStringPhoneNumberConfig().Max
		min = transformer_utils.MinInt(min, maxLen)
		max = transformer_utils.Ceil(max, maxLen)
		return fmt.Sprintf("generate_string_phone_number(min:%d,max:%d)", min, max), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_RANDOM_STRING:
		min := col.Transformer.Config.GetGenerateStringConfig().Min
		max := col.Transformer.Config.GetGenerateStringConfig().Max
		min = transformer_utils.MinInt(min, maxLen) // ensure the min is not larger than the max allowed length
		max = transformer_utils.Ceil(max, maxLen)
		// todo: we need to pull in the min from the database schema
		return fmt.Sprintf(`generate_string(min:%d,max:%d)`, min, max), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UNIXTIMESTAMP:
		return "generate_unixtimestamp()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_USERNAME:
		return fmt.Sprintf(`generate_username(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UTCTIMESTAMP:
		return "generate_utctimestamp()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID:
		ih := col.Transformer.Config.GetGenerateUuidConfig().IncludeHyphens
		return fmt.Sprintf("generate_uuid(include_hyphens:%t)", ih), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_ZIPCODE:
		return "generate_zipcode()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_E164_PHONE_NUMBER:
		pl := col.Transformer.Config.GetTransformE164PhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_e164_phone_number(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FIRST_NAME:
		pl := col.Transformer.Config.GetTransformFirstNameConfig().PreserveLength
		return fmt.Sprintf("transform_first_name(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FLOAT64:
		rMin := col.Transformer.Config.GetTransformFloat64Config().RandomizationRangeMin
		rMax := col.Transformer.Config.GetTransformFloat64Config().RandomizationRangeMax
		return fmt.Sprintf(`transform_float64(value:this.%q,randomization_range_min:%f,randomization_range_max:%f)`, col.Column, rMin, rMax), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FULL_NAME:
		pl := col.Transformer.Config.GetTransformFullNameConfig().PreserveLength
		return fmt.Sprintf("transform_full_name(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64_PHONE_NUMBER:
		pl := col.Transformer.Config.GetTransformInt64PhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_int64_phone_number(value:this.%q,preserve_length:%t)", col.Column, pl), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64:
		rMin := col.Transformer.Config.GetTransformInt64Config().RandomizationRangeMin
		rMax := col.Transformer.Config.GetTransformInt64Config().RandomizationRangeMax
		return fmt.Sprintf(`transform_int64(value:this.%q,randomization_range_min:%d,randomization_range_max:%d)`, col.Column, rMin, rMax), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_LAST_NAME:
		pl := col.Transformer.Config.GetTransformLastNameConfig().PreserveLength
		return fmt.Sprintf("transform_last_name(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_PHONE_NUMBER:
		pl := col.Transformer.Config.GetTransformPhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_phone_number(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_STRING:
		pl := col.Transformer.Config.GetTransformStringConfig().PreserveLength
		minLength := int64(3) // todo: we need to pull in this value from the database schema
		return fmt.Sprintf(`transform_string(value:this.%q,preserve_length:%t,min_length:%d,max_length:%d)`, col.Column, pl, minLength, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL:
		return shared.NullString, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT:
		return `"DEFAULT"`, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_CHARACTER_SCRAMBLE:
		regex := col.Transformer.Config.GetTransformCharacterScrambleConfig().UserProvidedRegex

		if regex != nil {
			regexValue := *regex
			return fmt.Sprintf(`transform_character_scramble(value:this.%q,user_provided_regex:%q)`, col.Column, regexValue), nil
		} else {
			return fmt.Sprintf(`transform_character_scramble(value:this.%q)`, col.Column), nil
		}

	default:
		return "", fmt.Errorf("unsupported transformer")
	}
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
