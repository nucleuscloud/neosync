package run_cmd

import (
	run_stripe_usage_cmd "github.com/nucleuscloud/neosync/backend/internal/cmds/mgmt/run/stripe-usage"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Parent command for running one-off commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(run_stripe_usage_cmd.NewCmd())
	return cmd
}
