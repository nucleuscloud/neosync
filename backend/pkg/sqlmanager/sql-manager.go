package sqlmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sqlmanager_mysql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mysql"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"

	_ "github.com/go-sql-driver/mysql"
)

type SqlDatabase interface {
	GetDatabaseSchema(ctx context.Context) ([]*sqlmanager_shared.DatabaseSchemaRow, error)
	GetSchemaColumnMap(ctx context.Context) (map[string]map[string]*sqlmanager_shared.ColumnInfo, error) // ex: {public.users: { id: struct{}{}, created_at: struct{}{}}}
	GetTableConstraintsBySchema(ctx context.Context, schemas []string) (*sqlmanager_shared.TableConstraints, error)
	GetForeignKeyConstraints(ctx context.Context, schemas []string) ([]*sqlmanager_shared.ForeignKeyConstraintsRow, error)
	GetForeignKeyConstraintsMap(ctx context.Context, schemas []string) (map[string][]*sqlmanager_shared.ForeignConstraint, error)
	GetPrimaryKeyConstraints(ctx context.Context, schemas []string) ([]*sqlmanager_shared.PrimaryKey, error)
	GetPrimaryKeyConstraintsMap(ctx context.Context, schemas []string) (map[string][]string, error)
	GetUniqueConstraintsMap(ctx context.Context, schemas []string) (map[string][][]string, error)
	GetCreateTableStatement(ctx context.Context, schema, table string) (string, error)
	GetTableInitStatements(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableInitStatement, error)
	GetRolePermissionsMap(ctx context.Context, role string) (map[string][]string, error)
	GetTableRowCount(ctx context.Context, schema, table string, whereClause *string) (int64, error)
	GetSchemaTableDataTypes(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) (*sqlmanager_shared.SchemaTableDataTypeResponse, error)
	GetSchemaTableTriggers(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableTrigger, error)
	BatchExec(ctx context.Context, batchSize int, statements []string, opts *sqlmanager_shared.BatchExecOpts) error
	Exec(ctx context.Context, statement string) error
	Close()
}

type SqlManager struct {
	pgpool    *sync.Map
	pgquerier pg_queries.Querier

	mysqlpool    *sync.Map
	mysqlquerier mysql_queries.Querier

	sqlconnector sqlconnect.SqlConnector
}

func NewSqlManager(
	pgpool *sync.Map,
	pgquerier pg_queries.Querier,
	mysqlpool *sync.Map,
	mysqlquerier mysql_queries.Querier,
	sqlconnector sqlconnect.SqlConnector,
) *SqlManager {
	return &SqlManager{
		pgpool:       pgpool,
		pgquerier:    pgquerier,
		mysqlpool:    mysqlpool,
		mysqlquerier: mysqlquerier,
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
			pgconfig := connection.ConnectionConfig.GetPgConfig()
			if pgconfig == nil {
				return nil, fmt.Errorf("source connection (%s) is not a postgres config", connection.Id)
			}
			pgconn, err := s.sqlconnector.NewPgPoolFromConnectionConfig(pgconfig, sqlmanager_shared.Ptr(uint32(5)), slogger)
			if err != nil {
				return nil, fmt.Errorf("unable to create new postgres pool from connection config: %w", err)
			}
			pool, err := pgconn.Open(ctx)
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
	default:
		return nil, errors.New("unsupported sql database connection: %s")
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
		timeout := uint32(*connectionTimeout)
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
		pgconn, err := s.sqlconnector.NewPgPoolFromConnectionConfig(pgconfig, connTimeout, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to create new postgres pool from connection config: %w", err)
		}
		pool, err := pgconn.Open(ctx)
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
	default:
		return nil, errors.New("unsupported sql database connection: %s")
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
	case sqlmanager_shared.PostgresDriver:
		pgconn, err := pgxpool.New(ctx, connectionUrl)
		if err != nil {
			return nil, err
		}
		db = sqlmanager_postgres.NewManager(s.pgquerier, pgconn, func() {
			if pgconn != nil {
				pgconn.Close()
			}
		})
		driver = sqlmanager_shared.PostgresDriver
	case sqlmanager_shared.MysqlDriver:
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
	default:
		return nil, errors.New("unsupported sql database connection: %s")
	}

	return &SqlConnection{
		Db:     db,
		Driver: driver,
	}, nil
}
