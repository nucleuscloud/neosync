package sqlmanager

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	sqlmanager_mssql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mssql"
	sqlmanager_mysql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mysql"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
)

type SqlDatabase interface {
	GetDatabaseSchema(ctx context.Context) ([]*sqlmanager_shared.DatabaseSchemaRow, error)
	GetSchemaColumnMap(ctx context.Context) (map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow, error) // ex: {public.users: { id: struct{}{}, created_at: struct{}{}}}
	GetTableConstraintsBySchema(ctx context.Context, schemas []string) (*sqlmanager_shared.TableConstraints, error)
	GetCreateTableStatement(ctx context.Context, schema, table string) (string, error)
	GetTableInitStatements(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableInitStatement, error)
	GetRolePermissionsMap(ctx context.Context) (map[string][]string, error)
	GetTableRowCount(ctx context.Context, schema, table string, whereClause *string) (int64, error)
	GetSchemaTableDataTypes(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) (*sqlmanager_shared.SchemaTableDataTypeResponse, error)
	GetSchemaTableTriggers(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableTrigger, error)
	GetSchemaInitStatements(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.InitSchemaStatements, error)
	GetSequencesByTables(ctx context.Context, schema string, tables []string) ([]*sqlmanager_shared.DataType, error)
	BatchExec(ctx context.Context, batchSize int, statements []string, opts *sqlmanager_shared.BatchExecOpts) error
	Exec(ctx context.Context, statement string) error
	Close()
}

type SqlManager struct {
	connectionManager connectionmanager.Interface[neosync_benthos_sql.SqlDbtx]
	config            *sqlManagerConfig
}

type sqlManagerConfig struct {
	pgQuerier    pg_queries.Querier
	mysqlQuerier mysql_queries.Querier
	mssqlQuerier mssql_queries.Querier
}
type SqlManagerOption func(*sqlManagerConfig)

func NewSqlManager(
	connectionManager connectionmanager.Interface[neosync_benthos_sql.SqlDbtx],
	opts ...SqlManagerOption,
) *SqlManager {
	config := &sqlManagerConfig{
		pgQuerier:    pg_queries.New(),
		mysqlQuerier: mysql_queries.New(),
		mssqlQuerier: mssql_queries.New(),
	}
	for _, opt := range opts {
		opt(config)
	}
	return &SqlManager{
		connectionManager: connectionManager,
		config:            config,
	}
}

func WithPostgresQuerier(querier pg_queries.Querier) SqlManagerOption {
	return func(smc *sqlManagerConfig) {
		smc.pgQuerier = querier
	}
}
func WithMysqlQuerier(querier mysql_queries.Querier) SqlManagerOption {
	return func(smc *sqlManagerConfig) {
		smc.mysqlQuerier = querier
	}
}
func WithMssqlQuerier(querier mssql_queries.Querier) SqlManagerOption {
	return func(smc *sqlManagerConfig) {
		smc.mssqlQuerier = querier
	}
}

type SqlManagerClient interface {
	NewSqlConnection(
		ctx context.Context,
		connection connectionmanager.ConnectionInput,
		slogger *slog.Logger,
	) (*SqlConnection, error)
}

var _ SqlManagerClient = &SqlManager{}

type SqlConnection struct {
	database SqlDatabase
	driver   string
}

func (s *SqlConnection) Db() SqlDatabase {
	return s.database
}
func (s *SqlConnection) Driver() string {
	return s.driver
}
func newSqlConnection(database SqlDatabase, driver string) *SqlConnection {
	return &SqlConnection{database: database, driver: driver}
}

func (s *SqlManager) NewSqlConnection(
	ctx context.Context,
	connection connectionmanager.ConnectionInput,
	slogger *slog.Logger,
) (*SqlConnection, error) {
	var db SqlDatabase
	var driver string
	session := uuid.NewString()
	connclient, err := s.connectionManager.GetConnection(session, connection, slogger)
	if err != nil {
		return nil, err
	}
	closer := func() {
		s.connectionManager.ReleaseSession(session)
	}

	switch connection.GetConnectionConfig().GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		db = sqlmanager_postgres.NewManager(s.config.pgQuerier, connclient, closer)
		driver = sqlmanager_shared.PostgresDriver
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		db = sqlmanager_mysql.NewManager(s.config.mysqlQuerier, connclient, closer)
		driver = sqlmanager_shared.MysqlDriver
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		db = sqlmanager_mssql.NewManager(s.config.mssqlQuerier, connclient, closer)
		driver = sqlmanager_shared.MssqlDriver
	default:
		return nil, fmt.Errorf("unsupported sql database connection: %T", connection.GetConnectionConfig().GetConfig())
	}

	return newSqlConnection(db, driver), nil
}

// func (s *SqlManager) NewSqlDbFromUrl(
// 	ctx context.Context,
// 	driver, connectionUrl string,
// ) (*SqlConnection, error) {
// 	var db SqlDatabase
// 	switch driver {
// 	case sqlmanager_shared.PostgresDriver, "postgres":
// 		pgxconfig, err := pgx.ParseConfig(connectionUrl)
// 		if err != nil {
// 			return nil, err
// 		}
// 		pgxconfig.Tracer = &tracelog.TraceLog{
// 			Logger:   pgxslog.NewLogger(slog.Default(), pgxslog.GetShouldOmitArgs()),
// 			LogLevel: pgxslog.GetDatabaseLogLevel(),
// 		}
// 		pgxconfig.DefaultQueryExecMode = pgx.QueryExecModeExec
// 		pgconn, err := pgx.ConnectConfig(ctx, pgxconfig)
// 		if err != nil {
// 			return nil, err
// 		}
// 		sqldb := stdlib.OpenDB(*pgxconfig)
// 		db = sqlmanager_postgres.NewManager(s.pgquerier, sqldb, func() {
// 			if pgconn != nil {
// 				sqldb.Close()
// 			}
// 		})
// 		driver = sqlmanager_shared.PostgresDriver
// 	case sqlmanager_shared.MysqlDriver:
// 		if strings.Contains(connectionUrl, "?") {
// 			connectionUrl = fmt.Sprintf("%s&multiStatements=true&parseTime=true", connectionUrl)
// 		} else {
// 			connectionUrl = fmt.Sprintf("%s?multiStatements=true&parseTime=true", connectionUrl)
// 		}

// 		conn, err := sql.Open(sqlmanager_shared.MysqlDriver, connectionUrl)
// 		if err != nil {
// 			return nil, err
// 		}
// 		db = sqlmanager_mysql.NewManager(s.mysqlquerier, conn, func() {
// 			if conn != nil {
// 				conn.Close()
// 			}
// 		})
// 		driver = sqlmanager_shared.MysqlDriver
// 	case sqlmanager_shared.MssqlDriver:
// 		conn, err := sql.Open(sqlmanager_shared.MssqlDriver, connectionUrl)
// 		if err != nil {
// 			return nil, err
// 		}
// 		db = sqlmanager_mssql.NewManager(s.mssqlquerier, conn, func() {
// 			if conn != nil {
// 				conn.Close()
// 			}
// 		})
// 		driver = sqlmanager_shared.MssqlDriver
// 	default:
// 		return nil, fmt.Errorf("unsupported sql driver: %s", driver)
// 	}

// 	return &SqlConnection{
// 		Db:     db,
// 		Driver: driver,
// 	}, nil
// }

func GetColumnOverrideAndResetProperties(driver string, cInfo *sqlmanager_shared.DatabaseSchemaRow) (needsOverride, needsReset bool, err error) {
	switch driver {
	case sqlmanager_shared.PostgresDriver, "postgres":
		needsOverride, needsReset := sqlmanager_postgres.GetPostgresColumnOverrideAndResetProperties(cInfo)
		return needsOverride, needsReset, nil
	case sqlmanager_shared.MysqlDriver:
		needsOverride, needsReset := sqlmanager_mysql.GetMysqlColumnOverrideAndResetProperties(cInfo)
		return needsOverride, needsReset, nil
	case sqlmanager_shared.MssqlDriver:
		needsOverride, needsReset := sqlmanager_mssql.GetMssqlColumnOverrideAndResetProperties(cInfo)
		return needsOverride, needsReset, nil
	default:
		return false, false, fmt.Errorf("unsupported sql driver: %s", driver)
	}
}
