package v1alpha1_useraccountservice

import (
	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	validTemporalConfigModel = &pg_models.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "bar",
		Url:              "http://localhost:7070",
	}
)

func (s *IntegrationTestSuite) Test_GetAccountTemporalConfig() {
	accountId := s.createPersonalAccount()

	s.mockTemporalClientMgr.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything).
		Return(validTemporalConfigModel, nil)

	resp, err := s.unauthUserClient.GetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{AccountId: accountId}))
	requireNoErrResp(s.T(), resp, err)

	tc := resp.Msg.GetConfig()
	require.NotNil(s.T(), tc)

	require.Equal(s.T(), validTemporalConfigModel.Namespace, tc.GetNamespace())
	require.Equal(s.T(), validTemporalConfigModel.SyncJobQueueName, tc.GetSyncJobQueueName())
	require.Equal(s.T(), validTemporalConfigModel.Url, tc.GetUrl())
}

func (s *IntegrationTestSuite) Test_GetAccountTemporalConfig_NoAccount() {
	resp, err := s.unauthUserClient.GetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{AccountId: uuid.NewString()}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_GetAccountTemporalConfig_NeosyncCloud() {
	accountId := s.createPersonalAccount()

	resp, err := s.ncunauthUserClient.GetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{AccountId: accountId}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodeUnimplemented)
}

func (s *IntegrationTestSuite) Test_SetAccountTemporalConfig_NoAccount() {
	resp, err := s.unauthUserClient.SetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{AccountId: uuid.NewString()}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_SetAccountTemporalConfig_NeosyncCloud() {
	accountId := s.createPersonalAccount()

	resp, err := s.ncunauthUserClient.SetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{AccountId: accountId}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodeUnimplemented)
}

func (s *IntegrationTestSuite) Test_SetAccountTemporalConfig_NoConfig() {
	accountId := s.createPersonalAccount()

	s.mockTemporalClientMgr.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything).
		Return(validTemporalConfigModel, nil)

	resp, err := s.unauthUserClient.SetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{AccountId: accountId, Config: nil}))
	requireNoErrResp(s.T(), resp, err)

	tc := resp.Msg.GetConfig()
	require.NotNil(s.T(), tc)

	require.Equal(s.T(), validTemporalConfigModel.Namespace, tc.GetNamespace())
	require.Equal(s.T(), validTemporalConfigModel.SyncJobQueueName, tc.GetSyncJobQueueName())
	require.Equal(s.T(), validTemporalConfigModel.Url, tc.GetUrl())
}

func (s *IntegrationTestSuite) Test_SetAccountTemporalConfig() {
	accountId := s.createPersonalAccount()

	// kind of a bad test since we are mocking this client wholesale, but it at least verifies we can write the config
	s.mockTemporalClientMgr.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything).
		Return(validTemporalConfigModel, nil)

	resp, err := s.unauthUserClient.SetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{
		AccountId: accountId, Config: &mgmtv1alpha1.AccountTemporalConfig{
			Url:              "test",
			Namespace:        "test",
			SyncJobQueueName: "test",
		}}))
	requireNoErrResp(s.T(), resp, err)

	tc := resp.Msg.GetConfig()
	require.NotNil(s.T(), tc)

	require.Equal(s.T(), validTemporalConfigModel.Namespace, tc.GetNamespace())
	require.Equal(s.T(), validTemporalConfigModel.SyncJobQueueName, tc.GetSyncJobQueueName())
	require.Equal(s.T(), validTemporalConfigModel.Url, tc.GetUrl())
}
