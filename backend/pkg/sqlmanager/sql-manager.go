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

	_ "github.com/go-sql-driver/mysql"
)

const (
	PostgresDriver = "postgres"
	MysqlDriver    = "mysql"
)

type BatchExecOpts struct {
	Prefix *string // this string will be added to the start of each statement
}

type ForeignKey struct {
	Table  string
	Column string
}
type ForeignConstraint struct {
	Column     string
	IsNullable bool
	ForeignKey *ForeignKey
}

type ReferenceKey struct {
	Table   string
	Columns []string
}
type ColumnConstraint struct {
	Columns     []string
	NotNullable []bool
	ForeignKey  *ReferenceKey
}

type ColumnInfo struct {
	OrdinalPosition        int32  // Specifies the sequence or order in which each column is defined within the table. Starts at 1 for the first column.
	ColumnDefault          string // Specifies the default value for a column, if any is set.
	IsNullable             bool   // Specifies if the column is nullable or not.
	DataType               string // Specifies the data type of the column, i.e., bool, varchar, int, etc.
	CharacterMaximumLength *int32 // Specifies the maximum allowable length of the column for character-based data types. For datatypes such as integers, boolean, dates etc. this is NULL.
	NumericPrecision       *int32 // Specifies the precision for numeric data types. It represents the TOTAL count of significant digits in the whole number, that is, the number of digits to BOTH sides of the decimal point. Null for non-numeric data types.
	NumericScale           *int32 // Specifies the scale of the column for numeric data types, specifically non-integers. It represents the number of digits to the RIGHT of the decimal point. Null for non-numeric data types and integers.
}

type SqlDatabase interface {
	GetDatabaseSchema(ctx context.Context) ([]*DatabaseSchemaRow, error)
	GetSchemaColumnMap(ctx context.Context) (map[string]map[string]*ColumnInfo, error) // ex: {public.users: { id: struct{}{}, created_at: struct{}{}}}
	GetForeignKeyConstraints(ctx context.Context, schemas []string) ([]*ForeignKeyConstraintsRow, error)
	GetForeignKeyConstraintsMap(ctx context.Context, schemas []string) (map[string][]*ForeignConstraint, error)
	GetForeignKeyReferencesMap(ctx context.Context, schemas []string) (map[string][]*ColumnConstraint, error)
	GetPrimaryKeyConstraints(ctx context.Context, schemas []string) ([]*PrimaryKey, error)
	GetPrimaryKeyConstraintsMap(ctx context.Context, schemas []string) (map[string][]string, error)
	GetUniqueConstraintsMap(ctx context.Context, schemas []string) (map[string][][]string, error)
	GetCreateTableStatement(ctx context.Context, schema, table string) (string, error)
	GetRolePermissionsMap(ctx context.Context, role string) (map[string][]string, error)
	BatchExec(ctx context.Context, batchSize int, statements []string, opts *BatchExecOpts) error
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
	GeneratedType          *string
}

type ForeignKeyConstraintsRow struct {
	ConstraintName    string
	SchemaName        string
	TableName         string
	ColumnName        string
	IsNullable        bool
	ForeignSchemaName string
	ForeignTableName  string
	ForeignColumnName string
}

type PrimaryKey struct {
	Schema  string
	Table   string
	Columns []string
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
		adapter := &PostgresManager{
			querier: s.pgquerier,
		}
		if _, ok := s.pgpool.Load(connection.Id); !ok {
			pgconfig := connection.ConnectionConfig.GetPgConfig()
			if pgconfig == nil {
				return nil, fmt.Errorf("source connection (%s) is not a postgres config", connection.Id)
			}
			pgconn, err := s.sqlconnector.NewPgPoolFromConnectionConfig(pgconfig, Ptr(uint32(5)), slogger)
			if err != nil {
				return nil, fmt.Errorf("unable to create new postgres pool from connection config: %w", err)
			}
			pool, err := pgconn.Open(ctx)
			if err != nil {
				return nil, fmt.Errorf("unable to open postgres connection: %w", err)
			}
			s.pgpool.Store(connection.Id, pool)
			adapter.close = func() {
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
		adapter := &MysqlManager{
			querier: s.mysqlquerier,
		}
		if _, ok := s.mysqlpool.Load(connection.Id); !ok {
			conn, err := s.sqlconnector.NewDbFromConnectionConfig(connection.ConnectionConfig, Ptr(uint32(5)), slogger)
			if err != nil {
				return nil, fmt.Errorf("unable to create new mysql pool from connection config: %w", err)
			}
			pool, err := conn.Open()
			if err != nil {
				return nil, fmt.Errorf("unable to open mysql connection: %w", err)
			}
			s.mysqlpool.Store(connection.Id, pool)
			adapter.close = func() {
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
	connTimeout := Ptr(uint32(5))
	if connectionTimeout != nil {
		timeout := uint32(*connectionTimeout)
		connTimeout = &timeout
	}

	var db SqlDatabase
	var driver string
	switch connectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		adapter := &PostgresManager{
			querier: s.pgquerier,
		}
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
		adapter.close = func() {
			if pgconn != nil {
				pgconn.Close()
			}
		}
		adapter.pool = pool
		db = adapter
		driver = PostgresDriver
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		adapter := &MysqlManager{
			querier: s.mysqlquerier,
		}
		conn, err := s.sqlconnector.NewDbFromConnectionConfig(connectionConfig, connTimeout, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to create new mysql pool from connection config: %w", err)
		}
		pool, err := conn.Open()
		if err != nil {
			return nil, fmt.Errorf("unable to open mysql connection: %w", err)
		}
		adapter.close = func() {
			if conn != nil {
				err := conn.Close()
				if err != nil {
					slogger.Error(fmt.Errorf("failed to close connection: %w", err).Error())
				}
			}
		}
		adapter.pool = pool
		db = adapter
		driver = MysqlDriver
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
	case PostgresDriver:
		adapter := &PostgresManager{
			querier: s.pgquerier,
		}
		pgconn, err := pgxpool.New(ctx, connectionUrl)
		if err != nil {
			return nil, err
		}
		adapter.close = func() {
			if pgconn != nil {
				pgconn.Close()
			}
		}
		adapter.pool = pgconn
		db = adapter
		driver = PostgresDriver
	case MysqlDriver:
		adapter := &MysqlManager{
			querier: s.mysqlquerier,
		}
		conn, err := sql.Open(MysqlDriver, connectionUrl)
		if err != nil {
			return nil, err
		}
		adapter.close = func() {
			if conn != nil {
				conn.Close()
			}
		}
		adapter.pool = conn
		db = adapter
		driver = MysqlDriver
	default:
		return nil, errors.New("unsupported sql database connection: %s")
	}

	return &SqlConnection{
		Db:     db,
		Driver: driver,
	}, nil
}
