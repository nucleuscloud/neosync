package integrationtests_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/temporal/clientmanager"
	integrationtests_test "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v79"
)

var (
	testAuthUserId  = "test-user"
	validAuthUser   = &authmgmt.User{Name: "foo", Email: "bar", Picture: "baz"}
	testAuthUserId2 = "test-user2"
)

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountOnboardingConfig() {
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().GetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountOnboardingConfigRequest{AccountId: accountId}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)
	onboardingConfig := resp.Msg.GetConfig()
	require.NotNil(s.T(), onboardingConfig)

	require.False(s.T(), onboardingConfig.GetHasCompletedOnboarding())
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountOnboardingConfig_NoAccount() {
	resp, err := s.OSSUnauthenticatedLicensedClients.Users().GetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountOnboardingConfigRequest{AccountId: uuid.NewString()}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountOnboardingConfig_NoAccount() {
	resp, err := s.OSSUnauthenticatedLicensedClients.Users().SetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{AccountId: uuid.NewString(), Config: &mgmtv1alpha1.AccountOnboardingConfig{}}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountOnboardingConfig_NoConfig() {
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().SetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{AccountId: accountId, Config: nil}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)
	onboardingConfig := resp.Msg.GetConfig()
	require.NotNil(s.T(), onboardingConfig)

	require.False(s.T(), onboardingConfig.GetHasCompletedOnboarding())
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountOnboardingConfig() {
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().SetAccountOnboardingConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{
		AccountId: accountId, Config: &mgmtv1alpha1.AccountOnboardingConfig{
			HasCompletedOnboarding: true,
		}},
	))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)

	onboardingConfig := resp.Msg.GetConfig()
	require.NotNil(s.T(), onboardingConfig)

	require.True(s.T(), onboardingConfig.GetHasCompletedOnboarding())
}

var (
	validTemporalConfigModel = &pg_models.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "bar",
		Url:              "http://localhost:7070",
	}
	validTemporalConfig = &clientmanager.TemporalConfig{
		Url:              validTemporalConfigModel.Url,
		Namespace:        validTemporalConfigModel.Namespace,
		SyncJobQueueName: validTemporalConfigModel.SyncJobQueueName,
	}
)

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountTemporalConfig() {
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	s.Mocks.TemporalConfigProvider.On("GetConfig", mock.Anything, mock.Anything).
		Return(validTemporalConfig, nil)

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().GetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{AccountId: accountId}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)

	tc := resp.Msg.GetConfig()
	require.NotNil(s.T(), tc)

	require.Equal(s.T(), validTemporalConfig.Namespace, tc.GetNamespace())
	require.Equal(s.T(), validTemporalConfig.SyncJobQueueName, tc.GetSyncJobQueueName())
	require.Equal(s.T(), validTemporalConfig.Url, tc.GetUrl())
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountTemporalConfig_NoAccount() {
	resp, err := s.OSSUnauthenticatedLicensedClients.Users().GetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{AccountId: uuid.NewString()}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountTemporalConfig_NeosyncCloud() {
	userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	s.setUser(s.ctx, userclient)
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), userclient)

	resp, err := userclient.GetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{AccountId: accountId}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodeUnimplemented)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountTemporalConfig_NoAccount() {
	resp, err := s.OSSUnauthenticatedLicensedClients.Users().SetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{AccountId: uuid.NewString()}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountTemporalConfig_NeosyncCloud() {
	userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	s.setUser(s.ctx, userclient)
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), userclient)

	resp, err := userclient.SetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{AccountId: accountId}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodeUnimplemented)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountTemporalConfig_NoConfig() {
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	s.Mocks.TemporalConfigProvider.On("GetConfig", mock.Anything, mock.Anything).
		Return(validTemporalConfig, nil)

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().SetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{AccountId: accountId, Config: nil}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)

	tc := resp.Msg.GetConfig()
	require.NotNil(s.T(), tc)

	require.Equal(s.T(), validTemporalConfig.Namespace, tc.GetNamespace())
	require.Equal(s.T(), validTemporalConfig.SyncJobQueueName, tc.GetSyncJobQueueName())
	require.Equal(s.T(), validTemporalConfig.Url, tc.GetUrl())
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetAccountTemporalConfig() {
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	// kind of a bad test since we are mocking this client wholesale, but it at least verifies we can write the config
	s.Mocks.TemporalConfigProvider.On("GetConfig", mock.Anything, mock.Anything).
		Return(validTemporalConfig, nil)

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().SetAccountTemporalConfig(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountTemporalConfigRequest{
		AccountId: accountId, Config: &mgmtv1alpha1.AccountTemporalConfig{
			Url:              "test",
			Namespace:        "test",
			SyncJobQueueName: "test",
		}}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)

	tc := resp.Msg.GetConfig()
	require.NotNil(s.T(), tc)

	require.Equal(s.T(), validTemporalConfig.Namespace, tc.GetNamespace())
	require.Equal(s.T(), validTemporalConfig.SyncJobQueueName, tc.GetSyncJobQueueName())
	require.Equal(s.T(), validTemporalConfig.Url, tc.GetUrl())
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetUser_Auth() {
	client := s.OSSAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId("test-user1"))
	userId := s.setUser(s.ctx, client)

	resp, err := client.GetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)
	require.Equal(s.T(), userId, resp.Msg.GetUserId())
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetUser_Auth_NotFound() {
	client := s.OSSAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId("test-user1"))
	resp, err := client.GetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodeNotFound)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetUser_Auth() {
	client := s.OSSAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	userId := s.setUser(s.ctx, client)
	require.NotEmpty(s.T(), userId)
	require.NotEqual(s.T(), "00000000-0000-0000-0000-000000000000", userId)
}

func (s *IntegrationTestSuite) Test_UserAccountService_CreateTeamAccount_Auth() {
	client := s.OSSAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	s.setUser(s.ctx, client)

	resp, err := client.CreateTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateTeamAccountRequest{Name: "test-name"}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)
	require.NotEmpty(s.T(), resp.Msg.GetAccountId())
}

func (s *IntegrationTestSuite) Test_UserAccountService_CreateTeamAccount_NeosyncCloud() {
	client := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	s.setUser(s.ctx, client)

	s.Mocks.Billingclient.On("NewCustomer", mock.Anything).Once().
		Return(&stripe.Customer{ID: "test-stripe-id"}, nil)
	s.Mocks.Billingclient.On("NewCheckoutSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
		Return(&stripe.CheckoutSession{URL: "test-url"}, nil)

	resp, err := client.CreateTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateTeamAccountRequest{Name: "test-name"}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)
	require.NotEmpty(s.T(), resp.Msg.GetAccountId())
	require.Equal(s.T(), "test-url", resp.Msg.GetCheckoutSessionUrl())
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetTeamAccountMembers_Auth() {
	client := s.OSSAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	userId := s.setUser(s.ctx, client)
	accountId := s.createTeamAccount(s.ctx, client, "test-team")

	s.Mocks.Authmanagerclient.On("GetUserBySub", mock.Anything, testAuthUserId).
		Return(validAuthUser, nil)

	resp, err := client.GetTeamAccountMembers(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetTeamAccountMembersRequest{AccountId: accountId}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)

	members := resp.Msg.GetUsers()
	require.Len(s.T(), members, 1)
	member := members[0]
	require.Equal(s.T(), userId, member.GetId())
	require.Equal(s.T(), validAuthUser.Email, member.GetEmail())
	require.Equal(s.T(), validAuthUser.Picture, member.GetImage())
	require.Equal(s.T(), validAuthUser.Name, member.GetName())
}

func (s *IntegrationTestSuite) setUser(ctx context.Context, client mgmtv1alpha1connect.UserAccountServiceClient) string {
	s.T().Helper()
	resp, err := client.SetUser(ctx, connect.NewRequest(&mgmtv1alpha1.SetUserRequest{}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)
	return resp.Msg.GetUserId()
}

func (s *IntegrationTestSuite) createTeamAccount(ctx context.Context, client mgmtv1alpha1connect.UserAccountServiceClient, name string) string {
	s.T().Helper()
	resp, err := client.CreateTeamAccount(ctx, connect.NewRequest(&mgmtv1alpha1.CreateTeamAccountRequest{Name: name}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)
	return resp.Msg.AccountId
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetUser() {
	resp, err := s.OSSUnauthenticatedLicensedClients.Users().GetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)

	userId := resp.Msg.GetUserId()
	require.NotEmpty(s.T(), userId)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetUser() {
	resp, err := s.OSSUnauthenticatedLicensedClients.Users().SetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetUserRequest{}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)

	userId := resp.Msg.UserId
	require.NotEmpty(s.T(), userId)
	require.Equal(s.T(), "00000000-0000-0000-0000-000000000000", userId)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccounts_Empty() {
	resp, err := s.OSSUnauthenticatedLicensedClients.Users().GetUserAccounts(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserAccountsRequest{}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)

	accounts := resp.Msg.GetAccounts()
	require.Empty(s.T(), accounts)
}

func (s *IntegrationTestSuite) Test_UserAccountService_SetPersonalAccount() {
	resp, err := s.OSSUnauthenticatedLicensedClients.Users().SetPersonalAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetPersonalAccountRequest{}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)

	accountId := resp.Msg.GetAccountId()
	require.NotEmpty(s.T(), accountId)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccounts_NotEmpty() {
	integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	accResp, err := s.OSSUnauthenticatedLicensedClients.Users().GetUserAccounts(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserAccountsRequest{}))
	integrationtests_test.RequireNoErrResp(s.T(), accResp, err)

	accounts := accResp.Msg.GetAccounts()
	require.NotEmpty(s.T(), accounts)
	require.Len(s.T(), accounts, 1)
}

func (s *IntegrationTestSuite) Test_UserAccountService_IsUserInAccount() {
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().IsUserInAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{
		AccountId: accountId,
	}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)
	require.True(s.T(), resp.Msg.GetOk())

	resp, err = s.OSSUnauthenticatedLicensedClients.Users().IsUserInAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{
		AccountId: uuid.NewString(),
	}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)
	require.False(s.T(), resp.Msg.GetOk())
}

func (s *IntegrationTestSuite) Test_UserAccountService_CreateTeamAccount_NoAuth() {
	integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().CreateTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateTeamAccountRequest{Name: "test-name"}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetTeamAccountMembers_NoAuth_Personal() {
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().GetTeamAccountMembers(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetTeamAccountMembersRequest{AccountId: accountId}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_RemoveTeamAccountMember_NoAuth_Personal() {
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().RemoveTeamAccountMember(s.ctx, connect.NewRequest(&mgmtv1alpha1.RemoveTeamAccountMemberRequest{AccountId: accountId, UserId: uuid.NewString()}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_InviteUserToTeamAccount_NoAuth_Personal() {
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().InviteUserToTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.InviteUserToTeamAccountRequest{AccountId: accountId, Email: "test@example.com"}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetTeamAccountInvites_NoAuth_Personal() {
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().GetTeamAccountInvites(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetTeamAccountInvitesRequest{AccountId: accountId}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_UserAccountService_RemoveTeamAccountInvite_NoAuth_Personal() {
	resp, err := s.OSSUnauthenticatedLicensedClients.Users().RemoveTeamAccountInvite(s.ctx, connect.NewRequest(&mgmtv1alpha1.RemoveTeamAccountInviteRequest{Id: uuid.NewString()}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)
}

func (s *IntegrationTestSuite) Test_UserAccountService_AcceptTeamAccountInvite_NoAuth_Personal() {
	resp, err := s.OSSUnauthenticatedLicensedClients.Users().AcceptTeamAccountInvite(s.ctx, connect.NewRequest(&mgmtv1alpha1.AcceptTeamAccountInviteRequest{Token: uuid.NewString()}))
	integrationtests_test.RequireErrResp(s.T(), resp, err)
	integrationtests_test.RequireConnectError(s.T(), err, connect.CodeUnauthenticated)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetSystemInformation() {
	resp, err := s.OSSUnauthenticatedLicensedClients.Users().GetSystemInformation(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetSystemInformationRequest{}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountStatus_NeosyncCloud_Personal() {
	userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	s.setUser(s.ctx, userclient)
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), userclient)

	resp, err := userclient.GetAccountStatus(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountStatusRequest{
		AccountId: accountId,
	}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)

	require.Equal(s.T(), mgmtv1alpha1.BillingStatus_BILLING_STATUS_TRIAL_ACTIVE, resp.Msg.GetSubscriptionStatus())
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
	userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	s.setUser(s.ctx, userclient)

	t := s.T()

	t.Run("active_sub", func(t *testing.T) {
		custId := "cust_id1"
		accountId := s.createBilledTeamAccount(s.ctx, userclient, "test-team", custId)
		s.Mocks.Billingclient.On("GetSubscriptions", custId).Once().Return(&testSubscriptionIter{subscriptions: []*stripe.Subscription{
			{Status: stripe.SubscriptionStatusIncompleteExpired},
			{Status: stripe.SubscriptionStatusActive},
		}}, nil)

		resp, err := userclient.GetAccountStatus(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountStatusRequest{
			AccountId: accountId,
		}))
		integrationtests_test.RequireNoErrResp(s.T(), resp, err)

		require.Equal(s.T(), mgmtv1alpha1.BillingStatus_BILLING_STATUS_ACTIVE, resp.Msg.GetSubscriptionStatus())
	})

	t.Run("no_active_subscriptions", func(t *testing.T) {
		custId := "cust_id2"
		accountId := s.createBilledTeamAccount(s.ctx, userclient, "test-team1", custId)
		err := s.setAccountCreatedAt(s.ctx, accountId, time.Now().UTC().Add(-30*24*time.Hour))
		assert.NoError(s.T(), err)

		s.Mocks.Billingclient.On("GetSubscriptions", custId).Once().Return(&testSubscriptionIter{subscriptions: []*stripe.Subscription{
			{Status: stripe.SubscriptionStatusIncompleteExpired},
			{Status: stripe.SubscriptionStatusIncompleteExpired},
		}}, nil)

		resp, err := userclient.GetAccountStatus(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountStatusRequest{
			AccountId: accountId,
		}))
		integrationtests_test.RequireNoErrResp(s.T(), resp, err)

		require.Equal(s.T(), mgmtv1alpha1.BillingStatus_BILLING_STATUS_EXPIRED, resp.Msg.GetSubscriptionStatus())
	})
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountStatus_OSS_Personal() {
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())

	resp, err := s.OSSUnauthenticatedLicensedClients.Users().GetAccountStatus(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountStatusRequest{
		AccountId: accountId,
	}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)

	require.Equal(s.T(), mgmtv1alpha1.BillingStatus_BILLING_STATUS_UNSPECIFIED, resp.Msg.GetSubscriptionStatus())
}

func (s *IntegrationTestSuite) Test_UserAccountService_IsAccountStatusValid_NeosyncCloud_Personal() {
	userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	s.setUser(s.ctx, userclient)
	accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), userclient)

	resp, err := userclient.IsAccountStatusValid(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
		AccountId: accountId,
	}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)

	require.True(s.T(), resp.Msg.GetIsValid())
	require.Empty(s.T(), resp.Msg.GetReason())
	require.Equal(s.T(), mgmtv1alpha1.AccountStatus_ACCOUNT_STATUS_ACCOUNT_TRIAL_ACTIVE, resp.Msg.GetAccountStatus())
}

func (s *IntegrationTestSuite) Test_UserAccountService_IsAccountStatusValid_NeosyncCloud_Billed() {
	userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	s.setUser(s.ctx, userclient)

	t := s.T()
	t.Run("active", func(t *testing.T) {
		custId := "cust_id1"
		accountId := s.createBilledTeamAccount(s.ctx, userclient, "test1", custId)
		s.Mocks.Billingclient.On("GetSubscriptions", custId).Once().Return(&testSubscriptionIter{subscriptions: []*stripe.Subscription{
			{Status: stripe.SubscriptionStatusActive},
		}}, nil)
		resp, err := userclient.IsAccountStatusValid(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
			AccountId: accountId,
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)

		assert.True(t, resp.Msg.GetIsValid())
		assert.Empty(t, resp.Msg.GetReason())
		require.Equal(t, mgmtv1alpha1.AccountStatus_ACCOUNT_STATUS_REASON_UNSPECIFIED, resp.Msg.GetAccountStatus())
		require.False(t, resp.Msg.GetShouldPoll())
	})
	t.Run("no_active_subs", func(t *testing.T) {
		custId := "cust_id2"
		accountId := s.createBilledTeamAccount(s.ctx, userclient, "test2", custId)
		err := s.setAccountCreatedAt(s.ctx, accountId, time.Now().UTC().Add(-30*24*time.Hour))
		assert.NoError(t, err)
		s.Mocks.Billingclient.On("GetSubscriptions", custId).Once().Return(&testSubscriptionIter{subscriptions: []*stripe.Subscription{
			{Status: stripe.SubscriptionStatusIncompleteExpired},
		}}, nil)

		resp, err := userclient.IsAccountStatusValid(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
			AccountId: accountId,
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)

		assert.False(t, resp.Msg.GetIsValid())
		assert.NotEmpty(t, resp.Msg.GetReason())
		assert.Equal(t, mgmtv1alpha1.AccountStatus_ACCOUNT_STATUS_ACCOUNT_IN_EXPIRED_STATE, resp.Msg.GetAccountStatus())
		assert.False(t, resp.Msg.GetShouldPoll())
	})
	t.Run("no_subs_active_trial", func(t *testing.T) {
		custId := "cust_id3"
		accountId := s.createBilledTeamAccount(s.ctx, userclient, "test3", custId)
		s.Mocks.Billingclient.On("GetSubscriptions", custId).Once().Return(&testSubscriptionIter{subscriptions: []*stripe.Subscription{}}, nil)

		resp, err := userclient.IsAccountStatusValid(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
			AccountId: accountId,
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)
		assert.True(t, resp.Msg.GetIsValid())
		assert.Empty(t, resp.Msg.GetReason())
		assert.Equal(t, mgmtv1alpha1.AccountStatus_ACCOUNT_STATUS_ACCOUNT_TRIAL_ACTIVE, resp.Msg.GetAccountStatus())
		assert.False(t, resp.Msg.GetShouldPoll())
	})
	t.Run("no_subs_expired_trial", func(t *testing.T) {
		custId := "cust_id4"
		accountId := s.createBilledTeamAccount(s.ctx, userclient, "test4", custId)
		err := s.setAccountCreatedAt(s.ctx, accountId, time.Now().UTC().Add(-30*24*time.Hour))
		assert.NoError(t, err)
		s.Mocks.Billingclient.On("GetSubscriptions", custId).Once().Return(&testSubscriptionIter{subscriptions: []*stripe.Subscription{}}, nil)

		resp, err := userclient.IsAccountStatusValid(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
			AccountId: accountId,
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)

		assert.False(t, resp.Msg.GetIsValid())
		assert.NotEmpty(t, resp.Msg.GetReason())
		assert.Equal(t, mgmtv1alpha1.AccountStatus_ACCOUNT_STATUS_ACCOUNT_TRIAL_EXPIRED, resp.Msg.GetAccountStatus())
		assert.False(t, resp.Msg.GetShouldPoll())
	})
	t.Run("no_active_subs_active_trial", func(t *testing.T) {
		custId := "cust_id5"
		accountId := s.createBilledTeamAccount(s.ctx, userclient, "test5", custId)
		s.Mocks.Billingclient.On("GetSubscriptions", custId).Once().Return(&testSubscriptionIter{subscriptions: []*stripe.Subscription{
			{Status: stripe.SubscriptionStatusIncompleteExpired},
		}}, nil)

		resp, err := userclient.IsAccountStatusValid(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
			AccountId: accountId,
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)
		assert.True(t, resp.Msg.GetIsValid())
		assert.Empty(t, resp.Msg.GetReason())
		assert.Equal(t, mgmtv1alpha1.AccountStatus_ACCOUNT_STATUS_ACCOUNT_TRIAL_ACTIVE, resp.Msg.GetAccountStatus())
		assert.False(t, resp.Msg.GetShouldPoll())
	})
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountBillingCheckoutSession() {
	userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	s.setUser(s.ctx, userclient)

	t := s.T()

	t.Run("billed account - allowed", func(t *testing.T) {
		teamAccountId := s.createBilledTeamAccount(s.ctx, userclient, "test-team", "test-stripe-id")

		s.Mocks.Billingclient.On("NewCheckoutSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
			Return(&stripe.CheckoutSession{URL: "new-test-url"}, nil)

		resp, err := userclient.GetAccountBillingCheckoutSession(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountBillingCheckoutSessionRequest{
			AccountId: teamAccountId,
		}))
		integrationtests_test.RequireNoErrResp(s.T(), resp, err)
		require.Equal(s.T(), "new-test-url", resp.Msg.GetCheckoutSessionUrl())
	})

	t.Run("personal account - disallowed", func(t *testing.T) {
		personalAccountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), userclient)
		resp, err := userclient.GetAccountBillingCheckoutSession(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountBillingCheckoutSessionRequest{
			AccountId: personalAccountId,
		}))
		integrationtests_test.RequireErrResp(s.T(), resp, err)
	})

	t.Run("non-neosynccloud - disallowed", func(t *testing.T) {
		personalAccountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())
		resp, err := userclient.GetAccountBillingCheckoutSession(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountBillingCheckoutSessionRequest{
			AccountId: personalAccountId,
		}))
		integrationtests_test.RequireErrResp(s.T(), resp, err)
	})
}

func (s *IntegrationTestSuite) Test_UserAccountService_GetAccountBillingPortalSession() {
	userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	s.setUser(s.ctx, userclient)

	t := s.T()

	t.Run("billed account - allowed", func(t *testing.T) {
		teamAccountId := s.createBilledTeamAccount(s.ctx, userclient, "test-team", "test-stripe-id")

		s.Mocks.Billingclient.On("NewBillingPortalSession", mock.Anything, mock.Anything).Once().
			Return(&stripe.BillingPortalSession{URL: "new-test-url"}, nil)

		resp, err := userclient.GetAccountBillingPortalSession(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountBillingPortalSessionRequest{
			AccountId: teamAccountId,
		}))
		integrationtests_test.RequireNoErrResp(s.T(), resp, err)
		require.Equal(s.T(), "new-test-url", resp.Msg.GetPortalSessionUrl())
	})

	t.Run("personal account - disallowed", func(t *testing.T) {
		personalAccountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), userclient)
		resp, err := userclient.GetAccountBillingPortalSession(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountBillingPortalSessionRequest{
			AccountId: personalAccountId,
		}))
		integrationtests_test.RequireErrResp(s.T(), resp, err)
	})

	t.Run("non-neosynccloud - disallowed", func(t *testing.T) {
		personalAccountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), s.OSSUnauthenticatedLicensedClients.Users())
		resp, err := s.OSSUnauthenticatedLicensedClients.Users().GetAccountBillingPortalSession(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountBillingPortalSessionRequest{
			AccountId: personalAccountId,
		}))
		integrationtests_test.RequireErrResp(s.T(), resp, err)
	})
}

func (s *IntegrationTestSuite) createBilledTeamAccount(ctx context.Context, client mgmtv1alpha1connect.UserAccountServiceClient, name, stripeCustomerId string) string {
	s.Mocks.Billingclient.On("NewCustomer", mock.Anything).Once().
		Return(&stripe.Customer{ID: stripeCustomerId}, nil)
	s.Mocks.Billingclient.On("NewCheckoutSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
		Return(&stripe.CheckoutSession{URL: "test-url"}, nil)
	return s.createTeamAccount(ctx, client, name)
}

func (s *IntegrationTestSuite) Test_GetBillingAccounts() {
	userclient1 := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	s.setUser(s.ctx, userclient1)

	userclient2 := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId2))
	s.setUser(s.ctx, userclient2)

	workerapikey := apikey.NewV1WorkerKey()
	workeruserclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(workerapikey))

	t := s.T()

	au1PersonalAccountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), userclient1)
	au1TeamAccountId1 := s.createBilledTeamAccount(s.ctx, userclient1, "test-team", "test-stripe-id")
	au1TeamAccountId2 := s.createBilledTeamAccount(s.ctx, userclient1, "test-team2", "test-stripe-id2")

	au2TeamAccountId1 := s.createBilledTeamAccount(s.ctx, userclient2, "test-team2", "test-stripeid-3")

	t.Run("all accounts", func(t *testing.T) {
		resp, err := workeruserclient.GetBillingAccounts(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetBillingAccountsRequest{}))
		integrationtests_test.RequireNoErrResp(t, resp, err)
		accounts := resp.Msg.GetAccounts()
		accountIds := getAccountIds(t, accounts)
		require.Len(t, accounts, 3)
		require.Contains(t, accountIds, au1TeamAccountId1)
		require.Contains(t, accountIds, au1TeamAccountId2)
		require.Contains(t, accountIds, au2TeamAccountId1)
		require.NotContains(t, accountIds, au1PersonalAccountId)
	})

	t.Run("filter accounts", func(t *testing.T) {
		resp, err := workeruserclient.GetBillingAccounts(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetBillingAccountsRequest{
			AccountIds: []string{au1TeamAccountId1, au2TeamAccountId1}, // one account from two different users
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)
		accounts := resp.Msg.GetAccounts()
		accountIds := getAccountIds(t, accounts)
		require.Len(t, accounts, 2)
		require.Contains(t, accountIds, au1TeamAccountId1)
		require.Contains(t, accountIds, au2TeamAccountId1)
	})

	t.Run("requires worker api key", func(t *testing.T) {
		resp, err := userclient1.GetBillingAccounts(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetBillingAccountsRequest{}))
		integrationtests_test.RequireErrResp(t, resp, err)
		unautherr := nucleuserrors.NewUnauthorized("")
		require.ErrorAs(t, err, &unautherr)
	})
}

func (s *IntegrationTestSuite) Test_ConvertPersonalToTeamAccount() {
	t := s.T()

	t.Run("OSS unauth", func(t *testing.T) {
		s.setUser(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())
		resp, err := s.OSSUnauthenticatedLicensedClients.Users().ConvertPersonalToTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.ConvertPersonalToTeamAccountRequest{
			Name: "unauthteamname",
		}))
		integrationtests_test.RequireErrResp(t, resp, err)
		integrationtests_test.RequireConnectError(t, err, connect.CodePermissionDenied)
	})

	t.Run("OSS auth success", func(t *testing.T) {
		userclient := s.OSSAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
		s.setUser(s.ctx, userclient)
		accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), userclient)

		resp, err := userclient.ConvertPersonalToTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.ConvertPersonalToTeamAccountRequest{
			Name:      "newname",
			AccountId: &accountId,
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)
		require.Empty(t, resp.Msg.GetCheckoutSessionUrl())
	})

	t.Run("cloud billing success", func(t *testing.T) {
		userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
		s.setUser(s.ctx, userclient)
		accountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), userclient)

		stripeCustomerId := "foo"
		s.Mocks.Billingclient.On("NewCustomer", mock.Anything).Once().
			Return(&stripe.Customer{ID: stripeCustomerId}, nil)
		s.Mocks.Billingclient.On("NewCheckoutSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
			Return(&stripe.CheckoutSession{URL: "test-url"}, nil)
		resp, err := userclient.ConvertPersonalToTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.ConvertPersonalToTeamAccountRequest{
			Name:      "newname2",
			AccountId: &accountId,
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)
		require.NotEmpty(t, resp.Msg.GetCheckoutSessionUrl())
	})

	t.Run("cloud success unspecified account", func(t *testing.T) {
		userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
		s.setUser(s.ctx, userclient)

		stripeCustomerId := "foo"
		s.Mocks.Billingclient.On("NewCustomer", mock.Anything).Once().
			Return(&stripe.Customer{ID: stripeCustomerId}, nil)
		s.Mocks.Billingclient.On("NewCheckoutSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
			Return(&stripe.CheckoutSession{URL: "test-url"}, nil)
		resp, err := userclient.ConvertPersonalToTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.ConvertPersonalToTeamAccountRequest{
			Name: "newname3",
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)
	})
}

func (s *IntegrationTestSuite) Test_SetBillingMeterEvent() {
	userclient1 := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	s.setUser(s.ctx, userclient1)

	workerapikey := apikey.NewV1WorkerKey()
	workeruserclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(workerapikey))

	t := s.T()

	au1PersonalAccountId := integrationtests_test.CreatePersonalAccount(s.ctx, s.T(), userclient1)
	au1TeamAccountId1 := s.createBilledTeamAccount(s.ctx, userclient1, "test-team", "test-stripe-id")

	t.Run("new event", func(t *testing.T) {
		s.Mocks.Billingclient.On("NewMeterEvent", mock.Anything).Once().Return(&stripe.BillingMeterEvent{}, nil)
		ts := uint64(1)
		resp, err := workeruserclient.SetBillingMeterEvent(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetBillingMeterEventRequest{
			AccountId: au1TeamAccountId1,
			EventName: "foo",
			Value:     "10",
			EventId:   "test-event-id",
			Timestamp: &ts,
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)
	})

	t.Run("needs valid stripe customer id", func(t *testing.T) {
		ts := uint64(1)
		resp, err := workeruserclient.SetBillingMeterEvent(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetBillingMeterEventRequest{
			AccountId: au1PersonalAccountId, // personal accounts don't have a stripe customer id
			EventName: "foo2",
			Value:     "10",
			EventId:   "test-event-id",
			Timestamp: &ts,
		}))
		integrationtests_test.RequireErrResp(t, resp, err)
		badreqerr := nucleuserrors.NewBadRequest("")
		require.ErrorAs(t, err, &badreqerr)
	})

	t.Run("requires worker api key", func(t *testing.T) {
		resp, err := userclient1.SetBillingMeterEvent(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetBillingMeterEventRequest{}))
		integrationtests_test.RequireErrResp(t, resp, err)
		unautherr := nucleuserrors.NewUnauthorized("")
		require.ErrorAs(t, err, &unautherr)
	})

	t.Run("squashes meter already existing", func(t *testing.T) {
		eventId := "test-event-id"
		stripeerr := &stripe.Error{Type: stripe.ErrorTypeInvalidRequest, Msg: fmt.Sprintf("An event already exists with identifier %s", eventId)}
		s.Mocks.Billingclient.On("NewMeterEvent", mock.Anything).Once().Return(nil, stripeerr)
		ts := uint64(1)
		resp, err := workeruserclient.SetBillingMeterEvent(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetBillingMeterEventRequest{
			AccountId: au1TeamAccountId1,
			EventName: "foo",
			Value:     "10",
			EventId:   eventId,
			Timestamp: &ts,
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)
	})
}
