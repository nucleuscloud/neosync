package worker_cmd

import (
	"fmt"

	serve_connect "github.com/nucleuscloud/neosync/worker/internal/cmds/worker/serve"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() {
	cobra.OnInitialize(func() { initConfig() })

	rootCmd := &cobra.Command{
		Use:   "worker",
		Short: "Terminal UI that interfaces with the Nucleus system.",
		Long:  "",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cmd.SilenceErrors = true
		},
	}

	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	// Wire up subcommands here
	rootCmd.AddCommand(serve_connect.NewCmd())

	cobra.CheckErr(rootCmd.Execute())
}

// initConfig reads in config file and ENV variables if set.
func initConfig(
// logger logr.Logger,
) {
	viper.AddConfigPath(".")
	viper.SetConfigType("dotenv")

	viper.AutomaticEnv() // read in environment variables that match

	viper.SetConfigName(".env")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		} else {
			panic(err)
		}
	}
	envType := viper.GetString("NUCLEUS_ENV")
	if envType != "" {
		viper.SetConfigName(fmt.Sprintf(".env.%s", envType))
		if err := viper.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			} else {
				panic(err)
			}
		}

		viper.SetConfigName(fmt.Sprintf(".env.%s.secrets", envType))
		if err := viper.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			} else {
				panic(err)
			}
		}
	}
}
