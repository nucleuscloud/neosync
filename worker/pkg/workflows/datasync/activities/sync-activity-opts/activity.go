package syncactivityopts_activity

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
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
	SyncActivityOptions  *workflow.ActivityOptions
	AccountId            string
	RequestedRecordCount *uint64
}

func (a *Activity) RetrieveActivityOptions(
	ctx context.Context,
	req *RetrieveActivityOptionsRequest,
) (*RetrieveActivityOptionsResponse, error) {
	activityInfo := activity.GetInfo(ctx)
	logger := log.With(
		activity.GetLogger(ctx),
		"jobId", req.JobId,
		"WorkflowID", activityInfo.WorkflowExecution.ID,
		"RunID", activityInfo.WorkflowExecution.RunID,
	)
	logger.Debug("retrieving activity options")

	jobResp, err := a.jobclient.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{Id: req.JobId}))
	if err != nil {
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}
	job := jobResp.Msg.GetJob()
	return &RetrieveActivityOptionsResponse{
		SyncActivityOptions:  getSyncActivityOptionsFromJob(job),
		AccountId:            job.GetAccountId(),
		RequestedRecordCount: getRequestedRecordCount(job),
	}, nil
}

func getRequestedRecordCount(job *mgmtv1alpha1.Job) *uint64 {
	switch config := job.GetSource().GetOptions().GetConfig().(type) {
	case *mgmtv1alpha1.JobSourceOptions_AiGenerate:
		return zeroToNilPointer(getAiGeneratedRequestedCount(config.AiGenerate))
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		return zeroToNilPointer(getGenerateRequestedCount(config.Generate))
	default:
		return nil
	}
}

func getAiGeneratedRequestedCount(config *mgmtv1alpha1.AiGenerateSourceOptions) uint64 {
	if config == nil {
		config = &mgmtv1alpha1.AiGenerateSourceOptions{}
	}
	total := uint64(0)
	for _, schema := range config.GetSchemas() {
		for _, table := range schema.GetTables() {
			count := table.GetRowCount()
			if count > 0 {
				total += uint64(count)
			}
		}
	}
	return total
}

func getGenerateRequestedCount(config *mgmtv1alpha1.GenerateSourceOptions) uint64 {
	if config == nil {
		config = &mgmtv1alpha1.GenerateSourceOptions{}
	}
	total := uint64(0)
	for _, schema := range config.GetSchemas() {
		for _, table := range schema.GetTables() {
			count := table.GetRowCount()
			if count > 0 {
				total += uint64(count)
			}
		}
	}
	return total
}

// if the input is less than or equal to 0, returns nil
func zeroToNilPointer[T uint64 | int64](value T) *T {
	if value <= 0 {
		return nil
	}
	return &value
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
