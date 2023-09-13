package serve_connect

import (
	"fmt"

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

	w := worker.New(temporalClient, "", worker.Options{})
	_ = w

	// todo: register workflows and activites

	err = w.Run(worker.InterruptCh())
	if err != nil {
		return err
	}
	return nil
}
