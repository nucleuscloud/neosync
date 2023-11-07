package v1alpha1_useraccountservice

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	jsonmodels "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	fakeAccountId = "00000000-0000-0000-0000-000000000001"
)

func Test_Service_GetAccountTemporalConfig(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("GetTemporalConfigByUserAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(&jsonmodels.TemporalConfig{Namespace: "foo", SyncJobQueueName: "foo-queue", Url: "localhost:1234"}, nil)

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier))

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

func Test_Service_GetAccountTemporalConfig_Defaults(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("GetTemporalConfigByUserAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, pgx.ErrNoRows)

	service := New(&Config{IsAuthEnabled: false, Temporal: &TemporalConfig{
		DefaultTemporalNamespace:        "default-ns",
		DefaultTemporalSyncJobQueueName: "default-sync",
		DefaultTemporalUrl:              "default-url",
	}}, nucleusdb.New(mockDbtx, mockQuerier))

	resp, err := service.GetAccountTemporalConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{
		AccountId: fakeAccountId,
	}))
	assert.Nil(t, err)
	assert.Equal(t, resp.Msg.Config, &mgmtv1alpha1.AccountTemporalConfig{
		Namespace:        "default-ns",
		SyncJobQueueName: "default-sync",
		Url:              "default-url",
	})
}

func Test_Service_SetAccountTemporalConfig(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	mgmtTc := &mgmtv1alpha1.AccountTemporalConfig{
		Namespace:        "foo",
		Url:              "foo-url",
		SyncJobQueueName: "foo-queue",
	}

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	mockQuerier.On("UpdateTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(db_queries.NeosyncApiAccount{
		TemporalConfig: &jsonmodels.TemporalConfig{
			Namespace:        mgmtTc.Namespace,
			SyncJobQueueName: mgmtTc.SyncJobQueueName,
			Url:              mgmtTc.Url,
		},
	}, nil)

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier))

	resp, err := service.SetAccountTemporalConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{
		AccountId: fakeAccountId,
		Config:    mgmtTc,
	}))
	assert.Nil(t, err)
	assert.Equal(t, resp.Msg.Config, mgmtTc)
}

func Test_Service_SetAccountTemporalConfig_NotInAccount(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	mgmtTc := &mgmtv1alpha1.AccountTemporalConfig{}

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil)

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier))

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
