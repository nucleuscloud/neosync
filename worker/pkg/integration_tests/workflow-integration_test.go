package integration_tests

import (
	"context"
	"testing"

	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	tcredis "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/redis"
	tcworkflow "github.com/nucleuscloud/neosync/worker/pkg/integration-test"
	"github.com/stretchr/testify/require"
)

const neosyncDbMigrationsPath = "../../../backend/sql/postgresql/schema"

func Test_Workflow(t *testing.T) {
	t.Parallel()
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	ctx := context.Background()

	neosyncApi, err := tcneosyncapi.NewNeosyncApiTestClient(ctx, t, tcneosyncapi.WithMigrationsDirectory(neosyncDbMigrationsPath))
	if err != nil {
		t.Fatal(err)
	}

	connclient := neosyncApi.UnauthdClients.Connections
	accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, neosyncApi.UnauthdClients.Users)
	dbManagers := tcworkflow.NewTestDatabaseManagers(t)

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()
		postgres, err := tcpostgres.NewPostgresTestSyncContainer(ctx, []tcpostgres.Option{}, []tcpostgres.Option{})
		if err != nil {
			t.Fatal(err)
		}
		sourceConn := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "postgres-source", postgres.Source.URL)
		destConn := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "postgres-dest", postgres.Target.URL)

		t.Run("all_types", func(t *testing.T) {
			t.Parallel()
			test_postgres_types(t, ctx, postgres, neosyncApi, dbManagers, accountId, sourceConn, destConn)
		})

		t.Run("edgecases", func(t *testing.T) {
			t.Parallel()
			test_postgres_edgecases(t, ctx, postgres, neosyncApi, dbManagers, accountId, sourceConn, destConn)
		})

		t.Run("virtual_foreign_keys", func(t *testing.T) {
			t.Parallel()
			test_postgres_virtual_foreign_keys(t, ctx, postgres, neosyncApi, dbManagers, accountId, sourceConn, destConn)
		})

		t.Run("javascript_transformers", func(t *testing.T) {
			t.Parallel()
			test_postgres_javascript_transformers(t, ctx, postgres, neosyncApi, dbManagers, accountId, sourceConn, destConn)
		})

		t.Run("skip_foreign_keys_violations", func(t *testing.T) {
			t.Parallel()
			test_postgres_skip_foreign_keys_violations(t, ctx, postgres, neosyncApi, dbManagers, accountId, sourceConn, destConn)
		})

		t.Run("foreign_keys_violations_error", func(t *testing.T) {
			t.Parallel()
			test_postgres_foreign_keys_violations_error(t, ctx, postgres, neosyncApi, dbManagers, accountId, sourceConn, destConn)
		})

		t.Run("subsetting", func(t *testing.T) {
			t.Parallel()
			test_postgres_subsetting(t, ctx, postgres, neosyncApi, dbManagers, accountId, sourceConn, destConn)
		})

		t.Run("primary_key_transformations", func(t *testing.T) {
			t.Parallel()
			redis, err := tcredis.NewRedisTestContainer(ctx)
			require.NoError(t, err)

			test_postgres_primary_key_transformations(t, ctx, postgres, redis, neosyncApi, dbManagers, accountId, sourceConn, destConn)

			t.Cleanup(func() {
				err := redis.TearDown(ctx)
				require.NoError(t, err)
			})
		})

		t.Cleanup(func() {
			err := postgres.TearDown(ctx)
			if err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Cleanup(func() {
		err = neosyncApi.TearDown(ctx)
		if err != nil {
			panic(err)
		}
	})
}
