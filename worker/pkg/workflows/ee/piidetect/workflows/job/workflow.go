package piidetect_job_workflow

import (
	"fmt"
	"time"

	piidetect_job_activities "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/job/activities"
	piidetect_table_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/table"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Workflow struct{}

func New() *Workflow {
	return &Workflow{}
}

type PiiDetectRequest struct {
	JobId string
}

type PiiDetectResponse struct{}

func (w *Workflow) JobPiiDetect(ctx workflow.Context, req *PiiDetectRequest) (*PiiDetectResponse, error) {
	logger := log.With(
		workflow.GetLogger(ctx),
		"jobId", req.JobId,
	)

	logger.Info("starting PII detection")

	var activities *piidetect_job_activities.Activities

	var jobDetailsResp *piidetect_job_activities.GetPiiDetectJobDetailsResponse
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 1 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 3,
			},
		}),
		activities.GetPiiDetectJobDetails,
		&piidetect_job_activities.GetPiiDetectJobDetailsRequest{
			JobId: req.JobId,
		},
	).Get(ctx, &jobDetailsResp)
	if err != nil {
		return nil, err
	}

	var tablesToScanResp *piidetect_job_activities.GetTablesToPiiScanResponse
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 1 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 3,
			},
		}),
		activities.GetTablesToPiiScan,
		&piidetect_job_activities.GetTablesToPiiScanRequest{
			SourceConnectionId: jobDetailsResp.SourceConnectionId,
			Filter:             nil, // todo
		},
	).Get(ctx, &tablesToScanResp)
	if err != nil {
		return nil, err
	}

	err = w.orchestrateTables(ctx, tablesToScanResp.Tables, req.JobId, jobDetailsResp.SourceConnectionId, logger)
	if err != nil {
		return nil, err
	}

	logger.Info("PII detection completed")
	return &PiiDetectResponse{}, nil
}

func (w *Workflow) orchestrateTables(
	ctx workflow.Context,
	tables []piidetect_job_activities.TableIdentifier,
	jobId string,
	connectionId string,
	logger log.Logger,
) error {
	workselector := workflow.NewNamedSelector(ctx, "job_pii_detect")

	maxConcurrency := 3
	inFlightLimiter := workflow.NewSemaphore(ctx, int64(maxConcurrency))

	tableWf := piidetect_table_workflow.New()

	processTable := func(table piidetect_job_activities.TableIdentifier) error {
		if err := inFlightLimiter.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("unable to acquire semaphore: %w", err)
		}
		workselector.AddFuture(
			workflow.ExecuteChildWorkflow(
				workflow.WithChildOptions(
					ctx,
					workflow.ChildWorkflowOptions{
						// WorkflowID: , // todo
						RetryPolicy: &temporal.RetryPolicy{
							MaximumAttempts: 1,
						},
						WorkflowRunTimeout: 5 * time.Minute,
					}),
				tableWf.TablePiiDetect,
				&piidetect_table_workflow.TablePiiDetectRequest{
					JobId:        jobId,
					ConnectionId: connectionId,
					TableSchema:  table.Schema,
					TableName:    table.Table,
				},
			),
			func(f workflow.Future) {
				var wfResult *piidetect_table_workflow.TablePiiDetectResponse
				err := f.Get(ctx, &wfResult)
				inFlightLimiter.Release(1)
				if err != nil {
					logger.Error("activity did not complete", "err", err)
				}
				// todo: handle result, good or bad
			},
		)
		return nil
	}

	for _, table := range tables {
		if err := processTable(table); err != nil {
			return err
		}
	}

	for workselector.HasPending() {
		workselector.Select(ctx)
	}

	logger.Debug("all tables processed")
	return nil
}
