package serve_connect

import (
	"errors"
	"log/slog"
	"os"

	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
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

	temporalClient, err := client.Dial(client.Options{
		Logger:    slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})),
		HostPort:  temporalUrl,
		Namespace: temporalNamespace,
		// Interceptors: ,
		// HeadersProvider: , // todo: set auth headers
	})
	if err != nil {
		return err
	}
	defer temporalClient.Close()

	w := worker.New(temporalClient, taskQueue, worker.Options{})
	_ = w

	w.RegisterWorkflow(datasync.Workflow)
	w.RegisterActivity(&datasync.Activities{})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		return err
	}
	return nil
}
