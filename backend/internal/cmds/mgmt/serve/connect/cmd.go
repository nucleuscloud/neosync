package serve_connect

import (
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
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"

	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	"github.com/nucleuscloud/neosync/backend/internal/authmw"
	auth_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/auth"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	auth_jwt "github.com/nucleuscloud/neosync/backend/internal/jwt"
	neosynclogger "github.com/nucleuscloud/neosync/backend/internal/logger"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	v1alpha1_authservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/auth-service"
	v1alpha1_connectionservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-service"
	v1alpha1_jobservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/job-service"
	v1alpha1_transformerservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/transformers-service"
	v1alpha1_useraccountservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/user-account-service"
	temporalclient "go.temporal.io/sdk/client"

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
			return serve()
		},
	}
}

func serve() error {
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

	validateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		return err
	}

	stdInterceptors := []connect.Interceptor{
		otelconnect.NewInterceptor(),
		logger_interceptor.NewInterceptor(logger),
		validateInterceptor,
	}

	isAuthEnabled := viper.GetBool("AUTH_ENABLED")
	if isAuthEnabled {
		jwtclient, err := auth_jwt.New(getJwtClientConfig())
		if err != nil {
			return err
		}
		stdInterceptors = append(stdInterceptors, auth_interceptor.NewInterceptor(authmw.New(jwtclient).ValidateAndInjectAll))
	}

	stdInterceptorConnectOpt := connect.WithInterceptors(
		stdInterceptors...,
	)

	api := http.NewServeMux()

	useraccountService := v1alpha1_useraccountservice.New(&v1alpha1_useraccountservice.Config{
		IsAuthEnabled: isAuthEnabled,
	}, db)
	api.Handle(
		mgmtv1alpha1connect.NewUserAccountServiceHandler(
			useraccountService,
			stdInterceptorConnectOpt,
		),
	)

	connectionService := v1alpha1_connectionservice.New(&v1alpha1_connectionservice.Config{}, db, useraccountService)
	api.Handle(
		mgmtv1alpha1connect.NewConnectionServiceHandler(
			connectionService,
			stdInterceptorConnectOpt,
		),
	)

	temporalConfig := getTemporalConfig(logger)
	temporalClient, err := temporalclient.NewLazyClient(*temporalConfig)
	if err != nil {
		return err
	}
	defer temporalClient.Close()
	jobServiceConfig := &v1alpha1_jobservice.Config{
		TemporalTaskQueue: getTemporalTaskQueue(),
		TemporalNamespace: getTemporalNamespace(),
	}
	jobService := v1alpha1_jobservice.New(jobServiceConfig, db, temporalClient, connectionService, useraccountService)
	api.Handle(
		mgmtv1alpha1connect.NewJobServiceHandler(
			jobService,
			stdInterceptorConnectOpt,
		),
	)

	transformerService := v1alpha1_transformerservice.New(&v1alpha1_transformerservice.Config{}, db, useraccountService)
	api.Handle(
		mgmtv1alpha1connect.NewTransformersServiceHandler(
			transformerService,
			stdInterceptorConnectOpt,
		),
	)

	authBaseUrl := getAuthBaseUrl()
	tokenUrl := getAuthTokenUrl()
	clientIdSecretMap := getAuthClientIdSecretMap()

	authclient := auth_client.New(tokenUrl, clientIdSecretMap)

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
		AuthorizeUrl:  getAuthAuthorizeUrl(),
		CliClientId:   viper.GetString("AUTH_CLI_CLIENT_ID"),
		CliAudience:   getAuthCliAudience(),
		IssuerUrl:     issuerStr,
	}, authclient)
	api.Handle(
		mgmtv1alpha1connect.NewAuthServiceHandler(
			authService,
			connect.WithInterceptors(
				otelconnect.NewInterceptor(),
				logger_interceptor.NewInterceptor(logger),
				validateInterceptor,
			),
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

func getTemporalConfig(
	logger *slog.Logger,
) *temporalclient.Options {
	temporalUrl := viper.GetString("TEMPORAL_URL")
	if temporalUrl == "" {
		temporalUrl = "localhost:7233"
	}

	temporalNamespace := getTemporalNamespace()

	return &temporalclient.Options{
		Logger:    logger.With("temporalClient", "true"),
		HostPort:  temporalUrl,
		Namespace: temporalNamespace,
		// Interceptors: ,
		// HeadersProvider: ,
	}
}

func getTemporalNamespace() string {
	temporalNamespace := viper.GetString("TEMPORAL_NAMESPACE")
	if temporalNamespace == "" {
		temporalNamespace = "default"
	}
	return temporalNamespace
}

func getTemporalTaskQueue() string {
	taskQueue := viper.GetString("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		return "default"
	}
	return taskQueue
}

func getJwtClientConfig() *auth_jwt.ClientConfig {
	authBaseUrl := getAuthBaseUrl()
	authAudiences := getAuthAudiences()

	return &auth_jwt.ClientConfig{
		BaseUrl:      authBaseUrl,
		ApiAudiences: authAudiences,
	}
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

func getAuthTokenUrl() string {
	baseUrl := getAuthBaseUrl()
	return fmt.Sprintf("%s/oauth/token", baseUrl)
}

func getAuthAuthorizeUrl() string {
	baseUrl := getAuthBaseUrl()
	return fmt.Sprintf("%s/authorize", baseUrl)
}

func getAuthClientIdSecretMap() map[string]string {
	return viper.GetStringMapString("AUTH_CLIENTID_SECRET")
}
