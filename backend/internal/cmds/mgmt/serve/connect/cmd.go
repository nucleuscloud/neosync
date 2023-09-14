package serve_connect

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/otelconnect"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"

	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	neosynclogger "github.com/nucleuscloud/neosync/backend/internal/logger"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	v1alpha1_connectionservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-service"
	v1alpha1_jobservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/job-service"
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

	temporalConfig := getTemporalConfig()
	temporalClient, err := temporalclient.Dial(*temporalConfig)
	if err != nil {
		panic(err)
	}
	defer temporalClient.Close()

	temporalTaskQueue, err := getTemporalTaskQueue()
	if err != nil {
		panic(err)
	}

	stdInterceptors := connect.WithInterceptors(
		otelconnect.NewInterceptor(),
		logger_interceptor.NewInterceptor(logger),
	)

	api := http.NewServeMux()

	connectionService := v1alpha1_connectionservice.New(&v1alpha1_connectionservice.Config{}, db)
	api.Handle(
		mgmtv1alpha1connect.NewConnectionServiceHandler(
			connectionService,
			stdInterceptors,
		),
	)

	jobService := v1alpha1_jobservice.New(&v1alpha1_jobservice.Config{TemporalTaskQueue: temporalTaskQueue}, db, temporalClient, connectionService)
	api.Handle(
		mgmtv1alpha1connect.NewJobServiceHandler(
			jobService,
			stdInterceptors,
		),
	)

	mux.Handle("/", api)

	addr := fmt.Sprintf("%s:%d", host, port)

	logger.Info(fmt.Sprintf("listening on %s", addr))
	httpServer := http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ErrorLog:          loglogger,
		ReadHeaderTimeout: 10 * time.Second,
	}

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

func getTemporalConfig() *temporalclient.Options {
	port := viper.GetInt32("PORT")
	if port == 0 {
		port = 7233
	}
	host := viper.GetString("HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	temporalNamespace := viper.GetString("TEMPORAL_NAMESPACE")
	if temporalNamespace == "" {
		temporalNamespace = "default"
	}

	address := fmt.Sprintf("%s:%d", host, port)

	return &temporalclient.Options{
		// Logger: ,
		HostPort:  address,
		Namespace: temporalNamespace,
		// Interceptors: ,
		// HeadersProvider: ,
	}, nil
}

func getTemporalTaskQueue() (string, error) {
	taskQueue := viper.GetString("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		return "", errors.New("must provide TEMPORAL_TASK_QUEUE environment variable")
	}
	return taskQueue, nil
}
