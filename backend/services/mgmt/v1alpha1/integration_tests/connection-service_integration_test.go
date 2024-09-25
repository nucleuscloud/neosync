package integrationtests_test

import (
	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_ConnectionService_IsConnectionNameAvailable_Available() {
	accountId := s.createPersonalAccount(s.ctx, s.unauthdClients.users)

	resp, err := s.unauthdClients.connections.IsConnectionNameAvailable(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.IsConnectionNameAvailableRequest{
			AccountId:      accountId,
			ConnectionName: "foo",
		}),
	)
	requireNoErrResp(s.T(), resp, err)
	require.True(s.T(), resp.Msg.GetIsAvailable())
}

func (s *IntegrationTestSuite) Test_ConnectionService_IsConnectionNameAvailable_NotAvailable() {
	accountId := s.createPersonalAccount(s.ctx, s.unauthdClients.users)
	s.createPostgresConnection(s.unauthdClients.connections, accountId, "foo", "test-url")

	resp, err := s.unauthdClients.connections.IsConnectionNameAvailable(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.IsConnectionNameAvailableRequest{
			AccountId:      accountId,
			ConnectionName: "foo",
		}),
	)
	requireNoErrResp(s.T(), resp, err)
	require.False(s.T(), resp.Msg.GetIsAvailable())
}
