package whoami_cmd

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	cli_logger "github.com/nucleuscloud/neosync/cli/internal/logger"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Find out who you are",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}
			debugMode, err := cmd.Flags().GetBool("debug")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return whoami(cmd.Context(), &apiKey, debugMode)
		},
	}

	return cmd
}

func whoami(ctx context.Context, apiKey *string, debug bool) error {
	logger := cli_logger.NewSLogger(cli_logger.GetCharmLevelOrDefault(debug))
	httpclient, err := auth.GetNeosyncHttpClient(ctx, logger, auth.WithApiKey(apiKey))
	if err != nil {
		return err
	}
	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(
		httpclient,
		auth.GetNeosyncUrl(),
	)
	resp, err := userclient.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return err
	}
	// todo: layer in account data and access/id token information for even more goodness
	logger.Info(fmt.Sprintf("UserId: %q", resp.Msg.GetUserId()))
	return nil
}
