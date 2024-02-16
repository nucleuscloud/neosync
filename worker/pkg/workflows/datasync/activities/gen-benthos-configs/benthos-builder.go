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
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
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
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}
	responses := []*BenthosConfigResponse{}

	groupedMappings := groupMappingsByTable(job.Mappings)
	groupedTableMapping := map[string]*tableMapping{}
	for _, tm := range groupedMappings {
		groupedTableMapping[neosync_benthos.BuildBenthosTable(tm.Schema, tm.Table)] = tm
	}
	uniqueSchemas := shared.GetUniqueSchemasFromMappings(job.Mappings)

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

	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		sourceConnection, err := b.getConnectionById(ctx, jobSourceConfig.Postgres.ConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get connection by id (%s): %w", jobSourceConfig.Postgres.ConnectionId, err)
		}
		pgconfig := sourceConnection.ConnectionConfig.GetPgConfig()
		if pgconfig == nil {
			return nil, fmt.Errorf("source connection (%s) is not a postgres config", jobSourceConfig.Postgres.ConnectionId)
		}
		sqlOpts := jobSourceConfig.Postgres
		var sourceTableOpts map[string]*sqlSourceTableOptions
		if sqlOpts != nil {
			sourceTableOpts = groupPostgresSourceOptionsByTable(sqlOpts.Schemas)
		}

		if _, ok := b.pgpool[sourceConnection.Id]; !ok {
			pgconn, err := b.sqlconnector.NewPgPoolFromConnectionConfig(pgconfig, shared.Ptr(uint32(5)), slogger)
			if err != nil {
				return nil, fmt.Errorf("unable to create new postgres pool from connection config: %w", err)
			}
			pool, err := pgconn.Open(ctx)
			if err != nil {
				return nil, fmt.Errorf("unable to open postgres connection: %w", err)
			}
			defer pgconn.Close()
			b.pgpool[sourceConnection.Id] = pool
		}
		pool := b.pgpool[sourceConnection.Id]

		// validate job mappings align with sql connections
		dbschemas, err := b.pgquerier.GetDatabaseSchema(ctx, pool)
		if err != nil {
			return nil, fmt.Errorf("unable to get database schema for postgres connection: %w", err)
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

		sourceResponses, err := buildBenthosSqlSourceConfigResponses(ctx, b.transformerclient, groupedMappings, jobSourceConfig.Postgres.ConnectionId, "postgres", sourceTableOpts, groupedSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to build postgres benthos sql source config responses: %w", err)
		}
		responses = append(responses, sourceResponses...)

		allConstraints, err := dbschemas_postgres.GetAllPostgresFkConstraints(b.pgquerier, ctx, pool, uniqueSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve postgres foreign key constraints: %w", err)
		}
		td := dbschemas_postgres.GetPostgresTableDependencies(allConstraints)
		tables := filterNullTables(groupedMappings)
		dependencyConfigs := tabledependency.GetRunConfigs(td, tables)
		primaryKeys, err := b.getAllPostgresPkConstraints(ctx, pool, uniqueSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to get all postgres primary key constraints: %w", err)
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
				return nil, fmt.Errorf("unexpected number of benthos configs")
			}
		}

	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		sourceConnection, err := b.getConnectionById(ctx, jobSourceConfig.Mysql.ConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get connection (%s) by id: %w", jobSourceConfig.Mysql.ConnectionId, err)
		}
		mysqlconfig := sourceConnection.ConnectionConfig.GetMysqlConfig()
		if mysqlconfig == nil {
			return nil, fmt.Errorf("source connection (%s) is not a mysql config", jobSourceConfig.Mysql.ConnectionId)
		}

		sqlOpts := jobSourceConfig.Mysql
		var sourceTableOpts map[string]*sqlSourceTableOptions
		if sqlOpts != nil {
			sourceTableOpts = groupMysqlSourceOptionsByTable(sqlOpts.Schemas)
		}

		if _, ok := b.mysqlpool[sourceConnection.Id]; !ok {
			conn, err := b.sqlconnector.NewDbFromConnectionConfig(sourceConnection.ConnectionConfig, shared.Ptr(uint32(5)), slogger)
			if err != nil {
				return nil, fmt.Errorf("unable to create new mysql pool from connection config: %w", err)
			}
			pool, err := conn.Open()
			if err != nil {
				return nil, fmt.Errorf("unable to open mysql connection: %w", err)
			}
			defer conn.Close()
			b.mysqlpool[sourceConnection.Id] = pool
		}
		pool := b.mysqlpool[sourceConnection.Id]

		// validate job mappings align with sql connections
		dbschemas, err := b.mysqlquerier.GetDatabaseSchema(ctx, pool)
		if err != nil {
			return nil, fmt.Errorf("unable to get database schema for mysql connection: %w", err)
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

		sourceResponses, err := buildBenthosSqlSourceConfigResponses(ctx, b.transformerclient, groupedMappings, jobSourceConfig.Mysql.ConnectionId, "mysql", sourceTableOpts, groupedSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to build mysql benthos sql source config responses: %w", err)
		}
		responses = append(responses, sourceResponses...)

		allConstraints, err := dbschemas_mysql.GetAllMysqlFkConstraints(b.mysqlquerier, ctx, pool, uniqueSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve mysql foreign key constraints: %w", err)
		}
		td := dbschemas_mysql.GetMysqlTableDependencies(allConstraints)
		tables := filterNullTables(groupedMappings)
		dependencyConfigs := tabledependency.GetRunConfigs(td, tables)
		primaryKeys, err := b.getAllMysqlPkConstraints(ctx, pool, uniqueSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to get all mysql primary key constraints: %w", err)
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
							return nil, fmt.Errorf("no primary keys found for table (%s). Unable to build update query", tableName)
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

			switch connection := destinationConnection.ConnectionConfig.Config.(type) {
			case *mgmtv1alpha1.ConnectionConfig_PgConfig:
				resp.BenthosDsns = append(resp.BenthosDsns, &shared.BenthosDsn{EnvVarKey: dstEnvVarKey, ConnectionId: destinationConnection.Id})

				if resp.Config.Input.SqlSelect != nil {
					colSourceMap := map[string]string{}
					for _, col := range tm.Mappings {
						colSourceMap[col.Column] = col.GetTransformer().Source
					}

					out := buildPostgresOutputQueryAndArgs(resp, tm, tableKey, colSourceMap)
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
						updateResp, err := createSqlUpdateBenthosConfig(ctx, b.transformerclient, resp, dsn, tableKey, tm, colSourceMap, groupedColInfoMap)
						if err != nil {
							return nil, fmt.Errorf("unable to create sql update benthos config: %w", err)
						}
						updateResponses = append(updateResponses, updateResp)
					}
				} else if resp.Config.Input.Generate != nil {
					cols := buildPlainColumns(tm.Mappings)
					colSourceMap := map[string]string{}
					for _, col := range tm.Mappings {
						colSourceMap[col.Column] = col.GetTransformer().Source
					}

					filteredCols := filterColsBySource(cols, colSourceMap) // filters out default columns

					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
						SqlRaw: &neosync_benthos.SqlRaw{
							Driver: "postgres",
							Dsn:    dsn,

							Query:       buildPostgresInsertQuery(tableKey, cols, colSourceMap),
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
				resp.BenthosDsns = append(resp.BenthosDsns, &shared.BenthosDsn{EnvVarKey: dstEnvVarKey, ConnectionId: destination.ConnectionId})

				if resp.Config.Input.SqlSelect != nil {
					colSourceMap := map[string]string{}
					for _, col := range tm.Mappings {
						colSourceMap[col.Column] = col.GetTransformer().Source
					}

					out := buildMysqlOutputQueryAndArgs(resp, tm, tableKey, colSourceMap)
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
						updateResp, err := createSqlUpdateBenthosConfig(ctx, b.transformerclient, resp, dsn, tableKey, tm, colSourceMap, groupedColInfoMap)
						if err != nil {
							return nil, fmt.Errorf("unable to create sql update benthos config: %w", err)
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
					filteredCols := filterColsBySource(cols, colSourceMap)

					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
						SqlRaw: &neosync_benthos.SqlRaw{
							Driver: "mysql",
							Dsn:    dsn,

							Query:       buildMysqlInsertQuery(tableKey, cols, colSourceMap),
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
	slogger.Info(fmt.Sprintf("successfully built %d benthos configs", len(responses)))
	return &GenerateBenthosConfigsResponse{
		BenthosConfigs: responses,
	}, nil
}

type sqlOutput struct {
	Query       string
	ArgsMapping string
	Columns     []string
}

func buildPostgresOutputQueryAndArgs(resp *BenthosConfigResponse, tm *tableMapping, tableKey string, colSourceMap map[string]string) *sqlOutput {
	if len(resp.excludeColumns) > 0 {
		filteredInsertMappings := []*mgmtv1alpha1.JobMapping{}
		for _, m := range tm.Mappings {
			if !slices.Contains(resp.excludeColumns, m.Column) {
				filteredInsertMappings = append(filteredInsertMappings, m)
			}
		}
		insertCols := buildPlainColumns(filteredInsertMappings)
		filteredInsertCols := filterColsBySource(insertCols, colSourceMap) // filters out default columns
		return &sqlOutput{
			Query:       buildPostgresInsertQuery(tableKey, insertCols, colSourceMap),
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
		filteredUpdateCols := filterColsBySource(updateCols, colSourceMap) // filters out default columns
		updateArgsMapping := []string{}
		updateArgsMapping = append(updateArgsMapping, filteredUpdateCols...)
		updateArgsMapping = append(updateArgsMapping, resp.primaryKeys...)

		return &sqlOutput{
			Query:       buildPostgresUpdateQuery(tableKey, updateCols, colSourceMap, resp.primaryKeys),
			ArgsMapping: buildPlainInsertArgs(updateArgsMapping),
			Columns:     updateCols,
		}
	} else {
		cols := buildPlainColumns(tm.Mappings)
		filteredCols := filterColsBySource(cols, colSourceMap) // filters out default columns
		return &sqlOutput{
			Query:       buildPostgresInsertQuery(tableKey, cols, colSourceMap),
			ArgsMapping: buildPlainInsertArgs(filteredCols),
			Columns:     cols,
		}
	}
}

func buildMysqlOutputQueryAndArgs(resp *BenthosConfigResponse, tm *tableMapping, tableKey string, colSourceMap map[string]string) *sqlOutput {
	if len(resp.excludeColumns) > 0 {
		filteredInsertMappings := []*mgmtv1alpha1.JobMapping{}
		for _, m := range tm.Mappings {
			if !slices.Contains(resp.excludeColumns, m.Column) {
				filteredInsertMappings = append(filteredInsertMappings, m)
			}
		}
		insertCols := buildPlainColumns(filteredInsertMappings)
		filteredInsertCols := filterColsBySource(insertCols, colSourceMap) // filters out default columns
		return &sqlOutput{
			Query:       buildMysqlInsertQuery(tableKey, insertCols, colSourceMap),
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
		filteredUpdateCols := filterColsBySource(updateCols, colSourceMap) // filters out default columns
		updateArgsMapping := []string{}
		updateArgsMapping = append(updateArgsMapping, filteredUpdateCols...)
		updateArgsMapping = append(updateArgsMapping, resp.primaryKeys...)

		return &sqlOutput{
			Query:       buildMysqlUpdateQuery(tableKey, updateCols, colSourceMap, resp.primaryKeys),
			ArgsMapping: buildPlainInsertArgs(updateArgsMapping),
			Columns:     updateCols,
		}
	} else {
		cols := buildPlainColumns(tm.Mappings)
		filteredCols := filterColsBySource(cols, colSourceMap) // filters out default columns
		return &sqlOutput{
			Query:       buildMysqlInsertQuery(tableKey, cols, colSourceMap),
			ArgsMapping: buildPlainInsertArgs(filteredCols),
			Columns:     cols,
		}
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

		mapping, err := buildMutationConfigs(ctx, transformerclient, tableMapping.Mappings, columnInfo)
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

func buildPostgresUpdateQuery(table string, columns []string, colSourceMap map[string]string, primaryKeys []string) string {
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

func buildPostgresInsertQuery(table string, columns []string, colSourceMap map[string]string) string {
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

func buildMysqlInsertQuery(table string, columns []string, colSourceMap map[string]string) string {
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

func buildMysqlUpdateQuery(table string, columns []string, colSourceMap map[string]string, primaryKeys []string) string {
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

func filterColsBySource(columns []string, colSourceMap map[string]string) []string {
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
func filterNullTables(mappings []*tableMapping) []string {
	tables := []string{}
	for _, group := range mappings {
		if !shared.AreAllColsNull(group.Mappings) {
			tables = append(tables, dbschemas_utils.BuildTable(group.Schema, group.Table))
		}
	}
	return tables
}

// creates copy of benthos insert config
// changes query and argsmapping to sql update statement
func createSqlUpdateBenthosConfig(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	insertConfig *BenthosConfigResponse,
	dsn, tableKey string,
	tm *tableMapping,
	colSourceMap map[string]string,
	groupedColInfo map[string]map[string]*dbschemas_utils.ColumnInfo,
) (*BenthosConfigResponse, error) {
	driver := insertConfig.Config.Input.SqlSelect.Driver
	sourceResponses, err := buildBenthosSqlSourceConfigResponses(
		ctx,
		transformerclient,
		[]*tableMapping{tm},
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
			out := buildPostgresOutputQueryAndArgs(newResp, tm, tableKey, colSourceMap)
			output = out
		} else if driver == "mysql" {
			out := buildMysqlOutputQueryAndArgs(newResp, tm, tableKey, colSourceMap)
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

type sqlSourceTableOptions struct {
	WhereClause *string
}

func buildBenthosSqlSourceConfigResponses(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	mappings []*tableMapping,
	dsnConnectionId string,
	driver string,
	sourceTableOpts map[string]*sqlSourceTableOptions,
	groupedColumnInfo map[string]map[string]*dbschemas_utils.ColumnInfo,
) ([]*BenthosConfigResponse, error) {
	responses := []*BenthosConfigResponse{}

	for i := range mappings {
		tableMapping := mappings[i]
		cols := buildPlainColumns(tableMapping.Mappings)
		if shared.AreAllColsNull(tableMapping.Mappings) {
			// skipping table as no columns are mapped
			continue
		}

		var where string
		tableOpt := sourceTableOpts[neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table)]
		if tableOpt != nil && tableOpt.WhereClause != nil {
			where = *tableOpt.WhereClause
		}

		table := neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table)
		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						SqlSelect: &neosync_benthos.SqlSelect{
							Driver: driver,
							Dsn:    "${SOURCE_CONNECTION_DSN}",

							Table:   table,
							Where:   where,
							Columns: cols,
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

		processorConfigs, err := buildProcessorConfigs(ctx, transformerclient, tableMapping.Mappings, groupedColumnInfo[table])
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

func groupPostgresSourceOptionsByTable(
	schemaOptions []*mgmtv1alpha1.PostgresSourceSchemaOption,
) map[string]*sqlSourceTableOptions {
	groupedMappings := map[string]*sqlSourceTableOptions{}

	for idx := range schemaOptions {
		schemaOpt := schemaOptions[idx]
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

func groupMysqlSourceOptionsByTable(
	schemaOptions []*mgmtv1alpha1.MysqlSourceSchemaOption,
) map[string]*sqlSourceTableOptions {
	groupedMappings := map[string]*sqlSourceTableOptions{}

	for idx := range schemaOptions {
		schemaOpt := schemaOptions[idx]
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

func buildProcessorConfigs(ctx context.Context, transformerclient mgmtv1alpha1connect.TransformersServiceClient, cols []*mgmtv1alpha1.JobMapping, tableColumnInfo map[string]*dbschemas_utils.ColumnInfo) ([]*neosync_benthos.ProcessorConfig, error) {
	jsCode, err := extractJsFunctionsAndOutputs(ctx, transformerclient, cols)
	if err != nil {
		return nil, err
	}

	mutations, err := buildMutationConfigs(ctx, transformerclient, cols, tableColumnInfo)
	if err != nil {
		return nil, err
	}

	var processorConfigs []*neosync_benthos.ProcessorConfig
	if mutations != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Mutation: &mutations})
	}
	if jsCode != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Javascript: &neosync_benthos.JavascriptConfig{Code: jsCode}})
	}

	return processorConfigs, err
}

func extractJsFunctionsAndOutputs(ctx context.Context, transformerclient mgmtv1alpha1connect.TransformersServiceClient, cols []*mgmtv1alpha1.JobMapping) (string, error) {
	var benthosOutputs []string
	var jsFunctions []string

	for _, col := range cols {
		if shouldProcessColumn(col.Transformer) {
			if _, ok := col.Transformer.Config.Config.(*mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig); ok {
				val, err := convertUserDefinedFunctionConfig(ctx, transformerclient, col.Transformer)
				if err != nil {
					return "", errors.New("unable to look up user defined transformer config by id")
				}
				col.Transformer = val
			}
			if col.Transformer.Source == "transform_javascript" {
				code := col.Transformer.Config.GetTransformJavascriptConfig().Code
				if code != "" {
					jsFunctions = append(jsFunctions, constructJsFunction(code, col.Column))
					benthosOutputs = append(benthosOutputs, constructBenthosOutput(col.Column))
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
			if col.Transformer.Source != "transform_javascript" {
				mutation, err := computeMutationFunction(col, colInfo)
				if err != nil {
					return "", fmt.Errorf("%s is not a supported transformer: %w", col.Transformer, err)
				}
				mutations = append(mutations, fmt.Sprintf("root.%s = %s", col.Column, mutation))
			}
		}
	}

	return strings.Join(mutations, "\n"), nil
}

func shouldProcessColumn(t *mgmtv1alpha1.JobMappingTransformer) bool {
	return t != nil &&
		t.Source != "" &&
		t.Source != "passthrough" &&
		t.Source != "generate_default"
}

func constructJsFunction(jsCode, col string) string {
	if jsCode != "" {
		return fmt.Sprintf(`
function fn_%s(value, input){
  %s
};
`, col, jsCode)
	} else {
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

func constructBenthosOutput(col string) string {
	return fmt.Sprintf(`output["%[1]s"] = fn_%[1]s(input["%[1]s"], input);`, col)
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
		pieces[idx] = fmt.Sprintf("this.%s", cols[idx])
	}
	return fmt.Sprintf("root = [%s]", strings.Join(pieces, ", "))
}

/*
function transformers
root.{destination_col} = transformerfunction(args)
*/

func computeMutationFunction(col *mgmtv1alpha1.JobMapping, colInfo *dbschemas_utils.ColumnInfo) (string, error) {
	var maxLen int32 = 10000
	if colInfo != nil && colInfo.CharacterMaximumLength != nil {
		maxLen = *colInfo.CharacterMaximumLength
	}

	switch col.Transformer.Source {
	case "generate_categorical":
		categories := col.Transformer.Config.GetGenerateCategoricalConfig().Categories
		return fmt.Sprintf(`generate_categorical(categories: %q)`, categories), nil
	case "generate_email":
		return fmt.Sprintf(`generate_email(max_length:%d)`, maxLen), nil
	case "transform_email":
		pd := col.Transformer.Config.GetTransformEmailConfig().PreserveDomain
		pl := col.Transformer.Config.GetTransformEmailConfig().PreserveLength
		excludedDomains := col.Transformer.Config.GetTransformEmailConfig().ExcludedDomains

		sliceBytes, err := json.Marshal(excludedDomains)
		if err != nil {
			return "", err
		}

		excludedDomainstStr := string(sliceBytes)
		return fmt.Sprintf("transform_email(email:this.%s,preserve_domain:%t,preserve_length:%t,excluded_domains:%v,max_length:%d)", col.Column, pd, pl, excludedDomainstStr, maxLen), nil
	case "generate_bool":
		return "generate_bool()", nil
	case "generate_card_number":
		luhn := col.Transformer.Config.GetGenerateCardNumberConfig().ValidLuhn
		return fmt.Sprintf(`generate_card_number(valid_luhn:%t)`, luhn), nil
	case "generate_city":
		return fmt.Sprintf(`generate_city(max_length:%d)`, maxLen), nil
	case "generate_e164_phone_number":
		min := col.Transformer.Config.GetGenerateE164PhoneNumberConfig().Min
		max := col.Transformer.Config.GetGenerateE164PhoneNumberConfig().Max
		return fmt.Sprintf(`generate_e164_phone_number(min:%d,max:%d)`, min, max), nil
	case "generate_first_name":
		return fmt.Sprintf(`generate_first_name(max_length:%d)`, maxLen), nil
	case "generate_float64":
		randomSign := col.Transformer.Config.GetGenerateFloat64Config().RandomizeSign
		min := col.Transformer.Config.GetGenerateFloat64Config().Min
		max := col.Transformer.Config.GetGenerateFloat64Config().Max
		precision := col.Transformer.Config.GetGenerateFloat64Config().Precision
		return fmt.Sprintf(`generate_float64(randomize_sign:%t, min:%f, max:%f, precision:%d)`, randomSign, min, max, precision), nil
	case "generate_full_address":
		return fmt.Sprintf(`generate_full_address(max_length:%d)`, maxLen), nil
	case "generate_full_name":
		return fmt.Sprintf(`generate_full_name(max_length:%d)`, maxLen), nil
	case "generate_gender":
		ab := col.Transformer.Config.GetGenerateGenderConfig().Abbreviate
		return fmt.Sprintf(`generate_gender(abbreviate:%t,max_length:%d)`, ab, maxLen), nil
	case "generate_int64_phone_number":
		return "generate_int64_phone_number()", nil
	case "generate_int64":
		sign := col.Transformer.Config.GetGenerateInt64Config().RandomizeSign
		min := col.Transformer.Config.GetGenerateInt64Config().Min
		max := col.Transformer.Config.GetGenerateInt64Config().Max
		return fmt.Sprintf(`generate_int64(randomize_sign:%t,min:%d, max:%d)`, sign, min, max), nil
	case "generate_last_name":
		return fmt.Sprintf(`generate_last_name(max_length:%d)`, maxLen), nil
	case "generate_sha256hash":
		return `generate_sha256hash()`, nil
	case "generate_ssn":
		return "generate_ssn()", nil
	case "generate_state":
		return "generate_state()", nil
	case "generate_street_address":
		return fmt.Sprintf(`generate_street_address(max_length:%d)`, maxLen), nil
	case "generate_string_phone_number":
		min := col.Transformer.Config.GetGenerateStringPhoneNumberConfig().Min
		max := col.Transformer.Config.GetGenerateStringPhoneNumberConfig().Max
		return fmt.Sprintf("generate_string_phone_number(min:%d,max:%d,max_length:%d)", min, max, maxLen), nil
	case "generate_string":
		min := col.Transformer.Config.GetGenerateStringConfig().Min
		max := col.Transformer.Config.GetGenerateStringConfig().Max
		return fmt.Sprintf(`generate_string(min:%d,max:%d,max_length:%d)`, min, max, maxLen), nil
	case "generate_unixtimestamp":
		return "generate_unixtimestamp()", nil
	case "generate_username":
		return fmt.Sprintf(`generate_username(max_length:%d)`, maxLen), nil
	case "generate_utctimestamp":
		return "generate_utctimestamp()", nil
	case "generate_uuid":
		ih := col.Transformer.Config.GetGenerateUuidConfig().IncludeHyphens
		return fmt.Sprintf("generate_uuid(include_hyphens:%t)", ih), nil
	case "generate_zipcode":
		return "generate_zipcode()", nil
	case "transform_e164_phone_number":
		pl := col.Transformer.Config.GetTransformE164PhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_e164_phone_number(value:this.%s,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case "transform_first_name":
		pl := col.Transformer.Config.GetTransformFirstNameConfig().PreserveLength
		return fmt.Sprintf("transform_first_name(value:this.%s,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case "transform_float64":
		rMin := col.Transformer.Config.GetTransformFloat64Config().RandomizationRangeMin
		rMax := col.Transformer.Config.GetTransformFloat64Config().RandomizationRangeMax
		return fmt.Sprintf(`transform_float64(value:this.%s,randomization_range_min:%f,randomization_range_max:%f)`, col.Column, rMin, rMax), nil
	case "transform_full_name":
		pl := col.Transformer.Config.GetTransformFullNameConfig().PreserveLength
		return fmt.Sprintf("transform_full_name(value:this.%s,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case "transform_int64_phone_number":
		pl := col.Transformer.Config.GetTransformInt64PhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_int64_phone_number(value:this.%s,preserve_length:%t)", col.Column, pl), nil
	case "transform_int64":
		rMin := col.Transformer.Config.GetTransformInt64Config().RandomizationRangeMin
		rMax := col.Transformer.Config.GetTransformInt64Config().RandomizationRangeMax
		return fmt.Sprintf(`transform_int64(value:this.%s,randomization_range_min:%d,randomization_range_max:%d)`, col.Column, rMin, rMax), nil
	case "transform_last_name":
		pl := col.Transformer.Config.GetTransformLastNameConfig().PreserveLength
		return fmt.Sprintf("transform_last_name(value:this.%s,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case "transform_phone_number":
		pl := col.Transformer.Config.GetTransformPhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_phone_number(value:this.%s,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case "transform_string":
		pl := col.Transformer.Config.GetTransformStringConfig().PreserveLength
		return fmt.Sprintf(`transform_string(value:this.%s,preserve_length:%t,max_length:%d)`, col.Column, pl, maxLen), nil
	case shared.NullString:
		return shared.NullString, nil
	case generateDefault:
		return "default", nil
	case "transform_character_scramble":
		regex := col.Transformer.Config.GetTransformCharacterScrambleConfig().UserProvidedRegex

		if regex != nil {
			regexValue := *regex
			return fmt.Sprintf(`transform_character_scramble(value:this.%s,user_provided_regex:%q)`, col.Column, regexValue), nil
		} else {
			return fmt.Sprintf(`transform_character_scramble(value:this.%s)`, col.Column), nil
		}

	default:
		return "", fmt.Errorf("unsupported transformer")
	}
}
