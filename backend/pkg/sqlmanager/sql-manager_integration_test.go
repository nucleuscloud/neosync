package sqlmanager

import (
	context "context"
	"testing"

	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/stretchr/testify/require"
	testmssql "github.com/testcontainers/testcontainers-go/modules/mssql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/microsoft/go-mssqldb"

	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/sqlprovider"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcmysql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/mysql"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
)

func Test_NewSqlConnection(t *testing.T) {
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	t.Parallel()
	ctx := context.Background()

	manager := NewSqlManager(connectionmanager.NewConnectionManager(
		sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{}),
	))

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()
		container, err := tcpostgres.NewPostgresTestContainer(ctx)
		require.NoError(t, err)
		defer container.TearDown(ctx)

		mgmtconn := &mgmtv1alpha1.Connection{
			Id: uuid.NewString(),
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: container.URL,
						},
					},
				},
			},
		}
		conn, err := manager.NewSqlConnection(ctx, mgmtconn, testutil.GetTestLogger(t))
		requireNoConnErr(t, conn, err)
		defer conn.Db().Close()
		requireValidDatabase(t, ctx, conn, "pgx", "SELECT 1")
	})

	t.Run("mysql", func(t *testing.T) {
		t.Parallel()
		container, err := tcmysql.NewMysqlTestContainer(ctx)
		require.NoError(t, err)
		defer container.TearDown(ctx)

		mgmtconn := &mgmtv1alpha1.Connection{
			Id: uuid.NewString(),
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: container.URL,
						},
					},
				},
			},
		}
		conn, err := manager.NewSqlConnection(ctx, mgmtconn, testutil.GetTestLogger(t))
		requireNoConnErr(t, conn, err)
		defer conn.Db().Close()
		requireValidDatabase(t, ctx, conn, "mysql", "SELECT 1")
	})

	t.Run("mssql", func(t *testing.T) {
		t.Parallel()
		container, err := testmssql.Run(ctx,
			"mcr.microsoft.com/mssql/server:2022-latest",
			testmssql.WithAcceptEULA(),
		)
		require.NoError(t, err)
		connstr, err := container.ConnectionString(ctx, "database=master", "encrypt=disable")
		require.NoError(t, err)

		mgmtconn := &mgmtv1alpha1.Connection{
			Id: uuid.NewString(),
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{
					MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
							Url: connstr,
						},
					},
				},
			},
		}

		conn, err := manager.NewSqlConnection(ctx, mgmtconn, testutil.GetTestLogger(t))
		requireNoConnErr(t, conn, err)
		defer conn.Db().Close()
		requireValidDatabase(t, ctx, conn, "sqlserver", "SELECT 1")
	})
}

func requireNoConnErr(t testing.TB, conn *SqlConnection, err error) {
	require.NoError(t, err)
	require.NotNil(t, conn)
}

func requireValidDatabase(t testing.TB, ctx context.Context, conn *SqlConnection, driver, statement string) { //nolint
	require.Equal(t, conn.Driver(), driver)
	err := conn.Db().Exec(ctx, statement)
	require.NoError(t, err)
}
