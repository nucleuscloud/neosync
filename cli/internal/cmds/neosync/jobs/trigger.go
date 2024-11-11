package jobs_cmd

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	cli_logger "github.com/nucleuscloud/neosync/cli/internal/logger"
	"github.com/spf13/cobra"
)

func newTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger [id]",
		Short: "trigger a job",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("must provide job uuid as argument")
			}

			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}

			accountId, err := cmd.Flags().GetString("account-id")
			if err != nil {
				return err
			}

			jobId := args[0]

			jobUuid, err := uuid.Parse(jobId)
			if err != nil {
				return err
			}

			debugMode, err := cmd.Flags().GetBool("debug")
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true
			return triggerJob(cmd.Context(), debugMode, jobUuid.String(), &apiKey, &accountId)
		},
	}
	cmd.Flags().String("account-id", "", "Account that job is in. Defaults to account id in cli context")
	return cmd
}

func triggerJob(
	ctx context.Context,
	debug bool,
	jobId string,
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
	job, err := jobclient.GetJob(ctx, connect.NewRequest[mgmtv1alpha1.GetJobRequest](&mgmtv1alpha1.GetJobRequest{
		Id: jobId,
	}))
	if err != nil {
		return err
	}
	if job.Msg.GetJob().GetAccountId() != accountId {
		return fmt.Errorf("Unable to trigger job run. Job not found. AccountId: %s", accountId)
	}
	_, err = jobclient.CreateJobRun(ctx, connect.NewRequest[mgmtv1alpha1.CreateJobRunRequest](&mgmtv1alpha1.CreateJobRunRequest{
		JobId: jobId,
	}))
	if err != nil {
		return err
	}
	return nil
}
