package serve_connect

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	"github.com/nucleuscloud/neosync/backend/internal/auth/authmw"
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	auth_jwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt/auth0"
	awsmanager "github.com/nucleuscloud/neosync/backend/internal/aws"
	up_cmd "github.com/nucleuscloud/neosync/backend/internal/cmds/mgmt/migrate/up"
	auth_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/auth"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	logging_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logging"
	neosynclogger "github.com/nucleuscloud/neosync/backend/internal/logger"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/nucleuscloud/neosync/backend/internal/sqlconnect"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	v1alpha1_apikeyservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/api-key-service"
	v1alpha1_authservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/auth-service"
	v1alpha1_connectiondataservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-data-service"
	v1alpha1_connectionservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-service"
	v1alpha1_jobservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/job-service"
	v1alpha1_transformerservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/transformers-service"
	v1alpha1_useraccountservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/user-account-service"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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
	environment := viper.GetString("NUCLEUS_ENV")
	if environment == "" {
		environment = "unknown"
	}

	logger := neosynclogger.New(neosynclogger.ShouldFormatAsJson(), nil).
		With("nucleusEnv", environment)
	loglogger := neosynclogger.NewLogLogger(neosynclogger.ShouldFormatAsJson(), nil)

	slog.SetDefault(logger) // set default logger for methods that can't easily access the configured logger

	mux := http.NewServeMux()

	services := []string{
		mgmtv1alpha1connect.UserAccountServiceName,
		mgmtv1alpha1connect.AuthServiceName,
		mgmtv1alpha1connect.ConnectionServiceName,
		mgmtv1alpha1connect.JobServiceName,
		mgmtv1alpha1connect.TransformersServiceName,
		mgmtv1alpha1connect.ApiKeyServiceName,
		mgmtv1alpha1connect.ConnectionDataServiceName,
	}

	checker := grpchealth.NewStaticChecker(services...)
	mux.Handle(grpchealth.NewHandler(checker))

	reflector := grpcreflect.NewStaticReflector(services...)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	dbconfig, err := getDbConfig()
	if err != nil {
		return err
	}

	db, err := nucleusdb.NewFromConfig(dbconfig)
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
		if err := up_cmd.Up(
			ctx,
			nucleusdb.GetDbUrl(dbMigConfig),
			schemaDir,
			logger,
		); err != nil {
			return err
		}
	}

	validateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		return err
	}

	otelInterceptor := otelconnect.NewInterceptor()
	loggerInterceptor := logger_interceptor.NewInterceptor(logger)
	loggingInterceptor := logging_interceptor.NewInterceptor(logger)

	stdInterceptors := []connect.Interceptor{
		otelInterceptor,
		loggerInterceptor,
		validateInterceptor,
		loggingInterceptor,
	}

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
		apikeyClient := auth_apikey.New(db.Q, db.Db)
		stdAuthInterceptors = append(
			stdAuthInterceptors,
			auth_interceptor.NewInterceptor(
				authmw.New(
					jwtclient,
					apikeyClient,
				).InjectTokenCtx,
			),
		)
		jwtOnlyAuthInterceptors = append(
			jwtOnlyAuthInterceptors,
			auth_interceptor.NewInterceptor(
				jwtclient.InjectTokenCtx,
			),
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

	var auth0Mgmt *auth0.Auth0MgmtClient
	if isAuthEnabled {
		authApiBaseUrl := getAuthApiBaseUrl()
		authApiClientId := getAuthApiClientId()
		authApiClientSecret := getAuthApiClientSecret()
		authmanagement, err := auth0.New(authApiBaseUrl, authApiClientId, authApiClientSecret)
		if err != nil {
			return err
		}
		auth0Mgmt = authmanagement
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
		),
	)

	authcerts, err := getTemporalAuthCertificate()
	if err != nil {
		return err
	}
	tfwfmgr := clientmanager.New(&clientmanager.Config{
		AuthCertificates: authcerts,
		DefaultTemporalConfig: &clientmanager.DefaultTemporalConfig{
			Url:              getDefaultTemporalUrl(),
			Namespace:        getDefaultTemporalNamespace(),
			SyncJobQueueName: getDefaultTemporalSyncJobQueue(),
		},
	}, db.Q, db.Db)

	useraccountService := v1alpha1_useraccountservice.New(&v1alpha1_useraccountservice.Config{
		IsAuthEnabled: isAuthEnabled,
	}, db, auth0Mgmt, tfwfmgr)
	api.Handle(
		mgmtv1alpha1connect.NewUserAccountServiceHandler(
			useraccountService,
			connect.WithInterceptors(stdInterceptors...),
			connect.WithInterceptors(stdAuthInterceptors...),
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
		),
	)

	sqlConnector := &sqlconnect.SqlOpenConnector{}
	connectionService := v1alpha1_connectionservice.New(&v1alpha1_connectionservice.Config{}, db, useraccountService, sqlConnector)
	api.Handle(
		mgmtv1alpha1connect.NewConnectionServiceHandler(
			connectionService,
			connect.WithInterceptors(stdInterceptors...),
			connect.WithInterceptors(stdAuthInterceptors...),
		),
	)

	jobServiceConfig := &v1alpha1_jobservice.Config{
		IsAuthEnabled: isAuthEnabled,
	}
	jobService := v1alpha1_jobservice.New(
		jobServiceConfig,
		db,
		tfwfmgr,
		connectionService,
		useraccountService,
	)
	api.Handle(
		mgmtv1alpha1connect.NewJobServiceHandler(
			jobService,
			connect.WithInterceptors(stdInterceptors...),
			connect.WithInterceptors(stdAuthInterceptors...),
		),
	)

	transformerService := v1alpha1_transformerservice.New(&v1alpha1_transformerservice.Config{}, db, useraccountService)
	api.Handle(
		mgmtv1alpha1connect.NewTransformersServiceHandler(
			transformerService,
			connect.WithInterceptors(stdInterceptors...),
			connect.WithInterceptors(stdAuthInterceptors...),
		),
	)

	pgquerier := pg_queries.New()
	mysqlquerier := mysql_queries.New()
	awsManager := awsmanager.New()
	connectionDataService := v1alpha1_connectiondataservice.New(
		&v1alpha1_connectiondataservice.Config{},
		useraccountService,
		connectionService,
		jobService,
		awsManager,
		sqlConnector,
		pgquerier,
		mysqlquerier,
	)
	api.Handle(
		mgmtv1alpha1connect.NewConnectionDataServiceHandler(
			connectionDataService,
			connect.WithInterceptors(stdInterceptors...),
			connect.WithInterceptors(stdAuthInterceptors...),
		),
	)

	mux.Handle("/", api)

	httpServer := http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ErrorLog:          loglogger,
		ReadHeaderTimeout: 10 * time.Second,
	}

	logger.Info(fmt.Sprintf("listening on %s", httpServer.Addr))

	if err = httpServer.ListenAndServe(); err != nil {
		logger.Error(err.Error())
	}
	return nil
}

func getDbConfig() (*nucleusdb.ConnectConfig, error) {
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

	return &nucleusdb.ConnectConfig{
		Host:     dbHost,
		Port:     dbPort,
		Database: dbName,
		User:     dbUser,
		Pass:     dbPass,
		SslMode:  &sslMode,
	}, nil
}

func getDbMigrationConfig() (*nucleusdb.ConnectConfig, error) {
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

	return &nucleusdb.ConnectConfig{
		Host:                  dbHost,
		Port:                  dbPort,
		Database:              dbName,
		User:                  dbUser,
		Pass:                  dbPass,
		SslMode:               &sslMode,
		MigrationsTableName:   migrationsTable,
		MigrationsTableQuoted: tableQuoted,
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
		cert, err := tls.X509KeyPair([]byte(key), []byte(cert))
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
