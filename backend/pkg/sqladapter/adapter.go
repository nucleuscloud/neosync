package sqladapter

import (
	"context"
	"fmt"
	"log/slog"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type SqlDatabase interface {
	GetDatabaseSchema(ctx context.Context) ([]*DatabaseSchemaRow, error)
	GetAllForeignKeyConstraints(ctx context.Context, schemas []string) ([]*ForeignKeyConstraintsRow, error)
	GetAllPrimaryKeyConstraints(ctx context.Context, schemas []string) ([]*PrimaryKeyConstraintsRow, error)
}

type SqlAdapter struct {
	pgpool    map[string]pg_queries.DBTX
	pgquerier pg_queries.Querier

	mysqlpool    map[string]mysql_queries.DBTX
	mysqlquerier mysql_queries.Querier

	sqlconnector sqlconnect.SqlConnector
}

func NewSqlAdapter(
	pgpool map[string]pg_queries.DBTX,
	pgquerier pg_queries.Querier,
	mysqlpool map[string]mysql_queries.DBTX,
	mysqlquerier mysql_queries.Querier,
	sqlconnector sqlconnect.SqlConnector,
) *SqlAdapter {
	return &SqlAdapter{
		pgpool:       pgpool,
		pgquerier:    pgquerier,
		mysqlpool:    mysqlpool,
		mysqlquerier: mysqlquerier,
	}
}

type SqlConnection struct {
	db SqlDatabase
}

func (s *SqlAdapter) NewSqlDb(
	ctx context.Context,
	driver string,
	slogger *slog.Logger,
	connection *mgmtv1alpha1.Connection,
) (*SqlConnection, error) {
	var db SqlDatabase
	switch driver {
	case "postgres":
		if _, ok := s.pgpool[connection.Id]; !ok {
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
			defer pgconn.Close()
			s.pgpool[connection.Id] = pool
		}
		pool := s.pgpool[connection.Id]
		db = &PostgresAdapter{
			querier: s.pgquerier,
			pool:    pool,
		}
	case "mysql":
		if _, ok := s.mysqlpool[connection.Id]; !ok {
			conn, err := s.sqlconnector.NewDbFromConnectionConfig(connection.ConnectionConfig, shared.Ptr(uint32(5)), slogger)
			if err != nil {
				return nil, fmt.Errorf("unable to create new mysql pool from connection config: %w", err)
			}
			pool, err := conn.Open()
			if err != nil {
				return nil, fmt.Errorf("unable to open mysql connection: %w", err)
			}
			defer conn.Close()
			s.mysqlpool[connection.Id] = pool
		}
		pool := s.mysqlpool[connection.Id]
		db = &MysqlAdapter{
			querier: s.mysqlquerier,
			pool:    pool,
		}
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	return &SqlConnection{
		db: db,
	}, nil
}

type DatabaseSchemaRow struct {
	TableSchema            string
	TableName              string
	ColumnName             string
	DataType               string
	ColumnDefault          interface{}
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
