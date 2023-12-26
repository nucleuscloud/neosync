package connections_cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list connections",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}

			accountId, err := cmd.Flags().GetString("account-id")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return listConnections(cmd.Context(), &apiKey, &accountId)
		},
	}
	cmd.Flags().String("account-id", "", "Account to list connections for. Defaults to account id in cli context")
	return cmd
}

func listConnections(
	ctx context.Context,
	apiKey, accountIdFlag *string,
) error {
	isAuthEnabled, err := auth.IsAuthEnabled(ctx)
	if err != nil {
		return err
	}

	var accountId = accountIdFlag
	if accountId == nil || *accountId == "" {
		aId, err := userconfig.GetAccountId()
		if err != nil {
			fmt.Println("Unable to retrieve account id. Please use account switch command to set account.") // nolint
			return err
		}
		accountId = &aId
	}

	if accountId == nil || *accountId == "" {
		return errors.New("Account Id not found. Please use account switch command to set account.")
	}

	connectionclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		http.DefaultClient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)
	res, err := connectionclient.GetConnections(ctx, connect.NewRequest[mgmtv1alpha1.GetConnectionsRequest](&mgmtv1alpha1.GetConnectionsRequest{
		AccountId: *accountId,
	}))
	if err != nil {
		return err
	}

	fmt.Println() // nolint
	printConnectionsTable(res.Msg.Connections)
	fmt.Println() // nolint
	return nil
}

func printConnectionsTable(
	connections []*mgmtv1alpha1.Connection,
) {
	tbl := table.
		New("Id", "Name", "Category", "Created At", "Updated At").
		WithHeaderFormatter(
			color.New(color.FgGreen, color.Underline).SprintfFunc(),
		).
		WithFirstColumnFormatter(
			color.New(color.FgYellow).SprintfFunc(),
		)

	for idx := range connections {
		connection := connections[idx]
		tbl.AddRow(
			connection.Id,
			connection.Name,
			getCategory(connection.GetConnectionConfig()),
			connection.CreatedAt.AsTime().Local().Format(time.RFC3339),
			connection.UpdatedAt.AsTime().Local().Format(time.RFC3339),
		)
	}
	tbl.Print()
}

func getCategory(cc *mgmtv1alpha1.ConnectionConfig) string {
	if cc == nil {
		return "Unknown"
	}
	switch cc.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return "Postgres"
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return "Mysql"
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		return "AWS S3"
	default:
		return "Unknown"
	}
}
