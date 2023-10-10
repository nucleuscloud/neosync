package jobs_cmd

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/spf13/cobra"
)

func newTriggerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trigger",
		Short: "trigger a job",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("must provide job uuid as argument")
			}

			jobId := args[0]

			jobUuid, err := uuid.Parse(jobId)
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true
			return triggerJob(cmd.Context(), jobUuid.String())
		},
	}
}

func triggerJob(
	ctx context.Context,
	jobId string,
) error {
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		http.DefaultClient,
		"localhost:8080",
	)
	_, err := jobclient.CreateJobRun(ctx, connect.NewRequest[mgmtv1alpha1.CreateJobRunRequest](&mgmtv1alpha1.CreateJobRunRequest{
		JobId: jobId,
	}))
	if err != nil {
		return err
	}
	return nil
}
