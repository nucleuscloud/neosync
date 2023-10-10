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
				fmt.Println(string(marshaled))
			} else if output == "yaml" {
				marshaled, err := yaml.Marshal(&versionInfo)
				if err != nil {
					return err
				}
				fmt.Println(string(marshaled))
			} else {
				fmt.Println("Git Version:", versionInfo.GitVersion)
				fmt.Println("Git Commit:", versionInfo.GitCommit)
				fmt.Println("Build Date:", versionInfo.BuildDate)
				fmt.Println("Go Version:", versionInfo.GoVersion)
				fmt.Println("Compiler:", versionInfo.Compiler)
				fmt.Println("Platform:", versionInfo.Platform)
			}
			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "json|yaml")

	return cmd
}
