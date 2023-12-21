package v1alpha1_connectiondataservice

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	anonymousUserId    = "00000000-0000-0000-0000-000000000000"
	mockAuthProvider   = "test-provider"
	mockUserId         = "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	mockAccountId      = "5629813e-1a35-4874-922c-9827d85f0378"
	mockConnectionName = "test-conn"
	mockConnectionId   = "884765c6-1708-488d-b03a-70a02b12c81e"
)

type ConnTypeMock string

const (
	MysqlMock    ConnTypeMock = "mysql"
	PostgresMock ConnTypeMock = "postgres"
	AwsS3Mock    ConnTypeMock = "awsS3"
)

func Test_GetConnectionSchema_Postgres(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	m.SqlMock.ExpectQuery(`SELECT (.+) FROM (.+) WHERE.*c.table_schema NOT IN \('pg_catalog', 'information_schema'\).*`).WillReturnRows(getRowsMock())

	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: mockConnectionId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Msg.GetSchemas()))
	assert.ElementsMatch(t, getDatabaseColumnsMock(), resp.Msg.Schemas)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_GetConnectionSchema_Mysql(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	m.SqlMock.ExpectQuery(`SELECT (.+) FROM (.+) WHERE.*c.table_schema NOT IN \('sys', 'performance_schema', 'mysql'\).*`).WillReturnRows(getRowsMock())

	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, MysqlMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: mockConnectionId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Msg.GetSchemas()))
	assert.ElementsMatch(t, getDatabaseColumnsMock(), resp.Msg.Schemas)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_GetConnectionSchema_NoRows(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	m.SqlMock.ExpectQuery(".*").WillReturnError(sql.ErrNoRows)

	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: mockConnectionId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, len(resp.Msg.GetSchemas()))
	assert.ElementsMatch(t, []*mgmtv1alpha1.DatabaseColumn{}, resp.Msg.Schemas)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_GetConnectionSchema_Error(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	m.SqlMock.ExpectQuery(".*").WillReturnError(errors.New("oh no"))

	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: mockConnectionId,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func getDatabaseColumnsMock() []*mgmtv1alpha1.DatabaseColumn {
	return []*mgmtv1alpha1.DatabaseColumn{
		{Schema: "schema1", Table: "table1", Column: "column1", DataType: "datatype1"},
		{Schema: "schema2", Table: "table2", Column: "column2", DataType: "datatype2"},
	}
}

func getRowsMock() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"table_schema", "table_name", "column_name", "data_type"}).
		AddRow("schema1", "table1", "column1", "datatype1").
		AddRow("schema2", "table2", "column2", "datatype2")
}

type mockConnector struct {
	db *sql.DB
}

func (mc *mockConnector) Open(driverName, dataSourceName string) (*sql.DB, error) {
	return mc.db, nil
}

type serviceMocks struct {
	Service                *Service
	DbtxMock               *nucleusdb.MockDBTX
	QuerierMock            *db_queries.MockQuerier
	UserAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient
	ConnectionServiceMock  *mgmtv1alpha1connect.MockConnectionServiceClient
	JobServiceMock         *mgmtv1alpha1connect.MockJobServiceClient
	SqlMock                sqlmock.Sqlmock
	SqlDbMock              *sql.DB
}

func createServiceMock(t *testing.T) *serviceMocks {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	mockConnectionService := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockJobService := mgmtv1alpha1connect.NewMockJobServiceClient(t)

	sqlDbMock, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(false))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	service := New(&Config{}, mockUserAccountService, mockConnectionService, mockJobService, &mockConnector{db: sqlDbMock})

	return &serviceMocks{
		Service:                service,
		DbtxMock:               mockDbtx,
		QuerierMock:            mockQuerier,
		UserAccountServiceMock: mockUserAccountService,
		ConnectionServiceMock:  mockConnectionService,
		JobServiceMock:         mockJobService,
		SqlMock:                sqlMock,
		SqlDbMock:              sqlDbMock,
	}
}

func mockIsUserInAccount(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) { // nolint
	userAccountServiceMock.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: isInAccount,
	}), nil)
}

//nolint:all
func getConnectionMock(accountId, name string, id string, connType ConnTypeMock) *mgmtv1alpha1.Connection {
	timestamp := timestamppb.New(time.Now())
	connection := &mgmtv1alpha1.Connection{
		AccountId:       accountId,
		Name:            name,
		Id:              id,
		CreatedByUserId: mockUserId,
		UpdatedByUserId: mockUserId,
		CreatedAt:       timestamp,
		UpdatedAt:       timestamp,
	}
	if connType == MysqlMock {
		connection.ConnectionConfig = &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
						Connection: &mgmtv1alpha1.MysqlConnection{
							Host:     "host",
							Port:     5432,
							Name:     "database",
							User:     "user",
							Pass:     "topsecret",
							Protocol: "tcp",
						},
					},
				},
			},
		}
	} else if connType == PostgresMock {
		sslMode := "disable"
		connection.ConnectionConfig = &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
						Connection: &mgmtv1alpha1.PostgresConnection{
							Host:    "host",
							Port:    5432,
							Name:    "database",
							User:    "user",
							Pass:    "topsecret",
							SslMode: &sslMode,
						},
					},
				},
			},
		}

	} else if connType == AwsS3Mock {
		connection.ConnectionConfig = &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{
				AwsS3Config: &mgmtv1alpha1.AwsS3ConnectionConfig{
					Bucket: "neosync",
				},
			},
		}
	}

	return connection
}
