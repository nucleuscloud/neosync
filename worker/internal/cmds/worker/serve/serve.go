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
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
	benthosstream "github.com/nucleuscloud/neosync/internal/benthos-stream"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/mongoprovider"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/sqlprovider"
	"github.com/nucleuscloud/neosync/internal/connectiondata"
	retry_interceptor "github.com/nucleuscloud/neosync/internal/connectrpc/interceptors/retry"
	cloudlicense "github.com/nucleuscloud/neosync/internal/ee/cloud-license"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	neosync_gcp "github.com/nucleuscloud/neosync/internal/gcp"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	neosyncotel "github.com/nucleuscloud/neosync/internal/otel"
	pyroscope_env "github.com/nucleuscloud/neosync/internal/pyroscope"
	neosync_redis "github.com/nucleuscloud/neosync/internal/redis"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	schemainit_workflow_register "github.com/nucleuscloud/neosync/worker/pkg/workflows/schemainit/workflow/register"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	datasync_workflow_register "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/register"
	accounthook_workflow_register "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/workflow/register"
	piidetect_workflow_register "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/register"
	tablesync_workflow_register "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/workflow/register"
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

	"github.com/grafana/pyroscope-go"
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
	slog.SetDefault(
		logger,
	) // set default logger for methods that can't easily access the configured logger

	eelicense, err := license.NewFromEnv()
	if err != nil {
		return fmt.Errorf("unable to initialize ee license from env: %w", err)
	}
	logger.Debug(fmt.Sprintf("ee license enabled: %t", eelicense.IsValid()))

	ncloudlicense, err := cloudlicense.NewFromEnv()
	if err != nil {
		return fmt.Errorf("unable to initialize neosync cloud license from env: %w", err)
	}
	logger.Debug(fmt.Sprintf("neosync cloud enabled: %t", ncloudlicense.IsValid()))

	pyroscopeConfig, isPyroscopeEnabled, err := pyroscope_env.NewFromEnv("neosync-worker", logger)
	if err != nil {
		return fmt.Errorf("unable to initialize pyroscope from env: %w", err)
	}
	if isPyroscopeEnabled {
		profiler, err := pyroscope.Start(*pyroscopeConfig)
		if err != nil {
			return fmt.Errorf("unable to start pyroscope profiler: %w", err)
		}
		defer profiler.Stop() //nolint:errcheck
	}

	var syncActivityMeter metric.Meter
	temporalClientInterceptors := []interceptor.ClientInterceptor{}
	var temopralMeterHandler client.MetricsHandler

	connectInterceptors := []connect.Interceptor{}

	otelconfig := neosyncotel.GetOtelConfigFromViperEnv()
	if otelconfig.IsEnabled {
		logger.Debug("otel is enabled")
		tmPropagator := neosyncotel.NewDefaultPropagator()
		otelconnopts := []otelconnect.Option{
			otelconnect.WithoutServerPeerAttributes(),
			otelconnect.WithPropagator(tmPropagator),
		}

		meterProviders := []neosyncotel.MeterProvider{}
		traceProviders := []neosyncotel.TracerProvider{}
		// Meter Provider that uses delta temporality for use with Benthos metrics
		// This meter provider is setup expire metrics after a specified time period for easy computation
		benthosMeterProvider, err := neosyncotel.NewMeterProvider(
			ctx,
			&neosyncotel.MeterProviderConfig{
				Exporter:   otelconfig.MeterExporter,
				AppVersion: otelconfig.ServiceVersion,
				Opts: neosyncotel.MeterExporterOpts{
					Otlp: []otlpmetricgrpc.Option{
						neosyncotel.GetBenthosMetricTemporalityOption(),
					},
					Console: []stdoutmetric.Option{stdoutmetric.WithPrettyPrint()},
				},
			},
		)
		if err != nil {
			return err
		}
		if benthosMeterProvider != nil {
			logger.Debug("otel metering for benthos has been configured")
			meterProviders = append(meterProviders, benthosMeterProvider)
			syncActivityMeter = benthosMeterProvider.Meter("sync_activity")
		}

		temporalMeterProvider, err := neosyncotel.NewMeterProvider(
			ctx,
			&neosyncotel.MeterProviderConfig{
				Exporter:   otelconfig.MeterExporter,
				AppVersion: otelconfig.ServiceVersion,
				Opts: neosyncotel.MeterExporterOpts{
					Otlp:    []otlpmetricgrpc.Option{},
					Console: []stdoutmetric.Option{stdoutmetric.WithPrettyPrint()},
				},
			},
		)
		if err != nil {
			return err
		}
		if temporalMeterProvider != nil {
			logger.Debug("otel metering for temporal has been configured")
			meterProviders = append(meterProviders, temporalMeterProvider)
			temopralMeterHandler = temporalotel.NewMetricsHandler(
				temporalotel.MetricsHandlerOptions{
					Meter: temporalMeterProvider.Meter("neosync-temporal-sdk"),
					OnError: func(err error) {
						logger.Error(fmt.Errorf("error with temporal metering: %w", err).Error())
					},
				},
			)
		}

		neosyncMeterProvider, err := neosyncotel.NewMeterProvider(
			ctx,
			&neosyncotel.MeterProviderConfig{
				Exporter:   otelconfig.MeterExporter,
				AppVersion: otelconfig.ServiceVersion,
				Opts: neosyncotel.MeterExporterOpts{
					Otlp:    []otlpmetricgrpc.Option{},
					Console: []stdoutmetric.Option{stdoutmetric.WithPrettyPrint()},
				},
			},
		)
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

		temporalTraceProvider, err := neosyncotel.NewTraceProvider(
			ctx,
			&neosyncotel.TraceProviderConfig{
				Exporter: otelconfig.TraceExporter,
				Opts: neosyncotel.TraceExporterOpts{
					Otlp:    []otlptracegrpc.Option{},
					Console: []stdouttrace.Option{stdouttrace.WithPrettyPrint()},
				},
			},
		)
		if err != nil {
			return err
		}
		if temporalTraceProvider != nil {
			logger.Debug("otel tracing for temporal has been configured")
			temporalTraceInterceptor, err := temporalotel.NewTracingInterceptor(
				temporalotel.TracerOptions{
					Tracer: temporalTraceProvider.Tracer("neosync-temporal-sdk"),
				},
			)
			if err != nil {
				return err
			}
			temporalClientInterceptors = append(
				temporalClientInterceptors,
				temporalTraceInterceptor,
			)
			traceProviders = append(traceProviders, temporalTraceProvider)
		}

		neosyncTraceProvider, err := neosyncotel.NewTraceProvider(
			ctx,
			&neosyncotel.TraceProviderConfig{
				Exporter: otelconfig.TraceExporter,
				Opts: neosyncotel.TraceExporterOpts{
					Otlp:    []otlptracegrpc.Option{},
					Console: []stdouttrace.Option{stdouttrace.WithPrettyPrint()},
				},
			},
		)
		if err != nil {
			return err
		}
		if neosyncTraceProvider != nil {
			logger.Debug("otel tracing for neosync clients has been configured")
			otelconnopts = append(
				otelconnopts,
				otelconnect.WithTracerProvider(neosyncTraceProvider),
			)
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
				logger.Error(
					fmt.Errorf("unable to gracefully shutdown otel providers: %w", err).Error(),
				)
			}
		}()
	}

	// Ensure that the retry interceptor comes after the otel interceptor
	connectInterceptors = append(
		connectInterceptors,
		retry_interceptor.DefaultRetryInterceptor(logger),
	)

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
		return fmt.Errorf("unable to get temporal auth certificate: %w", err)
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
	logger.Debug("temporal client dialed successfully")
	defer temporalClient.Close()

	w := worker.New(temporalClient, taskQueue, worker.Options{})
	_ = w

	cascadelicense := license.NewCascadeLicense(
		ncloudlicense,
		eelicense,
	)

	neosyncurl := shared.GetNeosyncUrl()
	httpclient := shared.GetNeosyncHttpClient()
	connectInterceptorOption := connect.WithInterceptors(connectInterceptors...)
	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(
		httpclient,
		neosyncurl,
		connectInterceptorOption,
	)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		httpclient,
		neosyncurl,
		connectInterceptorOption,
	)
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		httpclient,
		neosyncurl,
		connectInterceptorOption,
	)
	transformerclient := mgmtv1alpha1connect.NewTransformersServiceClient(
		httpclient,
		neosyncurl,
		connectInterceptorOption,
	)
	accounthookclient := mgmtv1alpha1connect.NewAccountHookServiceClient(
		httpclient,
		neosyncurl,
		connectInterceptorOption,
	)
	anonymizationclient := mgmtv1alpha1connect.NewAnonymizationServiceClient(
		httpclient,
		neosyncurl,
		connectInterceptorOption,
	)

	sqlConnector := &sqlconnect.SqlOpenConnector{}
	sqlconnmanager := connectionmanager.NewConnectionManager(sqlprovider.NewProvider(sqlConnector))
	go sqlconnmanager.Reaper(logger)
	defer sqlconnmanager.Shutdown(logger)

	mongoconnmanager := connectionmanager.NewConnectionManager(mongoprovider.NewProvider())
	go mongoconnmanager.Reaper(logger)
	defer mongoconnmanager.Shutdown(logger)

	sqlmanager := sql_manager.NewSqlManager(sql_manager.WithConnectionManager(sqlconnmanager))

	redisconfig := shared.GetRedisConfig()
	redisclient, err := neosync_redis.GetRedisClient(redisconfig)
	if err != nil {
		return fmt.Errorf("unable to get redis client: %w", err)
	}

	maxIterations := 100
	pageLimit := 100_000
	streamManager := benthosstream.NewBenthosStreamManager()
	tablesync_workflow_register.Register(
		w,
		connclient,
		jobclient,
		sqlconnmanager,
		mongoconnmanager,
		syncActivityMeter,
		streamManager,
		temporalClient,
		maxIterations,
		anonymizationclient,
		redisclient,
	)

	schemainit_workflow_register.Register(
		w,
		jobclient,
		connclient,
		sqlmanager,
		cascadelicense,
	)

	postgresSchemaDrift := false
	datasync_workflow_register.Register(
		w,
		userclient, jobclient, connclient, transformerclient,
		sqlmanager, redisconfig, cascadelicense, redisclient,
		otelconfig.IsEnabled,
		pageLimit,
		postgresSchemaDrift,
	)

	if cascadelicense.IsValid() {
		logger.Debug("ee license is valid, registering account hook activities")
		accounthook_workflow_register.Register(w, accounthookclient)

		openaiclient := openai.NewClient(option.WithAPIKey(viper.GetString("OPENAI_API_KEY")))

		neosynctyperegistry := neosynctypes.NewTypeRegistry(logger)
		conndatabuilder := connectiondata.NewConnectionDataBuilder(
			sqlConnector,
			sqlmanager,
			pg_queries.New(),
			mysql_queries.New(),
			awsmanager.New(),
			neosync_gcp.NewManager(),
			mongoconnect.NewConnector(),
			neosynctyperegistry,
		)

		piidetect_workflow_register.Register(
			w,
			connclient,
			jobclient,
			openaiclient,
			conndatabuilder,
			cascadelicense,
			temporalClient.ScheduleClient(),
		)
	}

	if err := w.Start(); err != nil {
		return fmt.Errorf("unable to start temporal worker: %w", err)
	}
	logger.Debug("temporal worker started successfully")

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
