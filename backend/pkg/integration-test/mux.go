package integrationtests_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/stdlib"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	auth_jwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	auth_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/auth"
	accounthooks "github.com/nucleuscloud/neosync/backend/internal/ee/hooks/accounts"
	jobhooks "github.com/nucleuscloud/neosync/backend/internal/ee/hooks/jobs"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	v1alpha1_accounthookservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/account-hooks-service"
	v1alpha_anonymizationservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/anonymization-service"
	v1alpha1_connectiondataservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-data-service"
	v1alpha1_connectionservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-service"
	v1alpha1_jobservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/job-service"
	v1alpha1_transformersservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/transformers-service"
	v1alpha1_useraccountservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/user-account-service"
	"github.com/nucleuscloud/neosync/internal/apikey"
	"github.com/nucleuscloud/neosync/internal/authmgmt"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
	"github.com/nucleuscloud/neosync/internal/billing"
	"github.com/nucleuscloud/neosync/internal/connectiondata"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	"github.com/nucleuscloud/neosync/internal/ee/rbac"
	"github.com/nucleuscloud/neosync/internal/ee/rbac/enforcer"
	neosync_gcp "github.com/nucleuscloud/neosync/internal/gcp"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
)

var (
	validAuthUser = &authmgmt.User{Name: "foo", Email: "bar", Picture: "baz"}

	authinterceptor = auth_interceptor.NewInterceptor(
		func(ctx context.Context, header http.Header, spec connect.Spec) (context.Context, error) {
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
		},
	)
)

const (
	// OSS, Unauthenticated, Licensed
	openSourceUnauthenticatedLicensedPostfix = "/oss-unauthenticated-licensed"
	// OSS, Authenticated, Licensed
	openSourceAuthenticatedLicensedPostfix = "/oss-authenticated-licensed"
	// OSS, Unauthenticated, Unlicensed
	openSourceUnauthenticatedUnlicensedPostfix = "/oss-unauthenticated-unlicensed"
	// NeoCloud, Licensed, Authenticated
	neoCloudAuthenticatedLicensedPostfix = "/neosynccloud-authenticated"
)

func (s *NeosyncApiTestClient) setupOssUnauthenticatedLicensedMux(
	ctx context.Context,
	pgcontainer *tcpostgres.PostgresTestContainer,
	logger *slog.Logger,
) (*http.ServeMux, error) {
	isLicensed := true
	isAuthEnabled := false
	isNeosyncCloud := false
	enforcedRbacClient, err := s.getEnforcedRbacClient(ctx, pgcontainer)
	if err != nil {
		return nil, fmt.Errorf("unable to get enforced rbac client: %w", err)
	}
	return s.setupMux(
		pgcontainer,
		isAuthEnabled,
		isLicensed,
		isNeosyncCloud,
		enforcedRbacClient,
		logger,
	)
}

func (s *NeosyncApiTestClient) setupOssLicensedAuthMux(
	ctx context.Context,
	pgcontainer *tcpostgres.PostgresTestContainer,
	logger *slog.Logger,
) (*http.ServeMux, error) {
	isLicensed := true
	isAuthEnabled := true
	isNeosyncCloud := false
	enforcedRbacClient, err := s.getEnforcedRbacClient(ctx, pgcontainer)
	if err != nil {
		return nil, fmt.Errorf("unable to get enforced rbac client: %w", err)
	}
	return s.setupMux(
		pgcontainer,
		isAuthEnabled,
		isLicensed,
		isNeosyncCloud,
		enforcedRbacClient,
		logger,
	)
}

func (s *NeosyncApiTestClient) setupOssUnlicensedMux(
	pgcontainer *tcpostgres.PostgresTestContainer,
	logger *slog.Logger,
) (*http.ServeMux, error) {
	isLicensed := false
	isAuthEnabled := false
	isNeosyncCloud := false
	permissiveRbacClient := rbac.NewAllowAllClient()
	return s.setupMux(
		pgcontainer,
		isAuthEnabled,
		isLicensed,
		isNeosyncCloud,
		permissiveRbacClient,
		logger,
	)
}

func (s *NeosyncApiTestClient) setupNeoCloudMux(
	ctx context.Context,
	pgcontainer *tcpostgres.PostgresTestContainer,
	logger *slog.Logger,
) (*http.ServeMux, error) {
	isLicensed := true
	isAuthEnabled := true
	isNeosyncCloud := true
	enforcedRbacClient, err := s.getEnforcedRbacClient(ctx, pgcontainer)
	if err != nil {
		return nil, fmt.Errorf("unable to get enforced rbac client: %w", err)
	}
	return s.setupMux(
		pgcontainer,
		isAuthEnabled,
		isLicensed,
		isNeosyncCloud,
		enforcedRbacClient,
		logger,
	)
}

func (s *NeosyncApiTestClient) setupMux(
	pgcontainer *tcpostgres.PostgresTestContainer,
	isAuthEnabled bool,
	isLicensed bool,
	isNeosyncCloud bool,
	rbacClient rbac.Interface,
	logger *slog.Logger,
) (*http.ServeMux, error) {
	isPresidioEnabled := isLicensed || isNeosyncCloud

	maxAllowed := int64(10000)
	var license *testutil.FakeEELicense
	if isLicensed {
		license = testutil.NewFakeEELicense(testutil.WithIsValid())
	} else {
		license = testutil.NewFakeEELicense()
	}

	neosyncDb := neosyncdb.New(pgcontainer.DB, db_queries.New())

	var billingclient billing.Interface
	if isNeosyncCloud {
		billingclient = s.Mocks.Billingclient
	} else {
		billingclient = nil
	}

	userService := v1alpha1_useraccountservice.New(
		&v1alpha1_useraccountservice.Config{
			IsAuthEnabled:            isAuthEnabled,
			IsNeosyncCloud:           isNeosyncCloud,
			DefaultMaxAllowedRecords: &maxAllowed,
		},
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		s.Mocks.TemporalConfigProvider,
		s.Mocks.Authclient,
		s.Mocks.Authmanagerclient,
		billingclient,
		rbacClient, // rbac client
		license,
	)
	userclient := userdata.NewClient(userService, rbacClient, license)

	transformerService := v1alpha1_transformersservice.New(
		&v1alpha1_transformersservice.Config{
			IsPresidioEnabled: isPresidioEnabled,
		},
		neosyncdb.New(pgcontainer.DB, db_queries.New()),
		s.Mocks.Presidio.Entities,
		userclient,
		license,
	)

	sqlmanagerclient := NewTestSqlManagerClient()

	connectionService := v1alpha1_connectionservice.New(
		&v1alpha1_connectionservice.Config{IsNeosyncCloud: isNeosyncCloud},
		neosyncDb,
		userclient,
		mongoconnect.NewConnector(),
		awsmanager.New(),
		sqlmanagerclient,
		&sqlconnect.SqlOpenConnector{},
	)

	var jobhookService *jobhooks.Service
	if isLicensed {
		jobhookService = jobhooks.New(
			neosyncDb,
			userclient,
			jobhooks.WithEnabled(),
		)
	} else {
		jobhookService = jobhooks.New(
			neosyncDb,
			userclient,
		)
	}

	awsManager := awsmanager.New()
	sqlConnector := &sqlconnect.SqlOpenConnector{}
	pgquerier := pg_queries.New()
	mysqlquerier := mysql_queries.New()
	mongoconnector := mongoconnect.NewConnector()
	sqlmanager := sqlmanagerclient
	gcpmanager := neosync_gcp.NewManager()
	neosynctyperegistry := neosynctypes.NewTypeRegistry(logger)

	connectiondatabuilder := connectiondata.NewConnectionDataBuilder(
		sqlConnector,
		sqlmanager,
		pgquerier,
		mysqlquerier,
		awsManager,
		gcpmanager,
		mongoconnector,
		neosynctyperegistry,
	)

	jobService := v1alpha1_jobservice.New(
		&v1alpha1_jobservice.Config{IsAuthEnabled: isAuthEnabled, IsNeosyncCloud: isNeosyncCloud},
		neosyncDb,
		s.Mocks.TemporalClientManager,
		connectionService,
		sqlmanagerclient,
		jobhookService,
		userclient,
		connectiondatabuilder,
	)

	var presAnalyzeClient presidioapi.AnalyzeInterface
	var presAnonClient presidioapi.AnonymizeInterface

	anonymizationService := v1alpha_anonymizationservice.New(
		&v1alpha_anonymizationservice.Config{
			IsPresidioEnabled: isPresidioEnabled,
			IsAuthEnabled:     isAuthEnabled,
			IsNeosyncCloud:    isNeosyncCloud,
		},
		nil, // meter
		userclient,
		userService,
		transformerService,
		presAnalyzeClient,
		presAnonClient,
		neosyncDb,
	)

	connectionDataService := v1alpha1_connectiondataservice.New(
		&v1alpha1_connectiondataservice.Config{},
		connectionService,
		connectiondatabuilder,
	)

	accountHookService := v1alpha1_accounthookservice.New(
		accounthooks.New(
			neosyncDb,
			userclient,
			accounthooks.WithSlackClient(s.Mocks.Slackclient),
		),
	)

	mux := http.NewServeMux()

	interceptors := []connect.Interceptor{}

	if isAuthEnabled {
		interceptors = append(interceptors, authinterceptor)
	}

	mux.Handle(mgmtv1alpha1connect.NewUserAccountServiceHandler(
		userService,
		connect.WithInterceptors(interceptors...),
	))
	mux.Handle(mgmtv1alpha1connect.NewTransformersServiceHandler(
		transformerService,
		connect.WithInterceptors(interceptors...),
	))
	mux.Handle(mgmtv1alpha1connect.NewConnectionServiceHandler(
		connectionService,
		connect.WithInterceptors(interceptors...),
	))
	mux.Handle(mgmtv1alpha1connect.NewJobServiceHandler(
		jobService,
		connect.WithInterceptors(interceptors...),
	))
	mux.Handle(mgmtv1alpha1connect.NewAnonymizationServiceHandler(
		anonymizationService,
		connect.WithInterceptors(interceptors...),
	))
	mux.Handle(mgmtv1alpha1connect.NewConnectionDataServiceHandler(
		connectionDataService,
		connect.WithInterceptors(interceptors...),
	))

	if isLicensed {
		mux.Handle(mgmtv1alpha1connect.NewAccountHookServiceHandler(
			accountHookService,
			connect.WithInterceptors(interceptors...),
		))
	} else {
		mux.Handle(mgmtv1alpha1connect.NewAccountHookServiceHandler(
			mgmtv1alpha1connect.UnimplementedAccountHookServiceHandler{},
			connect.WithInterceptors(interceptors...),
		))
	}

	return mux, nil
}

func (s *NeosyncApiTestClient) getEnforcedRbacClient(
	ctx context.Context,
	pgcontainer *tcpostgres.PostgresTestContainer,
) (rbac.Interface, error) {
	rbacenforcer, err := enforcer.NewActiveEnforcer(
		ctx,
		stdlib.OpenDBFromPool(pgcontainer.DB),
		"neosync_api.casbin_rule",
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create rbac enforcer: %w", err)
	}
	rbacenforcer.EnableAutoSave(true)
	err = rbacenforcer.LoadPolicy()
	if err != nil {
		return nil, fmt.Errorf("unable to load rbac policies: %w", err)
	}
	return rbac.New(rbacenforcer), nil
}
