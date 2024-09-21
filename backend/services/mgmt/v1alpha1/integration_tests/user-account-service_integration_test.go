package integrationtests_test

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v79"
)

var (
	testAuthUserId = "test-user"
	validAuthUser  = &authmgmt.User{Name: "foo", Email: "bar", Picture: "baz"}
)

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountOnboardingConfig() {
	accountId := s.createPersonalAccount(s.unauthdClients.users)

	resp, err := s.unauthdClients.users.GetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountOnboardingConfigRequest{AccountId: accountId}))
	requireNoErrResp(s.T(), resp, err)
	onboardingConfig := resp.Msg.GetConfig()
	require.NotNil(s.T(), onboardingConfig)

	require.False(s.T(), onboardingConfig.GetHasCreatedSourceConnection())
	require.False(s.T(), onboardingConfig.GetHasCreatedDestinationConnection())
	require.False(s.T(), onboardingConfig.GetHasCreatedJob())
	require.False(s.T(), onboardingConfig.GetHasInvitedMembers())
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountOnboardingConfig_NoAccount() {
	resp, err := s.unauthdClients.users.GetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountOnboardingConfigRequest{AccountId: uuid.NewString()}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountOnboardingConfig_NoAccount() {
	resp, err := s.unauthdClients.users.SetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{AccountId: uuid.NewString(), Config: &mgmtv1alpha1.AccountOnboardingConfig{}}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountOnboardingConfig_NoConfig() {
	accountId := s.createPersonalAccount(s.unauthdClients.users)

	resp, err := s.unauthdClients.users.SetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{AccountId: accountId, Config: nil}))
	requireNoErrResp(s.T(), resp, err)
	onboardingConfig := resp.Msg.GetConfig()
	require.NotNil(s.T(), onboardingConfig)

	require.False(s.T(), onboardingConfig.GetHasCreatedSourceConnection())
	require.False(s.T(), onboardingConfig.GetHasCreatedDestinationConnection())
	require.False(s.T(), onboardingConfig.GetHasCreatedJob())
	require.False(s.T(), onboardingConfig.GetHasInvitedMembers())
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountOnboardingConfig() {
	accountId := s.createPersonalAccount(s.unauthdClients.users)

	resp, err := s.unauthdClients.users.SetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{
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

var (
	validTemporalConfigModel = &pg_models.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "bar",
		Url:              "http://localhost:7070",
	}
)

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountTemporalConfig() {
	accountId := s.createPersonalAccount(s.unauthdClients.users)

	s.mocks.temporalClientManager.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything).
		Return(validTemporalConfigModel, nil)

	resp, err := s.unauthdClients.users.GetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{AccountId: accountId}))
	requireNoErrResp(s.T(), resp, err)

	tc := resp.Msg.GetConfig()
	require.NotNil(s.T(), tc)

	require.Equal(s.T(), validTemporalConfigModel.Namespace, tc.GetNamespace())
	require.Equal(s.T(), validTemporalConfigModel.SyncJobQueueName, tc.GetSyncJobQueueName())
	require.Equal(s.T(), validTemporalConfigModel.Url, tc.GetUrl())
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountTemporalConfig_NoAccount() {
	resp, err := s.unauthdClients.users.GetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{AccountId: uuid.NewString()}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountTemporalConfig_NeosyncCloud() {
	userclient := s.neosyncCloudClients.getUserClient(testAuthUserId)
	s.setUser(userclient)
	accountId := s.createPersonalAccount(userclient)

	resp, err := userclient.GetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{AccountId: accountId}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodeUnimplemented)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountTemporalConfig_NoAccount() {
	resp, err := s.unauthdClients.users.SetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{AccountId: uuid.NewString()}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountTemporalConfig_NeosyncCloud() {
	userclient := s.neosyncCloudClients.getUserClient(testAuthUserId)
	s.setUser(userclient)
	accountId := s.createPersonalAccount(userclient)

	resp, err := userclient.SetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{AccountId: accountId}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodeUnimplemented)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountTemporalConfig_NoConfig() {
	accountId := s.createPersonalAccount(s.unauthdClients.users)

	s.mocks.temporalClientManager.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything).
		Return(validTemporalConfigModel, nil)

	resp, err := s.unauthdClients.users.SetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{AccountId: accountId, Config: nil}))
	requireNoErrResp(s.T(), resp, err)

	tc := resp.Msg.GetConfig()
	require.NotNil(s.T(), tc)

	require.Equal(s.T(), validTemporalConfigModel.Namespace, tc.GetNamespace())
	require.Equal(s.T(), validTemporalConfigModel.SyncJobQueueName, tc.GetSyncJobQueueName())
	require.Equal(s.T(), validTemporalConfigModel.Url, tc.GetUrl())
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountTemporalConfig() {
	accountId := s.createPersonalAccount(s.unauthdClients.users)

	// kind of a bad test since we are mocking this client wholesale, but it at least verifies we can write the config
	s.mocks.temporalClientManager.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything).
		Return(validTemporalConfigModel, nil)

	resp, err := s.unauthdClients.users.SetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{
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

func (s *IntegrationTestSuite) Test_UserAccountService_GetUser_Auth() {
	client := s.authdClients.getUserClient("test-user1")
	userId := s.setUser(client)

	resp, err := client.GetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	requireNoErrResp(s.T(), resp, err)
	require.Equal(s.T(), userId, resp.Msg.GetUserId())
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetUser_Auth_NotFound() {
	client := s.authdClients.getUserClient(testAuthUserId)
	resp, err := client.GetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodeNotFound)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetUser_Auth() {
	client := s.authdClients.getUserClient(testAuthUserId)
	userId := s.setUser(client)
	require.NotEmpty(s.T(), userId)
	require.NotEqual(s.T(), "00000000-0000-0000-0000-000000000000", userId)
}

func (s *IntegrationTestSuite) Test_UserAccountService_CreateTeamAccount_Auth() {
	client := s.authdClients.getUserClient(testAuthUserId)
	s.setUser(client)

	resp, err := client.CreateTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateTeamAccountRequest{Name: "test-name"}))
	requireNoErrResp(s.T(), resp, err)
	require.NotEmpty(s.T(), resp.Msg.GetAccountId())
}

func (s *IntegrationTestSuite) Test_UserAccountService_CreateTeamAccount_NeosyncCloud() {
	client := s.neosyncCloudClients.getUserClient(testAuthUserId)
	s.setUser(client)

	s.mocks.billingclient.On("NewCustomer", mock.Anything).Once().
		Return(&stripe.Customer{ID: "test-stripe-id"}, nil)
	s.mocks.billingclient.On("NewCheckoutSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
		Return(&stripe.CheckoutSession{URL: "test-url"}, nil)

	resp, err := client.CreateTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateTeamAccountRequest{Name: "test-name"}))
	requireNoErrResp(s.T(), resp, err)
	require.NotEmpty(s.T(), resp.Msg.GetAccountId())
	require.Equal(s.T(), "test-url", resp.Msg.GetCheckoutSessionUrl())
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetTeamAccountMembers_Auth() {
	client := s.authdClients.getUserClient(testAuthUserId)
	userId := s.setUser(client)
	accountId := s.createTeamAccount(client, "test-team")

	s.mocks.authmanagerclient.On("GetUserBySub", mock.Anything, testAuthUserId).
		Return(validAuthUser, nil)

	resp, err := client.GetTeamAccountMembers(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetTeamAccountMembersRequest{AccountId: accountId}))
	requireNoErrResp(s.T(), resp, err)

	members := resp.Msg.GetUsers()
	require.Len(s.T(), members, 1)
	member := members[0]
	require.Equal(s.T(), userId, member.GetId())
	require.Equal(s.T(), validAuthUser.Email, member.GetEmail())
	require.Equal(s.T(), validAuthUser.Picture, member.GetImage())
	require.Equal(s.T(), validAuthUser.Name, member.GetName())
}

func (s *IntegrationTestSuite) setUser(client mgmtv1alpha1connect.UserAccountServiceClient) string {
	s.T().Helper()
	resp, err := client.SetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetUserRequest{}))
	requireNoErrResp(s.T(), resp, err)
	return resp.Msg.GetUserId()
}

func (s *IntegrationTestSuite) createTeamAccount(client mgmtv1alpha1connect.UserAccountServiceClient, name string) string {
	s.T().Helper()
	resp, err := client.CreateTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateTeamAccountRequest{Name: name}))
	requireNoErrResp(s.T(), resp, err)
	return resp.Msg.AccountId
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetUser() {
	resp, err := s.unauthdClients.users.GetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	requireNoErrResp(s.T(), resp, err)

	userId := resp.Msg.GetUserId()
	require.NotEmpty(s.T(), userId)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetUser() {
	resp, err := s.unauthdClients.users.SetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetUserRequest{}))
	requireNoErrResp(s.T(), resp, err)

	userId := resp.Msg.UserId
	require.NotEmpty(s.T(), userId)
	require.Equal(s.T(), "00000000-0000-0000-0000-000000000000", userId)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccounts_Empty() {
	resp, err := s.unauthdClients.users.GetUserAccounts(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserAccountsRequest{}))
	requireNoErrResp(s.T(), resp, err)

	accounts := resp.Msg.GetAccounts()
	require.Empty(s.T(), accounts)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetPersonalAccount() {
	resp, err := s.unauthdClients.users.SetPersonalAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetPersonalAccountRequest{}))
	requireNoErrResp(s.T(), resp, err)

	accountId := resp.Msg.GetAccountId()
	require.NotEmpty(s.T(), accountId)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccounts_NotEmpty() {
	s.createPersonalAccount(s.unauthdClients.users)

	accResp, err := s.unauthdClients.users.GetUserAccounts(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserAccountsRequest{}))
	requireNoErrResp(s.T(), accResp, err)

	accounts := accResp.Msg.GetAccounts()
	require.NotEmpty(s.T(), accounts)
	require.Len(s.T(), accounts, 1)
}

func (s *IntegrationTestSuite) Test_UserAccountService_IsUserInAccount() {
	accountId := s.createPersonalAccount(s.unauthdClients.users)

	resp, err := s.unauthdClients.users.IsUserInAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{
		AccountId: accountId,
	}))
	requireNoErrResp(s.T(), resp, err)
	require.True(s.T(), resp.Msg.GetOk())

	resp, err = s.unauthdClients.users.IsUserInAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{
		AccountId: uuid.NewString(),
	}))
	requireNoErrResp(s.T(), resp, err)
	require.False(s.T(), resp.Msg.GetOk())
}

func (s *IntegrationTestSuite) Test_UserAccountService_CreateTeamAccount_NoAuth() {
	s.createPersonalAccount(s.unauthdClients.users)

	resp, err := s.unauthdClients.users.CreateTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateTeamAccountRequest{Name: "test-name"}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetTeamAccountMembers_NoAuth_Personal() {
	accountId := s.createPersonalAccount(s.unauthdClients.users)

	resp, err := s.unauthdClients.users.GetTeamAccountMembers(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetTeamAccountMembersRequest{AccountId: accountId}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_RemoveTeamAccountMember_NoAuth_Personal() {
	accountId := s.createPersonalAccount(s.unauthdClients.users)

	resp, err := s.unauthdClients.users.RemoveTeamAccountMember(s.ctx, connect.NewRequest(&mgmtv1alpha1.RemoveTeamAccountMemberRequest{AccountId: accountId, UserId: uuid.NewString()}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_InviteUserToTeamAccount_NoAuth_Personal() {
	accountId := s.createPersonalAccount(s.unauthdClients.users)

	resp, err := s.unauthdClients.users.InviteUserToTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.InviteUserToTeamAccountRequest{AccountId: accountId, Email: "test@example.com"}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetTeamAccountInvites_NoAuth_Personal() {
	accountId := s.createPersonalAccount(s.unauthdClients.users)

	resp, err := s.unauthdClients.users.GetTeamAccountInvites(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetTeamAccountInvitesRequest{AccountId: accountId}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_RemoveTeamAccountInvite_NoAuth_Personal() {
	resp, err := s.unauthdClients.users.RemoveTeamAccountInvite(s.ctx, connect.NewRequest(&mgmtv1alpha1.RemoveTeamAccountInviteRequest{Id: uuid.NewString()}))
	requireNoErrResp(s.T(), resp, err)
}

func (s *IntegrationTestSuite) Test_UserAccountService_AcceptTeamAccountInvite_NoAuth_Personal() {
	resp, err := s.unauthdClients.users.AcceptTeamAccountInvite(s.ctx, connect.NewRequest(&mgmtv1alpha1.AcceptTeamAccountInviteRequest{Token: uuid.NewString()}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodeUnauthenticated)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetSystemInformation() {
	resp, err := s.unauthdClients.users.GetSystemInformation(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetSystemInformationRequest{}))
	requireNoErrResp(s.T(), resp, err)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountStatus_NeosyncCloud_Personal() {
	userclient := s.neosyncCloudClients.getUserClient(testAuthUserId)
	s.setUser(userclient)
	accountId := s.createPersonalAccount(userclient)

	err := s.setMaxAllowedRecords(s.ctx, accountId, 100)
	require.NoError(s.T(), err)

	s.mocks.prometheusclient.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Once().
		Return(model.Vector{{
			Value:     2,
			Timestamp: 0,
		}}, promv1.Warnings{}, nil)

	resp, err := userclient.GetAccountStatus(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountStatusRequest{
		AccountId: accountId,
	}))
	requireNoErrResp(s.T(), resp, err)

	require.Equal(s.T(), resp.Msg.GetUsedRecordCount(), uint64(2))
	require.NotNil(s.T(), resp.Msg.AllowedRecordCount)
	require.Equal(s.T(), uint64(100), resp.Msg.GetAllowedRecordCount())
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountStatus_NeosyncCloud_Personal_Unlimited() {
	userclient := s.neosyncCloudClients.getUserClient(testAuthUserId)
	s.setUser(userclient)
	accountId := s.createPersonalAccount(userclient)

	resp, err := userclient.GetAccountStatus(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountStatusRequest{
		AccountId: accountId,
	}))
	requireNoErrResp(s.T(), resp, err)

	require.Equal(s.T(), resp.Msg.GetUsedRecordCount(), uint64(0))
	require.Nil(s.T(), resp.Msg.AllowedRecordCount)
}

type testSubscriptionIter struct {
	subscriptions []*stripe.Subscription
	current       int
}

func (t *testSubscriptionIter) Next() bool {
	return t.current < len(t.subscriptions)
}
func (t *testSubscriptionIter) Subscription() *stripe.Subscription {
	sub := t.subscriptions[t.current]
	t.current++
	return sub
}
func (t *testSubscriptionIter) Err() error {
	return nil
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountStatus_NeosyncCloud_Billed() {
	userclient := s.neosyncCloudClients.getUserClient(testAuthUserId)
	s.setUser(userclient)

	t := s.T()

	t.Run("active sub", func(t *testing.T) {
		custId := "cust_id1"
		accountId := s.createBilledTeamAccount(userclient, "test-team", custId)
		s.mocks.billingclient.On("GetSubscriptions", custId).Once().Return(&testSubscriptionIter{subscriptions: []*stripe.Subscription{
			{Status: stripe.SubscriptionStatusIncompleteExpired},
			{Status: stripe.SubscriptionStatusActive},
		}}, nil)

		resp, err := userclient.GetAccountStatus(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountStatusRequest{
			AccountId: accountId,
		}))
		requireNoErrResp(s.T(), resp, err)

		require.Equal(s.T(), mgmtv1alpha1.BillingStatus_BILLING_STATUS_ACTIVE, resp.Msg.GetSubscriptionStatus())
	})

	t.Run("no active subscriptions", func(t *testing.T) {
		custId := "cust_id2"
		accountId := s.createBilledTeamAccount(userclient, "test-team1", custId)
		s.mocks.billingclient.On("GetSubscriptions", custId).Once().Return(&testSubscriptionIter{subscriptions: []*stripe.Subscription{
			{Status: stripe.SubscriptionStatusIncompleteExpired},
			{Status: stripe.SubscriptionStatusIncompleteExpired},
		}}, nil)

		resp, err := userclient.GetAccountStatus(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountStatusRequest{
			AccountId: accountId,
		}))
		requireNoErrResp(s.T(), resp, err)

		require.Equal(s.T(), mgmtv1alpha1.BillingStatus_BILLING_STATUS_EXPIRED, resp.Msg.GetSubscriptionStatus())
	})
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountStatus_OSS_Personal() {
	accountId := s.createPersonalAccount(s.unauthdClients.users)

	resp, err := s.unauthdClients.users.GetAccountStatus(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountStatusRequest{
		AccountId: accountId,
	}))
	requireNoErrResp(s.T(), resp, err)

	require.Equal(s.T(), resp.Msg.GetUsedRecordCount(), uint64(0))
	require.Nil(s.T(), resp.Msg.AllowedRecordCount)
}

func (s *IntegrationTestSuite) Test_UserAccountService_IsAccountStatusValid_NeosyncCloud_Personal() {
	userclient := s.neosyncCloudClients.getUserClient(testAuthUserId)
	s.setUser(userclient)
	accountId := s.createPersonalAccount(userclient)

	err := s.setMaxAllowedRecords(s.ctx, accountId, 100)
	require.NoError(s.T(), err)

	s.mocks.prometheusclient.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Once().
		Return(model.Vector{{
			Value:     2,
			Timestamp: 0,
		}}, promv1.Warnings{}, nil)

	resp, err := userclient.IsAccountStatusValid(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
		AccountId: accountId,
	}))
	requireNoErrResp(s.T(), resp, err)

	require.True(s.T(), resp.Msg.GetIsValid())
	require.Empty(s.T(), resp.Msg.GetReason())
}

func (s *IntegrationTestSuite) Test_UserAccountService_IsAccountStatusValid_NeosyncCloud_Personal_Overprovisioned() {
	userclient := s.neosyncCloudClients.getUserClient(testAuthUserId)
	s.setUser(userclient)
	accountId := s.createPersonalAccount(userclient)

	err := s.setMaxAllowedRecords(s.ctx, accountId, 100)
	require.NoError(s.T(), err)

	s.mocks.prometheusclient.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Once().
		Return(model.Vector{{
			Value:     100,
			Timestamp: 0,
		}}, promv1.Warnings{}, nil)

	resp, err := userclient.IsAccountStatusValid(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
		AccountId: accountId,
	}))
	requireNoErrResp(s.T(), resp, err)

	require.False(s.T(), resp.Msg.GetIsValid())
	require.NotEmpty(s.T(), resp.Msg.GetReason())
}

func (s *IntegrationTestSuite) Test_UserAccountService_IsAccountStatusValid_NeosyncCloud_Personal_RequestedRecords() {
	userclient := s.neosyncCloudClients.getUserClient(testAuthUserId)
	s.setUser(userclient)
	accountId := s.createPersonalAccount(userclient)

	err := s.setMaxAllowedRecords(s.ctx, accountId, 100)
	require.NoError(s.T(), err)

	s.mocks.prometheusclient.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Twice().
		Return(model.Vector{{
			Value:     50,
			Timestamp: 0,
		}}, promv1.Warnings{}, nil)

	t := s.T()
	t.Run("over the limit", func(t *testing.T) {
		resp, err := userclient.IsAccountStatusValid(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
			AccountId:            accountId,
			RequestedRecordCount: shared.Ptr(uint64(51)), // puts the user one over the limit
		}))
		requireNoErrResp(s.T(), resp, err)

		require.False(s.T(), resp.Msg.GetIsValid())
		require.NotEmpty(s.T(), resp.Msg.GetReason())
	})
	t.Run("under the limit", func(t *testing.T) {
		resp, err := userclient.IsAccountStatusValid(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
			AccountId:            accountId,
			RequestedRecordCount: shared.Ptr(uint64(50)),
		}))
		requireNoErrResp(s.T(), resp, err)

		require.True(s.T(), resp.Msg.GetIsValid())
		require.Empty(s.T(), resp.Msg.GetReason())
	})
}

func (s *IntegrationTestSuite) Test_UserAccountService_IsAccountStatusValid_NeosyncCloud_Billed() {
	userclient := s.neosyncCloudClients.getUserClient(testAuthUserId)
	s.setUser(userclient)

	t := s.T()
	t.Run("active", func(t *testing.T) {
		custId := "cust_id1"
		accountId := s.createBilledTeamAccount(userclient, "test1", custId)
		s.mocks.billingclient.On("GetSubscriptions", custId).Once().Return(&testSubscriptionIter{subscriptions: []*stripe.Subscription{
			{Status: stripe.SubscriptionStatusActive},
		}}, nil)
		resp, err := userclient.IsAccountStatusValid(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
			AccountId: accountId,
		}))
		requireNoErrResp(s.T(), resp, err)

		assert.True(s.T(), resp.Msg.GetIsValid())
		assert.Empty(s.T(), resp.Msg.GetReason())
	})
	t.Run("inactive", func(t *testing.T) {
		custId := "cust_id2"
		accountId := s.createBilledTeamAccount(userclient, "test2", custId)
		s.mocks.billingclient.On("GetSubscriptions", custId).Once().Return(&testSubscriptionIter{subscriptions: []*stripe.Subscription{
			{Status: stripe.SubscriptionStatusIncompleteExpired},
		}}, nil)

		resp, err := userclient.IsAccountStatusValid(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
			AccountId: accountId,
		}))
		requireNoErrResp(s.T(), resp, err)

		assert.False(s.T(), resp.Msg.GetIsValid())
		assert.NotEmpty(s.T(), resp.Msg.GetReason())
	})
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountBillingCheckoutSession() {
	userclient := s.neosyncCloudClients.getUserClient(testAuthUserId)
	s.setUser(userclient)

	t := s.T()

	t.Run("billed account - allowed", func(t *testing.T) {
		teamAccountId := s.createBilledTeamAccount(userclient, "test-team", "test-stripe-id")

		s.mocks.billingclient.On("NewCheckoutSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
			Return(&stripe.CheckoutSession{URL: "new-test-url"}, nil)

		resp, err := userclient.GetAccountBillingCheckoutSession(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountBillingCheckoutSessionRequest{
			AccountId: teamAccountId,
		}))
		requireNoErrResp(s.T(), resp, err)
		require.Equal(s.T(), "new-test-url", resp.Msg.GetCheckoutSessionUrl())
	})

	t.Run("personal account - disallowed", func(t *testing.T) {
		personalAccountId := s.createPersonalAccount(userclient)
		resp, err := userclient.GetAccountBillingCheckoutSession(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountBillingCheckoutSessionRequest{
			AccountId: personalAccountId,
		}))
		requireErrResp(s.T(), resp, err)
	})

	t.Run("non-neosynccloud - disallowed", func(t *testing.T) {
		personalAccountId := s.createPersonalAccount(s.unauthdClients.users)
		resp, err := userclient.GetAccountBillingCheckoutSession(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountBillingCheckoutSessionRequest{
			AccountId: personalAccountId,
		}))
		requireErrResp(s.T(), resp, err)
	})
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountBillingPortalSession() {
	userclient := s.neosyncCloudClients.getUserClient(testAuthUserId)
	s.setUser(userclient)

	t := s.T()

	t.Run("billed account - allowed", func(t *testing.T) {
		teamAccountId := s.createBilledTeamAccount(userclient, "test-team", "test-stripe-id")

		s.mocks.billingclient.On("NewBillingPortalSession", mock.Anything, mock.Anything).Once().
			Return(&stripe.BillingPortalSession{URL: "new-test-url"}, nil)

		resp, err := userclient.GetAccountBillingPortalSession(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountBillingPortalSessionRequest{
			AccountId: teamAccountId,
		}))
		requireNoErrResp(s.T(), resp, err)
		require.Equal(s.T(), "new-test-url", resp.Msg.GetPortalSessionUrl())
	})

	t.Run("personal account - disallowed", func(t *testing.T) {
		personalAccountId := s.createPersonalAccount(userclient)
		resp, err := userclient.GetAccountBillingPortalSession(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountBillingPortalSessionRequest{
			AccountId: personalAccountId,
		}))
		requireErrResp(s.T(), resp, err)
	})

	t.Run("non-neosynccloud - disallowed", func(t *testing.T) {
		personalAccountId := s.createPersonalAccount(s.unauthdClients.users)
		resp, err := s.unauthdClients.users.GetAccountBillingPortalSession(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountBillingPortalSessionRequest{
			AccountId: personalAccountId,
		}))
		requireErrResp(s.T(), resp, err)
	})
}

func (s *IntegrationTestSuite) createBilledTeamAccount(client mgmtv1alpha1connect.UserAccountServiceClient, name, stripeCustomerId string) string {
	s.mocks.billingclient.On("NewCustomer", mock.Anything).Once().
		Return(&stripe.Customer{ID: stripeCustomerId}, nil)
	s.mocks.billingclient.On("NewCheckoutSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
		Return(&stripe.CheckoutSession{URL: "test-url"}, nil)
	return s.createTeamAccount(client, name)
}
