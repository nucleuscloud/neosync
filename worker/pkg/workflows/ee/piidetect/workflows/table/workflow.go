package piidetect_table_workflow

import (
	"time"

	piidetect_table_activities "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/table/activities"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Workflow struct{}

func New() *Workflow {
	return &Workflow{}
}

type TablePiiDetectRequest struct {
	JobId            string
	ConnectionId     string
	TableSchema      string
	TableName        string
	ShouldSampleData bool

	PreviousResultsKey *string // incremental mode to only detect pii for new columns
}

type TablePiiDetectResponse struct {
	PiiColumns map[string]PiiDetectReport
}

type PiiDetectReport struct {
	Regex *piidetect_table_activities.PiiCategory
	LLM   *piidetect_table_activities.PiiDetectReport
}

func (w *Workflow) TablePiiDetect(ctx workflow.Context, req *TablePiiDetectRequest) (*TablePiiDetectResponse, error) {
	logger := log.With(
		workflow.GetLogger(ctx),
		"jobId", req.JobId,
		"tableSchema", req.TableSchema,
		"tableName", req.TableName,
	)

	logger.Info("starting PII detection")

	var activities *piidetect_table_activities.Activities

	var columDataResp *piidetect_table_activities.GetColumnDataResponse
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 1 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 3,
			},
		}),
		activities.GetColumnData,
		&piidetect_table_activities.GetColumnDataRequest{
			ConnectionId: req.ConnectionId,
			TableSchema:  req.TableSchema,
			TableName:    req.TableName,
		},
	).Get(ctx, &columDataResp)
	if err != nil {
		return nil, err
	}

	var regexResp *piidetect_table_activities.DetectPiiRegexResponse
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 1 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 3,
			},
		}),
		activities.DetectPiiRegex,
		&piidetect_table_activities.DetectPiiRegexRequest{
			ColumnData: columDataResp.ColumnData,
		},
	).Get(ctx, &regexResp)
	if err != nil {
		return nil, err
	}

	var llmResp *piidetect_table_activities.DetectPiiLLMResponse
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 1 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 3,
			},
		}),
		activities.DetectPiiLLM,
		&piidetect_table_activities.DetectPiiLLMRequest{
			TableName:  req.TableName,
			ColumnData: columDataResp.ColumnData,
		},
	).Get(ctx, &llmResp)
	if err != nil {
		return nil, err
	}

	logger.Debug("PII detection complete", "piiColumns", regexResp.PiiColumns)

	wfResp := &TablePiiDetectResponse{
		PiiColumns: make(map[string]PiiDetectReport),
	}

	for col, category := range regexResp.PiiColumns {
		wfResp.PiiColumns[col] = PiiDetectReport{
			Regex: &category,
		}
	}

	for col, report := range llmResp.PiiColumns {
		if _, ok := wfResp.PiiColumns[col]; ok {
			// Create a new report with both regex and LLM data
			existingReport := wfResp.PiiColumns[col]
			existingReport.LLM = &report
			wfResp.PiiColumns[col] = existingReport
		} else {
			wfResp.PiiColumns[col] = PiiDetectReport{
				LLM: &report,
			}
		}
	}

	// if sample data is enabled, we should query to retrieve some sample data as well, but maybe that should not come back to the workflow

	return wfResp, nil
}
