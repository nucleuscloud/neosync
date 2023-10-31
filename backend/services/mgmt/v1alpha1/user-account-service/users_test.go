package v1alpha1_useraccountservice

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	authjwt "github.com/nucleuscloud/neosync/backend/internal/jwt"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type tokenContextKey struct{}

func Test_GetUser_Anonymous(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier))

	userId := "00000000-0000-0000-0000-000000000000"
	idUuid, _ := nucleusdb.ToUuid(userId)
	user := db_queries.NeosyncApiUser{ID: idUuid}

	// getAnonymousUser
	mockQuerier.On("GetAnonymousUser", context.Background(), mock.Anything).Return(user, nil)

	resp, err := service.GetUser(context.Background(), &connect.Request[mgmtv1alpha1.GetUserRequest]{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userId, resp.Msg.GetUserId())
}

func Test_GetUser(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	service := New(&Config{IsAuthEnabled: true}, nucleusdb.New(mockDbtx, mockQuerier))

	userId := "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	idUuid, _ := nucleusdb.ToUuid(userId)
	user := db_queries.NeosyncApiUser{ID: idUuid}

	authProviderId := "test-provider"
	data := &authjwt.TokenContextData{AuthUserId: authProviderId}
	ctx := context.WithValue(context.Background(), authjwt.TokenContextKey{}, data)
	// getUserAssociationByAuth0Id
	mockQuerier.On("GetUserAssociationByAuth0Id", ctx, mock.Anything, authProviderId).Return(user, nil)

	resp, err := service.GetUser(ctx, &connect.Request[mgmtv1alpha1.GetUserRequest]{})
	jsonF, _ := json.MarshalIndent(resp, "", " ")
	fmt.Printf("\n\n  %s \n\n", string(jsonF))

	fmt.Println(resp.Msg.GetUserId())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userId, resp.Msg.GetUserId())
}
