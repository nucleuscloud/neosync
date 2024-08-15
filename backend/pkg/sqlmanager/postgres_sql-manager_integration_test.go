package sqlmanager

import (
	context "context"
	"fmt"
	slog "log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/testcontainers/testcontainers-go"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresIntegrationTestSuite struct {
	suite.Suite

	pgcontainer *testpg.PostgresContainer

	ctx context.Context

	sqlmanager SqlManagerClient

	// pg cfg
	pgcfg *mgmtv1alpha1.PostgresConnectionConfig
	// mgmt connection
	mgmtconn *mgmtv1alpha1.Connection
}

func (s *PostgresIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	pgcontainer, err := testpg.Run(
		s.ctx,
		"postgres:15",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		panic(err)
	}
	s.pgcontainer = pgcontainer

	connstr, err := pgcontainer.ConnectionString(s.ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	s.pgcfg = &mgmtv1alpha1.PostgresConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
			Url: connstr,
		},
	}
	s.mgmtconn = &mgmtv1alpha1.Connection{
		Id: uuid.NewString(),
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: s.pgcfg,
			},
		},
	}
}

func (s *PostgresIntegrationTestSuite) SetupTest() {
	s.sqlmanager = NewSqlManager(&sync.Map{}, pg_queries.New(), nil, nil, nil, nil, &sqlconnect.SqlOpenConnector{})
}

func (s *PostgresIntegrationTestSuite) TearDownTest() {
	if s.sqlmanager != nil {
		s.sqlmanager = nil
	}
}

func (s *PostgresIntegrationTestSuite) TearDownSuite() {
	if s.pgcontainer != nil {
		err := s.pgcontainer.Terminate(s.ctx)
		if err != nil {
			panic(err)
		}
	}
}

func TestPostgresIntegrationTestSuite(t *testing.T) {
	evkey := "INTEGRATION_TESTS_ENABLED"
	shouldRun := os.Getenv(evkey)
	if shouldRun != "1" {
		slog.Warn(fmt.Sprintf("skipping integration tests, set %s=1 to enable", evkey))
		return
	}
	suite.Run(t, new(PostgresIntegrationTestSuite))
}

func (s *PostgresIntegrationTestSuite) Test_NewPooledSqlDb() {
	t := s.T()

	conn, err := s.sqlmanager.NewPooledSqlDb(s.ctx, slog.Default(), s.mgmtconn)
	requireNoConnErr(t, conn, err)
	requireValidDatabase(t, s.ctx, conn, "postgres", "SELECT 1")
	conn.Db.Close()
}

func (s *PostgresIntegrationTestSuite) Test_NewSqlDb() {
	t := s.T()

	connTimeout := 5
	conn, err := s.sqlmanager.NewSqlDb(s.ctx, slog.Default(), s.mgmtconn, &connTimeout)
	requireNoConnErr(t, conn, err)

	requireValidDatabase(t, s.ctx, conn, "postgres", "SELECT 1")
	conn.Db.Close()
}

func (s *PostgresIntegrationTestSuite) Test_NewSqlDbFromUrl() {
	t := s.T()
	conn, err := s.sqlmanager.NewSqlDbFromUrl(s.ctx, "postgres", s.pgcfg.GetUrl())
	requireNoConnErr(t, conn, err)

	requireValidDatabase(t, s.ctx, conn, "postgres", "SELECT 1")
	conn.Db.Close()
}

func (s *PostgresIntegrationTestSuite) Test_NewSqlDbFromConnectionConfig() {
	t := s.T()
	connTimeout := 5
	conn, err := s.sqlmanager.NewSqlDbFromConnectionConfig(s.ctx, slog.Default(), s.mgmtconn.GetConnectionConfig(), &connTimeout)
	requireNoConnErr(t, conn, err)

	requireValidDatabase(t, s.ctx, conn, "postgres", "SELECT 1")
	conn.Db.Close()
}

func requireNoConnErr(t testing.TB, conn *SqlConnection, err error) {
	require.NoError(t, err)
	require.NotNil(t, conn)
}

func requireValidDatabase(t testing.TB, ctx context.Context, conn *SqlConnection, driver, statement string) { //nolint
	require.Equal(t, conn.Driver, driver)
	err := conn.Db.Exec(ctx, statement)
	require.NoError(t, err)
}
