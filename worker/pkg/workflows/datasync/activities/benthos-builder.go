package datasync_activities

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_mysql "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/mysql"
	dbschemas_postgres "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/postgres"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"golang.org/x/sync/errgroup"

	"go.temporal.io/sdk/log"
)

const (
	generateDefault = "generate_default"
	dbDefault       = "DEFAULT"
)

type benthosBuilder struct {
	pgpool    map[string]pg_queries.DBTX
	pgquerier pg_queries.Querier

	mysqlpool    map[string]mysql_queries.DBTX
	mysqlquerier mysql_queries.Querier

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

) *benthosBuilder {
	return &benthosBuilder{
		pgpool:            pgpool,
		pgquerier:         pgquerier,
		mysqlpool:         mysqlpool,
		mysqlquerier:      mysqlquerier,
		jobclient:         jobclient,
		connclient:        connclient,
		transformerclient: transformerclient,
	}
}

func (b *benthosBuilder) GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
	logger log.Logger,
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

	switch jobSourceConfig := job.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		sourceTableOpts := groupGenerateSourceOptionsByTable(jobSourceConfig.Generate.Schemas)
		sourceResponses, err := b.buildBenthosGenerateSourceConfigResponses(ctx, groupedMappings, sourceTableOpts)
		if err != nil {
			return nil, err
		}
		responses = append(responses, sourceResponses...)

		if jobSourceConfig.Generate.FkSourceConnectionId != nil {
			fkConnectionResp, err := b.connclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{Id: *jobSourceConfig.Generate.FkSourceConnectionId}))
			if err != nil {
				return nil, err
			}
			connection := fkConnectionResp.Msg.Connection

			var td map[string][]string
			switch fkconnconfig := connection.ConnectionConfig.Config.(type) {
			case *mgmtv1alpha1.ConnectionConfig_PgConfig:
				pgconfig := fkconnconfig.PgConfig
				if pgconfig == nil {
					return nil, errors.New("source connection is not a postgres config")
				}
				dsn, err := getPgDsn(pgconfig)
				if err != nil {
					return nil, err
				}

				if _, ok := b.pgpool[dsn]; !ok {
					pool, err := pgxpool.New(ctx, dsn)
					if err != nil {
						return nil, err
					}
					defer pool.Close()
					b.pgpool[dsn] = pool
				}
				pool := b.pgpool[dsn]

				allConstraints, err := b.getAllPostgresFkConstraintsFromMappings(ctx, pool, job.Mappings)
				if err != nil {
					return nil, err
				}
				td = dbschemas_postgres.GetPostgresTableDependencies(allConstraints)
			case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
				mysqlconfig := fkconnconfig.MysqlConfig
				if mysqlconfig == nil {
					return nil, errors.New("source connection is not a mysql config")
				}
				dsn, err := getMysqlDsn(mysqlconfig)
				if err != nil {
					return nil, err
				}
				if _, ok := b.mysqlpool[dsn]; !ok {
					pool, err := sql.Open("mysql", dsn)
					if err != nil {
						return nil, err
					}
					defer pool.Close()
					b.mysqlpool[dsn] = pool
				}
				pool := b.mysqlpool[dsn]

				allConstraints, err := b.getAllMysqlFkConstraintsFromMappings(ctx, pool, job.Mappings)
				if err != nil {
					return nil, err
				}
				td = dbschemas_mysql.GetMysqlTableDependencies(allConstraints)
			default:
				return nil, errors.New("unsupported fk connection")
			}

			for _, resp := range responses {
				dependsOn, ok := td[resp.Name]
				if ok {
					resp.DependsOn = dependsOn
				}
			}
		}

	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		sourceConnection, err := b.getConnectionById(ctx, jobSourceConfig.Postgres.ConnectionId)
		if err != nil {
			return nil, err
		}
		pgconfig := sourceConnection.ConnectionConfig.GetPgConfig()
		if pgconfig == nil {
			return nil, errors.New("source connection is not a postgres config")
		}
		dsn, err := getPgDsn(pgconfig)
		if err != nil {
			return nil, err
		}
		sqlOpts := jobSourceConfig.Postgres
		var sourceTableOpts map[string]*sqlSourceTableOptions
		if sqlOpts != nil {
			sourceTableOpts = groupPostgresSourceOptionsByTable(sqlOpts.Schemas)
		}

		sourceResponses, err := b.buildBenthosSqlSourceConfigReponses(ctx, groupedMappings, dsn, "postgres", sourceTableOpts)
		if err != nil {
			return nil, err
		}
		responses = append(responses, sourceResponses...)

		if _, ok := b.pgpool[dsn]; !ok {
			pool, err := pgxpool.New(ctx, dsn)
			if err != nil {
				return nil, err
			}
			defer pool.Close()
			b.pgpool[dsn] = pool
		}
		pool := b.pgpool[dsn]

		// validate job mappings align with sql connections
		dbschemas, err := b.pgquerier.GetDatabaseSchema(ctx, pool)
		if err != nil {
			return nil, err
		}
		groupedSchemas := dbschemas_postgres.GetUniqueSchemaColMappings(dbschemas)
		if !areMappingsSubsetOfSchemas(groupedSchemas, job.Mappings) {
			return nil, errors.New("job mappings are not equal to or a subset of the database schema found in the source connection")
		}
		if sqlOpts != nil && sqlOpts.HaltOnNewColumnAddition &&
			shouldHaltOnSchemaAddition(groupedSchemas, job.Mappings) {
			msg := "job mappings does not contain a column mapping for all " +
				"columns found in the source connection for the selected schemas and tables"
			return nil, errors.New(msg)
		}

		allConstraints, err := b.getAllPostgresFkConstraintsFromMappings(ctx, pool, job.Mappings)
		if err != nil {
			return nil, err
		}
		td := dbschemas_postgres.GetPostgresTableDependencies(allConstraints)

		for _, resp := range responses {
			dependsOn, ok := td[resp.Name]
			if ok {
				resp.DependsOn = dependsOn
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
		dsn, err := getMysqlDsn(mysqlconfig)
		if err != nil {
			return nil, err
		}

		sqlOpts := jobSourceConfig.Mysql
		var sourceTableOpts map[string]*sqlSourceTableOptions
		if sqlOpts != nil {
			sourceTableOpts = groupMysqlSourceOptionsByTable(sqlOpts.Schemas)
		}

		sourceResponses, err := b.buildBenthosSqlSourceConfigReponses(ctx, groupedMappings, dsn, "mysql", sourceTableOpts)
		if err != nil {
			return nil, err
		}
		responses = append(responses, sourceResponses...)

		if _, ok := b.mysqlpool[dsn]; !ok {
			pool, err := sql.Open("mysql", dsn)
			if err != nil {
				return nil, err
			}
			defer pool.Close()
			b.mysqlpool[dsn] = pool
		}
		pool := b.mysqlpool[dsn]

		// validate job mappings align with sql connections
		dbschemas, err := b.mysqlquerier.GetDatabaseSchema(ctx, pool)
		if err != nil {
			return nil, err
		}
		groupedSchemas := dbschemas_mysql.GetUniqueSchemaColMappings(dbschemas)
		if !areMappingsSubsetOfSchemas(groupedSchemas, job.Mappings) {
			return nil, errors.New("job mappings are not equal to or a subset of the database schema found in the source connection")
		}
		if sqlOpts != nil && sqlOpts.HaltOnNewColumnAddition &&
			shouldHaltOnSchemaAddition(groupedSchemas, job.Mappings) {
			msg := "job mappings does not contain a column mapping for all " +
				"columns found in the source connection for the selected schemas and tables"
			return nil, errors.New(msg)
		}

		allConstraints, err := b.getAllMysqlFkConstraintsFromMappings(ctx, pool, job.Mappings)
		if err != nil {
			return nil, err
		}
		td := dbschemas_mysql.GetMysqlTableDependencies(allConstraints)

		for _, resp := range responses {
			dependsOn, ok := td[resp.Name]
			if ok {
				resp.DependsOn = dependsOn
			}
		}
	default:
		return nil, errors.New("unsupported job source")
	}

	for _, destination := range job.Destinations {
		destinationConnection, err := b.getConnectionById(ctx, destination.ConnectionId)
		if err != nil {
			return nil, err
		}
		for _, resp := range responses {
			switch connection := destinationConnection.ConnectionConfig.Config.(type) {
			case *mgmtv1alpha1.ConnectionConfig_PgConfig:
				dsn, err := getPgDsn(connection.PgConfig)
				if err != nil {
					return nil, err
				}

				truncateBeforeInsert := false
				truncateCascade := false
				initSchema := false
				sqlOpts := destination.Options.GetPostgresOptions()
				if sqlOpts != nil {
					initSchema = sqlOpts.InitTableSchema
					if sqlOpts.TruncateTable != nil {
						truncateBeforeInsert = sqlOpts.TruncateTable.TruncateBeforeInsert
						truncateCascade = sqlOpts.TruncateTable.Cascade
					}
				}

				if resp.Config.Input.SqlSelect != nil {
					pool := b.pgpool[resp.Config.Input.SqlSelect.Dsn]
					// todo: make this more efficient to reduce amount of times we have to connect to the source database
					schema, table := splitTableKey(resp.Config.Input.SqlSelect.Table)
					initStmt, err := b.getInitStatementFromPostgres(
						ctx,
						pool,
						schema,
						table,
						&initStatementOpts{
							TruncateBeforeInsert: truncateBeforeInsert,
							TruncateCascade:      truncateCascade,
							InitSchema:           initSchema,
						},
					)
					if err != nil {
						return nil, err
					}
					tableKey := neosync_benthos.BuildBenthosTable(resp.tableSchema, resp.tableName)
					tm := groupedTableMapping[tableKey]
					if tm == nil {
						return nil, errors.New("unable to find table mapping for key")
					}
					colSourceMap := map[string]string{}
					for _, col := range tm.Mappings {
						colSourceMap[col.Column] = col.GetTransformer().Source
					}
					filteredCols := b.filterColsBySource(resp.Config.Input.SqlSelect.Columns, colSourceMap)
					logger.Info(fmt.Sprintf("sql batch count: %d", maxPgParamLimit/len(resp.Config.Input.SqlSelect.Columns)))
					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
						SqlRaw: &neosync_benthos.SqlRaw{

							Driver: "postgres",
							Dsn:    dsn,

							Query:         b.buildPostgresInsertQuery(resp.Config.Input.SqlSelect.Table, resp.Config.Input.SqlSelect.Columns, colSourceMap),
							ArgsMapping:   buildPlainInsertArgs(filteredCols),
							InitStatement: initStmt,

							ConnMaxIdle: 2,
							ConnMaxOpen: 2,

							Batching: &neosync_benthos.Batching{
								Period: "1s",
								// max allowed by postgres in a single batch
								Count: computeMaxPgBatchCount(len(resp.Config.Input.SqlSelect.Columns)),
							},
						},
					})
				} else if resp.Config.Input.Generate != nil {
					tableKey := neosync_benthos.BuildBenthosTable(resp.tableSchema, resp.tableName)
					tm := groupedTableMapping[tableKey]
					if tm == nil {
						return nil, errors.New("unable to find table mapping for key")
					}

					cols := buildPlainColumns(tm.Mappings)
					colSourceMap := map[string]string{}
					for _, col := range tm.Mappings {
						colSourceMap[col.Column] = col.GetTransformer().Source
					}
					// filters out default columns
					filteredCols := b.filterColsBySource(cols, colSourceMap)
					initStmt, err := b.getInitStatementFromPostgres(
						ctx,
						nil,
						resp.tableSchema,
						resp.tableName,
						&initStatementOpts{
							TruncateBeforeInsert: truncateBeforeInsert,
							TruncateCascade:      truncateCascade,
							InitSchema:           false, // todo
						},
					)
					if err != nil {
						return nil, err
					}

					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
						SqlRaw: &neosync_benthos.SqlRaw{
							Driver: "postgres",
							Dsn:    dsn,

							Query:       b.buildPostgresInsertQuery(tableKey, cols, colSourceMap),
							ArgsMapping: buildPlainInsertArgs(filteredCols),

							InitStatement: initStmt,

							ConnMaxIdle: 2,
							ConnMaxOpen: 2,

							Batching: &neosync_benthos.Batching{
								Period: "1s",
								// max allowed by postgres in a single batch
								Count: computeMaxPgBatchCount(len(cols)),
							},
						},
					})
				} else {
					return nil, errors.New("unable to build destination connection due to unsupported source connection")
				}
			case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
				dsn, err := getMysqlDsn(connection.MysqlConfig)
				if err != nil {
					return nil, err
				}

				truncateBeforeInsert := false
				initSchema := false
				sqlOpts := destination.Options.GetMysqlOptions()
				if sqlOpts != nil {
					initSchema = sqlOpts.InitTableSchema
					if sqlOpts.TruncateTable != nil {
						truncateBeforeInsert = sqlOpts.TruncateTable.TruncateBeforeInsert
					}
				}

				if resp.Config.Input.SqlSelect != nil {
					pool := b.mysqlpool[resp.Config.Input.SqlSelect.Dsn]
					// todo: make this more efficient to reduce amount of times we have to connect to the source database
					schema, table := splitTableKey(resp.Config.Input.SqlSelect.Table)
					initStmt, err := b.getInitStatementFromMysql(
						ctx,
						pool,
						schema,
						table,
						&initStatementOpts{
							TruncateBeforeInsert: truncateBeforeInsert,
							InitSchema:           initSchema,
						},
					)
					if err != nil {
						return nil, err
					}
					tableKey := neosync_benthos.BuildBenthosTable(resp.tableSchema, resp.tableName)
					tm := groupedTableMapping[tableKey]
					if tm == nil {
						return nil, errors.New("unable to find table mapping for key")
					}
					colSourceMap := map[string]string{}
					for _, col := range tm.Mappings {
						colSourceMap[col.Column] = col.GetTransformer().Source
					}
					filteredCols := b.filterColsBySource(resp.Config.Input.SqlSelect.Columns, colSourceMap)
					logger.Info(fmt.Sprintf("sql batch count: %d", maxPgParamLimit/len(resp.Config.Input.SqlSelect.Columns)))
					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
						SqlRaw: &neosync_benthos.SqlRaw{
							Driver: "mysql",
							Dsn:    dsn,

							Query:         b.buildMysqlInsertQuery(resp.Config.Input.SqlSelect.Table, resp.Config.Input.SqlSelect.Columns, colSourceMap),
							ArgsMapping:   buildPlainInsertArgs(filteredCols),
							InitStatement: initStmt,

							ConnMaxIdle: 2,
							ConnMaxOpen: 2,

							Batching: &neosync_benthos.Batching{
								Period: "1s",
								// max allowed by postgres in a single batch
								Count: computeMaxPgBatchCount(len(resp.Config.Input.SqlSelect.Columns)),
							},
						},
					})
				} else if resp.Config.Input.Generate != nil {
					tableKey := neosync_benthos.BuildBenthosTable(resp.tableSchema, resp.tableName)
					tm := groupedTableMapping[tableKey]
					if tm == nil {
						return nil, errors.New("unable to find table mapping for key")
					}

					cols := buildPlainColumns(tm.Mappings)
					colSourceMap := map[string]string{}
					for _, col := range tm.Mappings {
						colSourceMap[col.Column] = col.GetTransformer().Source
					}
					// filters out default columns
					filteredCols := b.filterColsBySource(cols, colSourceMap)
					initStmt, err := b.getInitStatementFromMysql(
						ctx,
						nil,
						resp.tableSchema,
						resp.tableName,
						&initStatementOpts{
							TruncateBeforeInsert: truncateBeforeInsert,
							TruncateCascade:      false,
							InitSchema:           false, // todo
						},
					)
					if err != nil {
						return nil, err
					}

					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
						SqlRaw: &neosync_benthos.SqlRaw{
							Driver: "mysql",
							Dsn:    dsn,

							Query:         b.buildMysqlInsertQuery(tableKey, cols, colSourceMap),
							ArgsMapping:   buildPlainInsertArgs(filteredCols),
							InitStatement: initStmt,

							ConnMaxIdle: 2,
							ConnMaxOpen: 2,

							Batching: &neosync_benthos.Batching{
								Period: "1s",
								// max allowed by postgres in a single batch
								Count: computeMaxPgBatchCount(len(cols)),
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
					req.JobId,
					req.WorkflowId,
					"activities",
					resp.Name, // may need to do more here
					"data",
					`${!count("files")}.txt.gz`,
				)

				// add job id here
				resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
					AwsS3: &neosync_benthos.AwsS3Insert{
						Bucket:      connection.AwsS3Config.Bucket,
						MaxInFlight: 64,
						Path:        fmt.Sprintf("/%s", strings.Join(s3pathpieces, "/")),
						Batching: &neosync_benthos.Batching{
							Count:  100,
							Period: "1s",
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

	return &GenerateBenthosConfigsResponse{
		BenthosConfigs: responses,
	}, nil
}

type generateSourceTableOptions struct {
	Count int
}

func (b *benthosBuilder) buildBenthosGenerateSourceConfigResponses(
	ctx context.Context,
	mappings []*TableMapping,
	sourceTableOpts map[string]*generateSourceTableOptions,
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

		mapping, err := b.buildProcessorMutation(ctx, tableMapping.Mappings)
		if err != nil {
			return nil, err
		}
		if mapping == "" {
			return nil, errors.New("unable to generate config mapping for table") // workshop this more
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
			DependsOn: []string{},

			tableSchema: tableMapping.Schema,
			tableName:   tableMapping.Table,
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

func (b *benthosBuilder) getAllPostgresFkConstraintsFromMappings(
	ctx context.Context,
	conn pg_queries.DBTX,
	mappings []*mgmtv1alpha1.JobMapping,
) ([]*pg_queries.GetForeignKeyConstraintsRow, error) {
	uniqueSchemas := getUniqueSchemasFromMappings(mappings)
	holder := make([][]*pg_queries.GetForeignKeyConstraintsRow, len(uniqueSchemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range uniqueSchemas {
		idx := idx
		schema := uniqueSchemas[idx]
		errgrp.Go(func() error {
			constraints, err := b.pgquerier.GetForeignKeyConstraints(errctx, conn, schema)
			if err != nil {
				return err
			}
			holder[idx] = constraints
			return nil
		})
	}

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	output := []*pg_queries.GetForeignKeyConstraintsRow{}
	for _, schemas := range holder {
		output = append(output, schemas...)
	}
	return output, nil
}

func (b *benthosBuilder) getAllMysqlFkConstraintsFromMappings(
	ctx context.Context,
	conn mysql_queries.DBTX,
	mappings []*mgmtv1alpha1.JobMapping,
) ([]*mysql_queries.GetForeignKeyConstraintsRow, error) {
	uniqueSchemas := getUniqueSchemasFromMappings(mappings)
	holder := make([][]*mysql_queries.GetForeignKeyConstraintsRow, len(uniqueSchemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range uniqueSchemas {
		idx := idx
		schema := uniqueSchemas[idx]
		errgrp.Go(func() error {
			constraints, err := b.mysqlquerier.GetForeignKeyConstraints(errctx, conn, schema)
			if err != nil {
				return err
			}
			holder[idx] = constraints
			return nil
		})
	}

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	output := []*mysql_queries.GetForeignKeyConstraintsRow{}
	for _, schemas := range holder {
		output = append(output, schemas...)
	}
	return output, nil
}

type initStatementOpts struct {
	TruncateBeforeInsert bool
	TruncateCascade      bool // only applied if truncatebeforeinsert is true
	InitSchema           bool
}

func (b *benthosBuilder) getInitStatementFromPostgres(
	ctx context.Context,
	conn pg_queries.DBTX,
	schema string,
	table string,
	opts *initStatementOpts,
) (string, error) {

	statements := []string{}
	if opts != nil && opts.InitSchema {
		stmt, err := dbschemas_postgres.GetTableCreateStatement(ctx, conn, b.pgquerier, schema, table)
		if err != nil {
			return "", err
		}
		statements = append(statements, stmt)
	}
	if opts != nil && opts.TruncateBeforeInsert {
		if opts.TruncateCascade {
			statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s.%s CASCADE;", schema, table))
		} else {
			statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s.%s;", schema, table))
		}
	}
	return strings.Join(statements, "\n"), nil
}

func (b *benthosBuilder) getInitStatementFromMysql(
	ctx context.Context,
	conn mysql_queries.DBTX,
	schema string,
	table string,
	opts *initStatementOpts,
) (string, error) {
	statements := []string{}
	if opts != nil && opts.InitSchema {
		stmt, err := dbschemas_mysql.GetTableCreateStatement(ctx, conn, &dbschemas_mysql.GetTableCreateStatementRequest{
			Schema: schema,
			Table:  table,
		})
		if err != nil {
			return "", err
		}
		statements = append(statements, stmt)
	}
	if opts != nil && opts.TruncateBeforeInsert {
		statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s.%s;", schema, table))
	}
	return strings.Join(statements, "\n"), nil
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
