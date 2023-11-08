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

			if output == "json" {
				marshaled, err := json.MarshalIndent(&versionInfo, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(marshaled)) //nolint
			} else if output == "yaml" {
				marshaled, err := yaml.Marshal(&versionInfo)
				if err != nil {
					return err
				}
				fmt.Println(string(marshaled)) //nolint
			} else {
				fmt.Println("Git Version:", versionInfo.GitVersion) //nolint
				fmt.Println("Git Commit:", versionInfo.GitCommit)   //nolint
				fmt.Println("Build Date:", versionInfo.BuildDate)   //nolint
				fmt.Println("Go Version:", versionInfo.GoVersion)   //nolint
				fmt.Println("Compiler:", versionInfo.Compiler)      //nolint
				fmt.Println("Platform:", versionInfo.Platform)      //nolint
			}
			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "json|yaml")

	return cmd
}
