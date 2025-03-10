package piidetect_job_activities

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/internal/connectiondata"
	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type Activities struct {
	jobclient             mgmtv1alpha1connect.JobServiceClient
	connclient            mgmtv1alpha1connect.ConnectionServiceClient
	connectiondatabuilder connectiondata.ConnectionDataBuilder
}

func New(
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	connectiondatabuilder connectiondata.ConnectionDataBuilder,
) *Activities {
	return &Activities{
		jobclient:             jobclient,
		connclient:            connclient,
		connectiondatabuilder: connectiondatabuilder,
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

type GetTablesToPiiScanRequest struct {
	SourceConnectionId string
	Filter             *mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter
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

type GetTablesToPiiScanResponse struct {
	Tables []TableIdentifier
}

func (a *Activities) GetTablesToPiiScan(ctx context.Context, req *GetTablesToPiiScanRequest) (*GetTablesToPiiScanResponse, error) {
	logger := log.With(activity.GetLogger(ctx), "sourceConnectionId", req.SourceConnectionId)
	slogger := temporallogger.NewSlogger(logger)

	allTables, err := a.getAllTablesFromConnection(ctx, req.SourceConnectionId, slogger)
	if err != nil {
		return nil, err
	}

	filteredTables := a.getFilteredTables(allTables, req.Filter)

	return &GetTablesToPiiScanResponse{Tables: filteredTables}, nil
}

func (a *Activities) getFilteredTables(allTables []TableIdentifier, filter *mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter) []TableIdentifier {
	if filter == nil {
		return allTables
	}

	var filteredTables []TableIdentifier

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
			if !excludedSchemas[table.Schema] && !excludedTables[table] {
				filteredTables = append(filteredTables, table)
			}
		}

	case *mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter_Include:
		patterns := filter.GetInclude()
		includedSchemas := makeStringSet(patterns.GetSchemas())
		includedTables := makeTableSet(convertProtoTablesToTableIdentifiers(patterns.GetTables()))

		for _, table := range allTables {
			// Include if schema is included or specific table is included
			if includedSchemas[table.Schema] || includedTables[table] {
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

func (a *Activities) getAllTablesFromConnection(ctx context.Context, sourceConnectionId string, logger *slog.Logger) ([]TableIdentifier, error) {
	connResp, err := a.connclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: sourceConnectionId,
	}))
	if err != nil {
		return nil, err
	}
	connection := connResp.Msg.GetConnection()

	connectionData, err := a.connectiondatabuilder.NewDataConnection(logger, connection)
	if err != nil {
		return nil, err
	}
	tables, err := connectionData.GetAllTables(ctx)
	if err != nil {
		return nil, err
	}

	tableIdentifiers := make([]TableIdentifier, len(tables))
	for i, table := range tables {
		tableIdentifiers[i] = TableIdentifier{
			Schema: table.Schema,
			Table:  table.Table,
		}
	}
	return tableIdentifiers, nil
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
	SuccessfulTableKeys []*mgmtv1alpha1.RunContextKey `json:"successfulTableKeys"`
}

func (a *Activities) SaveJobPiiDetectReport(ctx context.Context, req *SaveJobPiiDetectReportRequest) (*SaveJobPiiDetectReportResponse, error) {
	info := activity.GetInfo(ctx)
	jobRunId := info.WorkflowExecution.ID

	key := &mgmtv1alpha1.RunContextKey{
		AccountId:  req.AccountId,
		JobRunId:   jobRunId,
		ExternalId: fmt.Sprintf("%s--pii-report", req.JobId),
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
