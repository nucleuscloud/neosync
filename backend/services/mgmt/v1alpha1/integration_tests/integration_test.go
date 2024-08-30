package integrationtests_test

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

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	auth_jwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	up_cmd "github.com/nucleuscloud/neosync/backend/internal/cmds/mgmt/migrate/up"
	auth_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
	v1alpha1_useraccountservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/user-account-service"
	http_client "github.com/nucleuscloud/neosync/worker/pkg/http/client"
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

	unauthUserClient   mgmtv1alpha1connect.UserAccountServiceClient
	ncunauthUserClient mgmtv1alpha1connect.UserAccountServiceClient

	mockTemporalClientMgr *clientmanager.MockTemporalClientManagerClient
	mockAuthClient        *auth_client.MockInterface
	mockAuthMgmtClient    *authmgmt.MockInterface
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

	unauthdUserService := v1alpha1_useraccountservice.New(
		&v1alpha1_useraccountservice.Config{IsAuthEnabled: false, IsNeosyncCloud: false},
		nucleusdb.New(pool, db_queries.New()),
		s.mockTemporalClientMgr,
		s.mockAuthClient,
		s.mockAuthMgmtClient,
	)

	authdUserService := v1alpha1_useraccountservice.New(
		&v1alpha1_useraccountservice.Config{IsAuthEnabled: true, IsNeosyncCloud: false},
		nucleusdb.New(pool, db_queries.New()),
		s.mockTemporalClientMgr,
		s.mockAuthClient,
		s.mockAuthMgmtClient,
	)

	neoCloudUnauthdUserService := v1alpha1_useraccountservice.New(
		&v1alpha1_useraccountservice.Config{IsAuthEnabled: false, IsNeosyncCloud: true},
		nucleusdb.New(pool, db_queries.New()),
		s.mockTemporalClientMgr,
		s.mockAuthClient,
		s.mockAuthMgmtClient,
	)

	rootmux := http.NewServeMux()

	unauthmux := http.NewServeMux()
	unauthmux.Handle(mgmtv1alpha1connect.NewUserAccountServiceHandler(
		unauthdUserService,
	))
	rootmux.Handle("/unauth/", http.StripPrefix("/unauth", unauthmux))

	authmux := http.NewServeMux()
	authmux.Handle(mgmtv1alpha1connect.NewUserAccountServiceHandler(
		authdUserService,
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(func(ctx context.Context, header http.Header, spec connect.Spec) (context.Context, error) {
				// will need to further fill this out as the tests grow
				authuserid, err := utils.GetBearerTokenFromHeader(header, "Authorization")
				if err != nil {
					return nil, err
				}
				return auth_jwt.SetTokenData(ctx, &auth_jwt.TokenContextData{
					AuthUserId: authuserid,
				}), nil
			}),
		),
	))
	rootmux.Handle("/auth/", http.StripPrefix("/auth", authmux))

	ncnoauthmux := http.NewServeMux()
	ncnoauthmux.Handle(mgmtv1alpha1connect.NewUserAccountServiceHandler(
		neoCloudUnauthdUserService,
	))
	rootmux.Handle("/ncnoauth/", http.StripPrefix("/ncnoauth", ncnoauthmux))

	s.httpsrv = startHTTPServer(s.T(), rootmux)
	s.unauthUserClient = mgmtv1alpha1connect.NewUserAccountServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth")
	s.ncunauthUserClient = mgmtv1alpha1connect.NewUserAccountServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/ncnoauth")
}

func (s *IntegrationTestSuite) getAuthUserClient(authUserId string) mgmtv1alpha1connect.UserAccountServiceClient {
	return mgmtv1alpha1connect.NewUserAccountServiceClient(http_client.WithAuth(s.httpsrv.Client(), &authUserId), s.httpsrv.URL+"/auth")
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
	// Dropping here because 1) more efficient and 2) we have a bad down migration
	// _jobs-connection-id-null.down that breaks due to having a null connection_id column.
	// we should do something about that at some point. Running this single drop is easier though
	_, err := s.pgpool.Exec(s.ctx, "DROP SCHEMA IF EXISTS neosync_api CASCADE")
	if err != nil {
		panic(err)
	}
	_, err = s.pgpool.Exec(s.ctx, "DROP TABLE IF EXISTS public.schema_migrations")
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
