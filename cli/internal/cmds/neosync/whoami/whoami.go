package whoami_cmd

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Find out who you are",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return whoami(cmd.Context())
		},
	}

	return cmd
}

func whoami(ctx context.Context) error {
	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(
		http.DefaultClient,
		"http://localhost:8080",
	)
	resp, err := userclient.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return err
	}
	fmt.Println("UserId: ", resp.Msg.UserId)
	return nil
}
