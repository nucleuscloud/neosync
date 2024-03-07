package v1alpha1_useraccountservice

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Service_GetAccountOnboardingConfig(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuthClient := auth_client.NewMockInterface(t)
	mockAuthAdminClient := authmgmt.NewMockInterface(t)
	mockTfWfMgr := clientmanager.NewMockTemporalClientManagerClient(t)

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	mockQuerier.On("GetAccountOnboardingConfig", context.Background(), mock.Anything, mock.Anything).Return(&pg_models.AccountOnboardingConfig{
		HasCreatedSourceConnection:      true,
		HasCreatedDestinationConnection: true,
		HasCreatedJob:                   true,
		HasInvitedMembers:               true,
	}, nil)

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier), mockTfWfMgr, mockAuthClient, mockAuthAdminClient)

	resp, err := service.GetAccountOnboardingConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountOnboardingConfigRequest{
		AccountId: fakeAccountId,
	}))

	assert.Nil(t, err)
	assert.Equal(t, resp.Msg.Config, &mgmtv1alpha1.AccountOnboardingConfig{
		HasCreatedSourceConnection:      true,
		HasCreatedDestinationConnection: true,
		HasCreatedJob:                   true,
		HasInvitedMembers:               true,
	})
}

func Test_Service_GetAccountOnboardingConfig_GetConfig_Err(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuthClient := auth_client.NewMockInterface(t)
	mockAuthAdminClient := authmgmt.NewMockInterface(t)
	mockTfWfMgr := clientmanager.NewMockTemporalClientManagerClient(t)

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	mockQuerier.On("GetAccountOnboardingConfig", context.Background(), mock.Anything, mock.Anything).
		Return(nil, errors.New("test"))

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier), mockTfWfMgr, mockAuthClient, mockAuthAdminClient)

	resp, err := service.GetAccountOnboardingConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountOnboardingConfigRequest{
		AccountId: fakeAccountId,
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_GetAccountOnboardingConfig_NotInAccount(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuthClient := auth_client.NewMockInterface(t)
	mockAuthAdminClient := authmgmt.NewMockInterface(t)
	mockTfWfMgr := clientmanager.NewMockTemporalClientManagerClient(t)

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil)

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier), mockTfWfMgr, mockAuthClient, mockAuthAdminClient)

	resp, err := service.GetAccountOnboardingConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetAccountOnboardingConfigRequest{
		AccountId: fakeAccountId,
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_SetAccountOnboardingConfig(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuthClient := auth_client.NewMockInterface(t)
	mockAuthAdminClient := authmgmt.NewMockInterface(t)
	mockTfWfMgr := clientmanager.NewMockTemporalClientManagerClient(t)

	obConfig := &mgmtv1alpha1.AccountOnboardingConfig{
		HasCreatedSourceConnection:      true,
		HasCreatedDestinationConnection: true,
		HasCreatedJob:                   true,
		HasInvitedMembers:               true,
	}

	mockQuerier.On("GetAnonymousUser", mock.Anything, mock.Anything).Return(*getAnonTestApiUser(), nil)
	mockQuerier.On("IsUserInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	mockQuerier.On("UpdateAccountOnboardingConfig", mock.Anything, mock.Anything, mock.Anything).Return(db_queries.NeosyncApiAccount{
		OnboardingConfig: &pg_models.AccountOnboardingConfig{
			HasCreatedSourceConnection:      obConfig.HasCreatedSourceConnection,
			HasCreatedDestinationConnection: obConfig.HasCreatedDestinationConnection,
			HasCreatedJob:                   obConfig.HasCreatedJob,
			HasInvitedMembers:               obConfig.HasInvitedMembers,
		},
	}, nil)

	service := New(&Config{IsAuthEnabled: false}, nucleusdb.New(mockDbtx, mockQuerier), mockTfWfMgr, mockAuthClient, mockAuthAdminClient)

	resp, err := service.SetAccountOnboardingConfig(context.Background(), connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{
		AccountId: fakeAccountId,
		Config:    obConfig,
	}))
	assert.Nil(t, err)
	assert.Equal(t, resp.Msg.Config, obConfig)
}
