package runsqlinittablestmts_activity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type initStatementBuilder struct {
	pgpool    map[string]pg_queries.DBTX
	pgquerier pg_queries.Querier

	mysqlpool    map[string]mysql_queries.DBTX
	mysqlquerier mysql_queries.Querier

	jobclient  mgmtv1alpha1connect.JobServiceClient
	connclient mgmtv1alpha1connect.ConnectionServiceClient

	sqlconnector sqlconnect.SqlConnector
}

func newInitStatementBuilder(
	pgpool map[string]pg_queries.DBTX,
	pgquerier pg_queries.Querier,

	mysqlpool map[string]mysql_queries.DBTX,
	mysqlquerier mysql_queries.Querier,

	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,

	sqlconnector sqlconnect.SqlConnector,

) *initStatementBuilder {
	return &initStatementBuilder{
		pgpool:       pgpool,
		pgquerier:    pgquerier,
		mysqlpool:    mysqlpool,
		mysqlquerier: mysqlquerier,
		jobclient:    jobclient,
		connclient:   connclient,
		sqlconnector: sqlconnector,
	}
}

func (b *initStatementBuilder) RunSqlInitTableStatements(
	ctx context.Context,
	req *RunSqlInitTableStatementsRequest,
	slogger *slog.Logger,
) (*RunSqlInitTableStatementsResponse, error) {
	job, err := b.getJobById(ctx, req.JobId)
	if err != nil {
		return nil, err
	}

	var sourceConnectionId string
	var dependencyMap map[string][]string
	uniqueTables := shared.GetUniqueTablesFromMappings(job.Mappings)
	uniqueSchemas := shared.GetUniqueSchemasFromMappings(job.Mappings)

	switch jobSourceConfig := job.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		sourceConnection, err := b.getConnectionById(ctx, *jobSourceConfig.Generate.FkSourceConnectionId)
		if err != nil {
			return nil, err
		}
		switch connConfig := sourceConnection.ConnectionConfig.Config.(type) {
		case *mgmtv1alpha1.ConnectionConfig_PgConfig:
			if _, ok := b.pgpool[sourceConnection.Id]; !ok {
				pgconn, err := b.sqlconnector.NewPgPoolFromConnectionConfig(connConfig.PgConfig, shared.Ptr(uint32(5)), slogger)
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
			sourceConnectionId = sourceConnection.Id

		case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
			if _, ok := b.mysqlpool[sourceConnection.Id]; !ok {
				conn, err := b.sqlconnector.NewDbFromConnectionConfig(sourceConnection.ConnectionConfig, shared.Ptr(uint32(5)), slogger)
				if err != nil {
					return nil, err
				}
				db, err := conn.Open()
				if err != nil {
					return nil, err
				}
				defer conn.Close()
				b.mysqlpool[sourceConnection.Id] = db
			}
			sourceConnectionId = sourceConnection.Id
		default:
			return nil, errors.New("unsupported job source connection")
		}

	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		sourceConnection, err := b.getConnectionById(ctx, jobSourceConfig.Postgres.ConnectionId)
		if err != nil {
			return nil, err
		}

		if _, ok := b.pgpool[sourceConnection.Id]; !ok {
			pgconn, err := b.sqlconnector.NewPgPoolFromConnectionConfig(sourceConnection.ConnectionConfig.GetPgConfig(), shared.Ptr(uint32(5)), slogger)
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
		sourceConnectionId = sourceConnection.Id

		allConstraints, err := dbschemas_postgres.GetAllPostgresFkConstraints(b.pgquerier, ctx, pool, uniqueSchemas)
		if err != nil {
			return nil, err
		}
		td := dbschemas_postgres.GetPostgresTableDependencies(allConstraints)
		dependencyMap = getDependencyMap(td, uniqueTables)

	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		sourceConnection, err := b.getConnectionById(ctx, jobSourceConfig.Mysql.ConnectionId)
		if err != nil {
			return nil, err
		}

		if _, ok := b.pgpool[sourceConnection.Id]; !ok {
			conn, err := b.sqlconnector.NewDbFromConnectionConfig(sourceConnection.ConnectionConfig, shared.Ptr(uint32(5)), slogger)
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
		sourceConnectionId = sourceConnection.Id

		allConstraints, err := dbschemas_mysql.GetAllMysqlFkConstraints(b.mysqlquerier, ctx, pool, uniqueSchemas)
		if err != nil {
			return nil, err
		}
		td := dbschemas_mysql.GetMysqlTableDependencies(allConstraints)
		dependencyMap = getDependencyMap(td, uniqueTables)

	default:
		return nil, errors.New("unsupported job source")
	}

	for _, destination := range job.Destinations {
		destinationConnection, err := b.getConnectionById(ctx, destination.ConnectionId)
		if err != nil {
			return nil, err
		}
		switch connection := destinationConnection.ConnectionConfig.Config.(type) {
		case *mgmtv1alpha1.ConnectionConfig_PgConfig:
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

			if job.Source.Options.GetGenerate() != nil {
				initSchema = false
			}

			if !truncateBeforeInsert && !truncateCascade && !initSchema {
				continue
			}

			if job.Source.Options.GetPostgres() != nil || job.Source.Options.GetGenerate() != nil {
				sourcePool := b.pgpool[sourceConnectionId]
				tableInitMap := map[string]string{}
				for table := range uniqueTables {
					split := strings.Split(table, ".")
					// todo: make this more efficient to reduce amount of times we have to connect to the source database
					initStmt, err := b.getInitStatementFromPostgres(
						ctx,
						sourcePool,
						split[0],
						split[1],
						&initStatementOpts{
							TruncateBeforeInsert: truncateBeforeInsert,
							TruncateCascade:      truncateCascade,
							InitSchema:           initSchema,
						},
					)
					if err != nil {
						return nil, err
					}
					tableInitMap[table] = initStmt
				}

				sqlStatement := dbschemas_postgres.GetOrderedPostgresInitStatements(tableInitMap, dependencyMap)

				pgconn, err := b.sqlconnector.NewPgPoolFromConnectionConfig(connection.PgConfig, shared.Ptr(uint32(5)), slogger)
				if err != nil {
					return nil, err
				}
				pool, err := pgconn.Open(ctx)
				if err != nil {
					return nil, err
				}
				_, err = pool.Exec(ctx, strings.Join(sqlStatement, "\n"))
				if err != nil {
					pgconn.Close()
					return nil, err
				}
				pgconn.Close()
			} else {
				return nil, errors.New("unable to build destination connection due to unsupported source connection")
			}
		case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
			truncateBeforeInsert := false
			initSchema := false
			sqlOpts := destination.Options.GetMysqlOptions()
			if sqlOpts != nil {
				initSchema = sqlOpts.InitTableSchema
				if sqlOpts.TruncateTable != nil {
					truncateBeforeInsert = sqlOpts.TruncateTable.TruncateBeforeInsert
				}
			}
			if job.Source.Options.GetGenerate() != nil {
				initSchema = false
			}

			if !truncateBeforeInsert && !initSchema {
				continue
			}

			if job.Source.Options.GetMysql() != nil || job.Source.Options.GetGenerate() != nil {
				sourcePool := b.mysqlpool[sourceConnectionId]
				// todo: make this more efficient to reduce amount of times we have to connect to the source database
				tableInitMap := map[string][]string{}
				for table := range uniqueTables {
					split := strings.Split(table, ".")
					initStmt, err := b.getInitStatementFromMysql(
						ctx,
						sourcePool,
						split[0],
						split[1],
						&initStatementOpts{
							TruncateBeforeInsert: truncateBeforeInsert,
							InitSchema:           initSchema,
						},
					)
					if err != nil {
						return nil, err
					}
					tableInitMap[table] = initStmt
				}

				sqlStatements := dbschemas_mysql.GetOrderedMysqlInitStatements(tableInitMap, dependencyMap)
				conn, err := b.sqlconnector.NewDbFromConnectionConfig(destinationConnection.ConnectionConfig, shared.Ptr(uint32(5)), slogger)
				if err != nil {
					return nil, err
				}
				pool, err := conn.Open()
				if err != nil {
					return nil, err
				}

				for _, statement := range sqlStatements {
					_, err = pool.ExecContext(ctx, statement)
					if err != nil {
						if err := conn.Close(); err != nil {
							slogger.Error(err.Error())
						}
						return nil, err
					}
				}
				if err := conn.Close(); err != nil {
					slogger.Error(err.Error())
				}
			} else {
				return nil, errors.New("unable to build destination connection due to unsupported source connection")
			}

		case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
			// nothing to do here
		default:
			return nil, fmt.Errorf("unsupported destination connection config")
		}
	}

	return &RunSqlInitTableStatementsResponse{}, nil
}

func (b *initStatementBuilder) getJobById(
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

func (b *initStatementBuilder) getConnectionById(
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

type initStatementOpts struct {
	TruncateBeforeInsert bool
	TruncateCascade      bool // only applied if truncatebeforeinsert is true
	InitSchema           bool
}

func (b *initStatementBuilder) getInitStatementFromPostgres(
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

func (b *initStatementBuilder) getInitStatementFromMysql(
	ctx context.Context,
	conn mysql_queries.DBTX,
	schema string,
	table string,
	opts *initStatementOpts,
) ([]string, error) {
	statements := []string{}
	if opts != nil && opts.InitSchema {
		stmt, err := dbschemas_mysql.GetTableCreateStatement(ctx, conn, &dbschemas_mysql.GetTableCreateStatementRequest{
			Schema: schema,
			Table:  table,
		})
		if err != nil {
			return []string{}, err
		}
		statements = append(statements, stmt)
	}
	if opts != nil && opts.TruncateBeforeInsert {
		statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s.%s;", schema, table))
	}
	return statements, nil
}

func getDependencyMap(td map[string]*dbschemas_utils.TableConstraints, uniqueTables map[string]struct{}) map[string][]string {
	dpMap := map[string][]string{}
	for table, constraints := range td {
		_, ok := uniqueTables[table]
		if !ok {
			continue
		}
		for _, dep := range constraints.Constraints {
			_, ok := dpMap[table]
			if ok {
				dpMap[table] = append(dpMap[table], dep.ForeignKey.Table)
			} else {
				dpMap[table] = []string{dep.ForeignKey.Table}
			}
		}
	}
	return dpMap
}
