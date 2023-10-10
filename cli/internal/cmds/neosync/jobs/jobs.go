package jobs_cmd

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Parent command for jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newTriggerCmd())
	return cmd
}
