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

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/otelconnect"
	"github.com/go-logr/logr"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosynclogger "github.com/nucleuscloud/neosync/backend/pkg/logger"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	neosyncotel "github.com/nucleuscloud/neosync/internal/otel"
	accountstatus_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/account-status"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	jobhooks_by_timing_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/jobhooks-by-timing"
	posttablesync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/post-table-sync"
	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	datasync_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/client"
	temporalotel "go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
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

	eelicense, err := license.NewFromEnv()
	if err != nil {
		return fmt.Errorf("unable to initialize ee license from env: %w", err)
	}
	logger.Debug(fmt.Sprintf("ee license enabled: %t", eelicense.IsValid()))

	isNeosyncCloud := getIsNeosyncCloud()
	logger.Debug(fmt.Sprintf("neosync cloud enabled: %t", isNeosyncCloud))

	var syncActivityMeter metric.Meter
	temporalClientInterceptors := []interceptor.ClientInterceptor{}
	var temopralMeterHandler client.MetricsHandler

	connectInterceptors := []connect.Interceptor{}

	otelconfig := neosyncotel.GetOtelConfigFromViperEnv()
	if otelconfig.IsEnabled {
		logger.Debug("otel is enabled")
		tmPropagator := neosyncotel.NewDefaultPropagator()
		otelconnopts := []otelconnect.Option{otelconnect.WithoutServerPeerAttributes(), otelconnect.WithPropagator(tmPropagator)}

		meterProviders := []neosyncotel.MeterProvider{}
		traceProviders := []neosyncotel.TracerProvider{}
		// Meter Provider that uses delta temporality for use with Benthos metrics
		// This meter provider is setup expire metrics after a specified time period for easy computation
		benthosMeterProvider, err := neosyncotel.NewMeterProvider(ctx, &neosyncotel.MeterProviderConfig{
			Exporter:   otelconfig.MeterExporter,
			AppVersion: otelconfig.ServiceVersion,
			Opts: neosyncotel.MeterExporterOpts{
				Otlp:    []otlpmetricgrpc.Option{neosyncotel.GetBenthosMetricTemporalityOption()},
				Console: []stdoutmetric.Option{stdoutmetric.WithPrettyPrint()},
			},
		})
		if err != nil {
			return err
		}
		if benthosMeterProvider != nil {
			logger.Debug("otel metering for benthos has been configured")
			meterProviders = append(meterProviders, benthosMeterProvider)
			syncActivityMeter = benthosMeterProvider.Meter("sync_activity")
		}

		temporalMeterProvider, err := neosyncotel.NewMeterProvider(ctx, &neosyncotel.MeterProviderConfig{
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
		if temporalMeterProvider != nil {
			logger.Debug("otel metering for temporal has been configured")
			meterProviders = append(meterProviders, temporalMeterProvider)
			temopralMeterHandler = temporalotel.NewMetricsHandler(temporalotel.MetricsHandlerOptions{
				Meter: temporalMeterProvider.Meter("neosync-temporal-sdk"),
				OnError: func(err error) {
					logger.Error(fmt.Errorf("error with temporal metering: %w", err).Error())
				},
			})
		}

		neosyncMeterProvider, err := neosyncotel.NewMeterProvider(ctx, &neosyncotel.MeterProviderConfig{
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
		if neosyncMeterProvider != nil {
			logger.Debug("otel metering for neosync clients has been configured")
			meterProviders = append(meterProviders, neosyncMeterProvider)
			otelconnopts = append(otelconnopts, otelconnect.WithMeterProvider(neosyncMeterProvider))
		} else {
			otelconnopts = append(otelconnopts, otelconnect.WithoutMetrics())
		}

		temporalTraceProvider, err := neosyncotel.NewTraceProvider(ctx, &neosyncotel.TraceProviderConfig{
			Exporter: otelconfig.TraceExporter,
			Opts: neosyncotel.TraceExporterOpts{
				Otlp:    []otlptracegrpc.Option{},
				Console: []stdouttrace.Option{stdouttrace.WithPrettyPrint()},
			},
		})
		if err != nil {
			return err
		}
		if temporalTraceProvider != nil {
			logger.Debug("otel tracing for temporal has been configured")
			temporalTraceInterceptor, err := temporalotel.NewTracingInterceptor(temporalotel.TracerOptions{
				Tracer: temporalTraceProvider.Tracer("neosync-temporal-sdk"),
			})
			if err != nil {
				return err
			}
			temporalClientInterceptors = append(temporalClientInterceptors, temporalTraceInterceptor)
			traceProviders = append(traceProviders, temporalTraceProvider)
		}

		neosyncTraceProvider, err := neosyncotel.NewTraceProvider(ctx, &neosyncotel.TraceProviderConfig{
			Exporter: otelconfig.TraceExporter,
			Opts: neosyncotel.TraceExporterOpts{
				Otlp:    []otlptracegrpc.Option{},
				Console: []stdouttrace.Option{stdouttrace.WithPrettyPrint()},
			},
		})
		if err != nil {
			return err
		}
		if neosyncTraceProvider != nil {
			logger.Debug("otel tracing for neosync clients has been configured")
			otelconnopts = append(otelconnopts, otelconnect.WithTracerProvider(neosyncTraceProvider))
		} else {
			otelconnopts = append(otelconnopts, otelconnect.WithoutTracing(), otelconnect.WithoutTraceEvents())
		}

		otelConnectInterceptor, err := otelconnect.NewInterceptor(otelconnopts...)
		if err != nil {
			return err
		}
		connectInterceptors = append(connectInterceptors, otelConnectInterceptor)

		otelshutdown := neosyncotel.SetupOtelSdk(&neosyncotel.SetupConfig{
			TraceProviders:    traceProviders,
			MeterProviders:    meterProviders,
			Logger:            logr.FromSlogHandler(logger.Handler()),
			TextMapPropagator: tmPropagator,
		})
		defer func() {
			if err := otelshutdown(context.Background()); err != nil {
				logger.Error(fmt.Errorf("unable to gracefully shutdown otel providers: %w", err).Error())
			}
		}()
	}

	temporalUrl := viper.GetString("TEMPORAL_URL")
	if temporalUrl == "" {
		temporalUrl = client.DefaultHostPort
	}

	temporalNamespace := viper.GetString("TEMPORAL_NAMESPACE")
	if temporalNamespace == "" {
		temporalNamespace = client.DefaultNamespace
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
		logger.Debug("temporal TLS certificates have been attached")
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
		MetricsHandler: temopralMeterHandler,
		Interceptors:   temporalClientInterceptors,
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
	connectInterceptorOption := connect.WithInterceptors(connectInterceptors...)
	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(httpclient, neosyncurl, connectInterceptorOption)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(httpclient, neosyncurl, connectInterceptorOption)
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(httpclient, neosyncurl, connectInterceptorOption)
	transformerclient := mgmtv1alpha1connect.NewTransformersServiceClient(httpclient, neosyncurl, connectInterceptorOption)
	sqlconnector := &sqlconnect.SqlOpenConnector{}
	sqlmanager := sql_manager.NewSqlManager(pgpoolmap, pgquerier, mysqlpoolmap, mysqlquerier, mssqlpoolmap, mssqlquerier, sqlconnector)
	redisconfig := shared.GetRedisConfig()

	genbenthosActivity := genbenthosconfigs_activity.New(
		jobclient,
		connclient,
		transformerclient,
		sqlmanager,
		redisconfig,
		otelconfig.IsEnabled,
	)
	disableReaper := false
	syncActivity := sync_activity.New(connclient, jobclient, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, temporalClient, syncActivityMeter, sync_activity.NewBenthosStreamManager(), disableReaper)
	retrieveActivityOpts := syncactivityopts_activity.New(jobclient)
	runSqlInitTableStatements := runsqlinittablestmts_activity.New(jobclient, connclient, sqlmanager, eelicense, isNeosyncCloud)
	accountStatusActivity := accountstatus_activity.New(userclient)
	runPostTableSyncActivity := posttablesync_activity.New(jobclient, sqlmanager, connclient)
	jobhookByTimingActivity := jobhooks_by_timing_activity.New(jobclient, connclient, sqlmanager)

	w.RegisterWorkflow(datasync_workflow.Workflow)
	w.RegisterActivity(syncActivity.Sync)
	w.RegisterActivity(retrieveActivityOpts.RetrieveActivityOptions)
	w.RegisterActivity(runSqlInitTableStatements.RunSqlInitTableStatements)
	w.RegisterActivity(syncrediscleanup_activity.DeleteRedisHash)
	w.RegisterActivity(genbenthosActivity.GenerateBenthosConfigs)
	w.RegisterActivity(accountStatusActivity.CheckAccountStatus)
	w.RegisterActivity(runPostTableSyncActivity.RunPostTableSync)
	w.RegisterActivity(jobhookByTimingActivity.RunJobHooksByTiming)

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

func getIsNeosyncCloud() bool {
	return viper.GetBool("NEOSYNC_CLOUD")
}
