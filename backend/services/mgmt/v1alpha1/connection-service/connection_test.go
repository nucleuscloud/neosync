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
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"

	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	jsonmodels "github.com/nucleuscloud/neosync/backend/internal/nucleusdb/json-models"
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

type mockConnector struct {
	db *sql.DB
}

func (mc *mockConnector) Open(driverName, dataSourceName string) (*sql.DB, error) {
	return mc.db, nil // always return the mock db
}

// CheckConnectionConfig
func Test_CheckConnectionConfig_Mysql(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	db, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, &mockConnector{db: db})

	sqlMock.ExpectPing()

	resp, err := service.CheckConnectionConfig(context.Background(), &connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest]{
		Msg: &mgmtv1alpha1.CheckConnectionConfigRequest{
			ConnectionConfig: getMysqlConfigMock(),
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, true, resp.Msg.IsConnected)
	assert.Nil(t, resp.Msg.ConnectionError)
}

func Test_CheckConnectionConfig_Mysql_Fail(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	db, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, &mockConnector{db: db})

	sqlMock.ExpectPing().WillReturnError(errors.New("connection failed"))

	resp, err := service.CheckConnectionConfig(context.Background(), &connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest]{
		Msg: &mgmtv1alpha1.CheckConnectionConfigRequest{
			ConnectionConfig: getMysqlConfigMock(),
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, false, resp.Msg.IsConnected)
	assert.NotNil(t, resp.Msg.ConnectionError)
}

func Test_CheckConnectionConfig_Postgres(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	db, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, &mockConnector{db: db})

	sqlMock.ExpectPing()

	resp, err := service.CheckConnectionConfig(context.Background(), &connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest]{
		Msg: &mgmtv1alpha1.CheckConnectionConfigRequest{
			ConnectionConfig: getPostgresConfigMock(),
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, true, resp.Msg.IsConnected)
	assert.Nil(t, resp.Msg.ConnectionError)
}

func Test_CheckConnectionConfig_Postgres_Fail(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	db, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, &mockConnector{db: db})

	sqlMock.ExpectPing().WillReturnError(errors.New("connection failed"))

	resp, err := service.CheckConnectionConfig(context.Background(), &connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest]{
		Msg: &mgmtv1alpha1.CheckConnectionConfigRequest{
			ConnectionConfig: getPostgresConfigMock(),
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, false, resp.Msg.IsConnected)
	assert.NotNil(t, resp.Msg.ConnectionError)
}

func Test_CheckConnectionConfig_NotImplemented(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	resp, err := service.CheckConnectionConfig(context.Background(), &connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest]{
		Msg: &mgmtv1alpha1.CheckConnectionConfigRequest{
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{},
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// IsConnectionNameAvailable
func Test_IsConnectionNameAvailable_True(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	mockUserAccountService.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: true,
	}), nil)
	mockQuerier.On("IsConnectionNameAvailable", context.Background(), mock.Anything, db_queries.IsConnectionNameAvailableParams{
		AccountId:      accountUuid,
		ConnectionName: mockConnectionName,
	}).Return(int64(0), nil)

	resp, err := service.IsConnectionNameAvailable(context.Background(), &connect.Request[mgmtv1alpha1.IsConnectionNameAvailableRequest]{
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
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	mockUserAccountService.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: true,
	}), nil)
	mockQuerier.On("IsConnectionNameAvailable", context.Background(), mock.Anything, db_queries.IsConnectionNameAvailableParams{
		AccountId:      accountUuid,
		ConnectionName: mockConnectionName,
	}).Return(int64(1), nil)

	resp, err := service.IsConnectionNameAvailable(context.Background(), &connect.Request[mgmtv1alpha1.IsConnectionNameAvailableRequest]{
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
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connections := []db_queries.NeosyncApiConnection{getConnectionMock(mockAccountId, mockConnectionName, connectionUuid)}
	mockUserAccountService.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: true,
	}), nil)
	mockQuerier.On("GetConnectionsByAccount", context.Background(), mock.Anything, accountUuid).Return(connections, nil)

	resp, err := service.GetConnections(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionsRequest]{
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
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	var nilConnections []db_queries.NeosyncApiConnection
	mockUserAccountService.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: true,
	}), nil)
	mockQuerier.On("GetConnectionsByAccount", context.Background(), mock.Anything, accountUuid).Return(nilConnections, sql.ErrNoRows)

	resp, err := service.GetConnections(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionsRequest]{
		Msg: &mgmtv1alpha1.GetConnectionsRequest{
			AccountId: mockAccountId,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// GetConnection
func Test_GetConnection(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid)
	mockUserAccountService.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: true,
	}), nil)
	mockQuerier.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)

	resp, err := service.GetConnection(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionRequest]{
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
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	var nilConnection db_queries.NeosyncApiConnection
	mockQuerier.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(nilConnection, sql.ErrNoRows)

	_, err := service.GetConnection(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionRequest]{
		Msg: &mgmtv1alpha1.GetConnectionRequest{
			Id: mockConnectionId,
		},
	})

	assert.Error(t, err)
}

// CreateConnection
func Test_CreateConnection(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid)
	mockMgmtConnConfig := getPostgresConfigMock()
	mockConnectionConfig := &jsonmodels.ConnectionConfig{}
	mockConnectionConfig.FromDto(mockMgmtConnConfig)
	mockUserAccountService.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: true,
	}), nil)

	mockUserAccountService.On("GetUser", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: mockUserId,
	}), nil)
	mockQuerier.On("CreateConnection", context.Background(), mock.Anything, db_queries.CreateConnectionParams{
		AccountID:        accountUuid,
		Name:             mockConnectionName,
		ConnectionConfig: mockConnectionConfig,
		CreatedByID:      userUuid,
		UpdatedByID:      userUuid,
	}).Return(connection, nil)

	resp, err := service.CreateConnection(context.Background(), &connect.Request[mgmtv1alpha1.CreateConnectionRequest]{
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
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	mockMgmtConnConfig := getPostgresConfigMock()
	mockConnectionConfig := &jsonmodels.ConnectionConfig{}
	mockConnectionConfig.FromDto(mockMgmtConnConfig)
	mockUserAccountService.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: true,
	}), nil)

	var nilConnection db_queries.NeosyncApiConnection
	mockUserAccountService.On("GetUser", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: mockUserId,
	}), nil)
	mockQuerier.On("CreateConnection", context.Background(), mock.Anything, db_queries.CreateConnectionParams{
		AccountID:        accountUuid,
		Name:             mockConnectionName,
		ConnectionConfig: mockConnectionConfig,
		CreatedByID:      userUuid,
		UpdatedByID:      userUuid,
	}).Return(nilConnection, errors.New("help"))

	resp, err := service.CreateConnection(context.Background(), &connect.Request[mgmtv1alpha1.CreateConnectionRequest]{
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
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid)
	mockMgmtConnConfig := getPostgresConfigMock()
	mockConnectionConfig := &jsonmodels.ConnectionConfig{}
	mockConnectionConfig.FromDto(mockMgmtConnConfig)
	mockUserAccountService.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: true,
	}), nil)

	mockUserAccountService.On("GetUser", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: mockUserId,
	}), nil)
	mockQuerier.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)
	mockQuerier.On("UpdateConnection", context.Background(), mock.Anything, db_queries.UpdateConnectionParams{
		ID:               connectionUuid,
		ConnectionConfig: mockConnectionConfig,
		UpdatedByID:      userUuid,
	}).Return(connection, nil)

	resp, err := service.UpdateConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateConnectionRequest]{
		Msg: &mgmtv1alpha1.UpdateConnectionRequest{
			Id:               mockConnectionId,
			ConnectionConfig: mockMgmtConnConfig,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockConnectionId, resp.Msg.Connection.Id)
}

func Test_UpdateConnection_Error(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid)
	mockMgmtConnConfig := getPostgresConfigMock()
	mockConnectionConfig := &jsonmodels.ConnectionConfig{}
	mockConnectionConfig.FromDto(mockMgmtConnConfig)
	mockUserAccountService.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: true,
	}), nil)

	mockUserAccountService.On("GetUser", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: mockUserId,
	}), nil)
	var nilConnection db_queries.NeosyncApiConnection
	mockQuerier.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)
	mockQuerier.On("UpdateConnection", context.Background(), mock.Anything, db_queries.UpdateConnectionParams{
		ID:               connectionUuid,
		ConnectionConfig: mockConnectionConfig,
		UpdatedByID:      userUuid,
	}).Return(nilConnection, errors.New("boo"))

	resp, err := service.UpdateConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateConnectionRequest]{
		Msg: &mgmtv1alpha1.UpdateConnectionRequest{
			Id:               mockConnectionId,
			ConnectionConfig: mockMgmtConnConfig,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// DeleteConnection
func Test_DeleteConnection(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid)
	mockUserAccountService.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: true,
	}), nil)

	mockQuerier.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)
	mockQuerier.On("RemoveConnectionById", context.Background(), mock.Anything, connectionUuid).Return(nil)

	resp, err := service.DeleteConnection(context.Background(), &connect.Request[mgmtv1alpha1.DeleteConnectionRequest]{
		Msg: &mgmtv1alpha1.DeleteConnectionRequest{
			Id: mockConnectionId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_DeleteConnection_NotFound(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	var nilConnection db_queries.NeosyncApiConnection

	mockQuerier.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(nilConnection, sql.ErrNoRows)

	resp, err := service.DeleteConnection(context.Background(), &connect.Request[mgmtv1alpha1.DeleteConnectionRequest]{
		Msg: &mgmtv1alpha1.DeleteConnectionRequest{
			Id: mockConnectionId,
		},
	})

	mockQuerier.AssertNotCalled(t, "RemoveConnectionById", context.Background(), mock.Anything, mock.Anything)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_DeleteConnectionError(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid)
	mockUserAccountService.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: true,
	}), nil)

	mockQuerier.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)
	mockQuerier.On("RemoveConnectionById", context.Background(), mock.Anything, connectionUuid).Return(errors.New("sad"))

	resp, err := service.DeleteConnection(context.Background(), &connect.Request[mgmtv1alpha1.DeleteConnectionRequest]{
		Msg: &mgmtv1alpha1.DeleteConnectionRequest{
			Id: mockConnectionId,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// getConnectionUrl
func Test_GetConnectionUrl_Postgres(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	url, err := service.getConnectionUrl(getPostgresConfigMock())

	assert.NoError(t, err)
	assert.Equal(t, "postgres://user:topsecret@host:5432/database?sslmode=disable", url)
}

func Test_GetConnectionUrl_Mysql(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	url, err := service.getConnectionUrl(getMysqlConfigMock())

	assert.NoError(t, err)
	assert.Equal(t, "user:topsecret@tcp(host:5432)/database", url)
}

func Test_GetConnectionUrl_NotImplemented(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService, nil)

	_, err := service.getConnectionUrl(&mgmtv1alpha1.ConnectionConfig{})

	assert.Error(t, err)
}

func getConnectionMock(accountId, name string, id pgtype.UUID) db_queries.NeosyncApiConnection {
	accountUuid, _ := nucleusdb.ToUuid(accountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	timestamp := pgtype.Timestamp{
		Time: time.Now(),
	}
	return db_queries.NeosyncApiConnection{
		AccountID:        accountUuid,
		Name:             name,
		ID:               id,
		CreatedByID:      userUuid,
		UpdatedByID:      userUuid,
		CreatedAt:        timestamp,
		UpdatedAt:        timestamp,
		ConnectionConfig: &jsonmodels.ConnectionConfig{},
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
		Port:     5432,
		Name:     "database",
		User:     "user",
		Pass:     "topsecret",
		Protocol: "tcp",
	}
}
