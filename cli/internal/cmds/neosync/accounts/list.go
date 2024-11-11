package accounts_cmd

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	cli_logger "github.com/nucleuscloud/neosync/cli/internal/logger"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list accounts",
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
			return listAccounts(cmd.Context(), &apiKey, debugMode)
		},
	}
}

func listAccounts(
	ctx context.Context,
	apiKey *string,
	debug bool,
) error {
	logger := cli_logger.NewSLogger(cli_logger.GetCharmLevelOrDefault(debug))
	httpclient, err := auth.GetNeosyncHttpClient(ctx, logger, auth.WithApiKey(apiKey))
	if err != nil {
		return err
	}

	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(
		httpclient,
		auth.GetNeosyncUrl(),
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
	fmt.Println() //nolint:forbidigo
	printAccountTable(accounts)
	fmt.Println() //nolint:forbidigo
	return nil
}

func printAccountTable(
	accounts []*mgmtv1alpha1.UserAccount,
) {
	tbl := table.
		New("Id", "Name").
		WithHeaderFormatter(
			color.New(color.FgGreen, color.Underline).SprintfFunc(),
		).
		WithFirstColumnFormatter(
			color.New(color.FgYellow).SprintfFunc(),
		)

	for _, account := range accounts {
		tbl.AddRow(
			account.Id,
			account.Name,
		)
	}
	tbl.Print()
}
