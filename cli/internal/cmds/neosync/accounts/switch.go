package accounts_cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
	"github.com/spf13/cobra"
)

func newSwitchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch",
		Short: "switch accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}

			id, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true
			return switchAccount(cmd.Context(), &apiKey, &id, &name)
		},
	}
	cmd.Flags().String("id", "", "Account id to switch to")
	cmd.Flags().String("name", "", "Account name to switch to")
	cmd.MarkFlagsOneRequired("id", "name")
	return cmd
}

func switchAccount(
	ctx context.Context,
	apiKey, id, name *string,
) error {
	isAuthEnabled, err := auth.IsAuthEnabled(ctx)
	if err != nil {
		return err
	}

	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(
		http.DefaultClient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)

	accountsResp, err := userclient.GetUserAccounts(
		ctx,
		connect.NewRequest[mgmtv1alpha1.GetUserAccountsRequest](&mgmtv1alpha1.GetUserAccountsRequest{}),
	)
	if err != nil {
		return err
	}
	accounts := accountsResp.Msg.Accounts
	if len(accounts) == 0 {
		return errors.New("unable to find accounts for user")
	}

	var account *mgmtv1alpha1.UserAccount
	if id != nil && *id != "" {
		for _, a := range accounts {
			if a.Id == *id {
				account = a

			}
		}
	} else if name != nil && *name != "" {
		for _, a := range accounts {
			if a.Name == *name {
				account = a
			}
		}
	}

	if account == nil {
		return errors.New("unable to find account for user")
	}

	err = userconfig.SetAccountId(account.Id)
	if err != nil {
		fmt.Println("unable to switch accounts") // nolint
		return err
	}

	fmt.Println("Switched accounts")                            // nolint
	fmt.Printf("Name: %s  Id: %s \n", account.Name, account.Id) // nolint

	return nil
}
