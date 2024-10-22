package integrationtests_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	auth_jwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	auth_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	v1alpha_anonymizationservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/anonymization-service"
	v1alpha1_connectionservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-service"
	v1alpha1_jobservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/job-service"
	v1alpha1_transformersservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/transformers-service"
	v1alpha1_useraccountservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/user-account-service"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
	"github.com/nucleuscloud/neosync/internal/billing"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	neomigrate "github.com/nucleuscloud/neosync/internal/migrate"
	promapiv1mock "github.com/nucleuscloud/neosync/internal/mocks/github.com/prometheus/client_golang/api/prometheus/v1"
	http_client "github.com/nucleuscloud/neosync/worker/pkg/http/client"
	"github.com/testcontainers/testcontainers-go"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	validAuthUser = &authmgmt.User{Name: "foo", Email: "bar", Picture: "baz"}
)

type UnauthdClients struct {
	Users        mgmtv1alpha1connect.UserAccountServiceClient
	Transformers mgmtv1alpha1connect.TransformersServiceClient
	Connections  mgmtv1alpha1connect.ConnectionServiceClient
	Jobs         mgmtv1alpha1connect.JobServiceClient
	Anonymize    mgmtv1alpha1connect.AnonymizationServiceClient
}

type NeosyncApiTestClient struct {
	UnathdClients           *UnauthdClients
	AuthdClients            *AuthdClients
	NeosyncCloudClients     *NeosyncCloudClients
	apiIntegrationTestSuite *ApiIntegrationTestSuite
}

func NewNeosyncApiTestClient(ctx context.Context, t *testing.T) *NeosyncApiTestClient {
	a := &ApiIntegrationTestSuite{}
	a.SetupSuite(ctx, t)
	return &NeosyncApiTestClient{
		UnathdClients:       a.unauthdClients,
		AuthdClients:        a.authdClients,
		NeosyncCloudClients: a.neosyncCloudClients,

		apiIntegrationTestSuite: a,
	}
}

func (n *NeosyncApiTestClient) TearDown(ctx context.Context) error {
	return n.apiIntegrationTestSuite.TearDownSuite(ctx)
}

type NeosyncCloudClients struct {
	httpsrv  *httptest.Server
	basepath string
}

func (s *NeosyncCloudClients) GetUserClient(authUserId string) mgmtv1alpha1connect.UserAccountServiceClient {
	return mgmtv1alpha1connect.NewUserAccountServiceClient(http_client.WithBearerAuth(&http.Client{}, &authUserId), s.httpsrv.URL+s.basepath)
}

func (s *NeosyncCloudClients) GetConnectionClient(authUserId string) mgmtv1alpha1connect.ConnectionServiceClient {
	return mgmtv1alpha1connect.NewConnectionServiceClient(http_client.WithBearerAuth(&http.Client{}, &authUserId), s.httpsrv.URL+s.basepath)
}

type AuthdClients struct {
	httpsrv *httptest.Server
}

func (s *AuthdClients) GetUserClient(authUserId string) mgmtv1alpha1connect.UserAccountServiceClient {
	return mgmtv1alpha1connect.NewUserAccountServiceClient(http_client.WithBearerAuth(&http.Client{}, &authUserId), s.httpsrv.URL+"/auth")
}

func (s *AuthdClients) GetConnectionClient(authUserId string) mgmtv1alpha1connect.ConnectionServiceClient {
	return mgmtv1alpha1connect.NewConnectionServiceClient(http_client.WithBearerAuth(&http.Client{}, &authUserId), s.httpsrv.URL+"/auth")
}

type mocks struct {
	temporalClientManager *clientmanager.MockTemporalClientManagerClient
	authclient            *auth_client.MockInterface
	authmanagerclient     *authmgmt.MockInterface
	prometheusclient      *promapiv1mock.MockAPI
	billingclient         *billing.MockInterface
}

type ApiIntegrationTestSuite struct {
	pgpool         *pgxpool.Pool
	neosyncQuerier db_queries.Querier
	systemQuerier  pg_queries.Querier

	pgcontainer   *testpg.PostgresContainer
	connstr       string
	migrationsDir string

	httpsrv *httptest.Server

	unauthdClients      *UnauthdClients
	neosyncCloudClients *NeosyncCloudClients
	authdClients        *AuthdClients

	mocks *mocks
}

func (s *ApiIntegrationTestSuite) SetupSuite(ctx context.Context, t *testing.T) {
	pgcontainer, err := testpg.Run(
		ctx,
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
	connstr, err := pgcontainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}
	s.connstr = connstr

	pool, err := pgxpool.New(ctx, connstr)
	if err != nil {
		panic(err)
	}
	s.pgpool = pool
	s.neosyncQuerier = db_queries.New()
	s.systemQuerier = pg_queries.New()
	// TODO fix this or have it passed in
	s.migrationsDir = "../../../../../backend/sql/postgresql/schema"

	s.mocks = &mocks{
		temporalClientManager: clientmanager.NewMockTemporalClientManagerClient(t),
		authclient:            auth_client.NewMockInterface(t),
		authmanagerclient:     authmgmt.NewMockInterface(t),
		prometheusclient:      promapiv1mock.NewMockAPI(t),
		billingclient:         billing.NewMockInterface(t),
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

	authdConnectionService := v1alpha1_connectionservice.New(
		&v1alpha1_connectionservice.Config{},
		neosyncdb.New(pool, db_queries.New()),
		authdUserService,
		&sqlconnect.SqlOpenConnector{},
		pg_queries.New(),
		mysql_queries.New(),
		mssql_queries.New(),
		mongoconnect.NewConnector(),
		awsmanager.New(),
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
	neoCloudAuthdAnonymizeService := v1alpha_anonymizationservice.New(
		&v1alpha_anonymizationservice.Config{IsAuthEnabled: true, IsNeosyncCloud: true, IsPresidioEnabled: false},
		nil,
		neoCloudAuthdUserService,
		nil, // presidio
		nil, // presidio
		neosyncdb.New(pool, db_queries.New()),
	)

	neoCloudConnectionService := v1alpha1_connectionservice.New(
		&v1alpha1_connectionservice.Config{},
		neosyncdb.New(pool, db_queries.New()),
		neoCloudAuthdUserService,
		&sqlconnect.SqlOpenConnector{},
		pg_queries.New(),
		mysql_queries.New(),
		mssql_queries.New(),
		mongoconnect.NewConnector(),
		awsmanager.New(),
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

	var presAnalyzeClient presidioapi.AnalyzeInterface
	var presAnonClient presidioapi.AnonymizeInterface

	unauthdAnonymizationService := v1alpha_anonymizationservice.New(
		&v1alpha_anonymizationservice.Config{IsPresidioEnabled: false},
		nil,
		unauthdUserService,
		presAnalyzeClient, presAnonClient,
		neosyncdb.New(pool, db_queries.New()),
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

	unauthmux.Handle(mgmtv1alpha1connect.NewAnonymizationServiceHandler(
		unauthdAnonymizationService,
	))
	rootmux.Handle("/unauth/", http.StripPrefix("/unauth", unauthmux))

	authinterceptors := connect.WithInterceptors(
		auth_interceptor.NewInterceptor(func(ctx context.Context, header http.Header, spec connect.Spec) (context.Context, error) {
			// will need to further fill this out as the tests grow
			authuserid, err := utils.GetBearerTokenFromHeader(header, "Authorization")
			if err != nil {
				return nil, err
			}
			if apikey.IsValidV1WorkerKey(authuserid) {
				return auth_apikey.SetTokenData(ctx, &auth_apikey.TokenContextData{
					RawToken:   authuserid,
					ApiKey:     nil,
					ApiKeyType: apikey.WorkerApiKey,
				}), nil
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
	authmux.Handle(mgmtv1alpha1connect.NewConnectionServiceHandler(
		authdConnectionService,
		authinterceptors,
	))
	rootmux.Handle("/auth/", http.StripPrefix("/auth", authmux))

	ncauthmux := http.NewServeMux()
	ncauthmux.Handle(mgmtv1alpha1connect.NewUserAccountServiceHandler(
		neoCloudAuthdUserService,
		authinterceptors,
	))
	ncauthmux.Handle(mgmtv1alpha1connect.NewAnonymizationServiceHandler(
		neoCloudAuthdAnonymizeService,
		authinterceptors,
	))
	ncauthmux.Handle(mgmtv1alpha1connect.NewConnectionServiceHandler(
		neoCloudConnectionService,
		authinterceptors,
	))
	rootmux.Handle("/ncauth/", http.StripPrefix("/ncauth", ncauthmux))

	s.httpsrv = startHTTPServer(t, rootmux)

	s.unauthdClients = &UnauthdClients{
		Users:        mgmtv1alpha1connect.NewUserAccountServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
		Transformers: mgmtv1alpha1connect.NewTransformersServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
		Connections:  mgmtv1alpha1connect.NewConnectionServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
		Jobs:         mgmtv1alpha1connect.NewJobServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
		Anonymize:    mgmtv1alpha1connect.NewAnonymizationServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
	}

	s.authdClients = &AuthdClients{
		httpsrv: s.httpsrv,
	}
	s.neosyncCloudClients = &NeosyncCloudClients{
		httpsrv:  s.httpsrv,
		basepath: "/ncauth",
	}

	discardLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	err = neomigrate.Up(ctx, s.connstr, s.migrationsDir, discardLogger)
	if err != nil {
		panic(err)
	}
}

func (s *ApiIntegrationTestSuite) TearDownSuite(ctx context.Context) error {
	_, err := s.pgpool.Exec(ctx, "DROP SCHEMA IF EXISTS neosync_api CASCADE")
	if err != nil {
		return err
	}
	_, err = s.pgpool.Exec(ctx, "DROP TABLE IF EXISTS public.schema_migrations")
	if err != nil {
		return err
	}
	if s.pgpool != nil {
		s.pgpool.Close()
	}
	if s.pgcontainer != nil {
		err := s.pgcontainer.Terminate(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
