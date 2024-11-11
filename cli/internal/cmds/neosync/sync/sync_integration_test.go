package sync_cmd

import (
	"context"
	"testing"

	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/cli/internal/output"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcmysql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/mysql"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/require"
)

const neosyncDbMigrationsPath = "../../../../../backend/sql/postgresql/schema"

func Test_Sync(t *testing.T) {
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

	connclient := neosyncApi.UnauthdClients.Connections
	conndataclient := neosyncApi.UnauthdClients.ConnectionData
	sqlmanagerclient := tcneosyncapi.NewTestSqlManagerClient()

	accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, neosyncApi.UnauthdClients.Users)
	outputType := output.PlainOutput

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()
		postgres, err := tcpostgres.NewPostgresTestSyncContainer(ctx, []tcpostgres.Option{}, []tcpostgres.Option{})
		if err != nil {
			panic(err)
		}

		testdataFolder := "../../../../../internal/testutil/testdata/postgres"
		err = postgres.Source.RunSqlFiles(ctx, &testdataFolder, []string{"humanresources/create-tables.sql", "alltypes/create-tables.sql"})
		if err != nil {
			panic(err)
		}
		err = postgres.Target.RunSqlFiles(ctx, &testdataFolder, []string{"humanresources/create-schema.sql", "alltypes/create-schema.sql"})
		if err != nil {
			panic(err)
		}
		sourceConn := tcneosyncapi.CreatePostgresConnection(ctx, t, neosyncApi.UnauthdClients.Connections, accountId, "postgres-source", postgres.Source.URL)

		t.Run("sync", func(t *testing.T) {
			discardLogger := testutil.GetTestLogger(t)
			cmdConfig := &cmdConfig{
				Source: &sourceConfig{
					ConnectionId: sourceConn.Id,
				},
				Destination: &sqlDestinationConfig{
					ConnectionUrl:        postgres.Target.URL,
					Driver:               postgresDriver,
					InitSchema:           true,
					TruncateBeforeInsert: true,
					TruncateCascade:      true,
				},
				OutputType: &outputType,
				AccountId:  &accountId,
			}
			sync := &clisync{
				connectiondataclient: conndataclient,
				connectionclient:     connclient,
				sqlmanagerclient:     sqlmanagerclient,
				ctx:                  ctx,
				logger:               discardLogger,
				cmd:                  cmdConfig,
			}
			err := sync.configureAndRunSync()
			require.NoError(t, err)

			rows := postgres.Target.DB.QueryRow(ctx, "select count(*) from humanresources.employees;")
			var rowCount int
			err = rows.Scan(&rowCount)
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)

			rows = postgres.Target.DB.QueryRow(ctx, "select count(*) from humanresources.generated_table;")
			err = rows.Scan(&rowCount)
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)

			rows = postgres.Target.DB.QueryRow(ctx, "select count(*) from alltypes.all_postgres_types;")
			err = rows.Scan(&rowCount)
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)
		})

		t.Cleanup(func() {
			err := postgres.TearDown(ctx)
			if err != nil {
				panic(err)
			}
		})
	})

	t.Run("mysql", func(t *testing.T) {
		t.Parallel()
		mysql, err := tcmysql.NewMysqlTestSyncContainer(ctx, []tcmysql.Option{}, []tcmysql.Option{})
		if err != nil {
			panic(err)
		}

		testdataFolder := "../../../../../internal/testutil/testdata/mysql"
		err = mysql.Source.RunSqlFiles(ctx, &testdataFolder, []string{"humanresources/create-tables.sql", "alltypes/create-tables.sql"})
		if err != nil {
			panic(err)
		}
		err = mysql.Target.RunSqlFiles(ctx, &testdataFolder, []string{"humanresources/create-schema.sql", "alltypes/create-schema.sql"})
		if err != nil {
			panic(err)
		}
		sourceConn := tcneosyncapi.CreateMysqlConnection(ctx, t, neosyncApi.UnauthdClients.Connections, accountId, "mysql-source", mysql.Source.URL)

		t.Run("sync", func(t *testing.T) {
			discardLogger := testutil.GetTestLogger(t)
			cmdConfig := &cmdConfig{
				Source: &sourceConfig{
					ConnectionId: sourceConn.Id,
				},
				Destination: &sqlDestinationConfig{
					ConnectionUrl:        mysql.Target.URL,
					Driver:               mysqlDriver,
					InitSchema:           true,
					TruncateBeforeInsert: true,
				},
				OutputType: &outputType,
				AccountId:  &accountId,
			}
			sync := &clisync{
				connectiondataclient: conndataclient,
				connectionclient:     connclient,
				sqlmanagerclient:     sqlmanagerclient,
				ctx:                  ctx,
				logger:               discardLogger,
				cmd:                  cmdConfig,
			}
			err := sync.configureAndRunSync()
			require.NoError(t, err)

			rows := mysql.Target.DB.QueryRowContext(ctx, "select count(*) from humanresources.locations;")
			var rowCount int
			err = rows.Scan(&rowCount)
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)

			rows = mysql.Target.DB.QueryRowContext(ctx, "select count(*) from humanresources.generated_table;")
			err = rows.Scan(&rowCount)
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)

			rows = mysql.Target.DB.QueryRowContext(ctx, "select count(*) from alltypes.all_data_types;")
			err = rows.Scan(&rowCount)
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)
		})

		t.Cleanup(func() {
			err := mysql.TearDown(ctx)
			if err != nil {
				panic(err)
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
