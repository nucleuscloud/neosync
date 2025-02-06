package auth_apikey

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	pkg_utils "github.com/nucleuscloud/neosync/backend/pkg/utils"
	"github.com/nucleuscloud/neosync/internal/apikey"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
	"github.com/stretchr/testify/mock"
	"github.com/zeebo/assert"
)

func Test_Client_New(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	assert.NotNil(t, New(mockQuerier, mockDbTx, []string{}, []string{}))
}

func Test_Client_InjectTokenCtx_Account(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	client := New(mockQuerier, mockDbTx, []string{}, []string{})

	fakeToken := apikey.NewV1AccountKey()
	hashedFakeToken := pkg_utils.ToSha256(
		fakeToken,
	)
	expiresAt, err := neosyncdb.ToTimestamp(time.Now().Add(5 * time.Minute))
	assert.NoError(t, err)
	apiKeyRecord := db_queries.NeosyncApiAccountApiKey{
		ID:        pgtype.UUID{Valid: true},
		ExpiresAt: expiresAt,
	}
	mockQuerier.On("GetAccountApiKeyByKeyValue", mock.Anything, mock.Anything, hashedFakeToken).
		Return(apiKeyRecord, nil)

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", fakeToken)},
	}, connect.Spec{})
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

	client := New(mockQuerier, mockDbTx, []string{}, []string{})

	fakeToken := apikey.NewV1AccountKey()
	hashedFakeToken := pkg_utils.ToSha256(
		fakeToken,
	)
	expiresAt, err := neosyncdb.ToTimestamp(time.Now().Add(-5 * time.Second))
	assert.NoError(t, err)
	apiKeyRecord := db_queries.NeosyncApiAccountApiKey{
		ID:        pgtype.UUID{Valid: true},
		ExpiresAt: expiresAt,
	}
	mockQuerier.On("GetAccountApiKeyByKeyValue", mock.Anything, mock.Anything, hashedFakeToken).
		Return(apiKeyRecord, nil)

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", fakeToken)},
	}, connect.Spec{})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ApiKeyExpiredErr))
	assert.Nil(t, newctx)
}

func Test_Client_InjectTokenCtx_InvalidHeader(t *testing.T) {
	client := &Client{}
	_, err := client.InjectTokenCtx(context.Background(), http.Header{"Authorization": []string{}}, connect.Spec{})
	assert.Error(t, err)
}

func Test_Client_InjectTokenCtx_InvalidToken(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	client := New(mockQuerier, mockDbTx, []string{}, []string{})

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{"Bearer 123"},
	}, connect.Spec{})
	assert.Error(t, err)
	assert.Nil(t, newctx)
}

func Test_Client_InjectTokenCtx_Account_NotFoundKeyValue(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	client := New(mockQuerier, mockDbTx, []string{}, []string{})

	fakeToken := apikey.NewV1AccountKey()
	hashedFakeToken := pkg_utils.ToSha256(
		fakeToken,
	)

	mockQuerier.On("GetAccountApiKeyByKeyValue", mock.Anything, mock.Anything, hashedFakeToken).
		Return(db_queries.NeosyncApiAccountApiKey{}, pgx.ErrNoRows)

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", fakeToken)},
	}, connect.Spec{})
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

	client := New(mockQuerier, mockDbTx, []string{fakeToken, apikey.NewV1WorkerKey()}, []string{"/foo"})

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", fakeToken)},
	}, connect.Spec{Procedure: "/foo"})
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

func Test_Client_InjectTokenCtx_Worker_DisAllowed_ApiKey(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	fakeToken := apikey.NewV1WorkerKey()

	client := New(mockQuerier, mockDbTx, []string{apikey.NewV1WorkerKey(), apikey.NewV1WorkerKey()}, []string{})

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", fakeToken)},
	}, connect.Spec{})
	assert.Error(t, err)
	assert.Nil(t, newctx)
}

func Test_Client_InjectTokenCtx_Worker_DisAllowed_Procedure(t *testing.T) {
	mockQuerier := db_queries.NewMockQuerier(t)
	mockDbTx := db_queries.NewMockDBTX(t)

	fakeToken := apikey.NewV1WorkerKey()

	client := New(mockQuerier, mockDbTx, []string{fakeToken}, []string{"/foo"})

	newctx, err := client.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", fakeToken)},
	}, connect.Spec{Procedure: "/bar"})
	assert.Error(t, err)
	assert.Nil(t, newctx)
}

func Test_secureCompare(t *testing.T) {
	assert.True(t, secureCompare("a", "a"))
	assert.False(t, secureCompare("a", "b"))
	assert.False(t, secureCompare("a", "aa"))
}
