package serve_connect

import (
	"errors"
	"fmt"

	"github.com/nucleuscloud/neosync/worker/internal/workflows/datasync"
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

	taskQueue := viper.GetString("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		return errors.New("must provide TEMPORAL_TASK_QUEUE environment variable")
	}

	address := fmt.Sprintf("%s:%d", host, port)

	temporalClient, err := client.Dial(client.Options{
		// Logger: ,
		HostPort:  address,
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
