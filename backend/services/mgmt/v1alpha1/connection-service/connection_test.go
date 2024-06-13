package v1alpha1_connectionservice

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"

	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
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
)

// CheckConnectionConfig
func Test_CheckConnectionConfig_Postgres(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	pool, _ := pgxpool.New(context.Background(), "")
	m.PgPoolContainerMock.On("Open", mock.Anything).Return(pool, nil)
	m.PgPoolContainerMock.On("Close")
	m.SqlConnectorMock.On("NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.PgPoolContainerMock, nil)

	m.PgQuerierMock.On("GetPostgresRolePermissions", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetPostgresRolePermissionsRow{
			{
				TableSchema:   "Users",
				TableName:     "Users",
				PrivilegeType: "Insert",
			},
			{
				TableSchema:   "Users",
				TableName:     "Users",
				PrivilegeType: "Delete",
			},
		}, nil)

	resp, err := m.Service.CheckConnectionConfig(context.Background(), &connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest]{
		Msg: &mgmtv1alpha1.CheckConnectionConfigRequest{
			ConnectionConfig: getPostgresConfigMock(),
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Msg.Privileges[0].PrivilegeType), "There should be two privilege types for this connection")
}

func Test_CheckConnectionConfig_Mysql(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	m.SqlDbContainerMock.On("Open").Return(m.SqlDbMock, nil)
	m.SqlDbContainerMock.On("Close").Return(nil)
	m.SqlConnectorMock.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.SqlDbContainerMock, nil)
	m.MysqlQuerierMock.On("GetMysqlRolePermissions", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetMysqlRolePermissionsRow{
			{
				TableSchema:   "Users",
				TableName:     "Users",
				PrivilegeType: "Insert",
			},
			{
				TableSchema:   "Users",
				TableName:     "Users",
				PrivilegeType: "Delete",
			},
		}, nil)

	resp, err := m.Service.CheckConnectionConfig(context.Background(), &connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest]{
		Msg: &mgmtv1alpha1.CheckConnectionConfigRequest{
			ConnectionConfig: getMysqlConfigMock(),
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, true, resp.Msg.IsConnected)
	assert.Nil(t, resp.Msg.ConnectionError)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_CheckConnectionConfigs_Fail(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	m.SqlDbContainerMock.On("Open").Return(m.SqlDbMock, nil)
	m.SqlDbContainerMock.On("Close").Return(nil)
	m.SqlConnectorMock.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.SqlDbContainerMock, nil)
	m.MysqlQuerierMock.On("GetMysqlRolePermissions", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetMysqlRolePermissionsRow{}, errors.New("connection failed"))

	resp, err := m.Service.CheckConnectionConfig(context.Background(), &connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest]{
		Msg: &mgmtv1alpha1.CheckConnectionConfigRequest{
			ConnectionConfig: getMysqlConfigMock(),
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, false, resp.Msg.IsConnected)
	assert.NotNil(t, resp.Msg.ConnectionError)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_CheckConnectionConfig_NotImplemented(t *testing.T) {
	m := createServiceMock(t)

	resp, err := m.Service.CheckConnectionConfig(context.Background(), &connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest]{
		Msg: &mgmtv1alpha1.CheckConnectionConfigRequest{
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{},
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// IsConnectionNameAvailable
func Test_IsConnectionNameAvailable_True(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("IsConnectionNameAvailable", context.Background(), mock.Anything, db_queries.IsConnectionNameAvailableParams{
		AccountId:      accountUuid,
		ConnectionName: mockConnectionName,
	}).Return(int64(0), nil)

	resp, err := m.Service.IsConnectionNameAvailable(context.Background(), &connect.Request[mgmtv1alpha1.IsConnectionNameAvailableRequest]{
		Msg: &mgmtv1alpha1.IsConnectionNameAvailableRequest{
			AccountId:      mockAccountId,
			ConnectionName: mockConnectionName,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, true, resp.Msg.IsAvailable)
}

func Test_IsConnectionNameAvailable_False(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("IsConnectionNameAvailable", context.Background(), mock.Anything, db_queries.IsConnectionNameAvailableParams{
		AccountId:      accountUuid,
		ConnectionName: mockConnectionName,
	}).Return(int64(1), nil)

	resp, err := m.Service.IsConnectionNameAvailable(context.Background(), &connect.Request[mgmtv1alpha1.IsConnectionNameAvailableRequest]{
		Msg: &mgmtv1alpha1.IsConnectionNameAvailableRequest{
			AccountId:      mockAccountId,
			ConnectionName: mockConnectionName,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, false, resp.Msg.IsAvailable)
}

// GetConnections
func Test_GetConnections(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connections := []db_queries.NeosyncApiConnection{getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)}
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetConnectionsByAccount", context.Background(), mock.Anything, accountUuid).Return(connections, nil)

	resp, err := m.Service.GetConnections(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionsRequest]{
		Msg: &mgmtv1alpha1.GetConnectionsRequest{
			AccountId: mockAccountId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.Msg.GetConnections()))
	assert.Equal(t, mockConnectionId, resp.Msg.Connections[0].Id)
}

func Test_GetConnections_Error(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	var nilConnections []db_queries.NeosyncApiConnection
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetConnectionsByAccount", context.Background(), mock.Anything, accountUuid).Return(nilConnections, sql.ErrNoRows)

	resp, err := m.Service.GetConnections(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionsRequest]{
		Msg: &mgmtv1alpha1.GetConnectionsRequest{
			AccountId: mockAccountId,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// GetConnection
func Test_GetConnection(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)

	resp, err := m.Service.GetConnection(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionRequest]{
		Msg: &mgmtv1alpha1.GetConnectionRequest{
			Id: mockConnectionId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockAccountId, resp.Msg.Connection.AccountId)
	assert.Equal(t, mockConnectionId, resp.Msg.Connection.Id)
}

func Test_GetConnection_Supports_WorkerApiKeys(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)
	ctx := context.WithValue(context.Background(), auth_apikey.TokenContextKey{}, &auth_apikey.TokenContextData{
		ApiKeyType: apikey.WorkerApiKey,
	})
	m.QuerierMock.On("GetConnectionById", ctx, mock.Anything, connectionUuid).Return(connection, nil)

	resp, err := m.Service.GetConnection(ctx, &connect.Request[mgmtv1alpha1.GetConnectionRequest]{
		Msg: &mgmtv1alpha1.GetConnectionRequest{
			Id: mockConnectionId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockAccountId, resp.Msg.Connection.AccountId)
	assert.Equal(t, mockConnectionId, resp.Msg.Connection.Id)
}

func Test_GetConnection_Error(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	var nilConnection db_queries.NeosyncApiConnection
	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(nilConnection, sql.ErrNoRows)

	_, err := m.Service.GetConnection(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionRequest]{
		Msg: &mgmtv1alpha1.GetConnectionRequest{
			Id: mockConnectionId,
		},
	})

	assert.Error(t, err)
}

// CreateConnection
func Test_CreateConnection(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)
	mockMgmtConnConfig := getPostgresConfigMock()
	mockConnectionConfig := &pg_models.ConnectionConfig{}
	_ = mockConnectionConfig.FromDto(mockMgmtConnConfig)
	mockUserAccountCalls(m.UserAccountServiceMock, true)
	m.QuerierMock.On("CreateConnection", context.Background(), mock.Anything, db_queries.CreateConnectionParams{
		AccountID:        accountUuid,
		Name:             mockConnectionName,
		ConnectionConfig: mockConnectionConfig,
		CreatedByID:      userUuid,
		UpdatedByID:      userUuid,
	}).Return(connection, nil)

	resp, err := m.Service.CreateConnection(context.Background(), &connect.Request[mgmtv1alpha1.CreateConnectionRequest]{
		Msg: &mgmtv1alpha1.CreateConnectionRequest{
			AccountId:        mockAccountId,
			Name:             mockConnectionName,
			ConnectionConfig: mockMgmtConnConfig,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockAccountId, resp.Msg.Connection.AccountId)
	assert.Equal(t, mockConnectionName, resp.Msg.Connection.Name)
}

func Test_CreateConnection_Error(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	mockMgmtConnConfig := getPostgresConfigMock()
	mockConnectionConfig := &pg_models.ConnectionConfig{}
	_ = mockConnectionConfig.FromDto(mockMgmtConnConfig)
	mockIsUserInAccount(m.UserAccountServiceMock, true)

	var nilConnection db_queries.NeosyncApiConnection
	m.UserAccountServiceMock.On("GetUser", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: mockUserId,
	}), nil)
	m.QuerierMock.On("CreateConnection", context.Background(), mock.Anything, db_queries.CreateConnectionParams{
		AccountID:        accountUuid,
		Name:             mockConnectionName,
		ConnectionConfig: mockConnectionConfig,
		CreatedByID:      userUuid,
		UpdatedByID:      userUuid,
	}).Return(nilConnection, errors.New("help"))

	resp, err := m.Service.CreateConnection(context.Background(), &connect.Request[mgmtv1alpha1.CreateConnectionRequest]{
		Msg: &mgmtv1alpha1.CreateConnectionRequest{
			AccountId:        mockAccountId,
			Name:             mockConnectionName,
			ConnectionConfig: mockMgmtConnConfig,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// UpdateConnection
func Test_UpdateConnection(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)
	mockMgmtConnConfig := getPostgresConfigMock()
	mockConnectionConfig := &pg_models.ConnectionConfig{}
	_ = mockConnectionConfig.FromDto(mockMgmtConnConfig)
	mockUserAccountCalls(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)
	m.QuerierMock.On("UpdateConnection", context.Background(), mock.Anything, db_queries.UpdateConnectionParams{
		ID:               connectionUuid,
		Name:             mockConnectionName,
		ConnectionConfig: mockConnectionConfig,
		UpdatedByID:      userUuid,
	}).Return(connection, nil)

	resp, err := m.Service.UpdateConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateConnectionRequest]{
		Msg: &mgmtv1alpha1.UpdateConnectionRequest{
			Id:               mockConnectionId,
			Name:             mockConnectionName,
			ConnectionConfig: mockMgmtConnConfig,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockConnectionId, resp.Msg.Connection.Id)
}

func Test_UpdateConnection_UpdateError(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)
	mockMgmtConnConfig := getPostgresConfigMock()
	mockConnectionConfig := &pg_models.ConnectionConfig{}
	_ = mockConnectionConfig.FromDto(mockMgmtConnConfig)
	mockUserAccountCalls(m.UserAccountServiceMock, true)
	var nilConnection db_queries.NeosyncApiConnection
	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)
	m.QuerierMock.On("UpdateConnection", context.Background(), mock.Anything, db_queries.UpdateConnectionParams{
		ID:               connectionUuid,
		ConnectionConfig: mockConnectionConfig,
		Name:             mockConnectionName,
		UpdatedByID:      userUuid,
	}).Return(nilConnection, errors.New("boo"))

	resp, err := m.Service.UpdateConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateConnectionRequest]{
		Msg: &mgmtv1alpha1.UpdateConnectionRequest{
			Id:               mockConnectionId,
			Name:             mockConnectionName,
			ConnectionConfig: mockMgmtConnConfig,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_UpdateConnection_GetConnectionError(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	mockMgmtConnConfig := getPostgresConfigMock()

	var nilConnection db_queries.NeosyncApiConnection
	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(nilConnection, sql.ErrNoRows)

	resp, err := m.Service.UpdateConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateConnectionRequest]{
		Msg: &mgmtv1alpha1.UpdateConnectionRequest{
			Id:               mockConnectionId,
			Name:             mockConnectionName,
			ConnectionConfig: mockMgmtConnConfig,
		},
	})

	m.QuerierMock.AssertNotCalled(t, "UpdateConnection", mock.Anything, mock.Anything, mock.Anything)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_UpdateConnection_UnverifiedUser(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)
	mockMgmtConnConfig := getPostgresConfigMock()
	mockConnectionConfig := &pg_models.ConnectionConfig{}
	_ = mockConnectionConfig.FromDto(mockMgmtConnConfig)
	mockIsUserInAccount(m.UserAccountServiceMock, false)

	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)

	resp, err := m.Service.UpdateConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateConnectionRequest]{
		Msg: &mgmtv1alpha1.UpdateConnectionRequest{
			Id:               mockConnectionId,
			Name:             mockConnectionName,
			ConnectionConfig: mockMgmtConnConfig,
		},
	})

	m.QuerierMock.AssertNotCalled(t, "UpdateConnection", mock.Anything, mock.Anything, mock.Anything)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// DeleteConnection
func Test_DeleteConnection(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)

	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)
	m.QuerierMock.On("RemoveConnectionById", context.Background(), mock.Anything, connectionUuid).Return(nil)

	resp, err := m.Service.DeleteConnection(context.Background(), &connect.Request[mgmtv1alpha1.DeleteConnectionRequest]{
		Msg: &mgmtv1alpha1.DeleteConnectionRequest{
			Id: mockConnectionId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_DeleteConnection_NotFound(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	var nilConnection db_queries.NeosyncApiConnection

	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(nilConnection, sql.ErrNoRows)

	resp, err := m.Service.DeleteConnection(context.Background(), &connect.Request[mgmtv1alpha1.DeleteConnectionRequest]{
		Msg: &mgmtv1alpha1.DeleteConnectionRequest{
			Id: mockConnectionId,
		},
	})

	m.QuerierMock.AssertNotCalled(t, "RemoveConnectionById", context.Background(), mock.Anything, mock.Anything)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_DeleteConnection_RemoveError(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)

	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)
	m.QuerierMock.On("RemoveConnectionById", context.Background(), mock.Anything, connectionUuid).Return(errors.New("sad"))

	resp, err := m.Service.DeleteConnection(context.Background(), &connect.Request[mgmtv1alpha1.DeleteConnectionRequest]{
		Msg: &mgmtv1alpha1.DeleteConnectionRequest{
			Id: mockConnectionId,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_DeleteConnection_UnverifiedUserError(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, false)

	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)

	resp, err := m.Service.DeleteConnection(context.Background(), &connect.Request[mgmtv1alpha1.DeleteConnectionRequest]{
		Msg: &mgmtv1alpha1.DeleteConnectionRequest{
			Id: mockConnectionId,
		},
	})

	m.QuerierMock.AssertNotCalled(t, "RemoveConnectionById")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// CheckSqlQuery
func Test_CheckSqlQuery_Valid(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock), nil)

	mockQuery := "some query"
	m.SqlDbContainerMock.On("Open").Return(m.SqlDbMock, nil)
	m.SqlDbContainerMock.On("Close").Return(nil)
	m.SqlConnectorMock.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.SqlDbContainerMock, nil)
	m.SqlMock.ExpectBegin()
	m.SqlMock.ExpectPrepare(mockQuery)

	resp, err := m.Service.CheckSqlQuery(context.Background(), &connect.Request[mgmtv1alpha1.CheckSqlQueryRequest]{
		Msg: &mgmtv1alpha1.CheckSqlQueryRequest{
			Id:    mockConnectionId,
			Query: mockQuery,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, true, resp.Msg.IsValid)
	assert.Nil(t, resp.Msg.ErorrMessage)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_CheckSqlQuery_Invalid(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock), nil)

	mockQuery := "another query"
	m.SqlDbContainerMock.On("Open").Return(m.SqlDbMock, nil)
	m.SqlDbContainerMock.On("Close").Return(nil)
	m.SqlConnectorMock.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.SqlDbContainerMock, nil)
	m.SqlMock.ExpectBegin()
	m.SqlMock.ExpectPrepare(mockQuery).WillReturnError(errors.New("error"))
	m.SqlMock.ExpectRollback()

	resp, err := m.Service.CheckSqlQuery(context.Background(), &connect.Request[mgmtv1alpha1.CheckSqlQueryRequest]{
		Msg: &mgmtv1alpha1.CheckSqlQueryRequest{
			Id:    mockConnectionId,
			Query: mockQuery,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, false, resp.Msg.IsValid)
	assert.NotNil(t, resp.Msg.ErorrMessage)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_CheckSqlQuery_Error(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock), nil)

	mockQuery := "diff query"
	m.SqlDbContainerMock.On("Open").Return(m.SqlDbMock, nil)
	m.SqlDbContainerMock.On("Close").Return(nil)
	m.SqlConnectorMock.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.SqlDbContainerMock, nil)
	m.SqlMock.ExpectBegin().WillReturnError(errors.New("error"))

	resp, err := m.Service.CheckSqlQuery(context.Background(), &connect.Request[mgmtv1alpha1.CheckSqlQueryRequest]{
		Msg: &mgmtv1alpha1.CheckSqlQueryRequest{
			Id:    mockConnectionId,
			Query: mockQuery,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

type serviceMocks struct {
	Service                *Service
	DbtxMock               *nucleusdb.MockDBTX
	QuerierMock            *db_queries.MockQuerier
	UserAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient
	SqlConnectorMock       *sqlconnect.MockSqlConnector
	SqlMock                sqlmock.Sqlmock
	SqlDbMock              *sql.DB
	SqlDbContainerMock     *sqlconnect.MockSqlDbContainer
	PgPoolContainerMock    *sqlconnect.MockPgPoolContainer
	PgQuerierMock          *pg_queries.MockQuerier
	MysqlQuerierMock       *mysql_queries.MockQuerier
	MongoConnectorMock     *mongoconnect.MockInterface
}

func createServiceMock(t *testing.T) *serviceMocks {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	mockPgquerier := pg_queries.NewMockQuerier(t)
	mockMysqlquerier := mysql_queries.NewMockQuerier(t)
	mockMongoConnector := mongoconnect.NewMockInterface(t)

	sqlDbMock, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier),
		mockUserAccountService, mockSqlConnector, mockPgquerier, mockMysqlquerier, mockMongoConnector)

	return &serviceMocks{
		Service:                service,
		DbtxMock:               mockDbtx,
		QuerierMock:            mockQuerier,
		UserAccountServiceMock: mockUserAccountService,
		SqlConnectorMock:       mockSqlConnector,
		SqlMock:                sqlMock,
		SqlDbMock:              sqlDbMock,
		SqlDbContainerMock:     sqlconnect.NewMockSqlDbContainer(t),
		PgPoolContainerMock:    sqlconnect.NewMockPgPoolContainer(t),
		PgQuerierMock:          mockPgquerier,
		MysqlQuerierMock:       mockMysqlquerier,
		MongoConnectorMock:     mockMongoConnector,
	}
}

func mockIsUserInAccount(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) {
	userAccountServiceMock.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: isInAccount,
	}), nil)
}

func mockUserAccountCalls(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) {
	mockIsUserInAccount(userAccountServiceMock, isInAccount)
	userAccountServiceMock.On("GetUser", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: mockUserId,
	}), nil)
}

//nolint:all
func getConnectionMock(accountId, name string, id pgtype.UUID, connType ConnTypeMock) db_queries.NeosyncApiConnection {
	accountUuid, _ := nucleusdb.ToUuid(accountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	timestamp := pgtype.Timestamp{
		Time: time.Now(),
	}
	if connType == MysqlMock {
		return db_queries.NeosyncApiConnection{
			AccountID:   accountUuid,
			Name:        name,
			ID:          id,
			CreatedByID: userUuid,
			UpdatedByID: userUuid,
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
			ConnectionConfig: &pg_models.ConnectionConfig{
				MysqlConfig: &pg_models.MysqlConnectionConfig{
					Connection: &pg_models.MysqlConnection{
						Host:     "host",
						Port:     5432,
						Name:     "database",
						User:     "user",
						Pass:     "topsecret",
						Protocol: "tcp",
					},
				},
			},
		}
	}
	sslMode := "disable"
	return db_queries.NeosyncApiConnection{
		AccountID:   accountUuid,
		Name:        name,
		ID:          id,
		CreatedByID: userUuid,
		UpdatedByID: userUuid,
		CreatedAt:   timestamp,
		UpdatedAt:   timestamp,
		ConnectionConfig: &pg_models.ConnectionConfig{
			PgConfig: &pg_models.PostgresConnectionConfig{
				Connection: &pg_models.PostgresConnection{
					Host:    "host",
					Port:    5432,
					Name:    "database",
					User:    "user",
					Pass:    "topsecret",
					SslMode: &sslMode,
				},
			},
		},
	}
}

func getPostgresConfigMock() *mgmtv1alpha1.ConnectionConfig {
	return &mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
			PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
				ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
					Connection: getPostgresConnectionMock(),
				},
			},
		},
	}
}

func getMysqlConfigMock() *mgmtv1alpha1.ConnectionConfig {
	return &mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
			MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
				ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
					Connection: getMysqlConnectionMock(),
				},
			},
		},
	}
}

func getMysqlConnectionMock() *mgmtv1alpha1.MysqlConnection {
	return &mgmtv1alpha1.MysqlConnection{
		Host:     "host",
		Port:     3306,
		Name:     "database",
		User:     "user",
		Pass:     "topsecret",
		Protocol: "tcp",
	}
}

func getPostgresConnectionMock() *mgmtv1alpha1.PostgresConnection {
	sslMode := "disable"
	return &mgmtv1alpha1.PostgresConnection{
		Host:    "host",
		Port:    5432,
		Name:    "database",
		User:    "user",
		Pass:    "topsecret",
		SslMode: &sslMode,
	}
}

// func getMysqlConfigMock() *mgmtv1alpha1.ConnectionConfig {
// 	return &mgmtv1alpha1.ConnectionConfig{
// 		Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
// 			MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
// 				ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
// 					Connection: getMysqlConnectionMock(),
// 				},
// 			},
// 		},
// 	}
// }

// func getMysqlConnectionMock() *mgmtv1alpha1.MysqlConnection {
// 	return &mgmtv1alpha1.MysqlConnection{
// 		Host:     "host",
// 		Port:     5432,
// 		Name:     "database",
// 		User:     "user",
// 		Pass:     "topsecret",
// 		Protocol: "tcp",
// 	}
// }
