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
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
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
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}

	var sourceConnectionId string
	var tableDependencies map[string]*dbschemas_utils.TableConstraints
	uniqueTables := shared.GetUniqueTablesFromMappings(job.Mappings)
	uniqueSchemas := shared.GetUniqueSchemasFromMappings(job.Mappings)

	switch jobSourceConfig := job.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		sourceConnection, err := b.getConnectionById(ctx, *jobSourceConfig.Generate.FkSourceConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get connection by fk source connection id: %w", err)
		}
		switch connConfig := sourceConnection.ConnectionConfig.Config.(type) {
		case *mgmtv1alpha1.ConnectionConfig_PgConfig:
			if _, ok := b.pgpool[sourceConnection.Id]; !ok {
				pgconn, err := b.sqlconnector.NewPgPoolFromConnectionConfig(connConfig.PgConfig, shared.Ptr(uint32(5)), slogger)
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
			sourceConnectionId = sourceConnection.Id

		case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
			if _, ok := b.mysqlpool[sourceConnection.Id]; !ok {
				conn, err := b.sqlconnector.NewDbFromConnectionConfig(sourceConnection.ConnectionConfig, shared.Ptr(uint32(5)), slogger)
				if err != nil {
					return nil, fmt.Errorf("unable to create new mysql pool from connection config: %w", err)
				}
				db, err := conn.Open()
				if err != nil {
					return nil, fmt.Errorf("unable to open mysql connection: %w", err)
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
			return nil, fmt.Errorf("unable to get connection by id (%s): %w", jobSourceConfig.Postgres.ConnectionId, err)
		}

		if _, ok := b.pgpool[sourceConnection.Id]; !ok {
			pgconn, err := b.sqlconnector.NewPgPoolFromConnectionConfig(sourceConnection.ConnectionConfig.GetPgConfig(), shared.Ptr(uint32(5)), slogger)
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
		sourceConnectionId = sourceConnection.Id

		allConstraints, err := dbschemas_postgres.GetAllPostgresFkConstraints(b.pgquerier, ctx, pool, uniqueSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve postgres foreign key constraints: %w", err)
		}
		tableDependencies = dbschemas_postgres.GetPostgresTableDependencies(allConstraints)

	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		sourceConnection, err := b.getConnectionById(ctx, jobSourceConfig.Mysql.ConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get connection by id (%s): %w", jobSourceConfig.Mysql.ConnectionId, err)
		}

		if _, ok := b.pgpool[sourceConnection.Id]; !ok {
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
		sourceConnectionId = sourceConnection.Id

		allConstraints, err := dbschemas_mysql.GetAllMysqlFkConstraints(b.mysqlquerier, ctx, pool, uniqueSchemas)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve mysql foreign key constraints")
		}
		tableDependencies = dbschemas_mysql.GetMysqlTableDependencies(allConstraints)

	default:
		return nil, fmt.Errorf("unsupported job source: %T", jobSourceConfig)
	}

	for _, destination := range job.Destinations {
		destinationConnection, err := b.getConnectionById(ctx, destination.ConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get destination connection by id (%s): %w", destination.ConnectionId, err)
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
				slogger.Info("skipping truncate and schema init as none were set to true")
				continue
			}

			if job.Source.Options.GetPostgres() != nil || job.Source.Options.GetGenerate() != nil {
				sourcePool := b.pgpool[sourceConnectionId]
				pgconn, err := b.sqlconnector.NewPgPoolFromConnectionConfig(connection.PgConfig, shared.Ptr(uint32(5)), slogger)
				if err != nil {
					return nil, fmt.Errorf("unable to create new postgres pool from connection config: %w", err)
				}
				pool, err := pgconn.Open(ctx)
				if err != nil {
					return nil, err
				}

				// create statements
				if initSchema {
					tableForeignDependencyMap := getForeignToPrimaryTableMap(tableDependencies, uniqueTables)
					orderedTables := tabledependency.GetTablesOrderedByDependency(tableForeignDependencyMap)
					tableCreateStmts := []string{}
					for _, table := range orderedTables {
						split := strings.Split(table[0], ".") // fix
						// todo: make this more efficient to reduce amount of times we have to connect to the source database
						initStmt, err := b.getCreateStatementFromPostgres(
							ctx,
							sourcePool,
							split[0],
							split[1],
						)
						if err != nil {
							return nil, fmt.Errorf("unable to build init statement for postgres: %w", err)
						}
						tableCreateStmts = append(tableCreateStmts, initStmt)
					}
					slogger.Info(fmt.Sprintf("executing %d sql statements that will initialize tables", len(tableCreateStmts)))
					_, err = pool.Exec(ctx, strings.Join(tableCreateStmts, "\n"))
					if err != nil {
						pgconn.Close()
						return nil, fmt.Errorf("unable to open postgres connection: %w", err)
					}
				}

				// truncate statements
				if truncateCascade {
					tableTruncateStmts := []string{}
					for table := range uniqueTables {
						split := strings.Split(table, ".")
						tableTruncateStmts = append(tableTruncateStmts, b.getTruncateCascadeStatementFromPostgres(split[0], split[1]))
					}
					slogger.Info(fmt.Sprintf("executing %d sql statements that will truncate cascade tables", len(tableTruncateStmts)))
					_, err = pool.Exec(ctx, strings.Join(tableTruncateStmts, "\n"))
					if err != nil {
						pgconn.Close()
						return nil, fmt.Errorf("unable to open postgres connection: %w", err)
					}
				} else if truncateBeforeInsert {
					tablePrimaryDependencyMap := getPrimaryToForeignTableMap(tableDependencies, uniqueTables)
					orderedTables := tabledependency.GetTablesOrderedByDependency(tablePrimaryDependencyMap)
					orderedTableTruncate := []string{}
					for _, table := range orderedTables {
						split := strings.Split(table[0], ".") // fix
						orderedTableTruncate = append(orderedTableTruncate, fmt.Sprintf(`%q.%q`, split[0], split[1]))
					}
					slogger.Info(fmt.Sprintf("executing %d sql statements that will truncate tables", len(orderedTableTruncate)))
					truncateStmt := b.getTruncateStatementFromPostgres(orderedTableTruncate)
					_, err = pool.Exec(ctx, truncateStmt)
					if err != nil {
						pgconn.Close()
						return nil, fmt.Errorf("unable to open postgres connection: %w", err)
					}
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
				slogger.Info("skipping truncate and schema init as none were set to true")
				continue
			}

			if job.Source.Options.GetMysql() != nil || job.Source.Options.GetGenerate() != nil {
				sourcePool := b.mysqlpool[sourceConnectionId]
				conn, err := b.sqlconnector.NewDbFromConnectionConfig(destinationConnection.ConnectionConfig, shared.Ptr(uint32(5)), slogger)
				if err != nil {
					return nil, fmt.Errorf("unable to create new mysql pool from connection config: %w", err)
				}
				pool, err := conn.Open()
				if err != nil {
					return nil, fmt.Errorf("unable to open mysql connection: %w", err)
				}

				// create statements
				if initSchema {
					tableForeignDependencyMap := getForeignToPrimaryTableMap(tableDependencies, uniqueTables)
					orderedTables := tabledependency.GetTablesOrderedByDependency(tableForeignDependencyMap)
					// todo: make this more efficient to reduce amount of times we have to connect to the source database
					tableCreateStmts := []string{}
					for _, table := range orderedTables {
						split := strings.Split(table[0], ".") // fix
						initStmt, err := b.getCreateStatementFromMysql(
							ctx,
							sourcePool,
							split[0],
							split[1],
						)
						if err != nil {
							return nil, fmt.Errorf("unable to build init statement for mysql: %w", err)
						}
						tableCreateStmts = append(tableCreateStmts, initStmt)
					}

					slogger.Info(fmt.Sprintf("executing %d sql statements that will initialize tables", len(tableCreateStmts)))
					for _, statement := range tableCreateStmts {
						_, err = pool.ExecContext(ctx, statement)
						if err != nil {
							if err := conn.Close(); err != nil {
								slogger.Error(err.Error())
							}
							return nil, err
						}
					}
				}

				// truncate statements
				if truncateBeforeInsert {
					tablePrimaryDependencyMap := getPrimaryToForeignTableMap(tableDependencies, uniqueTables)
					orderedTables := tabledependency.GetTablesOrderedByDependency(tablePrimaryDependencyMap)
					orderedTableTruncate := []string{}
					for _, table := range orderedTables {
						split := strings.Split(table[0], ".") // fix
						orderedTableTruncate = append(orderedTableTruncate, b.getTruncateStatementFromMysql(split[0], split[1]))
					}
					slogger.Info(fmt.Sprintf("executing %d sql statements that will truncate tables", len(orderedTableTruncate)))
					for _, statement := range orderedTableTruncate {
						_, err = pool.ExecContext(ctx, statement)
						if err != nil {
							if err := conn.Close(); err != nil {
								slogger.Error(err.Error())
							}
							return nil, err
						}
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

func (b *initStatementBuilder) getCreateStatementFromPostgres(
	ctx context.Context,
	conn pg_queries.DBTX,
	schema string,
	table string,
) (string, error) {
	stmt, err := dbschemas_postgres.GetTableCreateStatement(ctx, conn, b.pgquerier, schema, table)
	if err != nil {
		return "", err
	}
	return stmt, nil
}

func (b *initStatementBuilder) getCreateStatementFromMysql(
	ctx context.Context,
	conn mysql_queries.DBTX,
	schema string,
	table string,
) (string, error) {
	stmt, err := dbschemas_mysql.GetTableCreateStatement(ctx, conn, &dbschemas_mysql.GetTableCreateStatementRequest{
		Schema: schema,
		Table:  table,
	})
	if err != nil {
		return "", err
	}
	return stmt, nil
}

func (b *initStatementBuilder) getTruncateStatementFromPostgres(
	tables []string,
) string {
	return fmt.Sprintf("TRUNCATE TABLE %s;", strings.Join(tables, ", "))
}
func (b *initStatementBuilder) getTruncateCascadeStatementFromPostgres(
	schema string,
	table string,
) string {
	return fmt.Sprintf("TRUNCATE TABLE %q.%q CASCADE;", schema, table)
}

func (b *initStatementBuilder) getTruncateStatementFromMysql(
	schema string,
	table string,
) string {
	return fmt.Sprintf("TRUNCATE TABLE `%s`.`%s`;", schema, table)
}

// dependencyMap: {
// 	"public.countries": [
// 	 "public.regions"
// 	],
// 	"public.departments": [
// 	 "public.locations"
// 	],
// 	"public.dependents": [
// 	 "public.employees"
// 	],
// 	"public.employees": [
// 	 "public.departments",
// 	 "public.jobs",
// 	 "public.employees"
// 	],
// 	"public.locations": [
// 	 "public.countries"
// 	]
//  }

func getForeignToPrimaryTableMap(td map[string]*dbschemas_utils.TableConstraints, uniqueTables map[string]struct{}) map[string][]string {
	dpMap := map[string][]string{}
	for table := range uniqueTables {
		_, dpOk := dpMap[table]
		if !dpOk {
			dpMap[table] = []string{}
		}
		constraints, ok := td[table]
		if !ok {
			continue
		}
		for _, dep := range constraints.Constraints {
			dpMap[table] = append(dpMap[table], dep.ForeignKey.Table)
		}
	}
	return dpMap
}

// dependencyMap: {
// 	"public.regions": [
// 	 "public.countries"
// 	],
// 	"public.departments": [
// 	 "public.employees"
// 	],
// 	"public.dependents": [
// 	],
// 	"public.employees": [
// 	 "public.dependents",
// 	],
// 	"public.jobs": [
// 	 "public.employees",
// 	],
// 	"public.locations": [
// 	 "public.departments"
// 	]
//  }

func getPrimaryToForeignTableMap(td map[string]*dbschemas_utils.TableConstraints, uniqueTables map[string]struct{}) map[string][]string {
	dpMap := map[string][]string{}
	for table, constraints := range td {
		_, ok := uniqueTables[table]
		if !ok {
			continue
		}
		for _, dep := range constraints.Constraints {
			_, ok := dpMap[dep.ForeignKey.Table]
			if ok {
				dpMap[dep.ForeignKey.Table] = append(dpMap[dep.ForeignKey.Table], table)
			} else {
				dpMap[dep.ForeignKey.Table] = []string{table}
			}
		}
	}
	return dpMap
}
