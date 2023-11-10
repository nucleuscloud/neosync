package v1alpha1_apikeyservice

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Service_GetAccountApiKeys(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	svc := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService)

	rawData := []db_queries.NeosyncApiAccountApiKey{
		{
			ID:          newPgUuid(t),
			AccountID:   newPgUuid(t),
			KeyValue:    "foo",
			CreatedByID: newPgUuid(t),
			UpdatedByID: newPgUuid(t),
			CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
			ExpiresAt:   pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
			KeyName:     "foo",
		},
		{
			ID:          newPgUuid(t),
			AccountID:   newPgUuid(t),
			KeyValue:    "foobar",
			CreatedByID: newPgUuid(t),
			UpdatedByID: newPgUuid(t),
			CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
			ExpiresAt:   pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
			KeyName:     "foobar",
		},
	}
	mockQuerier.On("GetAccountApiKeys", mock.Anything, mock.Anything, mock.Anything).
		Return(rawData, nil)
	mockIsUserInAccount(mockUserAccountService, true)

	resp, err := svc.GetAccountApiKeys(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountApiKeysRequest{
		AccountId: uuid.NewString(),
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Msg.ApiKeys)
	assert.Len(
		t,
		resp.Msg.ApiKeys,
		len(rawData),
	)
	for idx, apiKey := range resp.Msg.ApiKeys {
		dbApikey := rawData[idx]
		assert.Equal(t, apiKey.Id, nucleusdb.UUIDString(dbApikey.ID))
		assert.Nil(t, apiKey.KeyValue)
		assert.Equal(t, apiKey.Name, dbApikey.KeyName)
	}
}

func Test_Service_GetAccountApiKeys_ForbiddenAccount(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	svc := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService)

	mockIsUserInAccount(mockUserAccountService, false)

	resp, err := svc.GetAccountApiKeys(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountApiKeysRequest{
		AccountId: uuid.NewString(),
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_GetAccountApiKey_Found(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	svc := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService)

	rawData := db_queries.NeosyncApiAccountApiKey{
		ID:          newPgUuid(t),
		AccountID:   newPgUuid(t),
		KeyValue:    "foo",
		CreatedByID: newPgUuid(t),
		UpdatedByID: newPgUuid(t),
		CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
		ExpiresAt:   pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
		KeyName:     "foo",
	}
	mockQuerier.On("GetAccountApiKeyById", mock.Anything, mock.Anything, mock.Anything).
		Return(rawData, nil)
	mockIsUserInAccount(mockUserAccountService, true)

	resp, err := svc.GetAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.Msg.ApiKey.Id, nucleusdb.UUIDString(rawData.ID))
	assert.Nil(t, resp.Msg.ApiKey.KeyValue)
}

func Test_Service_GetAccountApiKey_NotFound(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	svc := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService)

	mockQuerier.On("GetAccountApiKeyById", mock.Anything, mock.Anything, mock.Anything).
		Return(db_queries.NeosyncApiAccountApiKey{}, pgx.ErrNoRows)

	resp, err := svc.GetAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_GetAccountApiKey_Found_ForbiddenAccount(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	svc := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService)

	rawData := db_queries.NeosyncApiAccountApiKey{
		ID:          newPgUuid(t),
		AccountID:   newPgUuid(t),
		KeyValue:    "foo",
		CreatedByID: newPgUuid(t),
		UpdatedByID: newPgUuid(t),
		CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
		ExpiresAt:   pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
		KeyName:     "foo",
	}
	mockQuerier.On("GetAccountApiKeyById", mock.Anything, mock.Anything, mock.Anything).
		Return(rawData, nil)
	mockIsUserInAccount(mockUserAccountService, false)

	resp, err := svc.GetAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_DeleteAccountApiKey_Existing(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	svc := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService)

	rawData := db_queries.NeosyncApiAccountApiKey{
		ID:          newPgUuid(t),
		AccountID:   newPgUuid(t),
		KeyValue:    "foo",
		CreatedByID: newPgUuid(t),
		UpdatedByID: newPgUuid(t),
		CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
		ExpiresAt:   pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
		KeyName:     "foo",
	}
	mockQuerier.On("GetAccountApiKeyById", mock.Anything, mock.Anything, mock.Anything).
		Return(rawData, nil)
	mockIsUserInAccount(mockUserAccountService, true)
	mockQuerier.On("RemoveAccountApiKey", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	resp, err := svc.DeleteAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.DeleteAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_Service_DeleteAccountApiKey_Existing_ForbiddenAccount(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	svc := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService)

	rawData := db_queries.NeosyncApiAccountApiKey{
		ID:          newPgUuid(t),
		AccountID:   newPgUuid(t),
		KeyValue:    "foo",
		CreatedByID: newPgUuid(t),
		UpdatedByID: newPgUuid(t),
		CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
		ExpiresAt:   pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
		KeyName:     "foo",
	}
	mockQuerier.On("GetAccountApiKeyById", mock.Anything, mock.Anything, mock.Anything).
		Return(rawData, nil)
	mockIsUserInAccount(mockUserAccountService, false)

	resp, err := svc.DeleteAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.DeleteAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_DeleteAccountApiKey_NotFound(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	svc := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService)

	mockQuerier.On("GetAccountApiKeyById", mock.Anything, mock.Anything, mock.Anything).
		Return(db_queries.NeosyncApiAccountApiKey{}, pgx.ErrNoRows)

	resp, err := svc.DeleteAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.DeleteAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_Service_DeleteAccountApiKey_Existing_DeleteRace(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)

	svc := New(&Config{}, nucleusdb.New(mockDbtx, mockQuerier), mockUserAccountService)

	rawData := db_queries.NeosyncApiAccountApiKey{
		ID:          newPgUuid(t),
		AccountID:   newPgUuid(t),
		KeyValue:    "foo",
		CreatedByID: newPgUuid(t),
		UpdatedByID: newPgUuid(t),
		CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
		ExpiresAt:   pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
		KeyName:     "foo",
	}
	mockQuerier.On("GetAccountApiKeyById", mock.Anything, mock.Anything, mock.Anything).
		Return(rawData, nil)
	mockIsUserInAccount(mockUserAccountService, true)
	mockQuerier.On("RemoveAccountApiKey", mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)

	resp, err := svc.DeleteAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.DeleteAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func newPgUuid(t *testing.T) pgtype.UUID {
	t.Helper()
	newuuid := uuid.NewString()
	val, err := nucleusdb.ToUuid(newuuid)
	assert.NoError(t, err)
	return val
}

func mockIsUserInAccount(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) {
	userAccountServiceMock.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: isInAccount,
	}), nil)
}

func mockUserAccountCalls(
	userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient,
	isInAccount bool,
	userId string,
) {
	mockIsUserInAccount(userAccountServiceMock, isInAccount)
	userAccountServiceMock.On("GetUser", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: userId,
	}), nil)
}
