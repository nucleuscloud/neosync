package integrationtests_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/stdlib"
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
	jobhooks "github.com/nucleuscloud/neosync/backend/internal/ee/hooks/jobs"
	"github.com/nucleuscloud/neosync/backend/internal/ee/rbac"
	"github.com/nucleuscloud/neosync/backend/internal/ee/rbac/enforcer"
	neosync_gcp "github.com/nucleuscloud/neosync/backend/internal/gcp"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/clientmanager"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	v1alpha_anonymizationservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/anonymization-service"
	v1alpha1_connectiondataservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-data-service"
	v1alpha1_connectionservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-service"
	v1alpha1_jobservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/job-service"
	v1alpha1_transformersservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/transformers-service"
	v1alpha1_useraccountservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/user-account-service"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
	"github.com/nucleuscloud/neosync/internal/billing"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	http_client "github.com/nucleuscloud/neosync/internal/http/client"
	neomigrate "github.com/nucleuscloud/neosync/internal/migrate"
	promapiv1mock "github.com/nucleuscloud/neosync/internal/mocks/github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/mock"
)

var (
	validAuthUser = &authmgmt.User{Name: "foo", Email: "bar", Picture: "baz"}
)

type UnauthdClients struct {
	Users          mgmtv1alpha1connect.UserAccountServiceClient
	Transformers   mgmtv1alpha1connect.TransformersServiceClient
	Connections    mgmtv1alpha1connect.ConnectionServiceClient
	ConnectionData mgmtv1alpha1connect.ConnectionDataServiceClient
	Jobs           mgmtv1alpha1connect.JobServiceClient
	Anonymize      mgmtv1alpha1connect.AnonymizationServiceClient
}

type Mocks struct {
	TemporalClientManager  *clientmanager.MockInterface
	TemporalConfigProvider *clientmanager.MockConfigProvider
	Authclient             *auth_client.MockInterface
	Authmanagerclient      *authmgmt.MockInterface
	Prometheusclient       *promapiv1mock.MockAPI
	Billingclient          *billing.MockInterface
	Presidio               Presidiomocks
}

type Presidiomocks struct {
	Analyzer   *presidioapi.MockAnalyzeInterface
	Anonymizer *presidioapi.MockAnonymizeInterface
	Entities   *presidioapi.MockEntityInterface
}

type NeosyncApiTestClient struct {
	NeosyncQuerier db_queries.Querier
	systemQuerier  pg_queries.Querier

	Pgcontainer   *tcpostgres.PostgresTestContainer
	migrationsDir string

	httpsrv *httptest.Server

	UnauthdClients      *UnauthdClients
	NeosyncCloudClients *NeosyncCloudClients
	AuthdClients        *AuthdClients

	Mocks *Mocks
}

// Option is a functional option for configuring Neosync Api Test Client
type Option func(*NeosyncApiTestClient)

func NewNeosyncApiTestClient(ctx context.Context, t testing.TB, opts ...Option) (*NeosyncApiTestClient, error) {
	neoApi := &NeosyncApiTestClient{
		migrationsDir: "../../../../sql/postgresql/schema",
	}
	for _, opt := range opts {
		opt(neoApi)
	}
	err := neoApi.Setup(ctx, t)
	if err != nil {
		return nil, err
	}
	return neoApi, nil
}

// Sets neosync database migrations directory path
func WithMigrationsDirectory(directoryPath string) Option {
	return func(a *NeosyncApiTestClient) {
		a.migrationsDir = directoryPath
	}
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

func (s *NeosyncCloudClients) GetAnonymizeClient(authUserId string) mgmtv1alpha1connect.AnonymizationServiceClient {
	return mgmtv1alpha1connect.NewAnonymizationServiceClient(http_client.WithBearerAuth(&http.Client{}, &authUserId), s.httpsrv.URL+s.basepath)
}

func (s *NeosyncCloudClients) GetJobClient(authUserId string) mgmtv1alpha1connect.JobServiceClient {
	return mgmtv1alpha1connect.NewJobServiceClient(http_client.WithBearerAuth(&http.Client{}, &authUserId), s.httpsrv.URL+s.basepath)
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

func (s *NeosyncApiTestClient) Setup(ctx context.Context, t testing.TB) error {
	pgcontainer, err := tcpostgres.NewPostgresTestContainer(ctx)
	if err != nil {
		return err
	}

	s.Pgcontainer = pgcontainer
	s.NeosyncQuerier = db_queries.New()
	s.systemQuerier = pg_queries.New()

	s.Mocks = &Mocks{
		TemporalClientManager:  clientmanager.NewMockInterface(t),
		TemporalConfigProvider: clientmanager.NewMockConfigProvider(t),
		Authclient:             auth_client.NewMockInterface(t),
		Authmanagerclient:      authmgmt.NewMockInterface(t),
		Prometheusclient:       promapiv1mock.NewMockAPI(t),
		Billingclient:          billing.NewMockInterface(t),
		Presidio: Presidiomocks{
			Analyzer:   presidioapi.NewMockAnalyzeInterface(t),
			Anonymizer: presidioapi.NewMockAnonymizeInterface(t),
			Entities:   presidioapi.NewMockEntityInterface(t),
		},
	}

	err = s.InitializeTest(ctx, t)
	if err != nil {
		return err
	}

	permissiveRbacClient := rbac.NewAllowAllClient()

	rbacenforcer, err := enforcer.NewActiveEnforcer(ctx, stdlib.OpenDBFromPool(pgcontainer.DB), "neosync_api.casbin_rule")
	if err != nil {
		return fmt.Errorf("unable to create rbac enforcer: %w", err)
	}
	rbacenforcer.EnableAutoSave(true)
	err = rbacenforcer.LoadPolicy()
	if err != nil {
		return fmt.Errorf("unable to load rbac policies: %w", err)
	}
	enforcedRbacClient := rbac.New(rbacenforcer)

	maxAllowed := int64(10000)
	validLicense := testutil.NewFakeEELicense(testutil.WithIsValid())
	unauthdUserService := v1alpha1_useraccountservice.New(
		&v1alpha1_useraccountservice.Config{IsAuthEnabled: false, IsNeosyncCloud: false, DefaultMaxAllowedRecords: &maxAllowed},
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		s.Mocks.TemporalConfigProvider,
		s.Mocks.Authclient,
		s.Mocks.Authmanagerclient,
		nil,                  // billing client
		permissiveRbacClient, // rbac client
		validLicense,
	)
	unauthdUserClient := userdata.NewClient(unauthdUserService, permissiveRbacClient)

	authdUserService := v1alpha1_useraccountservice.New(
		&v1alpha1_useraccountservice.Config{IsAuthEnabled: true, IsNeosyncCloud: false},
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		s.Mocks.TemporalConfigProvider,
		s.Mocks.Authclient,
		s.Mocks.Authmanagerclient,
		nil,                // billing client
		enforcedRbacClient, // rbac client
		validLicense,
	)

	sqlmanagerclient := NewTestSqlManagerClient()

	authdUserDataClient := userdata.NewClient(authdUserService, enforcedRbacClient)

	authdConnectionService := v1alpha1_connectionservice.New(
		&v1alpha1_connectionservice.Config{},
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		authdUserDataClient,
		mongoconnect.NewConnector(),
		awsmanager.New(),
		sqlmanagerclient,
		&sqlconnect.SqlOpenConnector{},
	)

	neoCloudAuthdUserService := v1alpha1_useraccountservice.New(
		&v1alpha1_useraccountservice.Config{IsAuthEnabled: true, IsNeosyncCloud: true},
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		s.Mocks.TemporalConfigProvider,
		s.Mocks.Authclient,
		s.Mocks.Authmanagerclient,
		s.Mocks.Billingclient,
		enforcedRbacClient, // rbac client
		validLicense,
	)
	neoCloudUserDataClient := userdata.NewClient(neoCloudAuthdUserService, enforcedRbacClient)
	neoCloudAuthdAnonymizeService := v1alpha_anonymizationservice.New(
		&v1alpha_anonymizationservice.Config{IsAuthEnabled: true, IsNeosyncCloud: true, IsPresidioEnabled: false},
		nil, // meter
		neoCloudUserDataClient,
		neoCloudAuthdUserService,
		s.Mocks.Presidio.Analyzer,
		s.Mocks.Presidio.Anonymizer,
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
	)

	neoCloudConnectionService := v1alpha1_connectionservice.New(
		&v1alpha1_connectionservice.Config{},
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		neoCloudUserDataClient,
		mongoconnect.NewConnector(),
		awsmanager.New(),
		sqlmanagerclient,
		&sqlconnect.SqlOpenConnector{},
	)
	neoCloudJobHookService := jobhooks.New(
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		neoCloudUserDataClient,
		jobhooks.WithEnabled(),
	)
	neoCloudJobService := v1alpha1_jobservice.New(
		&v1alpha1_jobservice.Config{IsNeosyncCloud: true, IsAuthEnabled: true},
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		s.Mocks.TemporalClientManager,
		neoCloudConnectionService,
		sqlmanagerclient,
		neoCloudJobHookService,
		neoCloudUserDataClient,
	)

	unauthdTransformersService := v1alpha1_transformersservice.New(
		&v1alpha1_transformersservice.Config{
			IsPresidioEnabled: true,
			IsNeosyncCloud:    false,
		},
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		s.Mocks.Presidio.Entities,
		unauthdUserClient,
	)

	unauthdConnectionsService := v1alpha1_connectionservice.New(
		&v1alpha1_connectionservice.Config{},
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		unauthdUserClient,
		mongoconnect.NewConnector(),
		awsmanager.New(),
		sqlmanagerclient,
		&sqlconnect.SqlOpenConnector{},
	)

	unAuthdjobhookService := jobhooks.New(
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		unauthdUserClient,
	)

	unauthdJobsService := v1alpha1_jobservice.New(
		&v1alpha1_jobservice.Config{},
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		s.Mocks.TemporalClientManager,
		unauthdConnectionsService,
		sqlmanagerclient,
		unAuthdjobhookService,
		unauthdUserClient,
	)

	unauthdConnectionDataService := v1alpha1_connectiondataservice.New(
		&v1alpha1_connectiondataservice.Config{},
		unauthdConnectionsService,
		unauthdJobsService,
		awsmanager.New(),
		&sqlconnect.SqlOpenConnector{},
		pg_queries.New(),
		mysql_queries.New(),
		mongoconnect.NewConnector(),
		sqlmanagerclient,
		neosync_gcp.NewManager(),
	)

	var presAnalyzeClient presidioapi.AnalyzeInterface
	var presAnonClient presidioapi.AnonymizeInterface

	unauthdAnonymizationService := v1alpha_anonymizationservice.New(
		&v1alpha_anonymizationservice.Config{IsPresidioEnabled: false},
		nil, // meter
		unauthdUserClient,
		unauthdUserService,
		presAnalyzeClient, presAnonClient,
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
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
	unauthmux.Handle(mgmtv1alpha1connect.NewConnectionDataServiceHandler(
		unauthdConnectionDataService,
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
	ncauthmux.Handle(mgmtv1alpha1connect.NewJobServiceHandler(
		neoCloudJobService,
		authinterceptors,
	))
	rootmux.Handle("/ncauth/", http.StripPrefix("/ncauth", ncauthmux))

	s.httpsrv = startHTTPServer(t, rootmux)

	s.UnauthdClients = &UnauthdClients{
		Users:          mgmtv1alpha1connect.NewUserAccountServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
		Transformers:   mgmtv1alpha1connect.NewTransformersServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
		Connections:    mgmtv1alpha1connect.NewConnectionServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
		ConnectionData: mgmtv1alpha1connect.NewConnectionDataServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
		Jobs:           mgmtv1alpha1connect.NewJobServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
		Anonymize:      mgmtv1alpha1connect.NewAnonymizationServiceClient(s.httpsrv.Client(), s.httpsrv.URL+"/unauth"),
	}

	s.AuthdClients = &AuthdClients{
		httpsrv: s.httpsrv,
	}
	s.NeosyncCloudClients = &NeosyncCloudClients{
		httpsrv:  s.httpsrv,
		basepath: "/ncauth",
	}
	return nil
}

func (s *NeosyncApiTestClient) MockTemporalForCreateJob(returnId string) {
	s.Mocks.TemporalClientManager.
		On(
			"DoesAccountHaveNamespace", mock.Anything, mock.Anything, mock.Anything,
		).
		Return(true, nil).
		Once()
	s.Mocks.TemporalClientManager.
		On(
			"GetSyncJobTaskQueue", mock.Anything, mock.Anything, mock.Anything,
		).
		Return("sync-job", nil).
		Once()
	s.Mocks.TemporalClientManager.
		On(
			"CreateSchedule", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).
		Return(returnId, nil).
		Once()
}

func (s *NeosyncApiTestClient) InitializeTest(ctx context.Context, t testing.TB) error {
	err := neomigrate.Up(ctx, s.Pgcontainer.URL, s.migrationsDir, testutil.GetTestLogger(t))
	if err != nil {
		return err
	}
	return nil
}

func (s *NeosyncApiTestClient) CleanupTest(ctx context.Context) error {
	// Dropping here because 1) more efficient and 2) we have a bad down migration
	// _jobs-connection-id-null.down that breaks due to having a null connection_id column.
	// we should do something about that at some point. Running this single drop is easier though
	_, err := s.Pgcontainer.DB.Exec(ctx, "DROP SCHEMA IF EXISTS neosync_api CASCADE")
	if err != nil {
		return err
	}
	_, err = s.Pgcontainer.DB.Exec(ctx, "DROP TABLE IF EXISTS public.schema_migrations")
	if err != nil {
		return err
	}
	return nil
}

func (s *NeosyncApiTestClient) TearDown(ctx context.Context) error {
	if s.Pgcontainer != nil {
		_, err := s.Pgcontainer.DB.Exec(ctx, "DROP SCHEMA IF EXISTS neosync_api CASCADE")
		if err != nil {
			return err
		}
		_, err = s.Pgcontainer.DB.Exec(ctx, "DROP TABLE IF EXISTS public.schema_migrations")
		if err != nil {
			return err
		}
		if s.Pgcontainer.DB != nil {
			s.Pgcontainer.DB.Close()
		}
		if s.Pgcontainer.TestContainer != nil {
			err := s.Pgcontainer.TestContainer.Terminate(ctx)
			if err != nil {
				return err
			}
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

func NewTestSqlManagerClient() *sqlmanager.SqlManager {
	return sqlmanager.NewSqlManager(
		sqlmanager.WithConnectionManagerOpts(connectionmanager.WithCloseOnRelease()),
	)
}
