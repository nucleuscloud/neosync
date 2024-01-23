package auth_apikey

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
	"github.com/stretchr/testify/mock"
	"github.com/zeebo/assert"
)

func Test_Client_New(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	assert.NotNil(t, New(mockQuerier, mockDbTx, []string{}))
}

func Test_Client_InjectTokenCtx_Account(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	client := New(mockQuerier, mockDbTx, []string{})

	fakeToken := apikey.NewV1AccountKey()
	hashedFakeToken := utils.ToSha256(
		fakeToken,
	)
	expiresAt, err := nucleusdb.ToTimestamp(time.Now().Add(5 * time.Minute))
	assert.NoError(t, err)
	apiKeyRecord := db_queries.NeosyncApiAccountApiKey{
		ID:        pgtype.UUID{Valid: true},
		ExpiresAt: expiresAt,
	}
	mockQuerier.On("GetAccountApiKeyByKeyValue", mock.Anything, mock.Anything, hashedFakeToken).
		Return(apiKeyRecord, nil)

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", fakeToken)},
	})
	assert.NoError(t, err)
	assert.NotNil(t, newctx)

	data, err := GetTokenDataFromCtx(newctx)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(
		t,
		data,
		&TokenContextData{
			RawToken:   fakeToken,
			ApiKey:     &apiKeyRecord,
			ApiKeyType: apikey.AccountApiKey,
		},
	)
}

func Test_Client_InjectTokenCtx_Account_Expired(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	client := New(mockQuerier, mockDbTx, []string{})

	fakeToken := apikey.NewV1AccountKey()
	hashedFakeToken := utils.ToSha256(
		fakeToken,
	)
	expiresAt, err := nucleusdb.ToTimestamp(time.Now().Add(-5 * time.Second))
	assert.NoError(t, err)
	apiKeyRecord := db_queries.NeosyncApiAccountApiKey{
		ID:        pgtype.UUID{Valid: true},
		ExpiresAt: expiresAt,
	}
	mockQuerier.On("GetAccountApiKeyByKeyValue", mock.Anything, mock.Anything, hashedFakeToken).
		Return(apiKeyRecord, nil)

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", fakeToken)},
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ApiKeyExpiredErr))
	assert.Nil(t, newctx)
}

func Test_Client_InjectTokenCtx_InvalidHeader(t *testing.T) {
	client := &Client{}
	_, err := client.InjectTokenCtx(context.Background(), http.Header{"Authorization": []string{}})
	assert.Error(t, err)
}

func Test_Client_InjectTokenCtx_InvalidToken(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	client := New(mockQuerier, mockDbTx, []string{})

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{"Bearer 123"},
	})
	assert.Error(t, err)
	assert.Nil(t, newctx)
}

func Test_Client_InjectTokenCtx_Account_NotFoundKeyValue(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	client := New(mockQuerier, mockDbTx, []string{})

	fakeToken := apikey.NewV1AccountKey()
	hashedFakeToken := utils.ToSha256(
		fakeToken,
	)

	mockQuerier.On("GetAccountApiKeyByKeyValue", mock.Anything, mock.Anything, hashedFakeToken).
		Return(db_queries.NeosyncApiAccountApiKey{}, pgx.ErrNoRows)

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", fakeToken)},
	})
	assert.Error(t, err)
	assert.Nil(t, newctx)
}

func Test_GetTokenDataFromCtx(t *testing.T) {
	ctx := context.WithValue(context.Background(), TokenContextKey{}, &TokenContextData{})
	_, err := GetTokenDataFromCtx(ctx)
	assert.NoError(t, err)
}

func Test_GetTokenDataFromCtx_UnAuthenticated(t *testing.T) {
	_, err := GetTokenDataFromCtx(context.Background())
	assert.Error(t, err)
}

func Test_Client_InjectTokenCtx_Worker_Allowed(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	fakeToken := apikey.NewV1WorkerKey()

	client := New(mockQuerier, mockDbTx, []string{fakeToken, apikey.NewV1WorkerKey()})

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", fakeToken)},
	})
	assert.NoError(t, err)
	assert.NotNil(t, newctx)

	data, err := GetTokenDataFromCtx(newctx)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(
		t,
		data,
		&TokenContextData{
			RawToken:   fakeToken,
			ApiKeyType: apikey.WorkerApiKey,
		},
	)
}

func Test_Client_InjectTokenCtx_Worker_DisAllowed(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	fakeToken := apikey.NewV1WorkerKey()

	client := New(mockQuerier, mockDbTx, []string{apikey.NewV1WorkerKey(), apikey.NewV1WorkerKey()})

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", fakeToken)},
	})
	assert.Error(t, err)
	assert.Nil(t, newctx)
}

func Test_secureCompare(t *testing.T) {
	assert.True(t, secureCompare("a", "a"))
	assert.False(t, secureCompare("a", "b"))
	assert.False(t, secureCompare("a", "aa"))
}
