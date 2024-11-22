package serve_connect

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/otelconnect"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/go-logr/logr"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/internal/connectrpc/validate"
	http_client "github.com/nucleuscloud/neosync/internal/http/client"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	"github.com/nucleuscloud/neosync/backend/internal/auth/authmw"
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	clientcredtokenprovider "github.com/nucleuscloud/neosync/backend/internal/auth/clientcred_token_provider"
	auth_jwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt/auth0"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt/keycloak"
	accountid_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/accountid"
	auth_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/auth"
	authlogging_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/auth_logging"
	bookend_logging_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/bookend"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	jobhooks "github.com/nucleuscloud/neosync/backend/internal/ee/hooks/jobs"
	neosync_gcp "github.com/nucleuscloud/neosync/backend/internal/gcp"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/internal/temporal/clientmanager"
	neosynclogger "github.com/nucleuscloud/neosync/backend/pkg/logger"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	v1alpha1_anonymizationservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/anonymization-service"
	v1alpha1_apikeyservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/api-key-service"
	v1alpha1_authservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/auth-service"
	v1alpha1_connectiondataservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-data-service"
	v1alpha1_connectionservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-service"
	v1alpha1_jobservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/job-service"
	v1alpha1_metricsservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/metrics-service"
	v1alpha1_transformerservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/transformers-service"
	v1alpha1_useraccountservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/user-account-service"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
	"github.com/nucleuscloud/neosync/internal/billing"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	neomigrate "github.com/nucleuscloud/neosync/internal/migrate"
	neosyncotel "github.com/nucleuscloud/neosync/internal/otel"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promconfig "github.com/prometheus/common/config"

	stripeapiclient "github.com/stripe/stripe-go/v79/client"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "connect",
		Short: "serves up connect",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return serve(cmd.Context())
		},
	}
}

func serve(ctx context.Context) error {
	port := viper.GetInt32("PORT")
	if port == 0 {
		port = 8080
	}
	host := viper.GetString("HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	slogger, loglogger := neosynclogger.NewLoggers()

	neoEnv := viper.GetString("NUCLEUS_ENV")
	if neoEnv != "" {
		slogger = slogger.With("nucleusEnv", neoEnv)
	}

	slog.SetDefault(slogger) // set default logger for methods that can't easily access the configured logger

	eelicense, err := license.NewFromEnv()
	if err != nil {
		return fmt.Errorf("unable to initialize ee license from env: %w", err)
	}
	slogger.Debug(fmt.Sprintf("ee license enabled: %t", eelicense.IsValid()))

	if getIsNeosyncCloud() {
		slogger.Debug("neosync cloud is enabled")
	}

	mux := http.NewServeMux()

	services := []string{
		mgmtv1alpha1connect.UserAccountServiceName,
		mgmtv1alpha1connect.AuthServiceName,
		mgmtv1alpha1connect.ConnectionServiceName,
		mgmtv1alpha1connect.JobServiceName,
		mgmtv1alpha1connect.TransformersServiceName,
		mgmtv1alpha1connect.ApiKeyServiceName,
		mgmtv1alpha1connect.ConnectionDataServiceName,
		mgmtv1alpha1connect.AnonymizationServiceName,
	}

	if shouldEnableMetricsService() {
		services = append(services, mgmtv1alpha1connect.MetricsServiceName)
	}

	checker := grpchealth.NewStaticChecker(services...)
	mux.Handle(grpchealth.NewHandler(checker))

	reflectorServices := append([]string{
		grpchealth.HealthV1ServiceName,
	}, services...)
	reflector := grpcreflect.NewStaticReflector(reflectorServices...)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// prevents the server from crashing on panics and returns a valid error response to the user
	recoverHandler := func(_ context.Context, _ connect.Spec, _ http.Header, r any) error {
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("panic: %v", r))
	}

	dbconfig, err := getDbConfig()
	if err != nil {
		return err
	}

	db, err := neosyncdb.NewFromConfig(dbconfig)
	if err != nil {
		return err
	}

	if viper.GetBool("DB_AUTO_MIGRATE") {
		schemaDir := viper.GetString("DB_SCHEMA_DIR")
		if schemaDir == "" {
			return errors.New("must provide DB_SCHEMA_DIR env var to run auto db migrations")
		}
		dbMigConfig, err := getDbMigrationConfig()
		if err != nil {
			return err
		}
		slogger.Debug("DB_AUTO_MIGRATE is enabled, running migrations...", "migrationDir", schemaDir)
		if err := neomigrate.Up(
			ctx,
			neosyncdb.GetDbUrl(dbMigConfig),
			schemaDir,
			slogger,
		); err != nil {
			return fmt.Errorf("unable to complete database migrations: %w", err)
		}
	}

	stdInterceptors := []connect.Interceptor{}

	var anonymizerMeter metric.Meter
	otelconfig := neosyncotel.GetOtelConfigFromViperEnv()
	if otelconfig.IsEnabled {
		slogger.Debug("otel is enabled")
		tmPropagator := neosyncotel.NewDefaultPropagator()
		otelconnopts := []otelconnect.Option{otelconnect.WithoutServerPeerAttributes(), otelconnect.WithPropagator(tmPropagator)}
		traceProviders := []neosyncotel.TracerProvider{}
		meterProviders := []neosyncotel.MeterProvider{}

		meterprovider, err := neosyncotel.NewMeterProvider(ctx, &neosyncotel.MeterProviderConfig{
			Exporter:   otelconfig.MeterExporter,
			AppVersion: otelconfig.ServiceVersion,
			Opts: neosyncotel.MeterExporterOpts{
				Otlp:    []otlpmetricgrpc.Option{},
				Console: []stdoutmetric.Option{stdoutmetric.WithPrettyPrint()},
			},
		})
		if err != nil {
			return err
		}
		if meterprovider != nil {
			slogger.Debug("otel metering has been configured")
			otelconnopts = append(otelconnopts, otelconnect.WithMeterProvider(meterprovider))
			meterProviders = append(meterProviders, meterprovider)
		} else {
			otelconnopts = append(otelconnopts, otelconnect.WithoutMetrics())
		}

		anonymizeMeterProvider, err := neosyncotel.NewMeterProvider(ctx, &neosyncotel.MeterProviderConfig{
			Exporter:   otelconfig.MeterExporter,
			AppVersion: otelconfig.ServiceVersion,
			Opts: neosyncotel.MeterExporterOpts{
				Otlp:    []otlpmetricgrpc.Option{neosyncotel.WithDefaultDeltaTemporalitySelector()},
				Console: []stdoutmetric.Option{stdoutmetric.WithPrettyPrint()},
			},
		})
		if err != nil {
			return err
		}
		if anonymizeMeterProvider != nil {
			slogger.Debug("otel metering for anonymize service has been configured")
			meterProviders = append(meterProviders, anonymizeMeterProvider)
			anonymizerMeter = anonymizeMeterProvider.Meter("anonymizer")
		}

		traceprovider, err := neosyncotel.NewTraceProvider(ctx, &neosyncotel.TraceProviderConfig{
			Exporter: otelconfig.TraceExporter,
			Opts: neosyncotel.TraceExporterOpts{
				Otlp:    []otlptracegrpc.Option{},
				Console: []stdouttrace.Option{stdouttrace.WithPrettyPrint()},
			},
		})
		if err != nil {
			return err
		}
		if traceprovider != nil {
			slogger.Debug("otel tracing has been configured")
			otelconnopts = append(otelconnopts, otelconnect.WithTracerProvider(traceprovider))
			traceProviders = append(traceProviders, traceprovider)
		} else {
			otelconnopts = append(otelconnopts, otelconnect.WithoutTracing(), otelconnect.WithoutTraceEvents())
		}

		otelInterceptor, err := otelconnect.NewInterceptor(otelconnopts...)
		if err != nil {
			return err
		}
		stdInterceptors = append(stdInterceptors, otelInterceptor)

		otelshutdown := neosyncotel.SetupOtelSdk(&neosyncotel.SetupConfig{
			TraceProviders:    traceProviders,
			MeterProviders:    meterProviders,
			Logger:            logr.FromSlogHandler(slogger.Handler()),
			TextMapPropagator: tmPropagator,
		})
		defer func() {
			if err := otelshutdown(context.Background()); err != nil {
				slogger.Error(fmt.Errorf("unable to gracefully shutdown otel providers: %w", err).Error())
			}
		}()
	}

	validateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		return err
	}
	loggerInterceptor := logger_interceptor.NewInterceptor(slogger)
	loggerAccountIdInterceptor := accountid_interceptor.NewInterceptor()
	handlerBookendInterceptor := bookend_logging_interceptor.NewInterceptor(
		bookend_logging_interceptor.WithLogLevel(slog.LevelInfo),
	)

	stdInterceptors = append(
		stdInterceptors,
		loggerInterceptor,
		validateInterceptor,
		loggerAccountIdInterceptor,
	)

	// standard auth interceptors that should be applied to most services
	stdAuthInterceptors := []connect.Interceptor{}
	// this will only authenticate jwts, not api keys. Mostly used by just the api key service
	jwtOnlyAuthInterceptors := []connect.Interceptor{}

	// interceptors for auth service.
	authSvcInterceptors := []connect.Interceptor{}
	authSvcInterceptors = append(authSvcInterceptors, stdAuthInterceptors...)

	isAuthEnabled := viper.GetBool("AUTH_ENABLED")
	if isAuthEnabled {
		jwtcfg, err := getJwtClientConfig()
		if err != nil {
			return err
		}
		jwtclient, err := auth_jwt.New(jwtcfg)
		if err != nil {
			return err
		}
		apikeyClient := auth_apikey.New(db.Q, db.Db, getAllowedWorkerApiKeys(getIsNeosyncCloud()), []string{
			mgmtv1alpha1connect.JobServiceGetJobProcedure,
			mgmtv1alpha1connect.JobServiceGetRunContextProcedure,
			mgmtv1alpha1connect.JobServiceSetRunContextProcedure,
			mgmtv1alpha1connect.JobServiceSetRunContextsProcedure,
			mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
			mgmtv1alpha1connect.TransformersServiceGetUserDefinedTransformerByIdProcedure,
			mgmtv1alpha1connect.ConnectionDataServiceGetConnectionForeignConstraintsProcedure,
			mgmtv1alpha1connect.ConnectionDataServiceGetConnectionPrimaryConstraintsProcedure,
			mgmtv1alpha1connect.ConnectionDataServiceGetConnectionInitStatementsProcedure,
			mgmtv1alpha1connect.UserAccountServiceIsAccountStatusValidProcedure,
			mgmtv1alpha1connect.UserAccountServiceGetBillingAccountsProcedure,
			mgmtv1alpha1connect.UserAccountServiceSetBillingMeterEventProcedure,
			mgmtv1alpha1connect.MetricsServiceGetDailyMetricCountProcedure,
			mgmtv1alpha1connect.AnonymizationServiceAnonymizeManyProcedure,
		})
		stdAuthInterceptors = append(
			stdAuthInterceptors,
			auth_interceptor.NewInterceptor(
				authmw.New(
					jwtclient,
					apikeyClient,
				).InjectTokenCtx,
			),
			authlogging_interceptor.NewInterceptor(db),
		)
		jwtOnlyAuthInterceptors = append(
			jwtOnlyAuthInterceptors,
			auth_interceptor.NewInterceptor(
				jwtclient.InjectTokenCtx,
			),
			authlogging_interceptor.NewInterceptor(db),
		)
		authSvcInterceptors = append(
			authSvcInterceptors,
			auth_interceptor.NewInterceptorWithExclude(
				jwtclient.InjectTokenCtx,
				[]string{
					mgmtv1alpha1connect.AuthServiceGetAuthStatusProcedure,
					mgmtv1alpha1connect.AuthServiceGetAuthorizeUrlProcedure,
					mgmtv1alpha1connect.AuthServiceGetCliIssuerProcedure,
					mgmtv1alpha1connect.AuthServiceLoginCliProcedure,
					mgmtv1alpha1connect.AuthServiceRefreshCliProcedure,
				},
			),
			authlogging_interceptor.NewInterceptor(db),
		)
	}

	api := http.NewServeMux()

	authBaseUrl := getAuthBaseUrl()
	clientIdSecretMap := getAuthClientIdSecretMap()

	authclient := auth_client.New(authBaseUrl, clientIdSecretMap)

	var issuerStr string
	issuerUrl, err := url.Parse(authBaseUrl + "/")
	if err != nil {
		if isAuthEnabled {
			return err
		}
	} else {
		issuerStr = issuerUrl.String()
	}

	authService := v1alpha1_authservice.New(&v1alpha1_authservice.Config{
		IsAuthEnabled: isAuthEnabled,
		CliClientId:   viper.GetString("AUTH_CLI_CLIENT_ID"),
		CliAudience:   getAuthCliAudience(),
		IssuerUrl:     issuerStr,
	}, authclient)
	api.Handle(
		mgmtv1alpha1connect.NewAuthServiceHandler(
			authService,
			connect.WithInterceptors(authSvcInterceptors...),
			connect.WithInterceptors(handlerBookendInterceptor),
			connect.WithRecover(recoverHandler),
		),
	)

	authcerts, err := getTemporalAuthCertificate()
	if err != nil {
		return err
	}

	var temporalTlsConfig *tls.Config
	if len(authcerts) > 0 {
		temporalTlsConfig = &tls.Config{
			Certificates: authcerts,
			MinVersion:   tls.VersionTLS13,
		}
	}
	temporalConfigProvider := clientmanager.NewDBConfigProvider(&clientmanager.TemporalConfig{
		Url:              getDefaultTemporalUrl(),
		Namespace:        getDefaultTemporalNamespace(),
		SyncJobQueueName: getDefaultTemporalSyncJobQueue(),
		TLSConfig:        temporalTlsConfig,
	}, db.Q, db.Db)
	tfwfmgr := clientmanager.NewClientManager(
		temporalConfigProvider,
		clientmanager.NewTemporalClientFactory(),
	)

	authadminclient, err := getAuthAdminClient(ctx, authclient, slogger)
	if err != nil {
		return err
	}
	promclient, err := getPromClientFromEnvironment()
	if err != nil {
		return err
	}

	stripeclient := getStripeApiClient()
	var billingClient billing.Interface
	if stripeclient != nil {
		slogger.Debug("stripe client is enabled")
		priceLookups, err := getStripePriceLookupMap()
		if err != nil {
			return err
		}
		billingClient = billing.New(stripeclient, &billing.Config{
			AppBaseUrl:   getAppBaseUrl(),
			PriceLookups: priceLookups,
		})
	}

	useraccountService := v1alpha1_useraccountservice.New(&v1alpha1_useraccountservice.Config{
		IsAuthEnabled:            isAuthEnabled,
		IsNeosyncCloud:           getIsNeosyncCloud(),
		DefaultMaxAllowedRecords: getDefaultMaxAllowedRecords(),
	}, db, temporalConfigProvider, authclient, authadminclient, billingClient)
	api.Handle(
		mgmtv1alpha1connect.NewUserAccountServiceHandler(
			useraccountService,
			connect.WithInterceptors(stdInterceptors...),
			connect.WithInterceptors(stdAuthInterceptors...),
			connect.WithInterceptors(handlerBookendInterceptor),
			connect.WithRecover(recoverHandler),
		),
	)

	apiKeyService := v1alpha1_apikeyservice.New(&v1alpha1_apikeyservice.Config{
		IsAuthEnabled: isAuthEnabled,
	}, db, useraccountService)
	api.Handle(
		mgmtv1alpha1connect.NewApiKeyServiceHandler(
			apiKeyService,
			connect.WithInterceptors(stdInterceptors...),
			connect.WithInterceptors(jwtOnlyAuthInterceptors...),
			connect.WithInterceptors(handlerBookendInterceptor),
			connect.WithRecover(recoverHandler),
		),
	)

	awsManager := awsmanager.New()
	pgquerier := pg_queries.New()
	mysqlquerier := mysql_queries.New()
	sqlConnector := &sqlconnect.SqlOpenConnector{}
	pgpoolmap := &sync.Map{}
	mysqlpoolmap := &sync.Map{}
	mssqlpoolmap := &sync.Map{}
	mssqlquerier := mssql_queries.New()
	sqlmanager := sql_manager.NewSqlManager(pgpoolmap, pgquerier, mysqlpoolmap, mysqlquerier, mssqlpoolmap, mssqlquerier, sqlConnector)
	mongoconnector := mongoconnect.NewConnector()
	connectionService := v1alpha1_connectionservice.New(&v1alpha1_connectionservice.Config{}, db, useraccountService, sqlConnector, pgquerier,
		mysqlquerier, mssqlquerier, mongoconnector, awsManager)
	api.Handle(
		mgmtv1alpha1connect.NewConnectionServiceHandler(
			connectionService,
			connect.WithInterceptors(stdInterceptors...),
			connect.WithInterceptors(stdAuthInterceptors...),
			connect.WithInterceptors(handlerBookendInterceptor),
			connect.WithRecover(recoverHandler),
		),
	)

	jobhookOpts := []jobhooks.Option{}
	// if getIsNeosyncCloud() || eelicense.IsValid() {
	jobhookOpts = append(jobhookOpts, jobhooks.WithEnabled())
	// }

	jobhookService := jobhooks.New(
		db,
		useraccountService,
		jobhookOpts...,
	)

	runLogConfig, err := getRunLogConfig()
	if err != nil {
		return err
	}

	jobServiceConfig := &v1alpha1_jobservice.Config{
		IsAuthEnabled:  isAuthEnabled,
		IsNeosyncCloud: getIsNeosyncCloud(),
		RunLogConfig:   runLogConfig,
	}
	jobService := v1alpha1_jobservice.New(
		jobServiceConfig,
		db,
		tfwfmgr,
		connectionService,
		useraccountService,
		sqlmanager,
		jobhookService,
	)
	api.Handle(
		mgmtv1alpha1connect.NewJobServiceHandler(
			jobService,
			connect.WithInterceptors(stdInterceptors...),
			connect.WithInterceptors(stdAuthInterceptors...),
			connect.WithInterceptors(handlerBookendInterceptor),
			connect.WithRecover(recoverHandler),
		),
	)

	var presAnalyzeClient presidioapi.AnalyzeInterface
	var presAnonClient presidioapi.AnonymizeInterface
	var presEntityClient presidioapi.EntityInterface
	if getIsNeosyncCloud() {
		analyzeClient, ok, err := getPresidioAnalyzeClient()
		if err != nil {
			return fmt.Errorf("unable to initialize presidio analyze client: %w", err)
		}
		if ok {
			slogger.Debug("presidio analyze client is enabled")
			presAnalyzeClient = analyzeClient
			presEntityClient = analyzeClient
		}
		anonClient, ok, err := getPresidioAnonymizeClient()
		if err != nil {
			return fmt.Errorf("unable to initialize presidio anonymize client: %w", err)
		}
		if ok {
			slogger.Debug("presidio anonymize client is enabled")
			presAnonClient = anonClient
		}
	}

	transformerService := v1alpha1_transformerservice.New(&v1alpha1_transformerservice.Config{
		IsPresidioEnabled: getIsNeosyncCloud(),
		IsNeosyncCloud:    getIsNeosyncCloud(),
	}, db, useraccountService, presEntityClient)
	api.Handle(
		mgmtv1alpha1connect.NewTransformersServiceHandler(
			transformerService,
			connect.WithInterceptors(stdInterceptors...),
			connect.WithInterceptors(stdAuthInterceptors...),
			connect.WithInterceptors(handlerBookendInterceptor),
			connect.WithRecover(recoverHandler),
		),
	)

	anonymizationService := v1alpha1_anonymizationservice.New(&v1alpha1_anonymizationservice.Config{
		IsPresidioEnabled: getIsNeosyncCloud(),
		IsAuthEnabled:     isAuthEnabled,
		IsNeosyncCloud:    getIsNeosyncCloud(),
	}, anonymizerMeter, useraccountService, presAnalyzeClient, presAnonClient, db)
	api.Handle(
		mgmtv1alpha1connect.NewAnonymizationServiceHandler(
			anonymizationService,
			connect.WithInterceptors(stdInterceptors...),
			connect.WithInterceptors(stdAuthInterceptors...),
			connect.WithInterceptors(handlerBookendInterceptor),
			connect.WithRecover(recoverHandler),
		),
	)

	gcpmanager := neosync_gcp.NewManager()
	connectionDataService := v1alpha1_connectiondataservice.New(
		&v1alpha1_connectiondataservice.Config{},
		useraccountService,
		connectionService,
		jobService,
		awsManager,
		sqlConnector,
		pgquerier,
		mysqlquerier,
		mongoconnector,
		sqlmanager,
		gcpmanager,
	)
	api.Handle(
		mgmtv1alpha1connect.NewConnectionDataServiceHandler(
			connectionDataService,
			connect.WithInterceptors(stdInterceptors...),
			connect.WithInterceptors(stdAuthInterceptors...),
			connect.WithInterceptors(handlerBookendInterceptor),
			connect.WithRecover(recoverHandler),
		),
	)

	if shouldEnableMetricsService() {
		metricsService := v1alpha1_metricsservice.New(
			&v1alpha1_metricsservice.Config{},
			useraccountService,
			jobService,
			promv1.NewAPI(promclient),
		)
		api.Handle(
			mgmtv1alpha1connect.NewMetricsServiceHandler(
				metricsService,
				connect.WithInterceptors(stdInterceptors...),
				connect.WithInterceptors(stdAuthInterceptors...),
				connect.WithInterceptors(handlerBookendInterceptor),
				connect.WithRecover(recoverHandler),
			),
		)
	}
	mux.Handle("/", api)

	httpServer := http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ErrorLog:          loglogger,
		ReadHeaderTimeout: 10 * time.Second,
	}

	slogger.Info(fmt.Sprintf("listening on %s", httpServer.Addr))

	if err = httpServer.ListenAndServe(); err != nil {
		slogger.Error(err.Error())
	}
	return nil
}

func getPromClientFromEnvironment() (promapi.Client, error) {
	roundTripper := promapi.DefaultRoundTripper
	promApiKey := getPromApiKey()
	if promApiKey != nil {
		roundTripper = promconfig.NewAuthorizationCredentialsRoundTripper("Bearer", promconfig.NewInlineSecret(*promApiKey), promapi.DefaultRoundTripper)
	}
	return promapi.NewClient(promapi.Config{
		Address:      getPromApiUrl(),
		RoundTripper: roundTripper,
	})
}

func getDbConfig() (*neosyncdb.ConnectConfig, error) {
	dbHost := viper.GetString("DB_HOST")
	if dbHost == "" {
		return nil, fmt.Errorf("must provide DB_HOST in environment")
	}

	dbPort := viper.GetInt("DB_PORT")
	if dbPort == 0 {
		return nil, fmt.Errorf("must provide DB_PORT in environment")
	}

	dbName := viper.GetString("DB_NAME")
	if dbName == "" {
		return nil, fmt.Errorf("must provide DB_NAME in environment")
	}

	dbUser := viper.GetString("DB_USER")
	if dbUser == "" {
		return nil, fmt.Errorf("must provide DB_USER in environment")
	}

	dbPass := viper.GetString("DB_PASS")
	if dbPass == "" {
		return nil, fmt.Errorf("must provide DB_PASS in environment")
	}

	sslMode := "require"
	if viper.IsSet("DB_SSL_DISABLE") && viper.GetBool("DB_SSL_DISABLE") {
		sslMode = "disable"
	}

	var dbOptions *string
	if viper.IsSet("DB_OPTIONS") {
		val := viper.GetString("DB_OPTIONS")
		dbOptions = &val
	}

	return &neosyncdb.ConnectConfig{
		Host:     dbHost,
		Port:     dbPort,
		Database: dbName,
		User:     dbUser,
		Pass:     dbPass,
		SslMode:  &sslMode,
		Options:  dbOptions,
	}, nil
}

func getDbMigrationConfig() (*neosyncdb.ConnectConfig, error) {
	dbHost := viper.GetString("DB_HOST")
	if dbHost == "" {
		return nil, fmt.Errorf("must provide DB_HOST in environment")
	}

	dbPort := viper.GetInt("DB_PORT")
	if dbPort == 0 {
		return nil, fmt.Errorf("must provide DB_PORT in environment")
	}

	dbName := viper.GetString("DB_NAME")
	if dbName == "" {
		return nil, fmt.Errorf("must provide DB_NAME in environment")
	}

	dbUser := viper.GetString("DB_USER")
	if dbUser == "" {
		return nil, fmt.Errorf("must provide DB_USER in environment")
	}

	dbPass := viper.GetString("DB_PASS")
	if dbPass == "" {
		return nil, fmt.Errorf("must provide DB_PASS in environment")
	}

	sslMode := "require"
	if viper.IsSet("DB_SSL_DISABLE") && viper.GetBool("DB_SSL_DISABLE") {
		sslMode = "disable"
	}

	var migrationsTable *string
	if viper.IsSet("DB_MIGRATIONS_TABLE") {
		table := viper.GetString("DB_MIGRATIONS_TABLE")
		migrationsTable = &table
	}

	var tableQuoted *bool
	if viper.IsSet("DB_MIGRATIONS_TABLE_QUOTED") {
		isQuoted := viper.GetBool("DB_MIGRATIONS_TABLE_QUOTED")
		tableQuoted = &isQuoted
	}

	var dbOptions *string
	if viper.IsSet("DB_MIGRATIONS_OPTIONS") {
		val := viper.GetString("DB_MIGRATIONS_OPTIONS")
		dbOptions = &val
	}

	return &neosyncdb.ConnectConfig{
		Host:                  dbHost,
		Port:                  dbPort,
		Database:              dbName,
		User:                  dbUser,
		Pass:                  dbPass,
		SslMode:               &sslMode,
		MigrationsTableName:   migrationsTable,
		MigrationsTableQuoted: tableQuoted,
		Options:               dbOptions,
	}, nil
}

func getTemporalAuthCertificate() ([]tls.Certificate, error) {
	keyPath := viper.GetString("TEMPORAL_CERT_KEY_PATH")
	certPath := viper.GetString("TEMPORAL_CERT_PATH")

	if keyPath != "" && certPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}
		return []tls.Certificate{cert}, nil
	}

	key := viper.GetString("TEMPORAL_CERT_KEY")
	cert := viper.GetString("TEMPORAL_CERT")
	if key != "" && cert != "" {
		cert, err := tls.X509KeyPair([]byte(cert), []byte(key))
		if err != nil {
			return nil, err
		}
		return []tls.Certificate{cert}, nil
	}
	return []tls.Certificate{}, nil
}

func getDefaultTemporalUrl() string {
	temporalUrl := viper.GetString("TEMPORAL_URL")
	if temporalUrl == "" {
		return "localhost:7233"
	}
	return temporalUrl
}
func getDefaultTemporalNamespace() string {
	ns := viper.GetString("TEMPORAL_DEFAULT_NAMESPACE")
	if ns == "" {
		return "default"
	}
	return ns
}

func getDefaultTemporalSyncJobQueue() string {
	name := viper.GetString("TEMPORAL_DEFAULT_SYNCJOB_QUEUE")
	if name == "" {
		return "sync-job"
	}
	return name
}

func getJwtClientConfig() (*auth_jwt.ClientConfig, error) {
	authBaseUrl := getAuthBaseUrl()
	authAudiences := getAuthAudiences()

	sigAlgo, err := getAuthSignatureAlgorithm()
	if err != nil {
		return nil, err
	}

	return &auth_jwt.ClientConfig{
		BackendIssuerUrl:   authBaseUrl,
		FrontendIssuerUrl:  getAuthExpectedIssUrl(),
		ApiAudiences:       authAudiences,
		SignatureAlgorithm: *sigAlgo,
	}, nil
}

var allowedSigningAlgorithms = map[validator.SignatureAlgorithm]bool{
	validator.EdDSA: true,
	validator.HS256: true,
	validator.HS384: true,
	validator.HS512: true,
	validator.RS256: true,
	validator.RS384: true,
	validator.RS512: true,
	validator.ES256: true,
	validator.ES384: true,
	validator.ES512: true,
	validator.PS256: true,
	validator.PS384: true,
	validator.PS512: true,
}

func getAuthSignatureAlgorithm() (*validator.SignatureAlgorithm, error) {
	algoStr := viper.GetString("AUTH_SIGNATURE_ALGORITHM")
	if algoStr == "" {
		rs256 := validator.RS256
		return &rs256, nil
	}
	if _, ok := allowedSigningAlgorithms[validator.SignatureAlgorithm(algoStr)]; !ok {
		return nil, errors.New("unsupported signature algorithm")
	}
	return (*validator.SignatureAlgorithm)(&algoStr), nil
}

func getAuthCliAudience() string {
	aud := viper.GetString("AUTH_CLI_AUDIENCE")
	if aud == "" {
		auds := getAuthAudiences()
		if len(auds) > 0 {
			return auds[0]
		}
	}
	return aud
}

func getAuthAudiences() []string {
	return viper.GetStringSlice("AUTH_AUDIENCE")
}

func getAuthBaseUrl() string {
	authBaseUrl := viper.GetString("AUTH_BASEURL")
	return authBaseUrl
}

func getAuthExpectedIssUrl() *string {
	iss := viper.GetString("AUTH_EXPECTED_ISS")
	if iss == "" {
		return nil
	}
	return &iss
}

func getAuthClientIdSecretMap() map[string]string {
	return viper.GetStringMapString("AUTH_CLIENTID_SECRET")
}

func getAuthApiBaseUrl() string {
	return viper.GetString("AUTH_API_BASEURL")
}

func getAuthApiClientId() string {
	return viper.GetString("AUTH_API_CLIENT_ID")
}

func getAuthApiClientSecret() string {
	return viper.GetString("AUTH_API_CLIENT_SECRET")
}

func getAuthApiProvider() string {
	return viper.GetString("AUTH_API_PROVIDER")
}

func getIsNeosyncCloud() bool {
	return viper.GetBool("NEOSYNC_CLOUD")
}

func getAllowedWorkerApiKeys(isNeosyncCloud bool) []string {
	if isNeosyncCloud {
		return viper.GetStringSlice("NEOSYNC_CLOUD_ALLOWED_WORKER_API_KEYS")
	}
	return []string{}
}

func getAuthAdminClient(ctx context.Context, authclient auth_client.Interface, logger *slog.Logger) (authmgmt.Interface, error) {
	authApiBaseUrl := getAuthApiBaseUrl()
	authApiClientId := getAuthApiClientId()
	authApiClientSecret := getAuthApiClientSecret()
	provider := getAuthApiProvider()
	if provider == "" || provider == "auth0" {
		return auth0.New(authApiBaseUrl, authApiClientId, authApiClientSecret)
	} else if provider == "keycloak" {
		tokenurl, err := authclient.GetTokenEndpoint(ctx)
		if err != nil {
			return nil, err
		}
		tokenProvider := clientcredtokenprovider.New(tokenurl, authApiClientId, authApiClientSecret, keycloak.DefaultTokenExpirationBuffer, logger)
		return keycloak.New(authApiBaseUrl, tokenProvider, logger)
	}
	logger.Warn(fmt.Sprintf("unable to initialize auth admin client due to unsupported provider: %q", provider))
	return &authmgmt.UnimplementedClient{}, nil
}

// whether or not to serve metrics via the metrics proto
// this is not the same as serving up prometheus compliant metrics endpoints
func shouldEnableMetricsService() bool {
	return viper.GetBool("METRICS_SERVICE_ENABLED")
}

func getPromApiUrl() string {
	baseurl := viper.GetString("METRICS_URL")
	if baseurl == "" {
		return "http://localhost:9090"
	}
	return baseurl
}
func getPromApiKey() *string {
	key := viper.GetString("METRICS_API_KEY")
	if key == "" {
		return nil
	}
	return &key
}

func getRunLogConfig() (*v1alpha1_jobservice.RunLogConfig, error) {
	isRunLogsEnabled := viper.GetBool("RUN_LOGS_ENABLED")
	if !isRunLogsEnabled {
		// look for fallback variables
		isKubernetes := getIsKubernetes()
		ksNs := getKubernetesNamespace()
		ksWorkerAppName := getKubernetesWorkerAppName()
		if isKubernetes {
			if ksNs == "" {
				ksNs = "neosync"
			}
			if ksWorkerAppName == "" {
				ksWorkerAppName = "neosync-worker"
			}
			runlogtype := v1alpha1_jobservice.KubePodRunLogType
			return &v1alpha1_jobservice.RunLogConfig{
				IsEnabled:  true,
				RunLogType: &runlogtype,
				RunLogPodConfig: &v1alpha1_jobservice.KubePodRunLogConfig{
					Namespace:     ksNs,
					WorkerAppName: ksWorkerAppName,
				},
			}, nil
		}
		return &v1alpha1_jobservice.RunLogConfig{
			IsEnabled: false,
		}, nil
	}
	runlogtype := getRunLogType()
	if runlogtype == nil {
		return nil, errors.New("run logs is enabled but run log type was unspecified or invalid")
	}
	switch *runlogtype {
	case v1alpha1_jobservice.KubePodRunLogType:
		ksNs := viper.GetString("RUN_LOGS_PODCONFIG_WORKER_NAMESPACE")
		ksWorkerAppName := viper.GetString("RUN_LOGS_PODCONFIG_WORKER_APPNAME")
		if ksNs == "" {
			ksNs = getKubernetesNamespace()
		}
		if ksNs == "" {
			ksNs = "neosync"
		}
		if ksWorkerAppName == "" {
			ksWorkerAppName = getKubernetesWorkerAppName()
		}
		if ksWorkerAppName == "" {
			ksWorkerAppName = "neosync-worker"
		}
		return &v1alpha1_jobservice.RunLogConfig{
			IsEnabled:  true,
			RunLogType: runlogtype,
			RunLogPodConfig: &v1alpha1_jobservice.KubePodRunLogConfig{
				Namespace:     ksNs,
				WorkerAppName: ksWorkerAppName,
			},
		}, nil
	case v1alpha1_jobservice.LokiRunLogType:
		lokibaseurl := viper.GetString("RUN_LOGS_LOKICONFIG_BASEURL")
		if lokibaseurl == "" {
			return nil, errors.New("must provide loki baseurl when loki run log type has been configured")
		}
		labelsQuery := viper.GetString("RUN_LOGS_LOKICONFIG_LABELSQUERY")
		if labelsQuery == "" {
			labelsQuery = `namespace="neosync", app="neosync-worker"`
		}
		keepLabels := viper.GetStringSlice("RUN_LOGS_LOKICONFIG_KEEPLABELS")
		return &v1alpha1_jobservice.RunLogConfig{
			IsEnabled:  true,
			RunLogType: runlogtype,
			LokiRunLogConfig: &v1alpha1_jobservice.LokiRunLogConfig{
				BaseUrl:     lokibaseurl,
				LabelsQuery: labelsQuery,
				KeepLabels:  keepLabels,
			},
		}, nil
	default:
		return nil, errors.New("unsupported or no run log type configured, but run logs are enabled.")
	}
}

func getRunLogType() *v1alpha1_jobservice.RunLogType {
	logtype := viper.GetString("RUN_LOGS_TYPE")
	switch logtype {
	case string(v1alpha1_jobservice.KubePodRunLogType):
		rt := v1alpha1_jobservice.KubePodRunLogType
		return &rt
	case string(v1alpha1_jobservice.LokiRunLogType):
		rt := v1alpha1_jobservice.LokiRunLogType
		return &rt
	default:
		return nil
	}
}

func getIsKubernetes() bool {
	return viper.GetBool("KUBERNETES_ENABLED")
}

func getKubernetesNamespace() string {
	return viper.GetString("KUBERNETES_NAMESPACE")
}

func getKubernetesWorkerAppName() string {
	return viper.GetString("KUBERNETES_WORKER_APP_NAME")
}

func getDefaultMaxAllowedRecords() *int64 {
	val := viper.GetInt64("MAX_ALLOWED_RECORDS")
	if val <= 0 {
		return nil
	}
	return &val
}

func getStripeApiClient() *stripeapiclient.API {
	apiKey := getStripeApiKey()
	if apiKey != nil {
		return stripeapiclient.New(*apiKey, nil)
	}
	return nil
}

func getStripeApiKey() *string {
	value := viper.GetString("STRIPE_API_KEY")
	if value == "" {
		return nil
	}
	return &value
}

func getStripePriceLookupMap() (billing.PriceQuantity, error) {
	value := viper.GetStringMapString("STRIPE_PRICE_LOOKUPS")

	output := billing.PriceQuantity{}
	for k, v := range value {
		if v == "" {
			output[k] = 0
			continue
		}
		quantity, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse value as int for billing quantity %q: %w", v, err)
		}
		output[k] = quantity
	}
	return output, nil
}

func getAppBaseUrl() string {
	return viper.GetString("APP_BASEURL")
}

func getPresidioAnalyzeClient() (*presidioapi.ClientWithResponses, bool, error) {
	endpoint := getPresidioAnalyzeEndpoint()
	if endpoint == "" {
		return nil, false, nil
	}
	return getPresidioClient(endpoint)
}

func getPresidioAnonymizeClient() (*presidioapi.ClientWithResponses, bool, error) {
	endpoint := getPresidioAnonymizeEndpoint()
	if endpoint == "" {
		return nil, false, nil
	}
	return getPresidioClient(endpoint)
}

func getPresidioClient(endpoint string) (*presidioapi.ClientWithResponses, bool, error) {
	httpclient := http_client.WithHeaders(&http.Client{}, getPresidioHttpHeaders())

	client, err := presidioapi.NewClientWithResponses(endpoint, presidioapi.WithHTTPClient(httpclient))
	if err != nil {
		return nil, false, err
	}

	return client, true, nil
}

func getPresidioAnalyzeEndpoint() string {
	return viper.GetString("PRESIDIO_ANALYZER_URL")
}
func getPresidioAnonymizeEndpoint() string {
	return viper.GetString("PRESIDIO_ANONYMIZER_URL")
}

func getPresidioHttpHeaders() map[string]string {
	output := map[string]string{}
	authtoken := getPresidioAuthTokenHeaderValue()
	if authtoken != nil && *authtoken != "" {
		output["Authorization"] = *authtoken
	}
	return output
}

func getPresidioAuthTokenHeaderValue() *string {
	val := viper.GetString("PRESIDIO_HEADER_AUTH_TOKEN")
	if val == "" {
		return nil
	}
	return &val
}
