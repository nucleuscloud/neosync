package jobs_cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return listJobs()
		},
	}
}

func listJobs() error {
	fmt.Println("todo")
	return nil
}
