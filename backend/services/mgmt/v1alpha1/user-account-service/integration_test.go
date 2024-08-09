package v1alpha1_useraccountservice

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	down_cmd "github.com/nucleuscloud/neosync/backend/internal/cmds/mgmt/migrate/down"
	up_cmd "github.com/nucleuscloud/neosync/backend/internal/cmds/mgmt/migrate/up"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type IntegrationTestSuite struct {
	suite.Suite

	pgpool  *pgxpool.Pool
	querier pg_queries.Querier

	ctx context.Context

	pgcontainer   *testpg.PostgresContainer
	connstr       string
	migrationsDir string

	service *Service

	httpsrv *httptest.Server

	userclient mgmtv1alpha1connect.UserAccountServiceClient
}

func (s *IntegrationTestSuite) SetupSuite() {
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
	s.connstr = connstr

	pool, err := pgxpool.New(s.ctx, connstr)
	if err != nil {
		panic(err)
	}
	s.pgpool = pool
	s.querier = pg_queries.New()
	s.migrationsDir = "../../../../sql/postgresql/schema"

	s.service = New(
		&Config{IsAuthEnabled: false, IsNeosyncCloud: false},
		nucleusdb.New(pool, db_queries.New()),
		clientmanager.NewMockTemporalClientManagerClient(s.T()),
		auth_client.NewMockInterface(s.T()),
		authmgmt.NewMockInterface(s.T()),
	)

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.NewUserAccountServiceHandler(
		s.service,
	))
	s.httpsrv = startHTTPServer(s.T(), mux)
	s.userclient = mgmtv1alpha1connect.NewUserAccountServiceClient(s.httpsrv.Client(), s.httpsrv.URL)
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	err := up_cmd.Up(s.ctx, s.connstr, s.migrationsDir, discardLogger)
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	err := down_cmd.Down(s.ctx, s.connstr, s.migrationsDir, discardLogger)
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.pgpool != nil {
		s.pgpool.Close()
	}
	if s.pgcontainer != nil {
		err := s.pgcontainer.Terminate(s.ctx)
		if err != nil {
			panic(err)
		}
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	evkey := "INTEGRATION_TESTS_ENABLED"
	shouldRun := os.Getenv(evkey)
	if shouldRun != "1" {
		slog.Warn(fmt.Sprintf("skipping integration tests, set %s=1 to enable", evkey))
		return
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
