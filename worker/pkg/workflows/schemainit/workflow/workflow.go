package schemainit_workflow

import (
	initschema_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/schemainit/activities/init-schema"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type SchemaInitRequest struct {
	AccountId     string
	JobId         string
	JobRunId      string
	DestinationId string

	SchemaInitActivityOptions *workflow.ActivityOptions
}

type SchemaInitResponse struct{}

type Workflow struct{}

func New() *Workflow {
	return &Workflow{}
}

func (w *Workflow) SchemaInit(ctx workflow.Context, req *SchemaInitRequest) (*SchemaInitResponse, error) {
	logger := log.With(
		workflow.GetLogger(ctx),
		"accountId", req.AccountId,
		"jobId", req.JobId,
		"jobRunId", req.JobRunId,
		"destinationId", req.DestinationId,
	)

	logger.Info("scheduling InitSchema activityfor execution.")
	var resp *initschema_activity.RunSqlInitTableStatementsResponse
	var initSchema *initschema_activity.Activity
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, *req.SchemaInitActivityOptions),
		initSchema.RunSqlInitTableStatements,
		&initschema_activity.RunSqlInitTableStatementsRequest{
			JobId:         req.JobId,
			JobRunId:      req.JobRunId,
			DestinationId: req.DestinationId,
		}).
		Get(ctx, &resp)
	if err != nil {
		return nil, err
	}
	logger.Info("completed InitSchema activity.")

	return &SchemaInitResponse{}, nil
}
