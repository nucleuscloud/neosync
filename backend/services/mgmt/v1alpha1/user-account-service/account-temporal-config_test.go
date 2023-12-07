package v1alpha1_useraccountservice

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt/auth0"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	fakeAccountId = "00000000-0000-0000-0000-000000000001"
)

func Test_Service_GetAccountTemporalConfig(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuth0MgmtClient := auth0.NewMockAuth0MgmtClientInterface(t)

	mockTfWfMgr := clientmanager.NewMockTemporalClientManagerClient(t)

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	mockTfWfMgr.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything).
		Return(&pg_models.TemporalConfig{Namespace: "foo", SyncJobQueueName: "foo-queue", Url: "localhost:1234"}, nil)

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier), mockAuth0MgmtClient, mockTfWfMgr)

	resp, err := service.GetAccountTemporalConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{
		AccountId: fakeAccountId,
	}))
	assert.Nil(t, err)
	assert.Equal(t, resp.Msg.Config, &mgmtv1alpha1.AccountTemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:1234",
	})
}

func Test_Service_GetAccountTemporalConfig_GetConfig_Err(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuth0MgmtClient := auth0.NewMockAuth0MgmtClientInterface(t)
	mockTfWfMgr := clientmanager.NewMockTemporalClientManagerClient(t)

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	mockTfWfMgr.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything).
		Return(nil, errors.New("test"))

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier), mockAuth0MgmtClient, mockTfWfMgr)

	resp, err := service.GetAccountTemporalConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{
		AccountId: fakeAccountId,
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_GetAccountTemporalConfig_NotInAccount(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuth0MgmtClient := auth0.NewMockAuth0MgmtClientInterface(t)

	mockTfWfMgr := clientmanager.NewMockTemporalClientManagerClient(t)

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil)

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier), mockAuth0MgmtClient, mockTfWfMgr)

	resp, err := service.GetAccountTemporalConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{
		AccountId: fakeAccountId,
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_GetAccountTemporalConfig_NotInAccount_Err(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuth0MgmtClient := auth0.NewMockAuth0MgmtClientInterface(t)

	mockTfWfMgr := clientmanager.NewMockTemporalClientManagerClient(t)

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(-1), errors.New("test"))

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier), mockAuth0MgmtClient, mockTfWfMgr)

	resp, err := service.GetAccountTemporalConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{
		AccountId: fakeAccountId,
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_SetAccountTemporalConfig(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuth0MgmtClient := auth0.NewMockAuth0MgmtClientInterface(t)

	mockTfWfMgr := clientmanager.NewMockTemporalClientManagerClient(t)

	mgmtTc := &mgmtv1alpha1.AccountTemporalConfig{
		Namespace:        "foo",
		Url:              "foo-url",
		SyncJobQueueName: "foo-queue",
	}

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	mockQuerier.On("UpdateTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(db_queries.NeosyncApiAccount{
		TemporalConfig: &pg_models.TemporalConfig{
			Namespace:        mgmtTc.Namespace,
			SyncJobQueueName: mgmtTc.SyncJobQueueName,
			Url:              mgmtTc.Url,
		},
	}, nil)

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier), mockAuth0MgmtClient, mockTfWfMgr)

	resp, err := service.SetAccountTemporalConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{
		AccountId: fakeAccountId,
		Config:    mgmtTc,
	}))
	assert.Nil(t, err)
	assert.Equal(t, resp.Msg.Config, mgmtTc)
}

func Test_Service_SetAccountTemporalConfig_Update_Err(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuth0MgmtClient := auth0.NewMockAuth0MgmtClientInterface(t)

	mockTfWfMgr := clientmanager.NewMockTemporalClientManagerClient(t)

	mgmtTc := &mgmtv1alpha1.AccountTemporalConfig{
		Namespace:        "foo",
		Url:              "foo-url",
		SyncJobQueueName: "foo-queue",
	}

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	mockQuerier.On("UpdateTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(db_queries.NeosyncApiAccount{}, errors.New("test"))

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier), mockAuth0MgmtClient, mockTfWfMgr)

	resp, err := service.SetAccountTemporalConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{
		AccountId: fakeAccountId,
		Config:    mgmtTc,
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_SetAccountTemporalConfig_NotInAccount(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuth0MgmtClient := auth0.NewMockAuth0MgmtClientInterface(t)

	mockTfWfMgr := clientmanager.NewMockTemporalClientManagerClient(t)

	mgmtTc := &mgmtv1alpha1.AccountTemporalConfig{}

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil)

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier), mockAuth0MgmtClient, mockTfWfMgr)

	resp, err := service.SetAccountTemporalConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{
		AccountId: fakeAccountId,
		Config:    mgmtTc,
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_SetAccountTemporalConfig_NotInAccount_Err(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuth0MgmtClient := auth0.NewMockAuth0MgmtClientInterface(t)

	mockTfWfMgr := clientmanager.NewMockTemporalClientManagerClient(t)

	mgmtTc := &mgmtv1alpha1.AccountTemporalConfig{}

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), errors.New("test"))

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier), mockAuth0MgmtClient, mockTfWfMgr)

	resp, err := service.SetAccountTemporalConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{
		AccountId: fakeAccountId,
		Config:    mgmtTc,
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func getAnonTestApiUser() *db_queries.NeosyncApiUser {
	return getTestApiUser("00000000-0000-0000-0000-000000000000")
}

func getTestApiUser(userId string) *db_queries.NeosyncApiUser {
	idUuid, _ := nucleusdb.ToUuid(userId)
	return &db_queries.NeosyncApiUser{ID: idUuid}
}
