package jobs_cmd

import (
	"context"
	"errors"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return listJobs(cmd.Context())
		},
	}
}

func listJobs(
	ctx context.Context,
) error {
	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(
		http.DefaultClient,
		"http://localhost:8080",
	)

	// todo: this should be settable via cli context in the future to allow users to switch active accounts
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
	account := accounts[0]

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		http.DefaultClient,
		"http://localhost:8080",
	)
	res, err := jobclient.GetJobs(ctx, connect.NewRequest[mgmtv1alpha1.GetJobsRequest](&mgmtv1alpha1.GetJobsRequest{
		AccountId: account.Id,
	}))
	if err != nil {
		return err
	}
	printJobTable(res.Msg.Jobs)
	return nil
}

func printJobTable(
	jobs []*mgmtv1alpha1.Job,
) {
	tbl := table.
		New("Id", "Name", "Status", "Created At", "Updated At").
		WithHeaderFormatter(
			color.New(color.FgGreen, color.Underline).SprintfFunc(),
		).
		WithFirstColumnFormatter(
			color.New(color.FgYellow).SprintfFunc(),
		)

	for _, job := range jobs {
		tbl.AddRow(
			job.Id,
			job.Name,
			job.Status,
			job.CreatedAt.AsTime().Local().Format(time.RFC3339),
			job.UpdatedAt.AsTime().Local().Format(time.RFC3339),
		)
	}
	tbl.Print()
}
