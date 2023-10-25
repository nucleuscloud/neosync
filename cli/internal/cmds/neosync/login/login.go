package login_cmd

import (
	"context"

	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Neosync",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return login(cmd.Context())
		},
	}
	return cmd
}

func login(ctx context.Context) error {
	return nil
}
