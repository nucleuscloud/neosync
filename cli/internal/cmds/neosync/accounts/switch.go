package accounts_cmd

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/spf13/cobra"
)

func newSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trigger",
		Short: "trigger a job",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("must provide job uuid as argument")
			}

			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}

			jobId := args[0]

			jobUuid, err := uuid.Parse(jobId)
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true
			return triggerJob(cmd.Context(), jobUuid.String(), &apiKey)
		},
	}
}

func triggerJob(
	ctx context.Context,
	jobId string,
	apiKey *string,
) error {
	isAuthEnabled, err := auth.IsAuthEnabled(ctx)
	if err != nil {
		return err
	}
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		http.DefaultClient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey))),
	)
	_, err = jobclient.CreateJobRun(ctx, connect.NewRequest[mgmtv1alpha1.CreateJobRunRequest](&mgmtv1alpha1.CreateJobRunRequest{
		JobId: jobId,
	}))
	if err != nil {
		return err
	}
	return nil
}
