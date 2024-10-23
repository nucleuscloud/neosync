package connections_cmd

import (
	"context"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/require"
)

const neosyncDbMigrationsPath = "../../../../../backend/sql/postgresql/schema"

func Test_Connections(t *testing.T) {
	t.Parallel()
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	ctx := context.Background()

	neosyncApi, err := tcneosyncapi.NewNeosyncApiTestClient(ctx, t, tcneosyncapi.WithMigrationsDirectory(neosyncDbMigrationsPath))
	if err != nil {
		panic(err)
	}
	postgresUrl := "postgresql://postgres:foofar@localhost:5434/neosync"

	t.Run("list_unauthed", func(t *testing.T) {
		accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, neosyncApi.UnauthdClients.Users)
		conn1 := tcneosyncapi.CreatePostgresConnection(ctx, t, neosyncApi.UnauthdClients.Connections, accountId, "conn1", postgresUrl)
		conn2 := tcneosyncapi.CreatePostgresConnection(ctx, t, neosyncApi.UnauthdClients.Connections, accountId, "conn2", postgresUrl)
		conns := []*mgmtv1alpha1.Connection{conn1, conn2}
		connections, err := getConnections(ctx, neosyncApi.UnauthdClients.Connections, accountId)
		require.NoError(t, err)
		require.Len(t, connections, len(conns))
	})

	t.Run("list_auth", func(t *testing.T) {
		testAuthUserId := "c3b32842-9b70-4f4e-ad45-9cab26c6f2f1"
		userclient := neosyncApi.AuthdClients.GetUserClient(testAuthUserId)
		connclient := neosyncApi.AuthdClients.GetConnectionClient(testAuthUserId)
		tcneosyncapi.SetUser(ctx, t, userclient)
		accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, userclient)
		conn1 := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "conn1", postgresUrl)
		conn2 := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "conn2", postgresUrl)
		conns := []*mgmtv1alpha1.Connection{conn1, conn2}
		connections, err := getConnections(ctx, connclient, accountId)
		require.NoError(t, err)
		require.Len(t, connections, len(conns))
	})

	t.Run("list_cloud", func(t *testing.T) {
		testAuthUserId := "34f3e404-c995-452b-89e4-9c486b491dab"
		userclient := neosyncApi.NeosyncCloudClients.GetUserClient(testAuthUserId)
		connclient := neosyncApi.NeosyncCloudClients.GetConnectionClient(testAuthUserId)
		tcneosyncapi.SetUser(ctx, t, userclient)
		accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, userclient)
		conn1 := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "conn1", postgresUrl)
		conn2 := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "conn2", postgresUrl)
		conns := []*mgmtv1alpha1.Connection{conn1, conn2}
		connections, err := getConnections(ctx, connclient, accountId)
		require.NoError(t, err)
		require.Len(t, connections, len(conns))
	})

	err = neosyncApi.TearDown(ctx)
	if err != nil {
		panic(err)
	}
}
