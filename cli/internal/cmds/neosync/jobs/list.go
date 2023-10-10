package jobs_cmd

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return listJobs(cmd.Context())
		},
	}
}

func listJobs(
	ctx context.Context,
) error {
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		http.DefaultClient,
		"http://localhost:8080",
	)
	res, err := jobclient.GetJobs(ctx, connect.NewRequest[mgmtv1alpha1.GetJobsRequest](&mgmtv1alpha1.GetJobsRequest{
		AccountId: "4f45a5ff-b1ff-47f5-9f89-f576ebd1d03c", // todo: pull from context
	}))
	if err != nil {
		return err
	}
	// todo: use table library from nucleus cli
	for _, job := range res.Msg.Jobs {
		fmt.Println(job.Id, job.Name)
	}
	return nil
}
