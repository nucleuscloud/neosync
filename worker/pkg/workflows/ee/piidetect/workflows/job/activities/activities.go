package piidetect_job_activities

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/internal/connectiondata"
	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"go.temporal.io/sdk/activity"
	tmprl "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
)

type Activities struct {
	jobclient             mgmtv1alpha1connect.JobServiceClient
	connclient            mgmtv1alpha1connect.ConnectionServiceClient
	connectiondatabuilder connectiondata.ConnectionDataBuilder
	tmprlScheduleClient   tmprl.ScheduleClient
}

func New(
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	connectiondatabuilder connectiondata.ConnectionDataBuilder,
	tmprlScheduleClient tmprl.ScheduleClient,
) *Activities {
	return &Activities{
		jobclient:             jobclient,
		connclient:            connclient,
		connectiondatabuilder: connectiondatabuilder,
		tmprlScheduleClient:   tmprlScheduleClient,
	}
}

type GetPiiDetectJobDetailsRequest struct {
	JobId string
}

type GetPiiDetectJobDetailsResponse struct {
	AccountId          string
	PiiDetectConfig    *mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect
	SourceConnectionId string
}

func (a *Activities) GetPiiDetectJobDetails(ctx context.Context, req *GetPiiDetectJobDetailsRequest) (*GetPiiDetectJobDetailsResponse, error) {
	logger := log.With(activity.GetLogger(ctx), "jobId", req.JobId)

	jobResp, err := a.jobclient.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.JobId,
	}))
	if err != nil {
		return nil, err
	}
	job := jobResp.Msg.GetJob()
	logger.Debug("retrieved job")

	switch jt := job.GetJobType().GetJobType().(type) {
	case *mgmtv1alpha1.JobTypeConfig_PiiDetect:
		piiDetectJob := jt.PiiDetect
		if piiDetectJob == nil {
			return nil, fmt.Errorf("pii detect job type config is nil")
		}
		sourceConnectionId, err := shared.GetJobSourceConnectionId(job.GetSource())
		if err != nil {
			return nil, fmt.Errorf("unable to get job source connection id: %w", err)
		}
		return &GetPiiDetectJobDetailsResponse{
			AccountId:          job.GetAccountId(),
			PiiDetectConfig:    piiDetectJob,
			SourceConnectionId: sourceConnectionId,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported job type, must be PiiDetect: %v", jt)
	}
}

type GetLastSuccessfulWorkflowIdRequest struct {
	AccountId string
	JobId     string
}

type GetLastSuccessfulWorkflowIdResponse struct {
	WorkflowId *string
}

func (a *Activities) GetLastSuccessfulWorkflowId(ctx context.Context, req *GetLastSuccessfulWorkflowIdRequest) (*GetLastSuccessfulWorkflowIdResponse, error) {
	logger := log.With(activity.GetLogger(ctx), "accountId", req.AccountId, "jobId", req.JobId)
	workflowIds, err := getRecentRunsFromHandle(ctx, a.tmprlScheduleClient.GetHandle(ctx, req.JobId))
	if err != nil {
		logger.Error("unable to get recent runs from handle", "error", err)
		return &GetLastSuccessfulWorkflowIdResponse{WorkflowId: nil}, err
	}
	logger.Debug("retrieved workflow ids", "workflowIds", workflowIds)

	lastSuccessfulRun, err := a.getMostRecentSuccessfulRun(ctx, req.AccountId, req.JobId, workflowIds, logger)
	if err != nil {
		logger.Error("unable to get most recent successful run", "error", err)
		return &GetLastSuccessfulWorkflowIdResponse{WorkflowId: nil}, err
	}

	logger.Debug("retrieved last workflow id", "workflowId", lastSuccessfulRun)

	return &GetLastSuccessfulWorkflowIdResponse{WorkflowId: &lastSuccessfulRun}, nil
}

func (a *Activities) getMostRecentSuccessfulRun(ctx context.Context, accountId, jobId string, workflowIds []string, logger log.Logger) (string, error) {
	for _, workflowId := range workflowIds {
		jobReport, found, err := a.getJobPiiDetectReport(ctx, accountId, workflowId, jobId)
		if err != nil {
			return "", fmt.Errorf("unable to get job pii detect report: %w", err)
		}
		if found && len(jobReport.SuccessfulTableReports) > 0 {
			return workflowId, nil
		} else {
			logger.Debug("run context does not contain successful table reports", "workflowId", workflowId)
		}
	}
	return "", fmt.Errorf("no successful run found")
}

func getRecentRunsFromHandle(ctx context.Context, handle tmprl.ScheduleHandle) ([]string, error) {
	description, err := handle.Describe(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to describe schedule when retrieving recent runs: %w", err)
	}
	recentRuns := description.Info.RecentActions
	// Sort recentRuns by ActualTime in descending order (most recent first)
	sort.Slice(recentRuns, func(i, j int) bool {
		return recentRuns[i].ActualTime.After(recentRuns[j].ActualTime)
	})
	workflowIds := make([]string, len(recentRuns))
	for i, run := range recentRuns {
		workflowIds[i] = run.StartWorkflowResult.WorkflowID
	}
	return workflowIds, nil
}

type GetTablesToPiiScanRequest struct {
	AccountId string
	JobId     string

	SourceConnectionId string
	Filter             *mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter

	IncrementalConfig *GetIncrementalTablesConfig
}

type GetIncrementalTablesConfig struct {
	LastWorkflowId string
}

type TablePatterns struct {
	// Match entire schemas
	Schemas []string
	// Match specific tables within schemas
	Tables []TableIdentifier
}

type TableIdentifier struct {
	Schema string
	Table  string
}

type TableIdentifierWithFingerprint struct {
	TableIdentifier
	Fingerprint string
}

type GetTablesToPiiScanResponse struct {
	Tables          []TableIdentifierWithFingerprint
	PreviousReports []*TableReport
}

func (a *Activities) GetTablesToPiiScan(ctx context.Context, req *GetTablesToPiiScanRequest) (*GetTablesToPiiScanResponse, error) {
	logger := log.With(activity.GetLogger(ctx), "sourceConnectionId", req.SourceConnectionId)
	slogger := temporallogger.NewSlogger(logger)

	logger.Debug("getting all tables from connection")

	allTables, err := a.getAllTablesFromConnection(ctx, req.SourceConnectionId, slogger)
	if err != nil {
		return nil, fmt.Errorf("unable to get all tables from connection: %w", err)
	}

	logger.Debug("filtering tables")

	filteredTables := a.getFilteredTables(allTables, req.Filter)

	var previousReports map[TableIdentifier]*TableReport
	if req.IncrementalConfig != nil {
		logger.Debug("getting tables from previous run to further filter tables")
		tableReports, err := a.getTableReportsFromPreviousRun(ctx, req.AccountId, req.IncrementalConfig.LastWorkflowId, req.JobId, logger)
		if err != nil {
			return nil, fmt.Errorf("unable to get tables from previous run: %w", err)
		}
		previousReports = tableReports
		oldTableCount := len(filteredTables)
		filteredTables = filterTablesByFingerprint(filteredTables, getFingerprintsFromReports(tableReports))
		newTableCount := len(filteredTables)
		logger.Debug("filtered tables in incremental scan", "oldTableCount", oldTableCount, "newTableCount", newTableCount)
		// might need to return unfiltered and filtered if we want to store the fingerprints in the job report
	}

	previousReportsArray := make([]*TableReport, 0, len(previousReports))
	for _, report := range previousReports {
		previousReportsArray = append(previousReportsArray, report)
	}

	return &GetTablesToPiiScanResponse{Tables: filteredTables, PreviousReports: previousReportsArray}, nil
}

func getFingerprintsFromReports(reports map[TableIdentifier]*TableReport) map[TableIdentifier]string {
	fingerprints := make(map[TableIdentifier]string)
	for identifier, report := range reports {
		fingerprints[identifier] = report.ScanFingerprint
	}
	return fingerprints
}

func filterTablesByFingerprint(tables []TableIdentifierWithFingerprint, fingerprints map[TableIdentifier]string) []TableIdentifierWithFingerprint {
	filteredTables := []TableIdentifierWithFingerprint{}
	for _, table := range tables {
		fingerprint, ok := fingerprints[table.TableIdentifier]
		// if the table is not in the fingerprints map, then it is a new table
		if !ok {
			filteredTables = append(filteredTables, table)
			continue
		}
		// if fingerprint is not the same, then the table has changed
		if fingerprint != table.Fingerprint {
			filteredTables = append(filteredTables, table)
		}
	}
	return filteredTables
}

func (a *Activities) getTableReportsFromPreviousRun(ctx context.Context, accountId, workflowId, jobId string, logger log.Logger) (map[TableIdentifier]*TableReport, error) {
	runCtx, found, err := a.getJobPiiDetectReport(ctx, accountId, workflowId, jobId)
	if err != nil {
		return nil, fmt.Errorf("unable to get job pii detect report: %w", err)
	}
	if !found {
		return nil, nil
	}
	successfulTables := runCtx.SuccessfulTableReports
	logger.Debug("found successful tables", "tables", successfulTables)

	// errgrp, errctx := errgroup.WithContext(ctx)
	// errgrp.SetLimit(5)

	tableReports := map[TableIdentifier]*TableReport{}
	for _, tableReport := range successfulTables {
		tableReports[TableIdentifier{Schema: tableReport.TableSchema, Table: tableReport.TableName}] = tableReport
	}
	// mu := sync.Mutex{}
	// for _, table := range successfulTables {
	// 	table := table
	// 	errgrp.Go(func() error {
	// 		tableReport, found, err := a.getTableReport(errctx, table)
	// 		if err != nil {
	// 			return fmt.Errorf("unable to get table report: %w", err)
	// 		}
	// 		if !found {
	// 			return nil
	// 		}
	// 		mu.Lock()
	// 		tableFingerprints[TableIdentifier{Schema: tableReport.TableSchema, Table: tableReport.TableName}] = getTableColumnFingerprint(tableReport.TableSchema, tableReport.TableName, tableReport.ScannedColumns)
	// 		mu.Unlock()
	// 		return nil
	// 	})
	// }
	// err = errgrp.Wait()
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to get table reports: %w", err)
	// }

	return tableReports, nil
}

func getTableColumnFingerprint(tableSchema, tableName string, columns []string) string {
	// Generate a hash from the schema, table, and columns
	h := sha256.New()

	// Write schema and table name to hash
	h.Write([]byte(tableSchema))
	h.Write([]byte(tableName))

	// Sort column names for consistent hashing
	columnNames := make([]string, 0, len(columns))
	columnNames = append(columnNames, columns...)
	sort.Strings(columnNames)

	// Write each column name to the hash
	for _, col := range columnNames {
		h.Write([]byte(col))
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (a *Activities) getJobPiiDetectReport(ctx context.Context, accountId, workflowId, jobId string) (*JobPiiDetectReport, bool, error) {
	runCtxResp, err := a.jobclient.GetRunContext(ctx, connect.NewRequest(&mgmtv1alpha1.GetRunContextRequest{
		Id: &mgmtv1alpha1.RunContextKey{
			AccountId:  accountId,
			JobRunId:   workflowId,
			ExternalId: fmt.Sprintf("%s--job-pii-report", jobId),
		},
	}))
	if err != nil {
		if isConnectNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("unable to get run context for job pii detect report: %w", err)
	}
	runCtxBytes := runCtxResp.Msg.GetValue()
	var runCtx JobPiiDetectReport
	err = json.Unmarshal(runCtxBytes, &runCtx)
	if err != nil {
		return nil, false, fmt.Errorf("unable to unmarshal run context: %w", err)
	}
	return &runCtx, true, nil
}

func isConnectNotFoundError(err error) bool {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) && connectErr.Code() == connect.CodeNotFound {
		return true
	}
	return false
}

func (a *Activities) getFilteredTables(allTables []TableIdentifierWithFingerprint, filter *mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter) []TableIdentifierWithFingerprint {
	if filter == nil {
		return allTables
	}

	var filteredTables []TableIdentifierWithFingerprint

	switch filter.GetMode().(type) {
	case *mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter_IncludeAll:
		return allTables
	case *mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter_Exclude:
		patterns := filter.GetExclude()
		// Create lookup maps for quick checking
		excludedSchemas := makeStringSet(patterns.GetSchemas())
		excludedTables := makeTableSet(convertProtoTablesToTableIdentifiers(patterns.GetTables()))

		for _, table := range allTables {
			// Skip if schema is excluded or specific table is excluded
			if !excludedSchemas[table.Schema] && !excludedTables[table.TableIdentifier] {
				filteredTables = append(filteredTables, table)
			}
		}

	case *mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter_Include:
		patterns := filter.GetInclude()
		includedSchemas := makeStringSet(patterns.GetSchemas())
		includedTables := makeTableSet(convertProtoTablesToTableIdentifiers(patterns.GetTables()))

		for _, table := range allTables {
			// Include if schema is included or specific table is included
			if includedSchemas[table.Schema] || includedTables[table.TableIdentifier] {
				filteredTables = append(filteredTables, table)
			}
		}
	}

	return filteredTables
}

// Helper function to convert proto TableIdentifier to our internal TableIdentifier
func convertProtoTablesToTableIdentifiers(protoTables []*mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableIdentifier) []TableIdentifier {
	tables := make([]TableIdentifier, len(protoTables))
	for i, pt := range protoTables {
		tables[i] = TableIdentifier{
			Schema: pt.Schema,
			Table:  pt.Table,
		}
	}
	return tables
}

func (a *Activities) getAllTablesFromConnection(ctx context.Context, sourceConnectionId string, logger *slog.Logger) ([]TableIdentifierWithFingerprint, error) {
	connResp, err := a.connclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: sourceConnectionId,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to get connection: %w", err)
	}
	connection := connResp.Msg.GetConnection()

	connectionData, err := a.connectiondatabuilder.NewDataConnection(logger, connection)
	if err != nil {
		return nil, fmt.Errorf("unable to build connection data: %w", err)
	}
	dbColumnSchemas, err := connectionData.GetSchema(ctx, &mgmtv1alpha1.ConnectionSchemaConfig{})
	if err != nil {
		return nil, fmt.Errorf("unable to get all tables from connection: %w", err)
	}

	dbCols := map[TableIdentifier][]string{}
	for _, dbColSchema := range dbColumnSchemas {
		identifier := TableIdentifier{
			Schema: dbColSchema.Schema,
			Table:  dbColSchema.Table,
		}
		dbCols[identifier] = append(dbCols[identifier], dbColSchema.Column)
	}

	tableFingerprints := make([]TableIdentifierWithFingerprint, len(dbCols))
	for identifier, columns := range dbCols {
		tableFingerprints = append(tableFingerprints, TableIdentifierWithFingerprint{
			TableIdentifier: identifier,
			Fingerprint:     getTableColumnFingerprint(identifier.Schema, identifier.Table, columns),
		})
	}

	return tableFingerprints, nil
}

// Helper functions for set operations
func makeStringSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}

func makeTableSet(tables []TableIdentifier) map[TableIdentifier]bool {
	set := make(map[TableIdentifier]bool, len(tables))
	for _, table := range tables {
		set[table] = true
	}
	return set
}

// Make TableIdentifier comparable
func (t TableIdentifier) Equals(other TableIdentifier) bool {
	return t.Schema == other.Schema && t.Table == other.Table
}

type SaveJobPiiDetectReportRequest struct {
	AccountId string
	JobId     string
	Report    *JobPiiDetectReport
}

type SaveJobPiiDetectReportResponse struct {
	Key *mgmtv1alpha1.RunContextKey
}

type JobPiiDetectReport struct {
	SuccessfulTableReports []*TableReport `json:"successfulTableReports"`
}

type TableReport struct {
	TableSchema     string                      `json:"tableSchema"`
	TableName       string                      `json:"tableName"`
	ReportKey       *mgmtv1alpha1.RunContextKey `json:"reportKey"`
	ScanFingerprint string                      `json:"scanFingerprint"`
}

func (a *Activities) SaveJobPiiDetectReport(ctx context.Context, req *SaveJobPiiDetectReportRequest) (*SaveJobPiiDetectReportResponse, error) {
	info := activity.GetInfo(ctx)
	jobRunId := info.WorkflowExecution.ID

	key := &mgmtv1alpha1.RunContextKey{
		AccountId:  req.AccountId,
		JobRunId:   jobRunId,
		ExternalId: fmt.Sprintf("%s--job-pii-report", req.JobId),
	}

	reportBytes, err := json.Marshal(req.Report)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal report: %w", err)
	}

	_, err = a.jobclient.SetRunContext(ctx, connect.NewRequest(&mgmtv1alpha1.SetRunContextRequest{
		Id:    key,
		Value: reportBytes,
	}))
	if err != nil {
		return nil, err
	}
	return &SaveJobPiiDetectReportResponse{Key: key}, nil
}
