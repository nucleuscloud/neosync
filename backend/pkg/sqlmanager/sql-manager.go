package sqlmanager

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

const (
	PostgresDriver = "postgres"
	MysqlDriver    = "mysql"
)

type SqlDatabase interface {
	GetDatabaseSchema(ctx context.Context) ([]*DatabaseSchemaRow, error)
	GetAllForeignKeyConstraints(ctx context.Context, schemas []string) ([]*ForeignKeyConstraintsRow, error)
	GetAllPrimaryKeyConstraints(ctx context.Context, schemas []string) ([]*PrimaryKeyConstraintsRow, error)
	ClosePool()
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

type SqlConnection struct {
	db     SqlDatabase
	Driver string
}

func (s *SqlManager) NewSqlDb(
	ctx context.Context,
	slogger *slog.Logger,
	connection *mgmtv1alpha1.Connection,
) (*SqlConnection, error) {
	var db SqlDatabase
	var driver string
	switch connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		adapter := &PostgresManager{
			querier: s.pgquerier,
		}
		if _, ok := s.pgpool.Load(connection.Id); !ok {
			pgconfig := connection.ConnectionConfig.GetPgConfig()
			if pgconfig == nil {
				return nil, fmt.Errorf("source connection (%s) is not a postgres config", connection.Id)
			}
			pgconn, err := s.sqlconnector.NewPgPoolFromConnectionConfig(pgconfig, shared.Ptr(uint32(5)), slogger)
			if err != nil {
				return nil, fmt.Errorf("unable to create new postgres pool from connection config: %w", err)
			}
			pool, err := pgconn.Open(ctx)
			if err != nil {
				return nil, fmt.Errorf("unable to open postgres connection: %w", err)
			}
			s.pgpool.Store(connection.Id, pool)
			adapter.closePool = func() {
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
		adapter.pool = pool
		db = adapter
		driver = PostgresDriver
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		adapter := &MySqlManager{
			querier: s.mysqlquerier,
		}
		if _, ok := s.mysqlpool.Load(connection.Id); !ok {
			conn, err := s.sqlconnector.NewDbFromConnectionConfig(connection.ConnectionConfig, shared.Ptr(uint32(5)), slogger)
			if err != nil {
				return nil, fmt.Errorf("unable to create new mysql pool from connection config: %w", err)
			}
			pool, err := conn.Open()
			if err != nil {
				return nil, fmt.Errorf("unable to open mysql connection: %w", err)
			}
			s.mysqlpool.Store(connection.Id, pool)
			adapter.closePool = func() {
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
		adapter.pool = pool
		db = adapter
		driver = MysqlDriver
	default:
		return nil, errors.New("unsupported sql database connection: %s")
	}

	return &SqlConnection{
		db:     db,
		Driver: driver,
	}, nil
}

type DatabaseSchemaRow struct {
	TableSchema            string
	TableName              string
	ColumnName             string
	DataType               string
	ColumnDefault          string
	IsNullable             string
	CharacterMaximumLength int32
	NumericPrecision       int32
	NumericScale           int32
	OrdinalPosition        int16
}

func (s *SqlConnection) GetDatabaseSchema(ctx context.Context) ([]*DatabaseSchemaRow, error) {
	return s.db.GetDatabaseSchema(ctx)
}

type ForeignKeyConstraintsRow struct {
	ConstraintName    string
	SchemaName        string
	TableName         string
	ColumnName        string
	IsNullable        string
	ForeignSchemaName string
	ForeignTableName  string
	ForeignColumnName string
}

func (s *SqlConnection) GetAllForeignKeyConstraints(ctx context.Context, schemas []string) ([]*ForeignKeyConstraintsRow, error) {
	return s.db.GetAllForeignKeyConstraints(ctx, schemas)
}

type PrimaryKeyConstraintsRow struct {
	SchemaName     string
	TableName      string
	ConstraintName string
	ColumnName     string
}

func (s *SqlConnection) GetAllPrimaryKeyConstraints(ctx context.Context, schemas []string) ([]*PrimaryKeyConstraintsRow, error) {
	return s.db.GetAllPrimaryKeyConstraints(ctx, schemas)
}

func (s *SqlConnection) ClosePool() {
	s.db.ClosePool()
}
