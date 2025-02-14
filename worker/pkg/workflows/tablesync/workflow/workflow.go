package tablesync_workflow

import (
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/activities/sync"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type TableSyncRequest struct {
	AccountId string
	Id        string

	ContinuationToken *string

	SyncActivityOptions *workflow.ActivityOptions
}

type TableSyncResponse struct {
}

type Workflow struct{}

const MAX_ITERATIONS = 10

func (*Workflow) TableSync(ctx workflow.Context, req *TableSyncRequest) (*TableSyncResponse, error) {
	logger := log.With(
		workflow.GetLogger(ctx),
		"accountId", req.AccountId,
		"id", req.Id,
		"isContinuation", req.ContinuationToken != nil,
	)

	var syncActivity *sync_activity.Activity

	var continuationToken *string
	var iterations int

	logger.Debug("starting table sync")

	for {
		iterations++

		var resp *sync_activity.SyncResponse
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, *req.SyncActivityOptions), // todo: check sync activity options nil
			syncActivity.Sync,
			sync_activity.SyncRequest{
				Id:                req.Id,
				AccountId:         req.AccountId,
				ContinuationToken: continuationToken,
			},
		).
			Get(ctx, &resp)
		if err != nil {
			return nil, err
		}
		continuationToken = resp.ContinuationToken
		if continuationToken == nil {
			logger.Debug("no continuation token, breaking")
			break
		}
		if iterations >= MAX_ITERATIONS {
			logger.Debug("max iterations reached, continuing as new")
			newReq := &TableSyncRequest{
				AccountId:         req.AccountId,
				Id:                req.Id,
				ContinuationToken: continuationToken,
			}
			return nil, workflow.NewContinueAsNewError(ctx, newReq)
		}
		logger.Debug("continuing")
	}
	return &TableSyncResponse{}, nil
}
