package integrationtests_test

import (
	"testing"

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

func (s *IntegrationTestSuite) Test_ConnectionService_CheckConnectionConfig() {
	t := s.T()
	accountId := s.createPersonalAccount(s.ctx, s.unauthdClients.users)
	pgconnstr, err := s.pgcontainer.ConnectionString(s.ctx, "sslmode=disable")
	require.NoError(t, err)

	conn := s.createPostgresConnection(s.unauthdClients.connections, accountId, "foo", pgconnstr)

	t.Run("valid-pg-connstr", func(t *testing.T) {
		t.Parallel()

		resp, err := s.unauthdClients.connections.CheckConnectionConfig(
			s.ctx,
			connect.NewRequest(&mgmtv1alpha1.CheckConnectionConfigRequest{
				ConnectionConfig: conn.GetConnectionConfig(),
			}),
		)
		requireNoErrResp(t, resp, err)
		require.True(t, resp.Msg.GetIsConnected())
		require.Empty(t, resp.Msg.GetConnectionError())
	})
}

func (s *IntegrationTestSuite) Test_ConnectionService_CreateConnection() {
	t := s.T()
	accountId := s.createPersonalAccount(s.ctx, s.unauthdClients.users)

	t.Run("postgres-success", func(t *testing.T) {
		pgconnstr, err := s.pgcontainer.ConnectionString(s.ctx, "sslmode=disable")
		require.NoError(t, err)
		s.createPostgresConnection(s.unauthdClients.connections, accountId, "foo", pgconnstr)
	})
}

func (s *IntegrationTestSuite) Test_ConnectionService_UpdateConnection() {
	t := s.T()
	accountId := s.createPersonalAccount(s.ctx, s.unauthdClients.users)

	t.Run("postgres-success", func(t *testing.T) {
		pgconnstr, err := s.pgcontainer.ConnectionString(s.ctx, "sslmode=disable")
		require.NoError(t, err)
		conn := s.createPostgresConnection(s.unauthdClients.connections, accountId, "foo", pgconnstr)

		resp, err := s.unauthdClients.connections.UpdateConnection(
			s.ctx,
			connect.NewRequest(&mgmtv1alpha1.UpdateConnectionRequest{
				Id:               conn.GetId(),
				Name:             "foo2",
				ConnectionConfig: conn.GetConnectionConfig(),
			}),
		)
		requireNoErrResp(t, resp, err)
		require.Equal(t, "foo2", resp.Msg.GetConnection().GetName())
	})
}

func (s *IntegrationTestSuite) Test_ConnectionService_GetConnection() {
	t := s.T()
	accountId := s.createPersonalAccount(s.ctx, s.unauthdClients.users)
	pgconnstr, err := s.pgcontainer.ConnectionString(s.ctx, "sslmode=disable")
	require.NoError(t, err)

	conn := s.createPostgresConnection(s.unauthdClients.connections, accountId, "foo", pgconnstr)

	resp, err := s.unauthdClients.connections.GetConnection(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: conn.GetId(),
		}),
	)
	requireNoErrResp(t, resp, err)
	require.NotNil(t, resp.Msg.GetConnection())
}

func (s *IntegrationTestSuite) Test_ConnectionService_GetConnections() {
	t := s.T()
	accountId := s.createPersonalAccount(s.ctx, s.unauthdClients.users)
	pgconnstr, err := s.pgcontainer.ConnectionString(s.ctx, "sslmode=disable")
	require.NoError(t, err)

	s.createPostgresConnection(s.unauthdClients.connections, accountId, "foo", pgconnstr)

	resp, err := s.unauthdClients.connections.GetConnections(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionsRequest{
			AccountId: accountId,
		}),
	)
	requireNoErrResp(t, resp, err)
	require.NotEmpty(t, resp.Msg.GetConnections())
}

func (s *IntegrationTestSuite) Test_ConnectionService_DeleteConnection() {
	t := s.T()
	accountId := s.createPersonalAccount(s.ctx, s.unauthdClients.users)
	pgconnstr, err := s.pgcontainer.ConnectionString(s.ctx, "sslmode=disable")
	require.NoError(t, err)

	conn := s.createPostgresConnection(s.unauthdClients.connections, accountId, "foo", pgconnstr)

	resp, err := s.unauthdClients.connections.GetConnections(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionsRequest{
			AccountId: accountId,
		}),
	)
	requireNoErrResp(t, resp, err)
	require.NotEmpty(t, resp.Msg.GetConnections())

	resp2, err := s.unauthdClients.connections.DeleteConnection(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.DeleteConnectionRequest{
			Id: conn.GetId(),
		}),
	)
	requireNoErrResp(t, resp2, err)

	// again to test idempotency
	resp2, err = s.unauthdClients.connections.DeleteConnection(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.DeleteConnectionRequest{
			Id: conn.GetId(),
		}),
	)
	requireNoErrResp(t, resp2, err)
}

func (s *IntegrationTestSuite) Test_ConnectionService_CheckSqlQuery() {
	t := s.T()
	accountId := s.createPersonalAccount(s.ctx, s.unauthdClients.users)
	pgconnstr, err := s.pgcontainer.ConnectionString(s.ctx, "sslmode=disable")
	require.NoError(t, err)

	conn := s.createPostgresConnection(s.unauthdClients.connections, accountId, "foo", pgconnstr)

	resp, err := s.unauthdClients.connections.CheckSqlQuery(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.CheckSqlQueryRequest{
			Id:    conn.GetId(),
			Query: "SELECT 1",
		}),
	)
	requireNoErrResp(t, resp, err)
	require.True(t, resp.Msg.GetIsValid())
	require.Empty(t, resp.Msg.GetErorrMessage())
}
