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
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return listJobs(cmd.Context(), &apiKey)
		},
	}
}

func listJobs(
	ctx context.Context,
	apiKey *string,
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
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)
	res, err := jobclient.GetJobs(ctx, connect.NewRequest[mgmtv1alpha1.GetJobsRequest](&mgmtv1alpha1.GetJobsRequest{
		AccountId: account.Id,
	}))
	if err != nil {
		return err
	}

	jobstatuses := make([]*mgmtv1alpha1.JobStatus, len(res.Msg.Jobs))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range res.Msg.Jobs {
		idx := idx
		errgrp.Go(func() error {
			jsres, err := jobclient.GetJobStatus(errctx, connect.NewRequest[mgmtv1alpha1.GetJobStatusRequest](&mgmtv1alpha1.GetJobStatusRequest{
				JobId: res.Msg.Jobs[idx].Id,
			}))
			if err != nil {
				return err
			}
			jobstatuses[idx] = &jsres.Msg.Status
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return err
	}

	printJobTable(res.Msg.Jobs, jobstatuses)
	return nil
}

func printJobTable(
	jobs []*mgmtv1alpha1.Job,
	jobstatuses []*mgmtv1alpha1.JobStatus,
) {
	tbl := table.
		New("Id", "Name", "Status", "Created At", "Updated At").
		WithHeaderFormatter(
			color.New(color.FgGreen, color.Underline).SprintfFunc(),
		).
		WithFirstColumnFormatter(
			color.New(color.FgYellow).SprintfFunc(),
		)

	for idx := range jobs {
		job := jobs[idx]
		js := jobstatuses[idx]
		tbl.AddRow(
			job.Id,
			job.Name,
			js.String(),
			job.CreatedAt.AsTime().Local().Format(time.RFC3339),
			job.UpdatedAt.AsTime().Local().Format(time.RFC3339),
		)
	}
	tbl.Print()
}
