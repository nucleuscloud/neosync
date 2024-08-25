package v1alpha1_jobservice

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	down_cmd "github.com/nucleuscloud/neosync/backend/internal/cmds/mgmt/migrate/down"
	up_cmd "github.com/nucleuscloud/neosync/backend/internal/cmds/mgmt/migrate/up"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	v1alpha1_connectionservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-service"
	v1alpha1_useraccountservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/user-account-service"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
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

	httpsrv *httptest.Server

	mockTemporalClientMgr *clientmanager.MockTemporalClientManagerClient
	mockAuthClient        *auth_client.MockInterface
	mockAuthMgmtClient    *authmgmt.MockInterface

	userclient mgmtv1alpha1connect.UserAccountServiceClient
	jobsclient mgmtv1alpha1connect.JobServiceClient
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

	s.mockTemporalClientMgr = clientmanager.NewMockTemporalClientManagerClient(s.T())
	s.mockAuthClient = auth_client.NewMockInterface(s.T())
	s.mockAuthMgmtClient = authmgmt.NewMockInterface(s.T())

	unauthUserService := v1alpha1_useraccountservice.New(
		&v1alpha1_useraccountservice.Config{IsAuthEnabled: false, IsNeosyncCloud: false},
		nucleusdb.New(pool, db_queries.New()),
		s.mockTemporalClientMgr,
		s.mockAuthClient,
		s.mockAuthMgmtClient,
	)
	unauthConnectionService := v1alpha1_connectionservice.New(
		&v1alpha1_connectionservice.Config{},
		nucleusdb.New(pool, db_queries.New()),
		unauthUserService,
		&sqlconnect.SqlOpenConnector{},
		s.querier,
		mysql_queries.New(),
		mssql_queries.New(),
		mongoconnect.NewConnector(),
		awsmanager.New(),
	)

	unauthJobsService := New(
		&Config{IsAuthEnabled: false, IsNeosyncCloud: false, RunLogConfig: &RunLogConfig{IsEnabled: false}},
		nucleusdb.New(pool, db_queries.New()),
		s.mockTemporalClientMgr,
		unauthConnectionService,
		unauthUserService,
		sqlmanager.NewSqlManager(
			&sync.Map{}, s.querier,
			&sync.Map{}, mysql_queries.New(),
			&sync.Map{}, mssql_queries.New(),
			&sqlconnect.SqlOpenConnector{},
		),
	)

	rootmux := http.NewServeMux()

	unauthmux := http.NewServeMux()
	unauthmux.Handle(mgmtv1alpha1connect.NewUserAccountServiceHandler(unauthUserService))
	unauthmux.Handle(mgmtv1alpha1connect.NewConnectionServiceHandler(unauthConnectionService))
	unauthmux.Handle(mgmtv1alpha1connect.NewJobServiceHandler(unauthJobsService))
	rootmux.Handle("/unauth/", http.StripPrefix("/unauth", unauthmux))

	s.httpsrv = startHTTPServer(s.T(), rootmux)

	s.userclient = mgmtv1alpha1connect.NewUserAccountServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth")
	s.jobsclient = mgmtv1alpha1connect.NewJobServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth")
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
