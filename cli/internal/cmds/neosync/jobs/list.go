package jobs_cmd

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
			debugMode, err := cmd.Flags().GetBool("debug")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return listJobs(cmd.Context(), debugMode, &apiKey, &accountId)
		},
	}
	cmd.Flags().
		String("account-id", "", "Account to list jobs for. Defaults to account id in cli context")
	return cmd
}

func listJobs(
	ctx context.Context,
	debug bool,
	apiKey,
	accountIdFlag *string,
) error {
	logger := cli_logger.NewSLogger(cli_logger.GetCharmLevelOrDefault(debug))

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

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		httpclient,
		neosyncurl,
	)
	res, err := jobclient.GetJobs(
		ctx,
		connect.NewRequest[mgmtv1alpha1.GetJobsRequest](&mgmtv1alpha1.GetJobsRequest{
			AccountId: accountId,
		}),
	)
	if err != nil {
		return err
	}

	jobstatuses := make([]*mgmtv1alpha1.JobStatus, len(res.Msg.Jobs))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range res.Msg.Jobs {
		idx := idx
		errgrp.Go(func() error {
			jsres, err := jobclient.GetJobStatus(
				errctx,
				connect.NewRequest[mgmtv1alpha1.GetJobStatusRequest](
					&mgmtv1alpha1.GetJobStatusRequest{
						JobId: res.Msg.Jobs[idx].Id,
					},
				),
			)
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
