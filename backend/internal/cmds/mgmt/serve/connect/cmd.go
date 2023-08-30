package serve_connect

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/otelconnect"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/auth"
	authjwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	authmw "github.com/nucleuscloud/neosync/backend/internal/auth/middleware"
	auth_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/auth"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	neosync_k8sclient "github.com/nucleuscloud/neosync/backend/internal/k8s/client"
	neosynclogger "github.com/nucleuscloud/neosync/backend/internal/logger"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	v1alpha1_authservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/auth-service"
	v1alpha1_connectionservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-service"
	v1alpha1_jobservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/job-service"

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

func serve(
	ctx context.Context,
) error {
	port := viper.GetInt32("PORT")
	if port == 0 {
		port = 8080
	}
	host := viper.GetString("HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	environment := viper.GetString("NUCLEUS_ENV")

	logger := neosynclogger.New(neosynclogger.ShouldFormatAsJson(), nil).
		With("nucleusEnv", environment)
	loglogger := neosynclogger.NewLogLogger(neosynclogger.ShouldFormatAsJson(), nil)

	mux := http.NewServeMux()

	services := []string{
		mgmtv1alpha1connect.UserAccountServiceName,
		mgmtv1alpha1connect.AuthServiceName,
		mgmtv1alpha1connect.ConnectionServiceName,
		mgmtv1alpha1connect.JobServiceName,
	}

	checker := grpchealth.NewStaticChecker(services...)
	mux.Handle(grpchealth.NewHandler(checker))

	reflector := grpcreflect.NewStaticReflector(services...)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	dbconfig, err := getDbConfig()
	if err != nil {
		panic(err)
	}
	db, err := nucleusdb.NewFromConfig(dbconfig)
	if err != nil {
		panic(err)
	}

	jwtClientConfig, err := getJwtClientConfig()
	if err != nil {
		panic(err)
	}
	jwtClient, err := authjwt.New(jwtClientConfig)
	if err != nil {
		panic(err)
	}
	authMiddleware := authmw.New(jwtClient, db)

	auth0Cfg, err := getAuthClientConfig()
	if err != nil {
		return err
	}
	auth0Client, err := auth.New(auth0Cfg.AuthBaseUrl, auth0Cfg.AuthClientIdSecretMap)
	if err != nil {
		return err
	}

	neosyncK8sClient, err := neosync_k8sclient.New()
	if err != nil {
		return err
	}

	stdInterceptors := connect.WithInterceptors(
		otelconnect.NewInterceptor(),
		auth_interceptor.NewInterceptor(authMiddleware.ValidateAndInjectAll),
		logger_interceptor.NewInterceptor(logger),
	)

	api := http.NewServeMux()

	userAccountService := v1alpha1_useraccountservice.New(&v1alpha1_useraccountservice.Config{}, db)
	api.Handle(
		mgmtv1alpha1connect.NewUserAccountServiceHandler(
			userAccountService,
			stdInterceptors,
		),
	)
	connectionService := v1alpha1_connectionservice.New(&v1alpha1_connectionservice.Config{}, db, neosyncK8sClient, userAccountService)
	api.Handle(
		mgmtv1alpha1connect.NewConnectionServiceHandler(
			connectionService,
			stdInterceptors,
		),
	)

	jobService := v1alpha1_jobservice.New(&v1alpha1_jobservice.Config{}, db, neosyncK8sClient, userAccountService, connectionService)
	api.Handle(
		mgmtv1alpha1connect.NewJobServiceHandler(
			jobService,
			stdInterceptors,
		),
	)

	api.Handle(
		mgmtv1alpha1connect.NewAuthServiceHandler(
			v1alpha1_authservice.New(&v1alpha1_authservice.Config{}, auth0Client),
			connect.WithInterceptors(
				otelconnect.NewInterceptor(),
				logger_interceptor.NewInterceptor(logger),
			),
		),
	)

	mux.Handle("/", api)

	addr := fmt.Sprintf("%s:%d", host, port)

	logger.Info(fmt.Sprintf("listening on %s", addr))
	httpServer := http.Server{
		Addr:     addr,
		Handler:  h2c.NewHandler(mux, &http2.Server{}),
		ErrorLog: loglogger,
	}

	if err = httpServer.ListenAndServe(); err != nil {
		logger.Error(err.Error())
	}
	return nil
}

func getJwtClientConfig() (*authjwt.JwtClientConfig, error) {
	authBaseUrl := viper.GetString("AUTH0_BASEURL")
	if authBaseUrl == "" {
		return nil, fmt.Errorf("must provide AUTH0_BASEURL in environment")
	}

	authAudience := viper.GetString("AUTH0_AUDIENCE")
	if authAudience == "" {
		return nil, fmt.Errorf("must provide AUTH0_AUDIENCE in environment")
	}
	return &authjwt.JwtClientConfig{
		BaseUrl:      authBaseUrl,
		ApiAudiences: []string{authAudience},
	}, nil
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

type authClientConfig struct {
	AuthBaseUrl           string
	AuthAudience          string
	AuthClientIdSecretMap map[string]string
}

func getAuthClientConfig() (*authClientConfig, error) {
	authBaseUrl := viper.GetString("AUTH0_BASEURL")
	if authBaseUrl == "" {
		return nil, fmt.Errorf("must provide AUTH0_BASEURL in environment")
	}

	authAudience := viper.GetString("AUTH0_AUDIENCE")
	if authAudience == "" {
		return nil, fmt.Errorf("must provide AUTH0_AUDIENCE in environment")
	}

	authClientIdSecretMap := viper.GetStringMapString("AUTH0_CLIENTID_SECRET")
	if len(authClientIdSecretMap) == 0 {
		return nil, fmt.Errorf("must provide AUTH0_CLIENTID_SECRET in environment with at least one clientId + secret pair")
	}
	return &authClientConfig{
		AuthBaseUrl:           authBaseUrl,
		AuthAudience:          authAudience,
		AuthClientIdSecretMap: authClientIdSecretMap,
	}, nil
}
