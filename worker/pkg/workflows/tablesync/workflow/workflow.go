package tablesync_workflow

import (
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/activities/sync"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type TableSyncRequest struct {
	AccountId         string
	Id                string
	JobRunId          string
	ContinuationToken *string

	SyncActivityOptions *workflow.ActivityOptions
	TableSchema         string
	TableName           string
}

type TableSyncResponse struct {
	// Here to make it easier to see in UI and logs
	Schema string
	Table  string
}

type Workflow struct {
	maxIterations int
}

func New(maxIterations int) *Workflow {
	return &Workflow{
		maxIterations: maxIterations,
	}
}

func (w *Workflow) TableSync(ctx workflow.Context, req *TableSyncRequest) (*TableSyncResponse, error) {
	logger := log.With(
		workflow.GetLogger(ctx),
		"accountId", req.AccountId,
		"id", req.Id,
		"isContinuation", req.ContinuationToken != nil,
	)

	var syncActivity *sync_activity.Activity

	continuationToken := req.ContinuationToken
	var iterations int

	logger.Debug("starting table sync")

	for {
		iterations++

		var resp *sync_activity.SyncTableResponse
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, *req.SyncActivityOptions), // todo: check sync activity options nil
			syncActivity.SyncTable,
			sync_activity.SyncTableRequest{
				Id:                req.Id,
				AccountId:         req.AccountId,
				JobRunId:          req.JobRunId,
				ContinuationToken: continuationToken,
			},
			&sync_activity.SyncMetadata{
				Schema: req.TableSchema,
				Table:  req.TableName,
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
		if iterations >= w.maxIterations {
			logger.Debug("max iterations reached, continuing as new")
			newReq := *req
			newReq.ContinuationToken = continuationToken
			var wf *Workflow
			return nil, workflow.NewContinueAsNewError(ctx, wf.TableSync, &newReq)
		}
		logger.Debug("continuing")
	}
	return &TableSyncResponse{
		Schema: req.TableSchema,
		Table:  req.TableName,
	}, nil
}
