package v1alpha1_useraccountservice

import (
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetUser() {
	resp, err := s.userclient.GetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	requireNoErrResp(s.T(), resp, err)

	userId := resp.Msg.GetUserId()
	require.NotEmpty(s.T(), userId)
}

func (s *IntegrationTestSuite) Test_SetUser() {
	resp, err := s.userclient.SetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetUserRequest{}))
	requireNoErrResp(s.T(), resp, err)

	userId := resp.Msg.UserId
	require.NotEmpty(s.T(), userId)
}

func (s *IntegrationTestSuite) Test_GetAccounts_Empty() {
	resp, err := s.userclient.GetUserAccounts(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserAccountsRequest{}))
	requireNoErrResp(s.T(), resp, err)

	accounts := resp.Msg.GetAccounts()
	require.Empty(s.T(), accounts)
}

func (s *IntegrationTestSuite) Test_SetPersonalAccount() {
	resp, err := s.userclient.SetPersonalAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetPersonalAccountRequest{}))
	requireNoErrResp(s.T(), resp, err)

	accountId := resp.Msg.GetAccountId()
	require.NotEmpty(s.T(), accountId)
}

func (s *IntegrationTestSuite) Test_GetAccounts_NotEmpty() {
	s.createPersonalAccount()

	accResp, err := s.userclient.GetUserAccounts(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserAccountsRequest{}))
	requireNoErrResp(s.T(), accResp, err)

	accounts := accResp.Msg.GetAccounts()
	require.NotEmpty(s.T(), accounts)
	require.Len(s.T(), accounts, 1)
}

func (s *IntegrationTestSuite) Test_IsUserInAccount() {
	accountId := s.createPersonalAccount()

	resp, err := s.userclient.IsUserInAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{
		AccountId: accountId,
	}))
	requireNoErrResp(s.T(), resp, err)
	require.True(s.T(), resp.Msg.GetOk())

	resp, err = s.userclient.IsUserInAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{
		AccountId: uuid.NewString(),
	}))
	requireNoErrResp(s.T(), resp, err)
	require.False(s.T(), resp.Msg.GetOk())
}

func (s *IntegrationTestSuite) Test_CreateTeamAccount_NoAuth() {
	s.createPersonalAccount()

	resp, err := s.userclient.CreateTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateTeamAccountRequest{Name: "test-name"}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodePermissionDenied)
}

func (s *IntegrationTestSuite) Test_GetSystemInformation() {
	resp, err := s.userclient.GetSystemInformation(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetSystemInformationRequest{}))
	requireNoErrResp(s.T(), resp, err)
}

func (s *IntegrationTestSuite) createPersonalAccount() string {
	s.T().Helper()
	resp, err := s.userclient.SetPersonalAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetPersonalAccountRequest{}))
	requireNoErrResp(s.T(), resp, err)
	return resp.Msg.AccountId
}

func requireNoErrResp[T any](t testing.TB, resp *connect.Response[T], err error) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func requireErrResp[T any](t testing.TB, resp *connect.Response[T], err error) {
	t.Helper()
	require.Error(t, err)
	require.Nil(t, resp)
}

func requireConnectError(t testing.TB, err error, code connect.Code) {
	t.Helper()
	connectErr, ok := err.(*connect.Error)
	require.True(t, ok, fmt.Sprintf("error was not connect error %T", err))
	require.Equal(t, code, connectErr.Code())
}
