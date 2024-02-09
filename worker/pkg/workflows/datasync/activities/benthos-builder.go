package datasync_activities

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"connectrpc.com/connect"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	dbschemas_mysql "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/mysql"
	dbschemas_postgres "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/postgres"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
)

const (
	generateDefault            = "generate_default"
	dbDefault                  = "DEFAULT"
	jobmappingSubsetErrMsg     = "job mappings are not equal to or a subset of the database schema found in the source connection"
	haltOnSchemaAdditionErrMsg = "job mappings does not contain a column mapping for all " +
		"columns found in the source connection for the selected schemas and tables"
)

type benthosBuilder struct {
	pgpool    map[string]pg_queries.DBTX
	pgquerier pg_queries.Querier

	mysqlpool    map[string]mysql_queries.DBTX
	mysqlquerier mysql_queries.Querier

	sqlconnector sqlconnect.SqlConnector

	jobclient         mgmtv1alpha1connect.JobServiceClient
	connclient        mgmtv1alpha1connect.ConnectionServiceClient
	transformerclient mgmtv1alpha1connect.TransformersServiceClient
}

func newBenthosBuilder(
	pgpool map[string]pg_queries.DBTX,
	pgquerier pg_queries.Querier,

	mysqlpool map[string]mysql_queries.DBTX,
	mysqlquerier mysql_queries.Querier,

	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,

	sqlconnector sqlconnect.SqlConnector,

) *benthosBuilder {
	return &benthosBuilder{
		pgpool:            pgpool,
		pgquerier:         pgquerier,
		mysqlpool:         mysqlpool,
		mysqlquerier:      mysqlquerier,
		sqlconnector:      sqlconnector,
		jobclient:         jobclient,
		connclient:        connclient,
		transformerclient: transformerclient,
	}
}

func (b *benthosBuilder) GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
	slogger *slog.Logger,
) (*GenerateBenthosConfigsResponse, error) {
	job, err := b.getJobById(ctx, req.JobId)
	if err != nil {
		return nil, err
	}
	responses := []*BenthosConfigResponse{}

	groupedMappings := groupMappingsByTable(job.Mappings)
	groupedTableMapping := map[string]*TableMapping{}
	for _, tm := range groupedMappings {
		groupedTableMapping[neosync_benthos.BuildBenthosTable(tm.Schema, tm.Table)] = tm
	}
	uniqueSchemas := getUniqueSchemasFromMappings(job.Mappings)

	var groupedColInfoMap map[string]map[string]*dbschemas_utils.ColumnInfo

	switch jobSourceConfig := job.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		sourceTableOpts := groupGenerateSourceOptionsByTable(jobSourceConfig.Generate.Schemas)
		// TODO this needs to be updated to get db schema
		sourceResponses, err := b.buildBenthosGenerateSourceConfigResponses(ctx, groupedMappings, sourceTableOpts, map[string]*dbschemas_utils.ColumnInfo{})
		if err != nil {
			return nil, err
		}
		responses = append(responses, sourceResponses...)

	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		sourceConnection, err := b.getConnectionById(ctx, jobSourceConfig.Postgres.ConnectionId)
		if err != nil {
			return nil, err
		}
		pgconfig := sourceConnection.ConnectionConfig.GetPgConfig()
		if pgconfig == nil {
			return nil, errors.New("source connection is not a postgres config")
		}
		sqlOpts := jobSourceConfig.Postgres
		var sourceTableOpts map[string]*sqlSourceTableOptions
		if sqlOpts != nil {
			sourceTableOpts = groupPostgresSourceOptionsByTable(sqlOpts.Schemas)
		}

		if _, ok := b.pgpool[sourceConnection.Id]; !ok {
			pgconn, err := b.sqlconnector.NewPgPoolFromConnectionConfig(pgconfig, ptr(uint32(5)), slogger)
			if err != nil {
				return nil, err
			}
			pool, err := pgconn.Open(ctx)
			if err != nil {
				return nil, err
			}
			defer pgconn.Close()
			b.pgpool[sourceConnection.Id] = pool
		}
		pool := b.pgpool[sourceConnection.Id]

		// validate job mappings align with sql connections
		dbschemas, err := b.pgquerier.GetDatabaseSchema(ctx, pool)
		if err != nil {
			return nil, err
		}
		groupedSchemas := dbschemas_postgres.GetUniqueSchemaColMappings(dbschemas)
		if !areMappingsSubsetOfSchemas(groupedSchemas, job.Mappings) {
			return nil, errors.New(jobmappingSubsetErrMsg)
		}
		if sqlOpts != nil && sqlOpts.HaltOnNewColumnAddition &&
			shouldHaltOnSchemaAddition(groupedSchemas, job.Mappings) {
			return nil, errors.New(haltOnSchemaAdditionErrMsg)
		}

		groupedColInfoMap = groupedSchemas

		sourceResponses, err := b.buildBenthosSqlSourceConfigResponses(ctx, groupedMappings, jobSourceConfig.Postgres.ConnectionId, "postgres", sourceTableOpts, groupedSchemas)
		if err != nil {
			return nil, err
		}
		responses = append(responses, sourceResponses...)

		allConstraints, err := dbschemas_postgres.GetAllPostgresFkConstraints(b.pgquerier, ctx, pool, uniqueSchemas)
		if err != nil {
			return nil, err
		}
		td := dbschemas_postgres.GetPostgresTableDependencies(allConstraints)
		tables := b.filterNullTables(groupedMappings)
		dependencyConfigs := tabledependency.GetRunConfigs(td, tables)
		primaryKeys, err := b.getAllPostgresPkConstraints(ctx, pool, uniqueSchemas)
		if err != nil {
			return nil, err
		}
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
						pks := primaryKeys[tableName]
						if len(pks) == 0 {
							return nil, fmt.Errorf("No primary keys found for table (%s). Unable to build update query.", tableName)
						}

						// config for sql update
						resp.updateConfig = c
						resp.primaryKeys = pks
					}
				}
			} else if len(configs) == 1 {
				resp.DependsOn = configs[0].DependsOn
			} else {
				return nil, fmt.Errorf("unexpected number of benthos configs")
			}
		}

	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		sourceConnection, err := b.getConnectionById(ctx, jobSourceConfig.Mysql.ConnectionId)
		if err != nil {
			return nil, err
		}
		mysqlconfig := sourceConnection.ConnectionConfig.GetMysqlConfig()
		if mysqlconfig == nil {
			return nil, errors.New("source connection is not a mysql config")
		}

		sqlOpts := jobSourceConfig.Mysql
		var sourceTableOpts map[string]*sqlSourceTableOptions
		if sqlOpts != nil {
			sourceTableOpts = groupMysqlSourceOptionsByTable(sqlOpts.Schemas)
		}

		if _, ok := b.mysqlpool[sourceConnection.Id]; !ok {
			conn, err := b.sqlconnector.NewDbFromConnectionConfig(sourceConnection.ConnectionConfig, ptr(uint32(5)), slogger)
			if err != nil {
				return nil, err
			}
			pool, err := conn.Open()
			if err != nil {
				return nil, err
			}
			defer conn.Close()
			b.mysqlpool[sourceConnection.Id] = pool
		}
		pool := b.mysqlpool[sourceConnection.Id]

		// validate job mappings align with sql connections
		dbschemas, err := b.mysqlquerier.GetDatabaseSchema(ctx, pool)
		if err != nil {
			return nil, err
		}
		groupedSchemas := dbschemas_mysql.GetUniqueSchemaColMappings(dbschemas)
		if !areMappingsSubsetOfSchemas(groupedSchemas, job.Mappings) {
			return nil, errors.New(jobmappingSubsetErrMsg)
		}
		if sqlOpts != nil && sqlOpts.HaltOnNewColumnAddition &&
			shouldHaltOnSchemaAddition(groupedSchemas, job.Mappings) {
			return nil, errors.New(haltOnSchemaAdditionErrMsg)
		}
		groupedColInfoMap = groupedSchemas

		sourceResponses, err := b.buildBenthosSqlSourceConfigResponses(ctx, groupedMappings, jobSourceConfig.Mysql.ConnectionId, "mysql", sourceTableOpts, groupedSchemas)
		if err != nil {
			return nil, err
		}
		responses = append(responses, sourceResponses...)

		allConstraints, err := dbschemas_mysql.GetAllMysqlFkConstraints(b.mysqlquerier, ctx, pool, uniqueSchemas)
		if err != nil {
			return nil, err
		}
		td := dbschemas_mysql.GetMysqlTableDependencies(allConstraints)
		tables := b.filterNullTables(groupedMappings)
		dependencyConfigs := tabledependency.GetRunConfigs(td, tables)
		primaryKeys, err := b.getAllMysqlPkConstraints(ctx, pool, uniqueSchemas)
		if err != nil {
			return nil, err
		}

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
						pks := primaryKeys[tableName]
						if len(pks) == 0 {
							return nil, fmt.Errorf("No primary keys found for table (%s). Unable to build update query.", tableName)
						}
						// config for sql update
						resp.updateConfig = c
						resp.primaryKeys = primaryKeys[tableName]
					}
				}
			} else if len(configs) == 1 {
				resp.DependsOn = configs[0].DependsOn
			} else {
				return nil, fmt.Errorf("unexpected number of benthos configs")
			}
		}

	default:
		return nil, errors.New("unsupported job source")
	}

	updateResponses := []*BenthosConfigResponse{} // update configs for circular dependecies
	for destIdx, destination := range job.Destinations {
		destinationConnection, err := b.getConnectionById(ctx, destination.ConnectionId)
		if err != nil {
			return nil, err
		}
		for _, resp := range responses {
			tableKey := neosync_benthos.BuildBenthosTable(resp.TableSchema, resp.TableName)
			tm := groupedTableMapping[tableKey]
			if tm == nil {
				return nil, errors.New("unable to find table mapping for key")
			}
			dstEnvVarKey := fmt.Sprintf("DESTINATION_%d_CONNECTION_DSN", destIdx)
			dsn := fmt.Sprintf("${%s}", dstEnvVarKey)

			switch connection := destinationConnection.ConnectionConfig.Config.(type) {
			case *mgmtv1alpha1.ConnectionConfig_PgConfig:
				resp.BenthosDsns = append(resp.BenthosDsns, &BenthosDsn{EnvVarKey: dstEnvVarKey, ConnectionId: destinationConnection.Id})

				if resp.Config.Input.SqlSelect != nil {
					colSourceMap := map[string]string{}
					for _, col := range tm.Mappings {
						colSourceMap[col.Column] = col.GetTransformer().Source
					}

					out := b.buildPostgresOutputQueryAndArgs(resp, tm, tableKey, colSourceMap)
					resp.Columns = out.Columns
					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
						SqlRaw: &neosync_benthos.SqlRaw{
							Driver: "postgres",
							Dsn:    dsn,

							Query:       out.Query,
							ArgsMapping: out.ArgsMapping,

							Batching: &neosync_benthos.Batching{
								Period: "5s",
								Count:  100,
							},
						},
					})
					if resp.updateConfig != nil {
						// circular dependency -> create update benthos config
						updateResp, err := b.createSqlUpdateBenthosConfig(ctx, resp, dsn, tableKey, tm, colSourceMap, groupedColInfoMap)
						if err != nil {
							return nil, err
						}
						updateResponses = append(updateResponses, updateResp)
					}
				} else if resp.Config.Input.Generate != nil {
					cols := buildPlainColumns(tm.Mappings)
					colSourceMap := map[string]string{}
					for _, col := range tm.Mappings {
						colSourceMap[col.Column] = col.GetTransformer().Source
					}

					filteredCols := b.filterColsBySource(cols, colSourceMap) // filters out default columns

					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
						SqlRaw: &neosync_benthos.SqlRaw{
							Driver: "postgres",
							Dsn:    dsn,

							Query:       b.buildPostgresInsertQuery(tableKey, cols, colSourceMap),
							ArgsMapping: buildPlainInsertArgs(filteredCols),

							Batching: &neosync_benthos.Batching{
								Period: "5s",
								Count:  100,
							},
						},
					})
				} else {
					return nil, errors.New("unable to build destination connection due to unsupported source connection")
				}
			case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
				resp.BenthosDsns = append(resp.BenthosDsns, &BenthosDsn{EnvVarKey: dstEnvVarKey, ConnectionId: destination.Id})

				if resp.Config.Input.SqlSelect != nil {
					colSourceMap := map[string]string{}
					for _, col := range tm.Mappings {
						colSourceMap[col.Column] = col.GetTransformer().Source
					}

					out := b.buildMysqlOutputQueryAndArgs(resp, tm, tableKey, colSourceMap)
					resp.Columns = out.Columns
					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
						SqlRaw: &neosync_benthos.SqlRaw{
							Driver: "mysql",
							Dsn:    dsn,

							Query:       out.Query,
							ArgsMapping: out.ArgsMapping,

							Batching: &neosync_benthos.Batching{
								Period: "5s",
								Count:  100,
							},
						},
					})
					if resp.updateConfig != nil {
						// circular dependency -> create update benthos config
						updateResp, err := b.createSqlUpdateBenthosConfig(ctx, resp, dsn, tableKey, tm, colSourceMap, groupedColInfoMap)
						if err != nil {
							return nil, err
						}
						updateResponses = append(updateResponses, updateResp)
					}
				} else if resp.Config.Input.Generate != nil {
					cols := buildPlainColumns(tm.Mappings)
					colSourceMap := map[string]string{}
					for _, col := range tm.Mappings {
						colSourceMap[col.Column] = col.GetTransformer().Source
					}
					// filters out default columns
					filteredCols := b.filterColsBySource(cols, colSourceMap)

					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
						SqlRaw: &neosync_benthos.SqlRaw{
							Driver: "mysql",
							Dsn:    dsn,

							Query:       b.buildMysqlInsertQuery(tableKey, cols, colSourceMap),
							ArgsMapping: buildPlainInsertArgs(filteredCols),

							Batching: &neosync_benthos.Batching{
								Period: "5s",
								Count:  100,
							},
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
				})
			default:
				return nil, fmt.Errorf("unsupported destination connection config")
			}
		}
	}

	responses = append(responses, updateResponses...)
	return &GenerateBenthosConfigsResponse{
		BenthosConfigs: responses,
	}, nil
}

type sqlOutput struct {
	Query       string
	ArgsMapping string
	Columns     []string
}

func (b *benthosBuilder) buildPostgresOutputQueryAndArgs(resp *BenthosConfigResponse, tm *TableMapping, tableKey string, colSourceMap map[string]string) *sqlOutput {
	if len(resp.excludeColumns) > 0 {
		filteredInsertMappings := []*mgmtv1alpha1.JobMapping{}
		for _, m := range tm.Mappings {
			if !slices.Contains(resp.excludeColumns, m.Column) {
				filteredInsertMappings = append(filteredInsertMappings, m)
			}
		}
		insertCols := buildPlainColumns(filteredInsertMappings)
		filteredInsertCols := b.filterColsBySource(insertCols, colSourceMap) // filters out default columns
		return &sqlOutput{
			Query:       b.buildPostgresInsertQuery(tableKey, insertCols, colSourceMap),
			ArgsMapping: buildPlainInsertArgs(filteredInsertCols),
			Columns:     insertCols,
		}
	} else if resp.updateConfig != nil && resp.updateConfig.Columns != nil && len(resp.updateConfig.Columns.Include) > 0 {
		filteredUpdateMappings := []*mgmtv1alpha1.JobMapping{}

		for _, m := range tm.Mappings {
			if slices.Contains(resp.updateConfig.Columns.Include, m.Column) {
				filteredUpdateMappings = append(filteredUpdateMappings, m)
			}
		}
		updateCols := buildPlainColumns(filteredUpdateMappings)
		filteredUpdateCols := b.filterColsBySource(updateCols, colSourceMap) // filters out default columns
		updateArgsMapping := []string{}
		updateArgsMapping = append(updateArgsMapping, filteredUpdateCols...)
		updateArgsMapping = append(updateArgsMapping, resp.primaryKeys...)

		return &sqlOutput{
			Query:       b.buildPostgresUpdateQuery(tableKey, updateCols, colSourceMap, resp.primaryKeys),
			ArgsMapping: buildPlainInsertArgs(updateArgsMapping),
			Columns:     updateCols,
		}
	} else {
		cols := buildPlainColumns(tm.Mappings)
		filteredCols := b.filterColsBySource(cols, colSourceMap) // filters out default columns
		return &sqlOutput{
			Query:       b.buildPostgresInsertQuery(tableKey, cols, colSourceMap),
			ArgsMapping: buildPlainInsertArgs(filteredCols),
			Columns:     cols,
		}
	}
}

func (b *benthosBuilder) buildMysqlOutputQueryAndArgs(resp *BenthosConfigResponse, tm *TableMapping, tableKey string, colSourceMap map[string]string) *sqlOutput {
	if len(resp.excludeColumns) > 0 {
		filteredInsertMappings := []*mgmtv1alpha1.JobMapping{}
		for _, m := range tm.Mappings {
			if !slices.Contains(resp.excludeColumns, m.Column) {
				filteredInsertMappings = append(filteredInsertMappings, m)
			}
		}
		insertCols := buildPlainColumns(filteredInsertMappings)
		filteredInsertCols := b.filterColsBySource(insertCols, colSourceMap) // filters out default columns
		return &sqlOutput{
			Query:       b.buildMysqlInsertQuery(tableKey, insertCols, colSourceMap),
			ArgsMapping: buildPlainInsertArgs(filteredInsertCols),
			Columns:     insertCols,
		}
	} else if resp.updateConfig != nil && resp.updateConfig.Columns != nil && len(resp.updateConfig.Columns.Include) > 0 {
		filteredUpdateMappings := []*mgmtv1alpha1.JobMapping{}

		for _, m := range tm.Mappings {
			if slices.Contains(resp.updateConfig.Columns.Include, m.Column) {
				filteredUpdateMappings = append(filteredUpdateMappings, m)
			}
		}
		updateCols := buildPlainColumns(filteredUpdateMappings)
		filteredUpdateCols := b.filterColsBySource(updateCols, colSourceMap) // filters out default columns
		updateArgsMapping := []string{}
		updateArgsMapping = append(updateArgsMapping, filteredUpdateCols...)
		updateArgsMapping = append(updateArgsMapping, resp.primaryKeys...)

		return &sqlOutput{
			Query:       b.buildMysqlUpdateQuery(tableKey, updateCols, colSourceMap, resp.primaryKeys),
			ArgsMapping: buildPlainInsertArgs(updateArgsMapping),
			Columns:     updateCols,
		}
	} else {
		cols := buildPlainColumns(tm.Mappings)
		filteredCols := b.filterColsBySource(cols, colSourceMap) // filters out default columns
		return &sqlOutput{
			Query:       b.buildMysqlInsertQuery(tableKey, cols, colSourceMap),
			ArgsMapping: buildPlainInsertArgs(filteredCols),
			Columns:     cols,
		}
	}
}

type generateSourceTableOptions struct {
	Count int
}

func (b *benthosBuilder) buildBenthosGenerateSourceConfigResponses(
	ctx context.Context,
	mappings []*TableMapping,
	sourceTableOpts map[string]*generateSourceTableOptions,
	columnInfo map[string]*dbschemas_utils.ColumnInfo,
) ([]*BenthosConfigResponse, error) {
	responses := []*BenthosConfigResponse{}

	for _, tableMapping := range mappings {
		if areAllColsNull(tableMapping.Mappings) {
			// skiping table as no columns are mapped
			continue
		}

		var count = 0
		tableOpt := sourceTableOpts[neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table)]
		if tableOpt != nil {
			count = tableOpt.Count
		}

		mapping, err := b.buildMutationConfigs(ctx, tableMapping.Mappings, columnInfo)
		if err != nil {
			return nil, err
		}

		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						Generate: &neosync_benthos.Generate{
							Interval: "",
							Count:    count,
							Mapping:  mapping,
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

		responses = append(responses, &BenthosConfigResponse{
			Name:      neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table), // todo: may need to expand on this
			Config:    bc,
			DependsOn: []*tabledependency.DependsOn{},

			TableSchema: tableMapping.Schema,
			TableName:   tableMapping.Table,
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

func (b *benthosBuilder) getAllPostgresPkConstraints(
	ctx context.Context,
	conn pg_queries.DBTX,
	uniqueSchemas []string,
) (map[string][]string, error) {
	primaryKeyConstraints, err := dbschemas_postgres.GetAllPostgresPkConstraints(b.pgquerier, ctx, conn, uniqueSchemas)
	if err != nil {
		return nil, err
	}
	pkMap := dbschemas_postgres.GetPostgresTablePrimaryKeys(primaryKeyConstraints)
	return pkMap, nil
}

func (b *benthosBuilder) getAllMysqlPkConstraints(
	ctx context.Context,
	conn mysql_queries.DBTX,
	uniqueSchemas []string,
) (map[string][]string, error) {
	primaryKeyConstraints, err := dbschemas_mysql.GetAllMysqlPkConstraints(b.mysqlquerier, ctx, conn, uniqueSchemas)
	if err != nil {
		return nil, err
	}
	pkMap := dbschemas_mysql.GetMysqlTablePrimaryKeys(primaryKeyConstraints)
	return pkMap, nil
}

func (b *benthosBuilder) buildPostgresUpdateQuery(table string, columns []string, colSourceMap map[string]string, primaryKeys []string) string {
	values := make([]string, len(columns))
	var where string
	paramCount := 1
	for i, col := range columns {
		colSource := colSourceMap[col]
		if colSource == generateDefault {
			values[i] = dbDefault
		} else {
			values[i] = fmt.Sprintf("%s = $%d", col, paramCount)
			paramCount++
		}
	}
	if len(primaryKeys) > 0 {
		clauses := []string{}
		for _, col := range primaryKeys {
			clauses = append(clauses, fmt.Sprintf("%s = $%d", col, paramCount))
			paramCount++
		}
		where = fmt.Sprintf("WHERE %s", strings.Join(clauses, " AND "))
	}
	return fmt.Sprintf("UPDATE %s SET %s %s;", table, strings.Join(values, ", "), where)
}

func (b *benthosBuilder) buildPostgresInsertQuery(table string, columns []string, colSourceMap map[string]string) string {
	values := make([]string, len(columns))
	paramCount := 1
	for i, col := range columns {
		colSource := colSourceMap[col]
		if colSource == generateDefault {
			values[i] = dbDefault
		} else {
			values[i] = fmt.Sprintf("$%d", paramCount)
			paramCount++
		}
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", table, strings.Join(columns, ", "), strings.Join(values, ", "))
}

func (b *benthosBuilder) buildMysqlInsertQuery(table string, columns []string, colSourceMap map[string]string) string {
	values := make([]string, len(columns))
	for i, col := range columns {
		colSource := colSourceMap[col]
		if colSource == generateDefault {
			values[i] = dbDefault
		} else {
			values[i] = "?"
		}
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", table, strings.Join(columns, ", "), strings.Join(values, ", "))
}

func (b *benthosBuilder) buildMysqlUpdateQuery(table string, columns []string, colSourceMap map[string]string, primaryKeys []string) string {
	values := make([]string, len(columns))
	var where string
	for i, col := range columns {
		colSource := colSourceMap[col]
		if colSource == generateDefault {
			values[i] = dbDefault
		} else {
			values[i] = fmt.Sprintf("%s = ?", col)
		}
	}
	if len(primaryKeys) > 0 {
		clauses := []string{}
		for _, col := range primaryKeys {
			clauses = append(clauses, fmt.Sprintf("%s = ?", col))
		}
		where = fmt.Sprintf("WHERE %s", strings.Join(clauses, " AND "))
	}
	return fmt.Sprintf("UPDATE %s SET %s %s;", table, strings.Join(values, ", "), where)
}

func (b *benthosBuilder) filterColsBySource(columns []string, colSourceMap map[string]string) []string {
	filteredCols := []string{}
	for _, col := range columns {
		colSource := colSourceMap[col]
		if colSource != generateDefault {
			filteredCols = append(filteredCols, col)
		}
	}
	return filteredCols
}

// filters out tables where all cols are set to null
func (b *benthosBuilder) filterNullTables(mappings []*TableMapping) []string {
	tables := []string{}
	for _, group := range mappings {
		if !areAllColsNull(group.Mappings) {
			tables = append(tables, dbschemas_utils.BuildTable(group.Schema, group.Table))
		}
	}
	return tables
}

// creates copy of benthos insert config
// changes query and argsmapping to sql update statement
func (b *benthosBuilder) createSqlUpdateBenthosConfig(
	ctx context.Context,
	insertConfig *BenthosConfigResponse,
	dsn, tableKey string,
	tm *TableMapping,
	colSourceMap map[string]string,
	groupedColInfo map[string]map[string]*dbschemas_utils.ColumnInfo,
) (*BenthosConfigResponse, error) {
	driver := insertConfig.Config.Input.SqlSelect.Driver
	sourceResponses, err := b.buildBenthosSqlSourceConfigResponses(
		ctx,
		[]*TableMapping{tm},
		"", // does not matter what is here. gets overwritten with insert config
		driver,
		map[string]*sqlSourceTableOptions{},
		groupedColInfo,
	)
	if err != nil {
		return nil, err
	}

	if len(sourceResponses) > 0 {
		newResp := sourceResponses[0]
		newResp.Config.Input.SqlSelect.Where = insertConfig.Config.Input.SqlSelect.Where // keep the where clause the same as insert
		newResp.updateConfig = insertConfig.updateConfig
		newResp.DependsOn = insertConfig.updateConfig.DependsOn
		newResp.Name = fmt.Sprintf("%s.update", insertConfig.Name)
		newResp.primaryKeys = insertConfig.primaryKeys
		var output *sqlOutput
		if driver == "postgres" {
			out := b.buildPostgresOutputQueryAndArgs(newResp, tm, tableKey, colSourceMap)
			output = out
		} else if driver == "mysql" {
			out := b.buildMysqlOutputQueryAndArgs(newResp, tm, tableKey, colSourceMap)
			output = out
		}
		newResp.Columns = output.Columns
		newResp.BenthosDsns = insertConfig.BenthosDsns
		newResp.Config.Output.Broker.Outputs = append(newResp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
			SqlRaw: &neosync_benthos.SqlRaw{
				Driver: driver,
				Dsn:    dsn,

				Query:       output.Query,
				ArgsMapping: output.ArgsMapping,

				Batching: &neosync_benthos.Batching{
					Period: "5s",
					Count:  100,
				},
			},
		})
		return newResp, nil
	}
	return nil, errors.New("unable to build sql update benthos config")
}
