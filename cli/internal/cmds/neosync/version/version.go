package version_cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nucleuscloud/neosync/cli/internal/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the client version information",
		Long:  "Print the client versio ninformation for the current context",
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}
			if output != "" && output != "json" && output != "yaml" {
				return fmt.Errorf("must provide valid output")
			}
			versionInfo := version.Get()

			switch output {
			case "json":
				marshaled, err := json.MarshalIndent(&versionInfo, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(marshaled)) //nolint:forbidigo
			case "yaml":
				marshaled, err := yaml.Marshal(&versionInfo)
				if err != nil {
					return err
				}
				fmt.Println(string(marshaled)) //nolint:forbidigo
			default:
				fmt.Println("Git Version:", versionInfo.GitVersion) //nolint:forbidigo
				fmt.Println("Git Commit:", versionInfo.GitCommit)   //nolint:forbidigo
				fmt.Println("Build Date:", versionInfo.BuildDate)   //nolint:forbidigo
				fmt.Println("Go Version:", versionInfo.GoVersion)   //nolint:forbidigo
				fmt.Println("Compiler:", versionInfo.Compiler)      //nolint:forbidigo
				fmt.Println("Platform:", versionInfo.Platform)      //nolint:forbidigo
			}
			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "json|yaml")

	return cmd
}
