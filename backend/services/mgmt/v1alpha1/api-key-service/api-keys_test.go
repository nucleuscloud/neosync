package v1alpha1_apikeyservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	pgxmock "github.com/nucleuscloud/neosync/internal/mocks/github.com/jackc/pgx/v5"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_Service_GetAccountApiKeys(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

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
	mockIsUserInAccount(t, mockUserService, true)

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
		assert.Equal(t, apiKey.Id, neosyncdb.UUIDString(dbApikey.ID))
		assert.Nil(t, apiKey.KeyValue)
		assert.Equal(t, apiKey.Name, dbApikey.KeyName)
	}
}

func Test_Service_GetAccountApiKeys_ForbiddenAccount(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

	mockIsUserInAccount(t, mockUserService, false)

	resp, err := svc.GetAccountApiKeys(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountApiKeysRequest{
		AccountId: uuid.NewString(),
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_GetAccountApiKey_Found(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

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
	mockIsUserInAccount(t, mockUserService, true)

	resp, err := svc.GetAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.Msg.ApiKey.Id, neosyncdb.UUIDString(rawData.ID))
	assert.Nil(t, resp.Msg.ApiKey.KeyValue)
}

func Test_Service_GetAccountApiKey_NotFound(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

	mockQuerier.On("GetAccountApiKeyById", mock.Anything, mock.Anything, mock.Anything).
		Return(db_queries.NeosyncApiAccountApiKey{}, pgx.ErrNoRows)

	resp, err := svc.GetAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_GetAccountApiKey_Found_ForbiddenAccount(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

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
	mockIsUserInAccount(t, mockUserService, false)

	resp, err := svc.GetAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_CreateAccountApiKey(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockTx := pgxmock.NewMockTx(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

	mockIsUserInAccount(t, mockUserService, true)

	mockDbtx.On("Begin", mock.Anything).Return(mockTx, nil)
	mockTx.On("Commit", mock.Anything).Return(nil)
	mockTx.On("Rollback", mock.Anything).Return(nil)
	user := db_queries.NeosyncApiUser{
		ID:       newPgUuid(t),
		UserType: 1,
	}
	mockQuerier.On("CreateMachineUser", mock.Anything, mock.Anything, mock.Anything).
		Return(user, nil)
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
		UserID:      user.ID,
	}
	mockQuerier.On("CreateAccountApiKey", mock.Anything, mock.Anything, mock.Anything).
		Return(rawData, nil)

	resp, err := svc.CreateAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.CreateAccountApiKeyRequest{
		AccountId: uuid.NewString(),
		Name:      "foo",
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Msg.ApiKey.KeyValue)
	assert.NotEqual(t, resp.Msg.ApiKey.KeyValue, rawData.KeyValue, "KeyValue return should be the clear text, not the hash")
}

func Test_Service_RegenerateAccountApiKey(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

	mockIsUserInAccount(t, mockUserService, true)

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
	mockQuerier.On("UpdateAccountApiKeyValue", mock.Anything, mock.Anything, mock.Anything).
		Return(rawData, nil)

	resp, err := svc.RegenerateAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.RegenerateAccountApiKeyRequest{
		Id:        uuid.NewString(),
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Msg.ApiKey.KeyValue)
	assert.NotEqual(t, resp.Msg.ApiKey.KeyValue, rawData.KeyValue, "KeyValue return should be the clear text, not the hash")
}

func Test_Service_RegenerateAccountApiKey_ForbiddenAccount(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

	mockIsUserInAccount(t, mockUserService, false)
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

	resp, err := svc.RegenerateAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.RegenerateAccountApiKeyRequest{
		Id:        uuid.NewString(),
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_RegenerateAccountApiKey_NotFound(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

	mockQuerier.On("GetAccountApiKeyById", mock.Anything, mock.Anything, mock.Anything).
		Return(db_queries.NeosyncApiAccountApiKey{}, pgx.ErrNoRows)

	resp, err := svc.RegenerateAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.RegenerateAccountApiKeyRequest{
		Id:        uuid.NewString(),
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_CreateAccountApiKey_ForbiddenAccount(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

	mockIsUserInAccount(t, mockUserService, false)

	resp, err := svc.CreateAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.CreateAccountApiKeyRequest{
		AccountId: uuid.NewString(),
		Name:      "foo",
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_DeleteAccountApiKey_Existing(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

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
	mockIsUserInAccount(t, mockUserService, true)
	mockQuerier.On("RemoveAccountApiKey", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	resp, err := svc.DeleteAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.DeleteAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_Service_DeleteAccountApiKey_Existing_ForbiddenAccount(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

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
	mockIsUserInAccount(t, mockUserService, false)

	resp, err := svc.DeleteAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.DeleteAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_DeleteAccountApiKey_NotFound(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

	mockQuerier.On("GetAccountApiKeyById", mock.Anything, mock.Anything, mock.Anything).
		Return(db_queries.NeosyncApiAccountApiKey{}, pgx.ErrNoRows)

	resp, err := svc.DeleteAccountApiKey(context.Background(), connect.NewRequest(&mgmtv1alpha1.DeleteAccountApiKeyRequest{
		Id: uuid.NewString(),
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_Service_DeleteAccountApiKey_Existing_DeleteRace(t *testing.T) {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserService := userdata.NewMockInterface(t)

	svc := New(&Config{}, neosyncdb.New(mockDbtx, mockQuerier), mockUserService)

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
	mockIsUserInAccount(t, mockUserService, true)
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
	val, err := neosyncdb.ToUuid(newuuid)
	assert.NoError(t, err)
	return val
}

func mockIsUserInAccount(t testing.TB, userServiceMock *userdata.MockInterface, isInAccount bool) {
	mockEntityEnforcer := userdata.NewMockEntityEnforcer(t)
	if isInAccount {
		mockEntityEnforcer.On("EnforceAccount", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
	} else {
		mockEntityEnforcer.On("EnforceAccount", mock.Anything, mock.Anything, mock.Anything).Once().Return(errors.New("test: not in account"))
	}
	userServiceMock.On("GetUser", mock.Anything).Once().Return(&userdata.User{
		EntityEnforcer: mockEntityEnforcer,
	}, nil)
}
