package serve_connect

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"github.com/go-logr/logr"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosynclogger "github.com/nucleuscloud/neosync/backend/pkg/logger"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	datasync_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"net/http/pprof"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "serves up the worker",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return serve(cmd.Context())
		},
	}
}

func serve(ctx context.Context) error {
	logger, loglogger := neosynclogger.NewLoggers()
	slog.SetDefault(logger) // set default logger for methods that can't easily access the configured logger

	var activityMeter metric.Meter
	if getIsOtelEnabled() {
		otel.SetLogger(logr.FromSlogHandler(logger.Handler()))
		metricProvider, ok, err := getConfiguredMeterProvider(ctx)
		if err != nil {
			return err
		}
		otelConfig := &otelSetupConfig{}
		if ok {
			otelConfig.MeterProvider = metricProvider
			activityMeter = metricProvider.Meter("sync_activity")
		}
		otelShutdown := setupOtelSdk(otelConfig)
		defer func() {
			if err := otelShutdown(context.Background()); err != nil {
				logger.Error(err.Error())
			}
		}()
	}

	temporalUrl := viper.GetString("TEMPORAL_URL")
	if temporalUrl == "" {
		temporalUrl = "localhost:7233"
	}

	temporalNamespace := viper.GetString("TEMPORAL_NAMESPACE")
	if temporalNamespace == "" {
		temporalNamespace = "default"
	}

	taskQueue := viper.GetString("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		return errors.New("must provide TEMPORAL_TASK_QUEUE environment variable")
	}

	certificates, err := getTemporalAuthCertificate()
	if err != nil {
		return err
	}

	var tlsConfig *tls.Config
	if len(certificates) > 0 {
		tlsConfig = &tls.Config{
			Certificates: certificates,
			MinVersion:   tls.VersionTLS13,
		}
	}

	temporalClient, err := client.Dial(client.Options{
		Logger:    logger,
		HostPort:  temporalUrl,
		Namespace: temporalNamespace,
		ConnectionOptions: client.ConnectionOptions{
			TLS: tlsConfig,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to dial temporal client: %w", err)
	}
	defer temporalClient.Close()

	w := worker.New(temporalClient, taskQueue, worker.Options{})
	_ = w

	pgpoolmap := &sync.Map{}
	mysqlpoolmap := &sync.Map{}
	mssqlpoolmap := &sync.Map{}
	pgquerier := pg_queries.New()
	mysqlquerier := mysql_queries.New()
	mssqlquerier := mssql_queries.New()

	neosyncurl := shared.GetNeosyncUrl()
	httpclient := shared.GetNeosyncHttpClient()
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(httpclient, neosyncurl)
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(httpclient, neosyncurl)
	transformerclient := mgmtv1alpha1connect.NewTransformersServiceClient(httpclient, neosyncurl)
	sqlconnector := &sqlconnect.SqlOpenConnector{}
	sqlmanager := sql_manager.NewSqlManager(pgpoolmap, pgquerier, mysqlpoolmap, mysqlquerier, mssqlpoolmap, mssqlquerier, sqlconnector)
	redisconfig := shared.GetRedisConfig()

	genbenthosActivity := genbenthosconfigs_activity.New(
		jobclient,
		connclient,
		transformerclient,
		sqlmanager,
		redisconfig,
		getIsOtelEnabled(),
	)
	disableReaper := false
	syncActivity := sync_activity.New(connclient, &sync.Map{}, temporalClient, activityMeter, sync_activity.NewBenthosStreamManager(), disableReaper)
	retrieveActivityOpts := syncactivityopts_activity.New(jobclient)
	runSqlInitTableStatements := runsqlinittablestmts_activity.New(jobclient, connclient, sqlmanager)

	w.RegisterWorkflow(datasync_workflow.Workflow)
	w.RegisterActivity(syncActivity.Sync)
	w.RegisterActivity(retrieveActivityOpts.RetrieveActivityOptions)
	w.RegisterActivity(runSqlInitTableStatements.RunSqlInitTableStatements)
	w.RegisterActivity(syncrediscleanup_activity.DeleteRedisHash)
	w.RegisterActivity(genbenthosActivity.GenerateBenthosConfigs)

	if err := w.Start(); err != nil {
		return fmt.Errorf("unable to start temporal worker: %w", err)
	}

	httpServer := getHttpServer(loglogger)

	go func() {
		logger.Info(fmt.Sprintf("listening on %s", httpServer.Addr))
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error(err.Error())
		}
	}()

	<-worker.InterruptCh()
	logger.Info("received interrupt, stopping worker...")
	w.Stop()
	logger.Info("temporal worker shut down, proceeding to shutting down http server")
	ctx, cancelHandler := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
	defer cancelHandler()
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}
	logger.Info("worker stopped successfully, fully shutting down")
	return nil
}

func getHttpServer(logger *log.Logger) *http.Server {
	port := viper.GetInt32("PORT")
	if port == 0 {
		port = 8080
	}
	host := viper.GetString("HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	mux := http.NewServeMux()
	mux.Handle(grpchealth.NewHandler(grpchealth.NewStaticChecker()))

	reflector := grpcreflect.NewStaticReflector()
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	api := http.NewServeMux()

	// Create a separate ServeMux for pprof

	// Mount the pprof mux at /debug/
	if viper.GetBool("ENABLE_PPROF") {
		mux.Handle("/debug/", getPprofMux("/debug"))
	}

	mux.Handle("/", api)

	httpServer := http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ErrorLog:          logger,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return &httpServer
}

func getPprofMux(prefix string) *http.ServeMux {
	mux := http.NewServeMux()

	// Ensure the prefix starts with a slash and doesn't end with one
	prefix = "/" + strings.Trim(prefix, "/")

	mux.HandleFunc(prefix+"/pprof/", http.HandlerFunc(pprof.Index))
	mux.HandleFunc(prefix+"/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc(prefix+"/pprof/profile", pprof.Profile)
	mux.HandleFunc(prefix+"/pprof/symbol", pprof.Symbol)
	mux.HandleFunc(prefix+"/pprof/trace", pprof.Trace)
	mux.Handle(prefix+"/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle(prefix+"/pprof/heap", pprof.Handler("heap"))
	mux.Handle(prefix+"/pprof/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle(prefix+"/pprof/block", pprof.Handler("block"))

	return mux
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

type otelSetupConfig struct {
	TraceProvider *trace.TracerProvider
	MeterProvider *metricsdk.MeterProvider
}

func setupOtelSdk(config *otelSetupConfig) func(context.Context) error {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	if config.TraceProvider != nil {
		shutdownFuncs = append(shutdownFuncs, config.TraceProvider.Shutdown)
		otel.SetTracerProvider(config.TraceProvider)
	}

	// Set up meter provider
	if config.MeterProvider != nil {
		shutdownFuncs = append(shutdownFuncs, config.MeterProvider.Shutdown)
		otel.SetMeterProvider(config.MeterProvider) // maybe dont set this as the global since it might be specific to a discrete part of the application
	}
	return shutdown
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func temporalitySelector(ik metricsdk.InstrumentKind) metricdata.Temporality {
	// Delta Temporality causes metrics to be reset after some time.
	// We are using this today for benthos metrics so that they don't persist indefinitely in the time series database
	return metricdata.DeltaTemporality
}

func getConfiguredMeterProvider(ctx context.Context) (*metricsdk.MeterProvider, bool, error) {
	if !getIsOtelEnabled() {
		return nil, false, nil
	}
	// todo: may want to conditionally allow http, prometheus metering based on env vars
	var exporter metricsdk.Exporter
	exporterType := getMetricsExporter()
	if exporterType == "otlp" {
		grpcExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithTemporalitySelector(temporalitySelector))
		if err != nil {
			return nil, false, err
		}
		exporter = grpcExporter
	} else {
		return nil, false, fmt.Errorf("that exporter type is not currently supported")
	}

	reader := metricsdk.WithReader(
		metricsdk.NewPeriodicReader(
			exporter,
		),
	)
	attrs := []attribute.KeyValue{
		semconv.ServiceVersion(getAppVersion()),
	}
	res := resource.NewWithAttributes(semconv.SchemaURL, attrs...)
	provider := metricsdk.NewMeterProvider(reader, metricsdk.WithResource(res))
	return provider, true, nil
}

func getIsOtelEnabled() bool {
	isDisabledStr := viper.GetString("OTEL_SDK_DISABLED")
	if isDisabledStr == "" {
		return false
	}
	return !viper.GetBool("OTEL_SDK_DISABLED")
}

func getAppVersion() string {
	return viper.GetString("OTEL_SERVICE_VERSION")
}

func getMetricsExporter() string {
	exporter := viper.GetString("OTEL_METRICS_EXPORTER")
	if exporter == "" {
		return "otlp"
	}
	return exporter
}
