package serve

import (
	serve_connect "github.com/nucleuscloud/neosync/backend/internal/cmds/mgmt/serve/connect"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Parent command for serving",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(serve_connect.NewCmd())
	return cmd
}
