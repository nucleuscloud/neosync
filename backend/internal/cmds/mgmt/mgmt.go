package mgmt_cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	serve "github.com/nucleuscloud/neosync/backend/internal/cmds/mgmt/serve"
)

func Execute() {
	// logger, err := l.NewLogger(true)
	// if err != nil {
	// 	panic(err)
	// }
	cobra.OnInitialize(func() { initConfig() })

	rootCmd := &cobra.Command{
		Use:   "mgmt",
		Short: "Terminal UI that interfaces with the Nucleus system.",
		Long:  "Terminal UI that allows authenticated access to the Nucleus system.\nThis CLI allows you to deploy and manage all of the environments and services within your Nucleus account or accounts.",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cmd.SilenceErrors = true
		},
	}

	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	// Wire up subcommands here
	rootCmd.AddCommand(serve.NewCmd())

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
			// logger.Info(".env file not found, skipping")
		} else {
			panic(err)
		}
	}
	envType := viper.GetString("NUCLEUS_ENV")
	if envType != "" {
		viper.SetConfigName(fmt.Sprintf(".env.%s", envType))
		if err := viper.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// logger.Info(fmt.Sprintf(".env.%s file not found, skipping", envType))
			} else {
				panic(err)
			}
		}

		viper.SetConfigName(fmt.Sprintf(".env.%s.secrets", envType))
		if err := viper.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// logger.Info(fmt.Sprintf(".env.%s.secrets file not found, skipping", envType))
			} else {
				panic(err)
			}
		}
	}
}
