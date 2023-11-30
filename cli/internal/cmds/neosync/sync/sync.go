package sync_cmd

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "One off sync job to local resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return sync(cmd.Context(), &apiKey)
		},
	}

	return cmd
}

func sync(ctx context.Context, apiKey *string) error {
	fmt.Println("CLI")
	isAuthEnabled, err := auth.IsAuthEnabled(ctx)
	if err != nil {
		return err
	}
	// userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(
	// 	http.DefaultClient,
	// 	serverconfig.GetApiBaseUrl(),
	// 	connect.WithInterceptors(
	// 		auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
	// 	),
	// )
	// resp, err := userclient.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	// if err != nil {
	// 	return err
	// }

	connectionclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		http.DefaultClient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)
	fmt.Println("connection client")

	stream, err := connectionclient.GetConnectionDataStream(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionDataStreamRequest{
		SourceConnectionId: "3b4db2af-ef33-4e26-b0b9-f6df7518e78b",
		Schema:             "public",
		Table:              "locations",
	}))
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("get data stream")

	for {
		response := stream.Receive()
		if response {
			fmt.Println(stream.Msg().Data)
		} else {
			return nil
		}

		// response

	}

}
