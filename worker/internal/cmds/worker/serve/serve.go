package serve_connect

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	datasync_activities "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities"
	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	datasync_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "serves up the worker",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return serve()
		},
	}
}

func serve() error {
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
	jsonloghandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	logger := slog.New(jsonloghandler)
	loglogger := slog.NewLogLogger(jsonloghandler, slog.LevelInfo)

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
		return err
	}
	defer temporalClient.Close()

	w := worker.New(temporalClient, taskQueue, worker.Options{})
	_ = w

	w.RegisterWorkflow(datasync_workflow.Workflow)
	w.RegisterActivity(sync_activity.Sync)
	w.RegisterActivity(syncactivityopts_activity.RetrieveActivityOptions)
	w.RegisterActivity(runsqlinittablestmts_activity.RunSqlInitTableStatements)
	w.RegisterActivity(syncrediscleanup_activity.DeleteRedisHash)
	w.RegisterActivity(&datasync_activities.Activities{})

	if err := w.Start(); err != nil {
		return err
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

	w.Stop()

	ctx, cancelHandler := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
	defer cancelHandler()
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}
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
	mux.Handle("/", api)

	httpServer := http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ErrorLog:          logger,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return &httpServer
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

type RedisConfig struct {
	Url    string
	Kind   *string
	Master *string
	Tls *RedisTlsConfig
}

type RedisTlsConfig struct {
	Enabled               bool
	skipCertVerify        bool
	enableRenegotiation   bool
	rootCertAuthority     *string
	rootCertAuthorityFile *string
}

// redis://<user>:<password>@<host>:<port>/<db_number>
func getRedisUrl() *RedisConfig {
	redisUrl := viper.GetString("REDIS_URL")
	if redisUrl == "" {
		return nil
	}

	kind := viper.GetString("REDIS_KIND")
	master := viper.GetString("REDIS_MASTER")
	return &RedisConfig{
		Url:    redisUrl,
		Kind:   &kind,
		Master: &master,
		Tls: &RedisTlsConfig{
			Enable: 
		},
	}

}
