package v1alpha1_useraccountservice

import (
	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetUser_Auth() {
	client := s.getAuthUserClient("test-user")
	userId := s.setUser(client)

	resp, err := client.GetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	requireNoErrResp(s.T(), resp, err)
	require.Equal(s.T(), userId, resp.Msg.GetUserId())
}

func (s *IntegrationTestSuite) Test_GetUser_Auth_NotFound() {
	client := s.getAuthUserClient("test-user")
	resp, err := client.GetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	requireErrResp(s.T(), resp, err)
	requireConnectError(s.T(), err, connect.CodeNotFound)
}

func (s *IntegrationTestSuite) Test_SetUser_Auth() {
	client := s.getAuthUserClient("test-user")
	userId := s.setUser(client)
	require.NotEmpty(s.T(), userId)
	require.NotEqual(s.T(), "00000000-0000-0000-0000-000000000000", userId)
}

func (s *IntegrationTestSuite) Test_CreateTeamAccount_Auth() {
	client := s.getAuthUserClient("test-user")
	s.setUser(client)

	resp, err := client.CreateTeamAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateTeamAccountRequest{Name: "test-name"}))
	requireNoErrResp(s.T(), resp, err)
	require.NotEmpty(s.T(), resp.Msg.GetAccountId())
}

func (s *IntegrationTestSuite) setUser(client mgmtv1alpha1connect.UserAccountServiceClient) string {
	s.T().Helper()
	resp, err := client.SetUser(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetUserRequest{}))
	requireNoErrResp(s.T(), resp, err)
	return resp.Msg.GetUserId()
}
