package tablesync_workflow

import (
	"errors"

	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/activities/sync"
	tablesync_shared "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/shared"
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

	ColumnIdentityCursors map[string]*tablesync_shared.IdentityCursor
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

	cursors := req.ColumnIdentityCursors

	if len(cursors) > 0 {
		err := setCursorUpdateHandler(ctx, cursors)
		if err != nil {
			return nil, err
		}
	}

	continuationToken := req.ContinuationToken
	var iterations int

	logger.Debug("starting table sync")

	var syncActivity *sync_activity.Activity
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
			// ensures that all update handlers have finished
			err = workflow.Await(ctx, func() bool { return workflow.AllHandlersFinished(ctx) })
			if err != nil {
				return nil, err
			}
			newReq := *req
			newReq.ContinuationToken = continuationToken
			newReq.ColumnIdentityCursors = cursors
			var wf *Workflow
			return nil, workflow.NewContinueAsNewError(ctx, wf.TableSync, &newReq)
		}
		logger.Debug("continuing")
	}
	// ensures that all update handlers have finished
	err := workflow.Await(ctx, func() bool { return workflow.AllHandlersFinished(ctx) })
	if err != nil {
		return nil, err
	}
	return &TableSyncResponse{
		Schema: req.TableSchema,
		Table:  req.TableName,
	}, nil
}

// Sets a temporal update handle for use with allocating identity blocks for auto increment columns
func setCursorUpdateHandler(ctx workflow.Context, cursors map[string]*tablesync_shared.IdentityCursor) error {
	cursorMutex := workflow.NewMutex(ctx)
	return workflow.SetUpdateHandlerWithOptions(
		ctx,
		tablesync_shared.AllocateIdentityBlock,
		func(ctx workflow.Context, req *tablesync_shared.AllocateIdentityBlockRequest) (*tablesync_shared.AllocateIdentityBlockResponse, error) {
			err := cursorMutex.Lock(ctx)
			if err != nil {
				return nil, err
			}
			defer cursorMutex.Unlock()
			cursor := cursors[req.Id]
			if cursor == nil {
				return nil, errors.New("cursor not found for provided id")
			}
			startValue := cursor.CurrentValue
			cursor.CurrentValue += req.BlockSize // prepare for next allocation
			cursors[req.Id] = cursor
			return &tablesync_shared.AllocateIdentityBlockResponse{
				StartValue: startValue,
				EndValue:   startValue + req.BlockSize,
			}, nil
		},
		workflow.UpdateHandlerOptions{
			Description: "Handles allocating blocks of integers to be used for auto increment columns",
			Validator: func(ctx workflow.Context, req *tablesync_shared.AllocateIdentityBlockRequest) error {
				if req == nil {
					return errors.New("request is nil, expected a valid *AllocateIdentityBlockRequest")
				}
				if req.Id == "" || req.BlockSize == 0 {
					return errors.New("id and block size are required")
				}
				err := cursorMutex.Lock(ctx)
				if err != nil {
					return err
				}
				defer cursorMutex.Unlock()
				if _, ok := cursors[req.Id]; !ok {
					return errors.New("cursor not found for provided id")
				}
				return nil
			},
		},
	)
}
