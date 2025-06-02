package piidetect_job_workflow

import (
	"fmt"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	piidetect_job_activities "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/job/activities"
	piidetect_table_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/table"
	workflow_shared "github.com/nucleuscloud/neosync/worker/pkg/workflows/shared"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Workflow struct {
	eelicense license.EEInterface
}

func New(eelicense license.EEInterface) *Workflow {
	return &Workflow{
		eelicense: eelicense,
	}
}

type PiiDetectRequest struct {
	JobId string
}

type PiiDetectResponse struct {
	ReportKey *mgmtv1alpha1.RunContextKey
}

func (w *Workflow) JobPiiDetect(
	ctx workflow.Context,
	req *PiiDetectRequest,
) (*PiiDetectResponse, error) {
	logger := log.With(
		workflow.GetLogger(ctx),
		"jobId", req.JobId,
	)

	if !w.eelicense.IsValid() {
		logger.Debug("ee license is not valid, skipping pii detect")
		return nil, fmt.Errorf("ee license is not valid, unable to run pii detect")
	}

	logger.Debug("starting PII detection")

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

	return workflow_shared.HandleWorkflowEventLifecycle(
		ctx,
		w.eelicense,
		req.JobId,
		workflow.GetInfo(ctx).WorkflowExecution.ID,
		logger,
		func() (string, error) {
			return jobDetailsResp.AccountId, nil
		},
		func(ctx workflow.Context, logger log.Logger) (*PiiDetectResponse, error) {
			return executeWorkflow(ctx, req, jobDetailsResp, logger, activities)
		},
	)
}

func executeWorkflow(
	ctx workflow.Context,
	req *PiiDetectRequest,
	jobDetailsResp *piidetect_job_activities.GetPiiDetectJobDetailsResponse,
	logger log.Logger,
	activities *piidetect_job_activities.Activities,
) (*PiiDetectResponse, error) {
	var filter *mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter
	if jobDetailsResp != nil && jobDetailsResp.PiiDetectConfig != nil &&
		jobDetailsResp.PiiDetectConfig.TableScanFilter != nil {
		logger.Debug(
			"using table scan filter",
			"filter",
			jobDetailsResp.PiiDetectConfig.TableScanFilter,
		)
		filter = jobDetailsResp.PiiDetectConfig.TableScanFilter
	}

	piiDetectConfig := jobDetailsResp.PiiDetectConfig
	if piiDetectConfig == nil {
		piiDetectConfig = &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect{}
	}

	var incrementalConfig *piidetect_job_activities.GetIncrementalTablesConfig
	if piiDetectConfig.GetIncremental().GetIsEnabled() {
		var lastSuccessfulWorkflowIdResp *piidetect_job_activities.GetLastSuccessfulWorkflowIdResponse
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: 1 * time.Minute,
				RetryPolicy: &temporal.RetryPolicy{
					MaximumAttempts: 3,
				},
			}),
			activities.GetLastSuccessfulWorkflowId,
			&piidetect_job_activities.GetLastSuccessfulWorkflowIdRequest{
				AccountId: jobDetailsResp.AccountId,
				JobId:     req.JobId,
			},
		).Get(ctx, &lastSuccessfulWorkflowIdResp)
		if err != nil {
			return nil, fmt.Errorf("unable to get last successful workflow id: %w", err)
		}
		if lastSuccessfulWorkflowIdResp.WorkflowId != nil {
			logger.Debug(
				"using last successful workflow id",
				"workflowId",
				*lastSuccessfulWorkflowIdResp.WorkflowId,
			)
			incrementalConfig = &piidetect_job_activities.GetIncrementalTablesConfig{
				LastWorkflowId: *lastSuccessfulWorkflowIdResp.WorkflowId,
			}
		} else {
			logger.Debug("no last successful workflow id found, skipping incremental pii detect")
		}
	}

	var tablesToScanResp *piidetect_job_activities.GetTablesToPiiScanResponse
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 1 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 3,
			},
		}),
		activities.GetTablesToPiiScan,
		&piidetect_job_activities.GetTablesToPiiScanRequest{
			AccountId:          jobDetailsResp.AccountId,
			JobId:              req.JobId,
			SourceConnectionId: jobDetailsResp.SourceConnectionId,
			Filter:             filter,
			IncrementalConfig:  incrementalConfig,
		},
	).Get(ctx, &tablesToScanResp)
	if err != nil {
		return nil, err
	}

	report, err := orchestrateTables(
		ctx,
		jobDetailsResp.AccountId,
		tablesToScanResp,
		req.JobId,
		jobDetailsResp.SourceConnectionId,
		piiDetectConfig.GetDataSampling().GetIsEnabled(),
		piiDetectConfig.GetUserPrompt(),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to orchestrate tables: %w", err)
	}

	report = buildFinalReport(tablesToScanResp.PreviousReports, report)

	logger.Debug("saving job pii detect report")

	var saveResp *piidetect_job_activities.SaveJobPiiDetectReportResponse
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 1 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 3,
			},
		}),
		activities.SaveJobPiiDetectReport,
		&piidetect_job_activities.SaveJobPiiDetectReportRequest{
			AccountId: jobDetailsResp.AccountId,
			JobId:     req.JobId,
			Report:    report,
		},
	).Get(ctx, &saveResp)
	if err != nil {
		return nil, fmt.Errorf("unable to save job pii detect report: %w", err)
	}

	logger.Info("PII detection completed")
	return &PiiDetectResponse{
		ReportKey: saveResp.Key,
	}, nil
}

// buildFinalReport builds the final report by combining the previous reports with the new report
func buildFinalReport(
	previousReports []*piidetect_job_activities.TableReport,
	newReport *piidetect_job_activities.JobPiiDetectReport,
) *piidetect_job_activities.JobPiiDetectReport {
	fullSuccessfulTableReports := map[piidetect_job_activities.TableIdentifier]*piidetect_job_activities.TableReport{}
	for _, report := range newReport.SuccessfulTableReports {
		fullSuccessfulTableReports[report.ToTableIdentifier()] = report
	}

	for _, report := range previousReports {
		if _, ok := fullSuccessfulTableReports[report.ToTableIdentifier()]; !ok {
			fullSuccessfulTableReports[report.ToTableIdentifier()] = report
		}
	}

	successfulTableReports := make(
		[]*piidetect_job_activities.TableReport,
		0,
		len(fullSuccessfulTableReports),
	)
	for _, report := range fullSuccessfulTableReports {
		successfulTableReports = append(successfulTableReports, report)
	}
	return &piidetect_job_activities.JobPiiDetectReport{
		SuccessfulTableReports: successfulTableReports,
	}
}

func orchestrateTables(
	ctx workflow.Context,
	accountId string,
	tablesToScanResp *piidetect_job_activities.GetTablesToPiiScanResponse,
	jobId string,
	connectionId string,
	shouldSampleData bool,
	userPrompt string,
	logger log.Logger,
) (*piidetect_job_activities.JobPiiDetectReport, error) {
	maxConcurrency := getTablePiiDetectMaxConcurrency()

	tableWf := piidetect_table_workflow.New()
	wfInfo := workflow.GetInfo(ctx)

	tableResultKeys := []*piidetect_job_activities.TableReport{}
	mu := workflow.NewMutex(ctx)

	previousReportsMap := make(
		map[piidetect_job_activities.TableIdentifier]*piidetect_job_activities.TableReport,
	)
	for _, report := range tablesToScanResp.PreviousReports {
		previousReportsMap[piidetect_job_activities.TableIdentifier{Schema: report.TableSchema, Table: report.TableName}] = report
	}

	logger.Debug("starting table processing")
	logger.Debug("total tables to process", "count", len(tablesToScanResp.Tables))

	// Create channels for coordination
	type tableWork struct {
		table          piidetect_job_activities.TableIdentifierWithFingerprint
		previousReport *piidetect_job_activities.TableReport
	}

	// Use a buffered channel as a work queue
	workQueue := workflow.NewBufferedChannel(ctx, len(tablesToScanResp.Tables))

	// Queue all work items
	for _, table := range tablesToScanResp.Tables {
		previousReport := previousReportsMap[table.TableIdentifier]
		workQueue.Send(ctx, tableWork{table: table, previousReport: previousReport})
		logger.Debug("queued table", "schema", table.Schema, "table", table.Table)
	}
	workQueue.Close()

	// Channel to track completion
	completionChannel := workflow.NewChannel(ctx)
	activeWorkers := 0

	// Start worker goroutines
	for i := 0; i < maxConcurrency; i++ {
		activeWorkers++
		workflow.Go(ctx, func(ctx workflow.Context) {
			defer func() {
				completionChannel.Send(ctx, true)
			}()

			for {
				var work tableWork
				more := workQueue.Receive(ctx, &work)
				if !more {
					logger.Debug("worker exiting, no more work", "workerIndex", i)
					return // Channel closed, no more work
				}

				logger.Debug("worker processing table", "workerIndex", i, "table", work.table.Table, "schema", work.table.Schema)

				var previousResultsKey *mgmtv1alpha1.RunContextKey
				if work.previousReport != nil {
					previousResultsKey = work.previousReport.ReportKey
				}

				// Execute child workflow synchronously within this goroutine
				var wfResult piidetect_table_workflow.TablePiiDetectResponse
				err := workflow.ExecuteChildWorkflow(
					workflow.WithChildOptions(
						ctx,
						workflow.ChildWorkflowOptions{
							WorkflowID: workflow_shared.BuildChildWorkflowId(
								wfInfo.WorkflowExecution.ID,
								fmt.Sprintf("%s.%s", work.table.Schema, work.table.Table),
								workflow.Now(ctx),
							),
							RetryPolicy: &temporal.RetryPolicy{
								MaximumAttempts: 1,
							},
							WorkflowRunTimeout: 5 * time.Minute,
						}),
					tableWf.TablePiiDetect,
					&piidetect_table_workflow.TablePiiDetectRequest{
						AccountId:          accountId,
						JobId:              jobId,
						ConnectionId:       connectionId,
						TableSchema:        work.table.Schema,
						TableName:          work.table.Table,
						ParentExecutionId:  &wfInfo.WorkflowExecution.ID,
						ShouldSampleData:   shouldSampleData,
						UserPrompt:         userPrompt,
						PreviousResultsKey: previousResultsKey,
					},
				).Get(ctx, &wfResult)

				if err != nil {
					logger.Error("child workflow did not complete",
						"table", work.table.Table,
						"schema", work.table.Schema,
						"err", err)
					continue
				}

				logger.Debug(
					"table pii detect completed",
					"table", work.table.Table,
					"schema", work.table.Schema,
				)

				// Store result
				err = mu.Lock(ctx)
				if err != nil {
					logger.Error(
						"unable to lock mutex after table pii detect completed",
						"err", err,
					)
					continue
				}
				tableResultKeys = append(tableResultKeys, &piidetect_job_activities.TableReport{
					TableSchema:     work.table.Schema,
					TableName:       work.table.Table,
					ScanFingerprint: work.table.Fingerprint,
					ReportKey:       wfResult.ResultKey,
				})
				mu.Unlock()
			}
		})
	}

	// Wait for all workers to complete
	for i := 0; i < activeWorkers; i++ {
		var completed bool
		completionChannel.Receive(ctx, &completed)
		logger.Debug("worker completed", "remaining", activeWorkers-i-1)
	}

	logger.Debug("all tables processed", "total_processed", len(tableResultKeys))
	return &piidetect_job_activities.JobPiiDetectReport{
		SuccessfulTableReports: tableResultKeys,
	}, nil
}

func getTablePiiDetectMaxConcurrency() int {
	maxConcurrency := viper.GetInt("TABLE_PII_DETECT_MAX_CONCURRENCY")
	if maxConcurrency <= 0 {
		return 3 // default max concurrency
	}
	return maxConcurrency
}
