package jobs_cmd

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
	"golang.org/x/sync/errgroup"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list jobs",
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
			return listJobs(cmd.Context(), &apiKey, &accountId)
		},
	}
	cmd.Flags().String("account-id", "", "Account to list jobs for. Defaults to account id in cli context")
	return cmd
}

func listJobs(
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
			fmt.Println("Unable to retrieve account id. Please use account switch command to set account.") //nolint:forbidigo
			return err
		}
		accountId = &aId
	}

	if accountId == nil || *accountId == "" {
		return errors.New("Account Id not found. Please use account switch command to set account.")
	}

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		http.DefaultClient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)
	res, err := jobclient.GetJobs(ctx, connect.NewRequest[mgmtv1alpha1.GetJobsRequest](&mgmtv1alpha1.GetJobsRequest{
		AccountId: *accountId,
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

	fmt.Println() //nolint:forbidigo
	printJobTable(res.Msg.Jobs, jobstatuses)
	fmt.Println() //nolint:forbidigo
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
