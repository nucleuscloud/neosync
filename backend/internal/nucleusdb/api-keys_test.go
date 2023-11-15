package nucleusdb

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	"github.com/stretchr/testify/mock"
	"github.com/zeebo/assert"
)

func Test_CreateAccountApiKey(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	service := New(dbtxMock, querierMock)
	dbtxMock.On("Begin", mock.Anything).Return(mockTx, nil)
	mockTx.On("Commit", mock.Anything).Return(nil)
	mockTx.On("Rollback", mock.Anything).Return(nil)
	user := db_queries.NeosyncApiUser{
		ID:       newPgUuid(t),
		UserType: 1,
	}
	querierMock.On("CreateMachineUser", mock.Anything, mock.Anything).
		Return(user, nil)
	apiKeyUuid := newPgUuid(t)
	querierMock.On("CreateAccountApiKey", mock.Anything, mock.Anything, mock.Anything).
		Return(db_queries.NeosyncApiAccountApiKey{
			ID: apiKeyUuid,
		}, nil)

	newApiKey, err := service.CreateAccountApikey(
		context.Background(),
		&CreateAccountApiKeyRequest{},
	)
	assert.NoError(t, err)
	assert.NotNil(t, newApiKey)
	assert.Equal(t, newApiKey.ID, apiKeyUuid)
}

func newPgUuid(t *testing.T) pgtype.UUID {
	t.Helper()
	newuuid := uuid.NewString()
	val, err := ToUuid(newuuid)
	assert.NoError(t, err)
	return val
}
