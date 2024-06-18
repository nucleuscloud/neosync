package whoami_cmd

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/nucleuscloud/neosync/cli/internal/version"
	http_client "github.com/nucleuscloud/neosync/worker/pkg/http/client"
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
			cmd.SilenceUsage = true
			return whoami(cmd.Context(), &apiKey)
		},
	}

	return cmd
}

func whoami(ctx context.Context, apiKey *string) error {
	isAuthEnabled, err := auth.IsAuthEnabled(ctx)
	if err != nil {
		return err
	}
	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(
		http_client.NewWithHeaders(version.Get().Headers()),
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)
	resp, err := userclient.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return err
	}
	// todo: layer in account data and access/id token information for even more goodness
	fmt.Println("UserId:", resp.Msg.UserId) //nolint:forbidigo
	return nil
}
