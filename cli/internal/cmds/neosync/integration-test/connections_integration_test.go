package integrationtest

import (
	"context"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	connections_cmd "github.com/nucleuscloud/neosync/cli/internal/cmds/neosync/connections"
	"github.com/stretchr/testify/require"
)

func Test_Connections(t *testing.T) {
	t.Parallel()
	ok := shouldRun()
	if !ok {
		return
	}
	ctx := context.Background()
	s := newCliIntegrationTest(ctx, t)

	t.Run("list_unauthed", func(t *testing.T) {
		accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, s.neosyncApi.UnathdClients.Users)
		conn1 := tcneosyncapi.CreatePostgresConnection(ctx, t, s.neosyncApi.UnathdClients.Connections, accountId, "conn1", s.postgres.URL)
		conn2 := tcneosyncapi.CreatePostgresConnection(ctx, t, s.neosyncApi.UnathdClients.Connections, accountId, "conn2", s.postgres.URL)
		conns := []*mgmtv1alpha1.Connection{conn1, conn2}
		connections, err := connections_cmd.GetConnections(ctx, s.neosyncApi.UnathdClients.Connections, accountId)
		require.NoError(t, err)
		require.Len(t, connections, len(conns))
	})

	t.Run("list_auth", func(t *testing.T) {
		testAuthUserId := "c3b32842-9b70-4f4e-ad45-9cab26c6f2f1"
		userclient := s.neosyncApi.AuthdClients.GetUserClient(testAuthUserId)
		connclient := s.neosyncApi.AuthdClients.GetConnectionClient(testAuthUserId)
		tcneosyncapi.SetUser(ctx, t, userclient)
		accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, userclient)
		conn1 := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "conn1", s.postgres.URL)
		conn2 := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "conn2", s.postgres.URL)
		conns := []*mgmtv1alpha1.Connection{conn1, conn2}
		connections, err := connections_cmd.GetConnections(ctx, connclient, accountId)
		require.NoError(t, err)
		require.Len(t, connections, len(conns))
	})

	t.Run("list_cloud", func(t *testing.T) {
		testAuthUserId := "34f3e404-c995-452b-89e4-9c486b491dab"
		userclient := s.neosyncApi.NeosyncCloudClients.GetUserClient(testAuthUserId)
		connclient := s.neosyncApi.NeosyncCloudClients.GetConnectionClient(testAuthUserId)
		tcneosyncapi.SetUser(ctx, t, userclient)
		accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, userclient)
		conn1 := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "conn1", s.postgres.URL)
		conn2 := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "conn2", s.postgres.URL)
		conns := []*mgmtv1alpha1.Connection{conn1, conn2}
		connections, err := connections_cmd.GetConnections(ctx, connclient, accountId)
		require.NoError(t, err)
		require.Len(t, connections, len(conns))
	})

	s.TearDownSuite(ctx)
}
