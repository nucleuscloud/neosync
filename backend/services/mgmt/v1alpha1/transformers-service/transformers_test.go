package v1alpha1_transformersservice

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
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	anonymousUserId            = "00000000-0000-0000-0000-000000000000"
	mockAuthProvider           = "test-provider"
	mockUserId                 = "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	mockAccountId              = "5629813e-1a35-4874-922c-9827d85f0378"
	mockTransformerName        = "transformer-name"
	mockTransformerDescription = "transformer-description"
	mockTransformerType        = "transformer-type"
	mockTransformerSource      = "transformer-source"
	mockTransformerId          = "884765c6-1708-488d-b03a-70a02b12c81e"
)

func Test_GetSystemTransformers(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	resp, err := m.Service.GetSystemTransformers(context.Background(), &connect.Request[mgmtv1alpha1.GetSystemTransformersRequest]{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Msg.GetTransformers())
}

func Test_GetUserDefinedTransformers(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	transformers := []db_queries.NeosyncApiTransformer{mockTransformer(mockAccountId, mockUserId, mockTransformerId)}
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetUserDefinedTransformersByAccount", context.Background(), mock.Anything, accountUuid).Return(transformers, nil)

	resp, err := m.Service.GetUserDefinedTransformers(context.Background(), &connect.Request[mgmtv1alpha1.GetUserDefinedTransformersRequest]{
		Msg: &mgmtv1alpha1.GetUserDefinedTransformersRequest{
			AccountId: mockAccountId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.Msg.GetTransformers()))
	assert.Equal(t, mockTransformerId, resp.Msg.Transformers[0].Id)
}

func Test_GetUserDefinedTransformersById(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	transformer := mockTransformer(mockAccountId, mockUserId, mockTransformerId)
	transformerId, _ := nucleusdb.ToUuid(mockTransformerId)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetUserDefinedTransformerById", context.Background(), mock.Anything, transformerId).Return(transformer, nil)

	resp, err := m.Service.GetUserDefinedTransformerById(context.Background(), &connect.Request[mgmtv1alpha1.GetUserDefinedTransformerByIdRequest]{
		Msg: &mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{
			TransformerId: mockTransformerId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockAccountId, resp.Msg.Transformer.AccountId)
	assert.Equal(t, mockTransformerId, resp.Msg.Transformer.Id)
}

func Test_CreateUserDefinedTransformer(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	transformer := mockTransformer(mockAccountId, mockUserId, mockTransformerId)
	mockMgmtTransformerConfig := getTransformerConfigMock()
	mockTransformerConfig := &pg_models.TransformerConfigs{}
	_ = mockTransformerConfig.FromTransformerConfigDto(mockMgmtTransformerConfig)
	mockUserAccountCalls(m.UserAccountServiceMock, true)
	m.QuerierMock.On("CreateUserDefinedTransformer", context.Background(), mock.Anything, db_queries.CreateUserDefinedTransformerParams{
		AccountID:         accountUuid,
		Name:              mockTransformerName,
		Description:       mockTransformerDescription,
		TransformerConfig: mockTransformerConfig,
		Type:              mockTransformerType,
		Source:            mockTransformerSource,
		CreatedByID:       userUuid,
		UpdatedByID:       userUuid,
	}).Return(transformer, nil)

	resp, err := m.Service.CreateUserDefinedTransformer(context.Background(), &connect.Request[mgmtv1alpha1.CreateUserDefinedTransformerRequest]{
		Msg: &mgmtv1alpha1.CreateUserDefinedTransformerRequest{
			AccountId:         mockAccountId,
			Name:              mockTransformerName,
			Description:       mockTransformerDescription,
			Type:              mockTransformerType,
			Source:            mockTransformerSource,
			TransformerConfig: mockMgmtTransformerConfig,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockAccountId, resp.Msg.Transformer.AccountId)
	assert.Equal(t, mockTransformerId, resp.Msg.Transformer.Id)
}

func Test_CreateUserDefinedTransformer_Error(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	mockMgmtTransformerConfig := getTransformerConfigMock()
	mockTransformerConfig := &pg_models.TransformerConfigs{}
	_ = mockTransformerConfig.FromTransformerConfigDto(mockMgmtTransformerConfig)
	mockUserAccountCalls(m.UserAccountServiceMock, true)

	var nilConnection db_queries.NeosyncApiTransformer

	m.UserAccountServiceMock.On("GetUser", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: mockUserId,
	}), nil)

	m.QuerierMock.On("CreateUserDefinedTransformer", context.Background(), mock.Anything, db_queries.
		CreateUserDefinedTransformerParams{
		AccountID:         accountUuid,
		Name:              mockTransformerName,
		Description:       mockTransformerDescription,
		TransformerConfig: mockTransformerConfig,
		Type:              mockTransformerType,
		Source:            mockTransformerSource,
		CreatedByID:       userUuid,
		UpdatedByID:       userUuid,
	}).Return(nilConnection, errors.New("help"))

	resp, err := m.Service.CreateUserDefinedTransformer(context.Background(), &connect.Request[mgmtv1alpha1.CreateUserDefinedTransformerRequest]{
		Msg: &mgmtv1alpha1.CreateUserDefinedTransformerRequest{
			AccountId:         mockAccountId,
			Name:              mockTransformerName,
			Description:       mockTransformerDescription,
			Type:              mockTransformerType,
			Source:            mockTransformerSource,
			TransformerConfig: mockMgmtTransformerConfig,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_DeleteUserDefinedTransformer(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	transformerUuid, _ := nucleusdb.ToUuid(mockTransformerId)
	transformer := mockTransformer(mockAccountId, mockUserId, mockTransformerId)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetUserDefinedTransformerById", context.Background(), mock.Anything, transformerUuid).Return(transformer, nil)
	m.QuerierMock.On("DeleteUserDefinedTransformerById", context.Background(), mock.Anything, transformerUuid).Return(nil)

	resp, err := m.Service.DeleteUserDefinedTransformer(context.Background(), &connect.Request[mgmtv1alpha1.DeleteUserDefinedTransformerRequest]{
		Msg: &mgmtv1alpha1.DeleteUserDefinedTransformerRequest{
			TransformerId: mockTransformerId,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_DeleteConnection_NotFound(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	transformerUuid, _ := nucleusdb.ToUuid(mockTransformerId)
	var nilConnection db_queries.NeosyncApiTransformer

	m.QuerierMock.On("GetUserDefinedTransformerById", context.Background(), mock.Anything, transformerUuid).Return(nilConnection, sql.ErrNoRows)

	resp, err := m.Service.DeleteUserDefinedTransformer(context.Background(), &connect.Request[mgmtv1alpha1.DeleteUserDefinedTransformerRequest]{
		Msg: &mgmtv1alpha1.DeleteUserDefinedTransformerRequest{
			TransformerId: mockTransformerId,
		},
	})

	m.QuerierMock.AssertNotCalled(t, "DeleteTransformerById", context.Background(), mock.Anything, mock.Anything)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_DeleteConnection_RemoveError(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	transformerUuid, _ := nucleusdb.ToUuid(mockTransformerId)
	transformer := mockTransformer(mockAccountId, mockUserId, mockTransformerId)
	mockIsUserInAccount(m.UserAccountServiceMock, true)

	m.QuerierMock.On("GetUserDefinedTransformerById", context.Background(), mock.Anything, transformerUuid).Return(transformer, nil)
	m.QuerierMock.On("DeleteUserDefinedTransformerById", context.Background(), mock.Anything, transformerUuid).Return(errors.New("sad"))

	resp, err := m.Service.DeleteUserDefinedTransformer(context.Background(), &connect.Request[mgmtv1alpha1.DeleteUserDefinedTransformerRequest]{
		Msg: &mgmtv1alpha1.DeleteUserDefinedTransformerRequest{
			TransformerId: mockTransformerId,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_DeleteConnection_UnverifiedUserError(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	transformerUuid, _ := nucleusdb.ToUuid(mockTransformerId)
	transformer := mockTransformer(mockAccountId, mockUserId, mockTransformerId)
	mockIsUserInAccount(m.UserAccountServiceMock, false)

	m.QuerierMock.On("GetUserDefinedTransformerById", context.Background(), mock.Anything, transformerUuid).Return(transformer, nil)

	resp, err := m.Service.DeleteUserDefinedTransformer(context.Background(), &connect.Request[mgmtv1alpha1.DeleteUserDefinedTransformerRequest]{
		Msg: &mgmtv1alpha1.DeleteUserDefinedTransformerRequest{
			TransformerId: mockTransformerId,
		},
	})

	m.QuerierMock.AssertNotCalled(t, "DeleteUserDefinedTransformerById")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_UpdateTransformer(t *testing.T) {

	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	transformerUuid, _ := nucleusdb.ToUuid(mockTransformerId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	transformer := mockTransformer(mockAccountId, mockUserId, mockTransformerId)
	mockMgmtTransformerConfig := getTransformerConfigMock()
	mockTransformerConfig := &pg_models.TransformerConfigs{}
	_ = mockTransformerConfig.FromTransformerConfigDto(mockMgmtTransformerConfig)
	mockUserAccountCalls(m.UserAccountServiceMock, true)

	m.QuerierMock.On("GetUserDefinedTransformerById", context.Background(), mock.Anything, transformerUuid).Return(transformer, nil)

	m.QuerierMock.On("UpdateUserDefinedTransformer", context.Background(), mock.Anything, db_queries.UpdateUserDefinedTransformerParams{
		ID:                transformerUuid,
		Name:              mockTransformerName,
		TransformerConfig: mockTransformerConfig,
		UpdatedByID:       userUuid,
		Description:       mockTransformerDescription,
	}).Return(transformer, nil)

	resp, err := m.Service.UpdateUserDefinedTransformer(context.Background(), &connect.Request[mgmtv1alpha1.UpdateUserDefinedTransformerRequest]{
		Msg: &mgmtv1alpha1.UpdateUserDefinedTransformerRequest{
			TransformerId:     mockTransformerId,
			TransformerConfig: mockMgmtTransformerConfig,
			Name:              mockTransformerName,
			Description:       mockTransformerDescription,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockTransformerId, resp.Msg.Transformer.Id)
}

func Test_UpdateTransformer_UpdateError(t *testing.T) {

	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	transformerUuid, _ := nucleusdb.ToUuid(mockTransformerId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	transformer := mockTransformer(mockAccountId, mockUserId, mockTransformerId)
	mockMgmtTransformerConfig := getTransformerConfigMock()
	mockTransformerConfig := &pg_models.TransformerConfigs{}
	_ = mockTransformerConfig.FromTransformerConfigDto(mockMgmtTransformerConfig)
	mockUserAccountCalls(m.UserAccountServiceMock, true)
	var nilTransformer db_queries.NeosyncApiTransformer

	m.QuerierMock.On("GetUserDefinedTransformerById", context.Background(), mock.Anything, transformerUuid).Return(transformer, nil)

	m.QuerierMock.On("UpdateUserDefinedTransformer", context.Background(), mock.Anything, db_queries.UpdateUserDefinedTransformerParams{
		ID:                transformerUuid,
		Name:              mockTransformerName,
		TransformerConfig: mockTransformerConfig,
		UpdatedByID:       userUuid,
		Description:       mockTransformerDescription,
	}).Return(nilTransformer, errors.New("boo"))

	resp, err := m.Service.UpdateUserDefinedTransformer(context.Background(), &connect.Request[mgmtv1alpha1.UpdateUserDefinedTransformerRequest]{
		Msg: &mgmtv1alpha1.UpdateUserDefinedTransformerRequest{
			TransformerId:     mockTransformerId,
			TransformerConfig: mockMgmtTransformerConfig,
			Name:              mockTransformerName,
			Description:       mockTransformerDescription,
		},
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_UpdateTransformer_GetTransformerError(t *testing.T) {

	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	transformerUuid, _ := nucleusdb.ToUuid(mockTransformerId)
	mockMgmtTransformerConfig := getTransformerConfigMock()

	var nilTransformer db_queries.NeosyncApiTransformer

	m.QuerierMock.On("GetUserDefinedTransformerById", context.Background(), mock.Anything, transformerUuid).Return(nilTransformer, sql.ErrNoRows)

	resp, err := m.Service.UpdateUserDefinedTransformer(context.Background(), &connect.Request[mgmtv1alpha1.UpdateUserDefinedTransformerRequest]{
		Msg: &mgmtv1alpha1.UpdateUserDefinedTransformerRequest{
			TransformerId:     mockTransformerId,
			TransformerConfig: mockMgmtTransformerConfig,
			Name:              mockTransformerName,
			Description:       mockTransformerDescription,
		},
	})
	m.QuerierMock.AssertNotCalled(t, "UpdateUserDefinedTransformer", mock.Anything, mock.Anything, mock.Anything)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_UpdateTransformer_UnverifiedUser(t *testing.T) {

	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	transformerUuid, _ := nucleusdb.ToUuid(mockTransformerId)
	transformer := mockTransformer(mockAccountId, mockUserId, mockTransformerId)
	mockMgmtTransformerConfig := getTransformerConfigMock()
	mockTransformerConfig := &pg_models.TransformerConfigs{}
	_ = mockTransformerConfig.FromTransformerConfigDto(mockMgmtTransformerConfig)
	mockIsUserInAccount(m.UserAccountServiceMock, false)

	m.QuerierMock.On("GetUserDefinedTransformerById", context.Background(), mock.Anything, transformerUuid).Return(transformer, nil)

	resp, err := m.Service.UpdateUserDefinedTransformer(context.Background(), &connect.Request[mgmtv1alpha1.UpdateUserDefinedTransformerRequest]{
		Msg: &mgmtv1alpha1.UpdateUserDefinedTransformerRequest{
			TransformerId:     mockTransformerId,
			TransformerConfig: mockMgmtTransformerConfig,
			Name:              mockTransformerName,
			Description:       mockTransformerDescription,
		},
	})

	m.QuerierMock.AssertNotCalled(t, "UpdateUserDefinedTransformer", mock.Anything, mock.Anything, mock.Anything)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_IsTtransformerNameAvailable_True(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("IsTransformerNameAvailable", context.Background(), mock.Anything, db_queries.IsTransformerNameAvailableParams{
		AccountId:       accountUuid,
		TransformerName: mockTransformerName,
	}).Return(int64(0), nil)

	resp, err := m.Service.IsTransformerNameAvailable(context.Background(), &connect.Request[mgmtv1alpha1.IsTransformerNameAvailableRequest]{
		Msg: &mgmtv1alpha1.IsTransformerNameAvailableRequest{
			AccountId:       mockAccountId,
			TransformerName: mockTransformerName,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, true, resp.Msg.IsAvailable)
}

func Test_IsTransformerNameAvailable_False(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("IsTransformerNameAvailable", context.Background(), mock.Anything, db_queries.IsTransformerNameAvailableParams{
		AccountId:       accountUuid,
		TransformerName: mockTransformerName,
	}).Return(int64(1), nil)

	resp, err := m.Service.IsTransformerNameAvailable(context.Background(), &connect.Request[mgmtv1alpha1.IsTransformerNameAvailableRequest]{
		Msg: &mgmtv1alpha1.IsTransformerNameAvailableRequest{
			AccountId:       mockAccountId,
			TransformerName: mockTransformerName,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, false, resp.Msg.IsAvailable)
}

func Test_ValidateUserJavascriptCode_True(t *testing.T) {
	m := createServiceMock(t)

	code := `var payload = value+=" hello";return payload;`

	mockIsUserInAccount(m.UserAccountServiceMock, true)

	resp, err := m.Service.ValidateUserJavascriptCode(context.Background(), &connect.Request[mgmtv1alpha1.ValidateUserJavascriptCodeRequest]{
		Msg: &mgmtv1alpha1.ValidateUserJavascriptCodeRequest{
			AccountId: mockAccountId,
			Code:      code,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, true, resp.Msg.Valid)
}

func Test_ValidateUserJavascriptCode_False(t *testing.T) {
	m := createServiceMock(t)

	code := `var payload = value" hello";return payload;`

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	resp, err := m.Service.ValidateUserJavascriptCode(context.Background(), &connect.Request[mgmtv1alpha1.ValidateUserJavascriptCodeRequest]{
		Msg: &mgmtv1alpha1.ValidateUserJavascriptCodeRequest{
			AccountId: mockAccountId,
			Code:      code,
		},
	})

	// Assert no error was returned and the response is as expected.
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.Msg.Valid, false)
}

//nolint:all
func mockTransformer(accountId, userId, transformerId string) db_queries.NeosyncApiTransformer {

	id, _ := nucleusdb.ToUuid(transformerId)
	accountUuid, _ := nucleusdb.ToUuid(accountId)
	userUuid, _ := nucleusdb.ToUuid(userId)
	currentTime := time.Now()
	var timestamp pgtype.Timestamp
	timestamp.Time = currentTime

	return db_queries.NeosyncApiTransformer{
		ID:                id,
		CreatedAt:         timestamp,
		UpdatedAt:         timestamp,
		Name:              mockTransformerName,
		Description:       mockTransformerDescription,
		Type:              mockTransformerType,
		Source:            mockTransformerSource,
		TransformerConfig: &pg_models.TransformerConfigs{},
		AccountID:         accountUuid,
		CreatedByID:       userUuid,
		UpdatedByID:       userUuid,
	}
}

func mockIsUserInAccount(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) {
	userAccountServiceMock.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: isInAccount,
	}), nil)
}

//nolint:all
func mockUserAccountCalls(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) {
	mockIsUserInAccount(userAccountServiceMock, isInAccount)
	userAccountServiceMock.On("GetUser", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: mockUserId,
	}), nil)
}

type serviceMocks struct {
	Service                *Service
	DbtxMock               *nucleusdb.MockDBTX
	QuerierMock            *db_queries.MockQuerier
	UserAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient
	SqlMock                sqlmock.Sqlmock
	SqlDbMock              *sql.DB
}

func createServiceMock(t *testing.T) *serviceMocks {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	sqlDbMock, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	service := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService)

	return &serviceMocks{
		Service:                service,
		DbtxMock:               mockDbtx,
		QuerierMock:            mockQuerier,
		UserAccountServiceMock: mockUserAccountService,
		SqlMock:                sqlMock,
		SqlDbMock:              sqlDbMock,
	}
}

func getTransformerConfigMock() *mgmtv1alpha1.TransformerConfig {
	return &mgmtv1alpha1.TransformerConfig{
		Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
			GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{},
		},
	}
}
