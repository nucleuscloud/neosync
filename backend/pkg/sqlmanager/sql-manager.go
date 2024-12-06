package sqlmanager

import (
	"context"
	"fmt"
	"log/slog"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sqlmanager_mssql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mssql"
	sqlmanager_mysql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mysql"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/sqlprovider"
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
	config *sqlManagerConfig
}

type sqlManagerConfig struct {
	pgQuerier    pg_queries.Querier
	mysqlQuerier mysql_queries.Querier
	mssqlQuerier mssql_queries.Querier

	mgr connectionmanager.Interface[neosync_benthos_sql.SqlDbtx]
}
type SqlManagerOption func(*sqlManagerConfig)

func NewSqlManager(
	opts ...SqlManagerOption,
) *SqlManager {
	config := &sqlManagerConfig{
		pgQuerier:    pg_queries.New(),
		mysqlQuerier: mysql_queries.New(),
		mssqlQuerier: mssql_queries.New(),
		mgr:          connectionmanager.NewConnectionManager(sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{})),
	}
	for _, opt := range opts {
		opt(config)
	}
	return &SqlManager{
		config: config,
	}
}

func WithConnectionManager(manager connectionmanager.Interface[neosync_benthos_sql.SqlDbtx]) SqlManagerOption {
	return func(smc *sqlManagerConfig) {
		smc.mgr = manager
	}
}

// Initializes a default SQL-enabled connection manager, but allows for providing options
func WithConnectionManagerOpts(opts ...connectionmanager.ManagerOption) SqlManagerOption {
	return func(smc *sqlManagerConfig) {
		smc.mgr = connectionmanager.NewConnectionManager(sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{}), opts...)
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
		session connectionmanager.SessionInterface,
		connection connectionmanager.ConnectionInput,
		slogger *slog.Logger,
	) (*SqlConnection, error)
}

var _ SqlManagerClient = &SqlManager{}

func (s *SqlManager) NewSqlConnection(
	ctx context.Context,
	session connectionmanager.SessionInterface,
	connection connectionmanager.ConnectionInput,
	slogger *slog.Logger,
) (*SqlConnection, error) {
	connclient, err := s.config.mgr.GetConnection(session, connection, slogger)
	if err != nil {
		return nil, err
	}
	closer := func() {
		s.config.mgr.ReleaseSession(session, slogger)
	}

	switch connection.GetConnectionConfig().GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		db := sqlmanager_postgres.NewManager(s.config.pgQuerier, connclient, closer)
		return NewPostgresSqlConnection(db), nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		db := sqlmanager_mysql.NewManager(s.config.mysqlQuerier, connclient, closer)
		return NewMysqlSqlConnection(db), nil
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		db := sqlmanager_mssql.NewManager(s.config.mssqlQuerier, connclient, closer)
		return NewMssqlSqlConnection(db), nil
	default:
		closer()
		return nil, fmt.Errorf("unsupported sql database connection: %T", connection.GetConnectionConfig().GetConfig())
	}
}

func GetColumnOverrideAndResetProperties(driver string, cInfo *sqlmanager_shared.DatabaseSchemaRow) (needsOverride, needsReset bool, err error) {
	switch driver {
	case sqlmanager_shared.PostgresDriver:
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
