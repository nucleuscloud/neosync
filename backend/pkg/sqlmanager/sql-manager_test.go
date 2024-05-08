package sqlmanager

import (
	"context"
	"database/sql"
	slog "log/slog"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v4/pgxpool"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_NewPooledSqlDb_NewPostgresConnection(t *testing.T) {
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := &sync.Map{}
	pgcache.Store("123", pg_queries.NewMockDBTX(t))
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := &sync.Map{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockSqlmanager := NewSqlManager(pgcache, pgquerier, mysqlcache, mysqlquerier, mockSqlConnector)
	mockPool := sqlconnect.NewMockPgPoolContainer(t)

	mockSqlConnector.On("NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(mockPool, nil)
	mockPool.On("Open", mock.Anything).Return(pg_queries.NewMockDBTX(t), nil)

	slogger := &slog.Logger{}
	connection := &mgmtv1alpha1.Connection{
		Id: "456",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
						Url: "fake-stage-url",
					},
				},
			},
		},
	}
	mockPool.On("Close").Return(nil)

	result, err := mockSqlmanager.NewPooledSqlDb(context.Background(), slogger, connection)
	require.Equal(t, syncMapLength(pgcache), 2)
	result.Db.Close()
	require.Equal(t, syncMapLength(pgcache), 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Db)
	require.Equal(t, PostgresDriver, result.Driver)
	mockPool.AssertCalled(t, "Close")
}

func Test_NewPooledSqlDb_ExistingPostgresConnection(t *testing.T) {
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := &sync.Map{}
	pgcache.Store("123", pg_queries.NewMockDBTX(t))
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := &sync.Map{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockSqlmanager := NewSqlManager(pgcache, pgquerier, mysqlcache, mysqlquerier, mockSqlConnector)
	mockPool := sqlconnect.NewMockPgPoolContainer(t)

	mockSqlConnector.AssertNotCalled(t, "NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything)
	mockPool.AssertNotCalled(t, "Open", mock.Anything)

	slogger := &slog.Logger{}
	connection := &mgmtv1alpha1.Connection{
		Id: "123",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
						Url: "fake-stage-url",
					},
				},
			},
		},
	}

	result, err := mockSqlmanager.NewPooledSqlDb(context.Background(), slogger, connection)
	require.Equal(t, syncMapLength(pgcache), 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Db)
	require.Equal(t, PostgresDriver, result.Driver)
	require.Equal(t, syncMapLength(pgcache), 1)
}

func Test_NewPooledSqlDb_NewMysqlConnection(t *testing.T) {
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := &sync.Map{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := &sync.Map{}
	mysqlcache.Store("123", mysql_queries.NewMockDBTX(t))

	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockSqlmanager := NewSqlManager(pgcache, pgquerier, mysqlcache, mysqlquerier, mockSqlConnector)
	mockPool := sqlconnect.NewMockSqlDbContainer(t)
	sqlDbMock, _, err := sqlmock.New(sqlmock.MonitorPingsOption(false))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mockSqlConnector.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(mockPool, nil)
	mockPool.On("Open").Return(sqlDbMock, nil)
	mockPool.On("Close").Return(nil)

	slogger := &slog.Logger{}
	connection := &mgmtv1alpha1.Connection{
		Id: "456",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
						Url: "fake-stage-url",
					},
				},
			},
		},
	}

	result, err := mockSqlmanager.NewPooledSqlDb(context.Background(), slogger, connection)
	require.Equal(t, syncMapLength(mysqlcache), 2)
	result.Db.Close()
	require.Equal(t, syncMapLength(mysqlcache), 1)

	mockPool.AssertCalled(t, "Close")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Db)
	require.Equal(t, MysqlDriver, result.Driver)
}

func Test_NewPooledSqlDb_ExistingMysqlConnection(t *testing.T) {
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := &sync.Map{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := &sync.Map{}
	mysqlcache.Store("123", mysql_queries.NewMockDBTX(t))
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockSqlmanager := NewSqlManager(pgcache, pgquerier, mysqlcache, mysqlquerier, mockSqlConnector)
	mockPool := sqlconnect.NewMockSqlDbContainer(t)

	mockSqlConnector.AssertNotCalled(t, "NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything)
	mockPool.AssertNotCalled(t, "Open")

	slogger := &slog.Logger{}
	connection := &mgmtv1alpha1.Connection{
		Id: "123",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
						Url: "fake-stage-url",
					},
				},
			},
		},
	}

	result, err := mockSqlmanager.NewPooledSqlDb(context.Background(), slogger, connection)
	require.Equal(t, syncMapLength(mysqlcache), 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Db)
	require.Equal(t, MysqlDriver, result.Driver)
}

func Test_NewSqlDbFromConnectionConfig_PostgresConnection(t *testing.T) {
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := &sync.Map{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := &sync.Map{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockSqlmanager := NewSqlManager(pgcache, pgquerier, mysqlcache, mysqlquerier, mockSqlConnector)
	mockPool := sqlconnect.NewMockPgPoolContainer(t)

	mockSqlConnector.On("NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(mockPool, nil)
	mockPool.On("Open", mock.Anything).Return(pg_queries.NewMockDBTX(t), nil)

	slogger := &slog.Logger{}
	connection := &mgmtv1alpha1.Connection{
		Id: "456",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
						Url: "fake-stage-url",
					},
				},
			},
		},
	}
	mockPool.On("Close").Return(nil)

	result, err := mockSqlmanager.NewSqlDbFromConnectionConfig(context.Background(), slogger, connection.GetConnectionConfig(), Ptr(5))
	result.Db.Close()
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Db)
	require.Equal(t, PostgresDriver, result.Driver)
	mockPool.AssertCalled(t, "Close")
	require.Equal(t, syncMapLength(pgcache), 0)
}

func Test_NewSqlDbFromConnectionConfig_MysqlConnection(t *testing.T) {
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := &sync.Map{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := &sync.Map{}

	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockSqlmanager := NewSqlManager(pgcache, pgquerier, mysqlcache, mysqlquerier, mockSqlConnector)
	mockPool := sqlconnect.NewMockSqlDbContainer(t)
	sqlDbMock, _, err := sqlmock.New(sqlmock.MonitorPingsOption(false))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mockSqlConnector.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(mockPool, nil)
	mockPool.On("Open").Return(sqlDbMock, nil)
	mockPool.On("Close").Return(nil)

	slogger := &slog.Logger{}
	connection := &mgmtv1alpha1.Connection{
		Id: "456",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
						Url: "fake-stage-url",
					},
				},
			},
		},
	}

	result, err := mockSqlmanager.NewSqlDbFromConnectionConfig(context.Background(), slogger, connection.GetConnectionConfig(), Ptr(5))
	result.Db.Close()

	mockPool.AssertCalled(t, "Close")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Db)
	require.Equal(t, MysqlDriver, result.Driver)
	require.Equal(t, syncMapLength(mysqlcache), 0)
}

type MockPgPoolConnector struct {
	mock.Mock
}

func (m *MockPgPoolConnector) New(ctx context.Context, connectionUrl string) (*pgxpool.Pool, error) {
	args := m.Called(ctx, connectionUrl)
	return args.Get(0).(*pgxpool.Pool), args.Error(1)
}

type MockSqlConnector struct {
	mock.Mock
}

func (m *MockSqlConnector) Open(driverName, dataSourceName string) (*sql.DB, error) {
	args := m.Called(driverName, dataSourceName)
	return args.Get(0).(*sql.DB), args.Error(1)
}

func Test_NewSqlDbFromUrl(t *testing.T) {
	mockPgPoolConnector := &MockPgPoolConnector{}
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := &sync.Map{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := &sync.Map{}

	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockSqlmanager := NewSqlManager(pgcache, pgquerier, mysqlcache, mysqlquerier, mockSqlConnector)

	// Test for PostgreSQL
	postgresURL := "postgres://user:password@localhost/db"
	mockPgPool := pg_queries.NewMockDBTX(t) // Assuming pgxpool.Pool is an interface or you have a mockable implementation
	mockPgPoolConnector.On("New", mock.Anything, mock.Anything).Return(mockPgPool, nil)

	sqlConnection, err := mockSqlmanager.NewSqlDbFromUrl(context.Background(), PostgresDriver, postgresURL)
	require.NoError(t, err)
	require.NotNil(t, sqlConnection)
	require.Equal(t, PostgresDriver, sqlConnection.Driver)

	// Test for MySQL
	mysqlURL := "user:password@/dbname"

	sqlConnection, err = mockSqlmanager.NewSqlDbFromUrl(context.Background(), MysqlDriver, mysqlURL)
	require.NoError(t, err)
	require.NotNil(t, sqlConnection)
	require.Equal(t, MysqlDriver, sqlConnection.Driver)

	// Test for unsupported driver
	unsupportedURL := "unsupported://user:password@localhost/db"
	sqlConnection, err = mockSqlmanager.NewSqlDbFromUrl(context.Background(), "unsupported", unsupportedURL)
	require.Error(t, err)
	require.Nil(t, sqlConnection)
}

func syncMapLength(m *sync.Map) int {
	length := 0
	m.Range(func(_, _ interface{}) bool { //nolint:gofmt
		length++
		return true
	})
	return length
}
