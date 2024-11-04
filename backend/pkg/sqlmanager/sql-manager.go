package sqlmanager

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	pgxslog "github.com/nucleuscloud/neosync/backend/internal/pgx-slog"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sqlmanager_mssql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mssql"
	sqlmanager_mysql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mysql"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
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
	pgpool    *sync.Map
	pgquerier pg_queries.Querier

	mysqlpool    *sync.Map
	mysqlquerier mysql_queries.Querier

	mssqlpool    *sync.Map
	mssqlquerier mssql_queries.Querier

	sqlconnector sqlconnect.SqlConnector
}

func NewSqlManager(
	pgpool *sync.Map,
	pgquerier pg_queries.Querier,
	mysqlpool *sync.Map,
	mysqlquerier mysql_queries.Querier,
	mssqlpool *sync.Map,
	mssqlquerier mssql_queries.Querier,
	sqlconnector sqlconnect.SqlConnector,
) *SqlManager {
	return &SqlManager{
		pgpool:       pgpool,
		pgquerier:    pgquerier,
		mysqlpool:    mysqlpool,
		mysqlquerier: mysqlquerier,
		mssqlpool:    mssqlpool,
		mssqlquerier: mssqlquerier,
		sqlconnector: sqlconnector,
	}
}

type SqlManagerClient interface {
	NewPooledSqlDb(
		ctx context.Context,
		slogger *slog.Logger,
		connection *mgmtv1alpha1.Connection,
	) (*SqlConnection, error)
	NewSqlDb(
		ctx context.Context,
		slogger *slog.Logger,
		connection *mgmtv1alpha1.Connection,
		connectionTimeout *int,
	) (*SqlConnection, error)
	NewSqlDbFromUrl(
		ctx context.Context,
		driver, connectionUrl string,
	) (*SqlConnection, error)
	NewSqlDbFromConnectionConfig(
		ctx context.Context,
		slogger *slog.Logger,
		connectionConfig *mgmtv1alpha1.ConnectionConfig,
		connectionTimeout *int,
	) (*SqlConnection, error)
}

var _ SqlManagerClient = &SqlManager{}

type SqlConnection struct {
	Db     SqlDatabase
	Driver string
}

func (s *SqlManager) NewPooledSqlDb(
	ctx context.Context,
	slogger *slog.Logger,
	connection *mgmtv1alpha1.Connection,
) (*SqlConnection, error) {
	var db SqlDatabase
	var driver string
	switch connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		var closer func()
		if _, ok := s.pgpool.Load(connection.Id); !ok {
			pgconn, err := s.sqlconnector.NewDbFromConnectionConfig(connection.GetConnectionConfig(), sqlmanager_shared.Ptr(uint32(5)), slogger)
			if err != nil {
				return nil, fmt.Errorf("unable to create new postgres pool from connection config: %w", err)
			}
			pool, err := pgconn.Open()
			if err != nil {
				return nil, fmt.Errorf("unable to open postgres connection: %w", err)
			}
			s.pgpool.Store(connection.Id, pool)
			closer = func() {
				if pgconn != nil {
					pgconn.Close()
					s.pgpool.Delete(connection.Id)
				}
			}
		}
		val, _ := s.pgpool.Load(connection.Id)
		pool, ok := val.(pg_queries.DBTX)
		if !ok {
			return nil, fmt.Errorf("pool found, but type assertion to pg_queries.DBTX failed")
		}
		db = sqlmanager_postgres.NewManager(s.pgquerier, pool, closer)
		driver = sqlmanager_shared.PostgresDriver
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		var closer func()
		if _, ok := s.mysqlpool.Load(connection.Id); !ok {
			conn, err := s.sqlconnector.NewDbFromConnectionConfig(connection.ConnectionConfig, sqlmanager_shared.Ptr(uint32(5)), slogger)
			if err != nil {
				return nil, fmt.Errorf("unable to create new mysql pool from connection config: %w", err)
			}
			pool, err := conn.Open()
			if err != nil {
				return nil, fmt.Errorf("unable to open mysql connection: %w", err)
			}
			s.mysqlpool.Store(connection.Id, pool)
			closer = func() {
				if conn != nil {
					err := conn.Close()
					if err != nil {
						slogger.Error(fmt.Errorf("failed to close connection: %w", err).Error())
					}
					s.mysqlpool.Delete(connection.Id)
				}
			}
		}
		val, _ := s.mysqlpool.Load(connection.Id)
		pool, ok := val.(mysql_queries.DBTX)
		if !ok {
			return nil, fmt.Errorf("pool found, but type assertion to mysql_queries.DBTX failed")
		}

		db = sqlmanager_mysql.NewManager(s.mysqlquerier, pool, closer)
		driver = sqlmanager_shared.MysqlDriver
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		var closer func()
		if _, ok := s.mssqlpool.Load(connection.Id); !ok {
			conn, err := s.sqlconnector.NewDbFromConnectionConfig(connection.ConnectionConfig, sqlmanager_shared.Ptr(uint32(5)), slogger)
			if err != nil {
				return nil, fmt.Errorf("unable to create new mssql pool from connection config: %w", err)
			}
			pool, err := conn.Open()
			if err != nil {
				return nil, fmt.Errorf("unable to open mssql connection: %w", err)
			}
			s.mssqlpool.Store(connection.Id, pool)
			closer = func() {
				if conn != nil {
					err := conn.Close()
					if err != nil {
						slogger.Error(fmt.Errorf("failed to close connection: %w", err).Error())
					}
					s.mssqlpool.Delete(connection.Id)
				}
			}
		}
		val, _ := s.mssqlpool.Load(connection.Id)
		pool, ok := val.(mysql_queries.DBTX)
		if !ok {
			return nil, fmt.Errorf("pool found, but type assertion to mssql_queries.DBTX failed")
		}

		db = sqlmanager_mssql.NewManager(s.mssqlquerier, pool, closer)
		driver = sqlmanager_shared.MssqlDriver
	default:
		return nil, fmt.Errorf("unsupported sql database connection: %T", connection.ConnectionConfig.Config)
	}

	return &SqlConnection{
		Db:     db,
		Driver: driver,
	}, nil
}

func (s *SqlManager) NewSqlDb(
	ctx context.Context,
	slogger *slog.Logger,
	connection *mgmtv1alpha1.Connection,
	connectionTimeout *int,
) (*SqlConnection, error) {
	return s.NewSqlDbFromConnectionConfig(ctx, slogger, connection.GetConnectionConfig(), connectionTimeout)
}

func (s *SqlManager) NewSqlDbFromConnectionConfig(
	ctx context.Context,
	slogger *slog.Logger,
	connectionConfig *mgmtv1alpha1.ConnectionConfig,
	connectionTimeout *int,
) (*SqlConnection, error) {
	connTimeout := sqlmanager_shared.Ptr(uint32(5))
	if connectionTimeout != nil {
		timeout := uint32(*connectionTimeout) //nolint:gosec // Ignoring for now
		connTimeout = &timeout
	}

	var db SqlDatabase
	var driver string
	switch connectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		pgconfig := connectionConfig.GetPgConfig()
		if pgconfig == nil {
			return nil, fmt.Errorf("source connection is not a postgres config")
		}
		pgconn, err := s.sqlconnector.NewDbFromConnectionConfig(connectionConfig, connTimeout, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to create new postgres pool from connection config: %w", err)
		}
		pool, err := pgconn.Open()
		if err != nil {
			return nil, fmt.Errorf("unable to open postgres connection: %w", err)
		}
		db = sqlmanager_postgres.NewManager(s.pgquerier, pool, func() {
			if pgconn != nil {
				pgconn.Close()
			}
		})
		driver = sqlmanager_shared.PostgresDriver
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		conn, err := s.sqlconnector.NewDbFromConnectionConfig(connectionConfig, connTimeout, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to create new mysql pool from connection config: %w", err)
		}
		pool, err := conn.Open()
		if err != nil {
			return nil, fmt.Errorf("unable to open mysql connection: %w", err)
		}
		db = sqlmanager_mysql.NewManager(s.mysqlquerier, pool, func() {
			if conn != nil {
				err := conn.Close()
				if err != nil {
					slogger.Error(fmt.Errorf("failed to close connection: %w", err).Error())
				}
			}
		})
		driver = sqlmanager_shared.MysqlDriver
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		conn, err := s.sqlconnector.NewDbFromConnectionConfig(connectionConfig, connTimeout, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to create new mssql client from connection config: %w", err)
		}
		pool, err := conn.Open()
		if err != nil {
			return nil, fmt.Errorf("unable to open mssql connection: %w", err)
		}
		db = sqlmanager_mssql.NewManager(s.mssqlquerier, pool, func() {
			if conn != nil {
				err := conn.Close()
				if err != nil {
					slogger.Error(fmt.Errorf("failed to close connection: %w", err).Error())
				}
			}
		})
		driver = sqlmanager_shared.MssqlDriver
	default:
		return nil, fmt.Errorf("unsupported sql database connection: %T", connectionConfig.Config)
	}

	return &SqlConnection{
		Db:     db,
		Driver: driver,
	}, nil
}

func (s *SqlManager) NewSqlDbFromUrl(
	ctx context.Context,
	driver, connectionUrl string,
) (*SqlConnection, error) {
	var db SqlDatabase
	switch driver {
	case sqlmanager_shared.PostgresDriver, "postgres":
		pgxconfig, err := pgx.ParseConfig(connectionUrl)
		if err != nil {
			return nil, err
		}
		pgxconfig.Tracer = &tracelog.TraceLog{
			Logger:   pgxslog.NewLogger(slog.Default(), pgxslog.GetShouldOmitArgs()),
			LogLevel: pgxslog.GetDatabaseLogLevel(),
		}
		pgxconfig.DefaultQueryExecMode = pgx.QueryExecModeExec
		pgconn, err := pgx.ConnectConfig(ctx, pgxconfig)
		if err != nil {
			return nil, err
		}
		sqldb := stdlib.OpenDB(*pgxconfig)
		db = sqlmanager_postgres.NewManager(s.pgquerier, sqldb, func() {
			if pgconn != nil {
				sqldb.Close()
			}
		})
		driver = sqlmanager_shared.PostgresDriver
	case sqlmanager_shared.MysqlDriver:
		if strings.Contains(connectionUrl, "?") {
			connectionUrl = fmt.Sprintf("%s&multiStatements=true&parseTime=true", connectionUrl)
		} else {
			connectionUrl = fmt.Sprintf("%s?multiStatements=true&parseTime=true", connectionUrl)
		}

		conn, err := sql.Open(sqlmanager_shared.MysqlDriver, connectionUrl)
		if err != nil {
			return nil, err
		}
		db = sqlmanager_mysql.NewManager(s.mysqlquerier, conn, func() {
			if conn != nil {
				conn.Close()
			}
		})
		driver = sqlmanager_shared.MysqlDriver
	case sqlmanager_shared.MssqlDriver:
		conn, err := sql.Open(sqlmanager_shared.MssqlDriver, connectionUrl)
		if err != nil {
			return nil, err
		}
		db = sqlmanager_mssql.NewManager(s.mssqlquerier, conn, func() {
			if conn != nil {
				conn.Close()
			}
		})
		driver = sqlmanager_shared.MssqlDriver
	default:
		return nil, fmt.Errorf("unsupported sql driver: %s", driver)
	}

	return &SqlConnection{
		Db:     db,
		Driver: driver,
	}, nil
}

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
