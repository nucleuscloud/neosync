package connections_cmd

import (
	"context"
	"fmt"
	"time"

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

			debugMode, err := cmd.Flags().GetBool("debug")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return listConnections(cmd.Context(), debugMode, &apiKey, &accountId)
		},
	}
	cmd.Flags().
		String("account-id", "", "Account to list connections for. Defaults to account id in cli context")
	return cmd
}

func listConnections(
	ctx context.Context,
	debugMode bool,
	apiKey,
	accountIdFlag *string,
) error {
	logger := cli_logger.NewSLogger(cli_logger.GetCharmLevelOrDefault(debugMode))

	neosyncurl := auth.GetNeosyncUrl()
	httpclient, err := auth.GetNeosyncHttpClient(ctx, logger, auth.WithApiKey(apiKey))
	if err != nil {
		return err
	}

	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(httpclient, neosyncurl)

	accountId, err := auth.ResolveAccountIdFromFlag(ctx, userclient, accountIdFlag, apiKey, logger)
	if err != nil {
		return err
	}

	connectionclient := mgmtv1alpha1connect.NewConnectionServiceClient(httpclient, neosyncurl)

	connections, err := getConnections(ctx, connectionclient, accountId)
	if err != nil {
		return err
	}

	fmt.Println() //nolint:forbidigo
	printConnectionsTable(connections)
	fmt.Println() //nolint:forbidigo
	return nil
}

func getConnections(
	ctx context.Context,
	connectionclient mgmtv1alpha1connect.ConnectionServiceClient,
	accountId string,
) ([]*mgmtv1alpha1.Connection, error) {
	res, err := connectionclient.GetConnections(
		ctx,
		connect.NewRequest[mgmtv1alpha1.GetConnectionsRequest](&mgmtv1alpha1.GetConnectionsRequest{
			AccountId: accountId,
		}),
	)
	if err != nil {
		return nil, err
	}
	return res.Msg.GetConnections(), nil
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
	switch cc.GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return "PostgreSQL"
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return "MySQL"
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		return "AWS S3"
	case *mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig:
		return "GCP Cloud Storage"
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		return "MongoDB"
	case *mgmtv1alpha1.ConnectionConfig_OpenaiConfig:
		return "OpenAI"
	case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
		return "DynamoDB"
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		return "MSSQL"
	default:
		return "Unknown"
	}
}
