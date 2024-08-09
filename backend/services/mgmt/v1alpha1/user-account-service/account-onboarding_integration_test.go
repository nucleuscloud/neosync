package v1alpha1_useraccountservice

import (
	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetAccountOnboardingConfig() {
	accountId := s.createPersonalAccount()

	resp, err := s.userclient.GetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountOnboardingConfigRequest{AccountId: accountId}))
	requireNoErrResp(s.T(), resp, err)
	onboardingConfig := resp.Msg.GetConfig()
	require.NotNil(s.T(), onboardingConfig)

	require.False(s.T(), onboardingConfig.GetHasCreatedSourceConnection())
	require.False(s.T(), onboardingConfig.GetHasCreatedDestinationConnection())
	require.False(s.T(), onboardingConfig.GetHasCreatedJob())
	require.False(s.T(), onboardingConfig.GetHasInvitedMembers())
}

func (s *IntegrationTestSuite) Test_GetAccountOnboardingConfig_NoAccount() {
	resp, err := s.userclient.GetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountOnboardingConfigRequest{AccountId: uuid.NewString()}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_SetAccountOnboardingConfig_NoAccount() {
	resp, err := s.userclient.SetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{AccountId: uuid.NewString(), Config: &mgmtv1alpha1.AccountOnboardingConfig{}}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_SetAccountOnboardingConfig_NoConfig() {
	accountId := s.createPersonalAccount()

	resp, err := s.userclient.SetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{AccountId: accountId, Config: nil}))
	requireNoErrResp(s.T(), resp, err)
	onboardingConfig := resp.Msg.GetConfig()
	require.NotNil(s.T(), onboardingConfig)

	require.False(s.T(), onboardingConfig.GetHasCreatedSourceConnection())
	require.False(s.T(), onboardingConfig.GetHasCreatedDestinationConnection())
	require.False(s.T(), onboardingConfig.GetHasCreatedJob())
	require.False(s.T(), onboardingConfig.GetHasInvitedMembers())
}

func (s *IntegrationTestSuite) Test_SetAccountOnboardingConfig() {
	accountId := s.createPersonalAccount()

	resp, err := s.userclient.SetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{
		AccountId: accountId, Config: &mgmtv1alpha1.AccountOnboardingConfig{
			HasCreatedSourceConnection:      true,
			HasCreatedDestinationConnection: true,
			HasCreatedJob:                   true,
			HasInvitedMembers:               true,
		}},
	))
	requireNoErrResp(s.T(), resp, err)

	onboardingConfig := resp.Msg.GetConfig()
	require.NotNil(s.T(), onboardingConfig)

	require.True(s.T(), onboardingConfig.GetHasCreatedSourceConnection())
	require.True(s.T(), onboardingConfig.GetHasCreatedDestinationConnection())
	require.True(s.T(), onboardingConfig.GetHasCreatedJob())
	require.True(s.T(), onboardingConfig.GetHasInvitedMembers())
}
