package neosync_cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	jobs_cmd "github.com/nucleuscloud/neosync/cli/internal/cmds/neosync/jobs"
	login_cmd "github.com/nucleuscloud/neosync/cli/internal/cmds/neosync/login"
	version_cmd "github.com/nucleuscloud/neosync/cli/internal/cmds/neosync/version"
	whoami_cmd "github.com/nucleuscloud/neosync/cli/internal/cmds/neosync/whoami"
	"github.com/nucleuscloud/neosync/cli/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc/metadata"
)

const (
	neosyncDirName           = ".neosync"
	cliSettingsFileNameNoExt = ".neosync-cli"
	cliSettingsFileExt       = "yaml"
)

func Execute() {
	var cfgFile string
	cobra.OnInitialize(func() { initConfig(cfgFile) })

	rootCmd := &cobra.Command{
		Use:   "neosync",
		Short: "Terminal UI that interfaces with the Neosync system.",
		Long:  "",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cmd.SilenceErrors = true

			versionInfo := version.Get()
			md := metadata.New(map[string]string{
				"cliVersion":  versionInfo.GitVersion,
				"cliPlatform": versionInfo.Platform,
				"cliCommit":   versionInfo.GitCommit,
			})
			cmd.SetContext(metadata.NewOutgoingContext(cmd.Context(), md))
		},
	}

	rootCmd.Version = version.Get().GitVersion
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/%s.%s)", cliSettingsFileNameNoExt, cliSettingsFileExt),
	)
	rootCmd.AddCommand(jobs_cmd.NewCmd())
	rootCmd.AddCommand(version_cmd.NewCmd())
	rootCmd.AddCommand(whoami_cmd.NewCmd())
	rootCmd.AddCommand(login_cmd.NewCmd())

	cobra.CheckErr(rootCmd.Execute())
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cfgFilePath string) {
	if cfgFilePath != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFilePath)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		fullNeosyncSettingsDir := filepath.Join(home, neosyncDirName)
		neosyncConfigDir := os.Getenv("NEOSYNC_CONFIG_DIR") // helpful for tools such as direnv and people who want it somewhere interesting
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")       // linux users expect this to be respected

		viper.AddConfigPath(".")
		viper.AddConfigPath(fullNeosyncSettingsDir)
		viper.AddConfigPath(home)
		viper.AddConfigPath(neosyncConfigDir)
		viper.AddConfigPath(xdgConfigHome)

		viper.SetConfigType(cliSettingsFileExt)
		viper.SetConfigName(cliSettingsFileNameNoExt)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
			return
		}
	}
}
