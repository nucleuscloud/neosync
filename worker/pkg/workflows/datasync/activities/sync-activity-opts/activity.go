package syncactivityopts_activity

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Activity struct {
	jobclient mgmtv1alpha1connect.JobServiceClient
}

func New(
	jobclient mgmtv1alpha1connect.JobServiceClient,
) *Activity {
	return &Activity{
		jobclient: jobclient,
	}
}

type RetrieveActivityOptionsRequest struct {
	JobId string
}
type RetrieveActivityOptionsResponse struct {
	SyncActivityOptions *workflow.ActivityOptions
}

func (a *Activity) RetrieveActivityOptions(
	ctx context.Context,
	req *RetrieveActivityOptionsRequest,
	wfmetadata *shared.WorkflowMetadata,
) (*RetrieveActivityOptionsResponse, error) {
	logger := log.With(
		activity.GetLogger(ctx),
		"jobId", req.JobId,
		"WorkflowID", wfmetadata.WorkflowId,
		"RunID", wfmetadata.RunId,
	)
	_ = logger

	jobResp, err := a.jobclient.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{Id: req.JobId}))
	if err != nil {
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}
	job := jobResp.Msg.Job
	return &RetrieveActivityOptionsResponse{
		SyncActivityOptions: getSyncActivityOptionsFromJob(job),
	}, nil
}

const (
	defaultStartCloseTimeout = 10 * time.Minute
	defaultMaxAttempts       = 1
)

func getSyncActivityOptionsFromJob(job *mgmtv1alpha1.Job) *workflow.ActivityOptions {
	syncActivityOptions := &workflow.ActivityOptions{
		HeartbeatTimeout: 1 * time.Minute,
	}
	if job.SyncOptions != nil {
		if job.SyncOptions.StartToCloseTimeout != nil {
			syncActivityOptions.StartToCloseTimeout = time.Duration(*job.SyncOptions.StartToCloseTimeout)
		}
		if job.SyncOptions.ScheduleToCloseTimeout != nil {
			syncActivityOptions.ScheduleToCloseTimeout = time.Duration(*job.SyncOptions.ScheduleToCloseTimeout)
		}
		if job.SyncOptions.RetryPolicy != nil {
			if job.SyncOptions.RetryPolicy.MaximumAttempts != nil {
				if syncActivityOptions.RetryPolicy == nil {
					syncActivityOptions.RetryPolicy = &temporal.RetryPolicy{}
				}
				syncActivityOptions.RetryPolicy.MaximumAttempts = *job.SyncOptions.RetryPolicy.MaximumAttempts
			}
		}
	} else {
		return &workflow.ActivityOptions{
			StartToCloseTimeout: defaultStartCloseTimeout, // backwards compatible default for pre-existing jobs that do not have sync options defined
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: defaultMaxAttempts, // backwards compatible default for pre-existing jobs that do not have sync options defined
			},
		}
	}
	if syncActivityOptions.StartToCloseTimeout == 0 && syncActivityOptions.ScheduleToCloseTimeout == 0 {
		syncActivityOptions.StartToCloseTimeout = defaultStartCloseTimeout
	}
	if syncActivityOptions.RetryPolicy == nil {
		syncActivityOptions.RetryPolicy = &temporal.RetryPolicy{MaximumAttempts: defaultMaxAttempts}
	}
	return syncActivityOptions
}
