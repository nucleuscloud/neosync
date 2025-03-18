package piidetect_table_workflow

import (
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
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
	AccountId          string
	JobId              string
	ConnectionId       string
	TableSchema        string
	TableName          string
	ShouldSampleData   bool
	UserPrompt         string
	PreviousResultsKey *mgmtv1alpha1.RunContextKey // incremental mode to only detect pii for new columns
	ParentExecutionId  *string                     // present if this is running as a child workflow
}

type TablePiiDetectResponse struct {
	PiiColumns map[string]piidetect_table_activities.CombinedPiiDetectReport
	ResultKey  *mgmtv1alpha1.RunContextKey
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
			TableSchema:  req.TableSchema,
			TableName:    req.TableName,
			ColumnData:   columDataResp.ColumnData,
			ShouldSample: req.ShouldSampleData,
			ConnectionId: req.ConnectionId,
			UserPrompt:   req.UserPrompt,
		},
	).Get(ctx, &llmResp)
	if err != nil {
		return nil, err
	}

	logger.Debug("PII detection complete")

	report := buildFinalReport(regexResp, llmResp)

	scannedColumns := make([]string, 0, len(columDataResp.ColumnData))
	for _, col := range columDataResp.ColumnData {
		scannedColumns = append(scannedColumns, col.Column)
	}

	var saveResp *piidetect_table_activities.SaveTablePiiDetectReportResponse
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 1 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 3,
			},
		}),
		activities.SaveTablePiiDetectReport,
		&piidetect_table_activities.SaveTablePiiDetectReportRequest{
			ParentRunId:    req.ParentExecutionId,
			AccountId:      req.AccountId,
			TableSchema:    req.TableSchema,
			TableName:      req.TableName,
			Report:         report,
			ScannedColumns: scannedColumns,
		},
	).Get(ctx, &saveResp)
	if err != nil {
		return nil, err
	}

	return &TablePiiDetectResponse{
		PiiColumns: report,
		ResultKey:  saveResp.Key,
	}, nil
}

func buildFinalReport(
	regexResp *piidetect_table_activities.DetectPiiRegexResponse,
	llmResp *piidetect_table_activities.DetectPiiLLMResponse,
) map[string]piidetect_table_activities.CombinedPiiDetectReport {
	reportByColumn := make(map[string]piidetect_table_activities.CombinedPiiDetectReport)

	for col, category := range regexResp.PiiColumns {
		reportByColumn[col] = piidetect_table_activities.CombinedPiiDetectReport{
			Regex: &piidetect_table_activities.RegexPiiDetectReport{
				Category: category,
			},
		}
	}

	for col, report := range llmResp.PiiColumns {
		if _, ok := reportByColumn[col]; ok {
			// Create a new report with both regex and LLM data
			existingReport := reportByColumn[col]
			existingReport.LLM = &report
			reportByColumn[col] = existingReport
		} else {
			reportByColumn[col] = piidetect_table_activities.CombinedPiiDetectReport{
				LLM: &report,
			}
		}
	}

	return reportByColumn
}
