package integrationtests_test

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

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	auth_jwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	auth_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/auth"
	mockPromV1 "github.com/nucleuscloud/neosync/backend/internal/mocks/github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	v1alpha1_connectionservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-service"
	v1alpha1_jobservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/job-service"
	v1alpha1_transformersservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/transformers-service"
	v1alpha1_useraccountservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/user-account-service"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
	"github.com/nucleuscloud/neosync/internal/billing"
	neomigrate "github.com/nucleuscloud/neosync/internal/migrate"
	http_client "github.com/nucleuscloud/neosync/worker/pkg/http/client"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type unauthdClients struct {
	users        mgmtv1alpha1connect.UserAccountServiceClient
	transformers mgmtv1alpha1connect.TransformersServiceClient
	connections  mgmtv1alpha1connect.ConnectionServiceClient
	jobs         mgmtv1alpha1connect.JobServiceClient
}

type neosyncCloudClients struct {
	httpsrv  *httptest.Server
	basepath string
}

func (s *neosyncCloudClients) getUserClient(authUserId string) mgmtv1alpha1connect.UserAccountServiceClient {
	return mgmtv1alpha1connect.NewUserAccountServiceClient(http_client.WithAuth(s.httpsrv.Client(), &authUserId), s.httpsrv.URL+s.basepath)
}

type authdClients struct {
	httpsrv *httptest.Server
}

func (s *authdClients) getUserClient(authUserId string) mgmtv1alpha1connect.UserAccountServiceClient {
	return mgmtv1alpha1connect.NewUserAccountServiceClient(http_client.WithAuth(s.httpsrv.Client(), &authUserId), s.httpsrv.URL+"/auth")
}

type mocks struct {
	temporalClientManager *clientmanager.MockTemporalClientManagerClient
	authclient            *auth_client.MockInterface
	authmanagerclient     *authmgmt.MockInterface
	prometheusclient      *mockPromV1.MockAPI
	billingclient         *billing.MockInterface
}

type IntegrationTestSuite struct {
	suite.Suite

	pgpool         *pgxpool.Pool
	neosyncQuerier db_queries.Querier
	systemQuerier  pg_queries.Querier

	ctx context.Context

	pgcontainer   *testpg.PostgresContainer
	connstr       string
	migrationsDir string

	httpsrv *httptest.Server

	unauthdClients      *unauthdClients
	neosyncCloudClients *neosyncCloudClients
	authdClients        *authdClients

	mocks *mocks
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
	s.neosyncQuerier = db_queries.New()
	s.systemQuerier = pg_queries.New()
	s.migrationsDir = "../../../../sql/postgresql/schema"

	s.mocks = &mocks{
		temporalClientManager: clientmanager.NewMockTemporalClientManagerClient(s.T()),
		authclient:            auth_client.NewMockInterface(s.T()),
		authmanagerclient:     authmgmt.NewMockInterface(s.T()),
		prometheusclient:      mockPromV1.NewMockAPI(s.T()),
		billingclient:         billing.NewMockInterface(s.T()),
	}

	maxAllowed := int64(10000)
	unauthdUserService := v1alpha1_useraccountservice.New(
		&v1alpha1_useraccountservice.Config{IsAuthEnabled: false, IsNeosyncCloud: false, DefaultMaxAllowedRecords: &maxAllowed},
		neosyncdb.New(pool, db_queries.New()),
		s.mocks.temporalClientManager,
		s.mocks.authclient,
		s.mocks.authmanagerclient,
		s.mocks.prometheusclient,
		nil,
	)

	authdUserService := v1alpha1_useraccountservice.New(
		&v1alpha1_useraccountservice.Config{IsAuthEnabled: true, IsNeosyncCloud: false},
		neosyncdb.New(pool, db_queries.New()),
		s.mocks.temporalClientManager,
		s.mocks.authclient,
		s.mocks.authmanagerclient,
		s.mocks.prometheusclient,
		nil,
	)

	neoCloudAuthdUserService := v1alpha1_useraccountservice.New(
		&v1alpha1_useraccountservice.Config{IsAuthEnabled: true, IsNeosyncCloud: true},
		neosyncdb.New(pool, db_queries.New()),
		s.mocks.temporalClientManager,
		s.mocks.authclient,
		s.mocks.authmanagerclient,
		s.mocks.prometheusclient,
		s.mocks.billingclient,
	)

	unauthdTransformersService := v1alpha1_transformersservice.New(
		&v1alpha1_transformersservice.Config{},
		neosyncdb.New(pool, db_queries.New()),
		unauthdUserService,
	)

	unauthdConnectionsService := v1alpha1_connectionservice.New(
		&v1alpha1_connectionservice.Config{},
		neosyncdb.New(pool, db_queries.New()),
		unauthdUserService,
		&sqlconnect.SqlOpenConnector{},
		pg_queries.New(),
		mysql_queries.New(),
		mssql_queries.New(),
		mongoconnect.NewConnector(),
		awsmanager.New(),
	)
	unauthdJobsService := v1alpha1_jobservice.New(
		&v1alpha1_jobservice.Config{},
		neosyncdb.New(pool, db_queries.New()),
		s.mocks.temporalClientManager,
		unauthdConnectionsService,
		unauthdUserService,
		sqlmanager.NewSqlManager(
			&sync.Map{}, pg_queries.New(),
			&sync.Map{}, mysql_queries.New(),
			&sync.Map{}, mssql_queries.New(),
			&sqlconnect.SqlOpenConnector{},
		),
	)

	rootmux := http.NewServeMux()

	unauthmux := http.NewServeMux()
	unauthmux.Handle(mgmtv1alpha1connect.NewUserAccountServiceHandler(
		unauthdUserService,
	))
	unauthmux.Handle(mgmtv1alpha1connect.NewTransformersServiceHandler(
		unauthdTransformersService,
	))
	unauthmux.Handle(mgmtv1alpha1connect.NewConnectionServiceHandler(
		unauthdConnectionsService,
	))
	unauthmux.Handle(mgmtv1alpha1connect.NewJobServiceHandler(
		unauthdJobsService,
	))
	rootmux.Handle("/unauth/", http.StripPrefix("/unauth", unauthmux))

	authinterceptors := connect.WithInterceptors(
		auth_interceptor.NewInterceptor(func(ctx context.Context, header http.Header, spec connect.Spec) (context.Context, error) {
			// will need to further fill this out as the tests grow
			authuserid, err := utils.GetBearerTokenFromHeader(header, "Authorization")
			if err != nil {
				return nil, err
			}
			return auth_jwt.SetTokenData(ctx, &auth_jwt.TokenContextData{
				AuthUserId: authuserid,
				Claims:     &auth_jwt.CustomClaims{Email: &validAuthUser.Email},
			}), nil
		}),
	)

	authmux := http.NewServeMux()
	authmux.Handle(mgmtv1alpha1connect.NewUserAccountServiceHandler(
		authdUserService,
		authinterceptors,
	))
	rootmux.Handle("/auth/", http.StripPrefix("/auth", authmux))

	ncauthmux := http.NewServeMux()
	ncauthmux.Handle(mgmtv1alpha1connect.NewUserAccountServiceHandler(
		neoCloudAuthdUserService,
		authinterceptors,
	))
	rootmux.Handle("/ncauth/", http.StripPrefix("/ncauth", ncauthmux))

	s.httpsrv = startHTTPServer(s.T(), rootmux)

	s.unauthdClients = &unauthdClients{
		users:        mgmtv1alpha1connect.NewUserAccountServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
		transformers: mgmtv1alpha1connect.NewTransformersServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
		connections:  mgmtv1alpha1connect.NewConnectionServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
		jobs:         mgmtv1alpha1connect.NewJobServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
	}

	s.authdClients = &authdClients{
		httpsrv: s.httpsrv,
	}
	s.neosyncCloudClients = &neosyncCloudClients{
		httpsrv:  s.httpsrv,
		basepath: "/ncauth",
	}
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	err := neomigrate.Up(s.ctx, s.connstr, s.migrationsDir, discardLogger)
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
