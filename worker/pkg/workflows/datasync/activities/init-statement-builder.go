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
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	dbschemas_mysql "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/mysql"
	dbschemas_postgres "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/postgres"

	"go.temporal.io/sdk/log"
)

type initStatementBuilder struct {
	pgpool    map[string]pg_queries.DBTX
	pgquerier pg_queries.Querier

	mysqlpool    map[string]mysql_queries.DBTX
	mysqlquerier mysql_queries.Querier

	jobclient  mgmtv1alpha1connect.JobServiceClient
	connclient mgmtv1alpha1connect.ConnectionServiceClient
}

func newInitStatementBuilder(
	pgpool map[string]pg_queries.DBTX,
	pgquerier pg_queries.Querier,

	mysqlpool map[string]mysql_queries.DBTX,
	mysqlquerier mysql_queries.Querier,

	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,

) *initStatementBuilder {
	return &initStatementBuilder{
		pgpool:       pgpool,
		pgquerier:    pgquerier,
		mysqlpool:    mysqlpool,
		mysqlquerier: mysqlquerier,
		jobclient:    jobclient,
		connclient:   connclient,
	}
}

func (b *initStatementBuilder) RunSqlInitTableStatements(
	ctx context.Context,
	req *RunSqlInitTableStatementsRequest,
	logger log.Logger,
) (*RunSqlInitTableStatementsResponse, error) {

	job, err := b.getJobById(ctx, req.JobId)
	if err != nil {
		return nil, err
	}

	var sourceDsn string
	var dependencyMap map[string][]string
	uniqueTables := getUniqueTablesFromMappings(job.Mappings)
	uniqueSchemas := getUniqueSchemasFromMappings(job.Mappings)

	switch jobSourceConfig := job.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		sourceConnection, err := b.getConnectionById(ctx, *jobSourceConfig.Generate.FkSourceConnectionId)
		if err != nil {
			return nil, err
		}
		switch connConfig := sourceConnection.ConnectionConfig.Config.(type) {
		case *mgmtv1alpha1.ConnectionConfig_PgConfig:
			dsn, err := getPgDsn(connConfig.PgConfig)
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
			sourceDsn = dsn

		case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
			dsn, err := getMysqlDsn(connConfig.MysqlConfig)
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
			sourceDsn = dsn
		default:
			return nil, errors.New("unsupported job source connection")
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

		sourceDsn = dsn

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

		if jobSourceConfig.Postgres != nil && jobSourceConfig.Postgres.HaltOnNewColumnAddition &&
			shouldHaltOnSchemaAddition(groupedSchemas, job.Mappings) {
			msg := "job mappings does not contain a column mapping for all " +
				"columns found in the source connection for the selected schemas and tables"
			return nil, errors.New(msg)
		}

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
		mysqlconfig := sourceConnection.ConnectionConfig.GetMysqlConfig()
		if mysqlconfig == nil {
			return nil, errors.New("source connection is not a mysql config")
		}
		dsn, err := getMysqlDsn(mysqlconfig)
		if err != nil {
			return nil, err
		}

		sourceDsn = dsn

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
		if jobSourceConfig.Mysql != nil && jobSourceConfig.Mysql.HaltOnNewColumnAddition &&
			shouldHaltOnSchemaAddition(groupedSchemas, job.Mappings) {
			msg := "job mappings does not contain a column mapping for all " +
				"columns found in the source connection for the selected schemas and tables"
			return nil, errors.New(msg)
		}
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

			if !truncateBeforeInsert && !truncateCascade && !initSchema {
				continue
			}
			if job.Source.Options.GetGenerate() != nil {
				initSchema = false
			}

			if job.Source.Options.GetPostgres() != nil || job.Source.Options.GetGenerate() != nil {

				sourcePool := b.pgpool[sourceDsn]
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

				sqlStatement := getOrderedSqlInitStatement(tableInitMap, dependencyMap)

				pool, err := pgxpool.New(ctx, dsn)
				if err != nil {
					return nil, err
				}
				_, err = pool.Exec(ctx, sqlStatement)
				if err != nil {
					return nil, err
				}
				pool.Close()

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
			if job.Source.Options.GetGenerate() != nil {
				initSchema = false
			}

			if job.Source.Options.GetMysql() != nil || job.Source.Options.GetGenerate() != nil {
				sourcePool := b.mysqlpool[sourceDsn]
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

				sqlStatements := getOrderedMysqlInitStatements(tableInitMap, dependencyMap)

				pool, err := sql.Open("mysql", dsn)
				if err != nil {
					return nil, err
				}
				for _, statement := range sqlStatements {
					_, err = pool.ExecContext(ctx, statement)
					if err != nil {
						return nil, err
					}
				}
				pool.Close()

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

// filters out tables where all col mappings are set to null
// returns unique list of tables
func getUniqueTablesFromMappings(mappings []*mgmtv1alpha1.JobMapping) map[string]struct{} {
	groupedMappings := map[string][]*mgmtv1alpha1.JobMapping{}
	for _, mapping := range mappings {
		tableName := dbschemas_utils.BuildTable(mapping.Schema, mapping.Table)
		_, ok := groupedMappings[tableName]
		if ok {
			groupedMappings[tableName] = append(groupedMappings[tableName], mapping)
		} else {
			groupedMappings[tableName] = []*mgmtv1alpha1.JobMapping{mapping}
		}
	}

	filteredTables := map[string]struct{}{}

	for table, mappings := range groupedMappings {
		if !areAllColsNull(mappings) {
			filteredTables[table] = struct{}{}
		}
	}
	return filteredTables
}

func getOrderedSqlInitStatement(tableInitMap map[string]string, dependencyMap map[string][]string) string {
	orderedStatements := []string{}
	seenTables := map[string]struct{}{}
	for table, statement := range tableInitMap {
		dep, ok := dependencyMap[table]
		if !ok || len(dep) == 0 {
			orderedStatements = append(orderedStatements, statement)
			seenTables[table] = struct{}{}
			delete(tableInitMap, table)
		}
	}

	maxCount := len(tableInitMap) * 2
	for len(tableInitMap) > 0 && maxCount > 0 {
		maxCount--
		for table, statement := range tableInitMap {
			deps := dependencyMap[table]
			if isReady(seenTables, deps, table) {
				orderedStatements = append(orderedStatements, statement)
				seenTables[table] = struct{}{}
				delete(tableInitMap, table)
			}
		}
	}

	return strings.Join(orderedStatements, "\n")
}

func getOrderedMysqlInitStatements(tableInitMap, dependencyMap map[string][]string) []string {
	orderedStatements := []string{}
	seenTables := map[string]struct{}{}
	for table, statements := range tableInitMap {
		dep, ok := dependencyMap[table]
		if !ok || len(dep) == 0 {
			orderedStatements = append(orderedStatements, statements...)
			seenTables[table] = struct{}{}
			delete(tableInitMap, table)
		}
	}

	maxCount := len(tableInitMap) * 2
	for len(tableInitMap) > 0 && maxCount > 0 {
		maxCount--
		for table, statements := range tableInitMap {
			deps := dependencyMap[table]
			if isReady(seenTables, deps, table) {
				orderedStatements = append(orderedStatements, statements...)
				seenTables[table] = struct{}{}
				delete(tableInitMap, table)
			}
		}
	}

	return orderedStatements
}

func isReady(seen map[string]struct{}, deps []string, table string) bool {
	for _, d := range deps {
		_, ok := seen[d]
		// allow self dependencies
		if !ok && d != table {
			return false
		}
	}
	return true
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
