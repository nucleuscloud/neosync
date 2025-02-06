package integrationtests_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	"github.com/nucleuscloud/neosync/internal/authmgmt"
	"github.com/nucleuscloud/neosync/internal/billing"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	neomigrate "github.com/nucleuscloud/neosync/internal/migrate"
	promapiv1mock "github.com/nucleuscloud/neosync/internal/mocks/github.com/prometheus/client_golang/api/prometheus/v1"
	clientmanager "github.com/nucleuscloud/neosync/internal/temporal/clientmanager"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/mock"
)

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

	// OSS, Unauthenticated, Licensed
	OSSUnauthenticatedLicensedClients *NeosyncClients
	// OSS, Authenticated, Licensed
	OSSAuthenticatedLicensedClients *NeosyncClients
	// OSS, Unauthenticated, Unlicensed
	OSSUnauthenticatedUnlicensedClients *NeosyncClients
	// NeoCloud, Authenticated, Licensed
	NeosyncCloudAuthenticatedLicensedClients *NeosyncClients

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

	rootmux := http.NewServeMux()

	logger := testutil.GetConcurrentTestLogger(t)

	ossUnauthLicensedMux, err := s.setupOssUnauthenticatedLicensedMux(ctx, pgcontainer, logger)
	if err != nil {
		return fmt.Errorf("unable to setup oss unauthenticated licensed mux: %w", err)
	}
	rootmux.Handle(openSourceUnauthenticatedLicensedPostfix+"/", http.StripPrefix(openSourceUnauthenticatedLicensedPostfix, ossUnauthLicensedMux))

	ossAuthLicensedMux, err := s.setupOssLicensedAuthMux(ctx, pgcontainer, logger)
	if err != nil {
		return fmt.Errorf("unable to setup oss authenticated licensed mux: %w", err)
	}
	rootmux.Handle(openSourceAuthenticatedLicensedPostfix+"/", http.StripPrefix(openSourceAuthenticatedLicensedPostfix, ossAuthLicensedMux))

	ossUnauthUnlicensedMux, err := s.setupOssUnlicensedMux(pgcontainer, logger)
	if err != nil {
		return fmt.Errorf("unable to setup oss unauthenticated unlicensed mux: %w", err)
	}
	rootmux.Handle(openSourceUnauthenticatedUnlicensedPostfix+"/", http.StripPrefix(openSourceUnauthenticatedUnlicensedPostfix, ossUnauthUnlicensedMux))

	neoCloudAuthdMux, err := s.setupNeoCloudMux(ctx, pgcontainer, logger)
	if err != nil {
		return fmt.Errorf("unable to setup neo cloud authenticated mux: %w", err)
	}
	rootmux.Handle(neoCloudAuthenticatedLicensedPostfix+"/", http.StripPrefix(neoCloudAuthenticatedLicensedPostfix, neoCloudAuthdMux))

	s.httpsrv = startHTTPServer(t, rootmux)
	rootmux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("404 for URL: %s\n", r.URL.Path)
		http.NotFound(w, r)
	})

	s.OSSUnauthenticatedLicensedClients = newNeosyncClients(s.httpsrv.URL + openSourceUnauthenticatedLicensedPostfix)
	s.OSSAuthenticatedLicensedClients = newNeosyncClients(s.httpsrv.URL + openSourceAuthenticatedLicensedPostfix)
	s.OSSUnauthenticatedUnlicensedClients = newNeosyncClients(s.httpsrv.URL + openSourceUnauthenticatedUnlicensedPostfix)
	s.NeosyncCloudAuthenticatedLicensedClients = newNeosyncClients(s.httpsrv.URL + neoCloudAuthenticatedLicensedPostfix)

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
