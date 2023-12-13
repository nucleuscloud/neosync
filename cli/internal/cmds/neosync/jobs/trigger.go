package jobs_cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
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

			cmd.SilenceUsage = true
			return triggerJob(cmd.Context(), jobUuid.String(), &apiKey, &accountId)
		},
	}
	cmd.Flags().String("account-id", "", "Account that job is in. Defaults to account id in cli context")
	return cmd
}

func triggerJob(
	ctx context.Context,
	jobId string,
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

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		http.DefaultClient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey))),
	)
	job, err := jobclient.GetJob(ctx, connect.NewRequest[mgmtv1alpha1.GetJobRequest](&mgmtv1alpha1.GetJobRequest{
		Id: jobId,
	}))
	if err != nil {
		return err
	}
	if job.Msg.Job.AccountId != *accountId {
		return fmt.Errorf("Unable to trigger job run. Job not found. AccountId: %s", *accountId)
	}
	_, err = jobclient.CreateJobRun(ctx, connect.NewRequest[mgmtv1alpha1.CreateJobRunRequest](&mgmtv1alpha1.CreateJobRunRequest{
		JobId: jobId,
	}))
	if err != nil {
		return err
	}
	return nil
}
