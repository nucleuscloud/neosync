package piidetect_job_activities

import (
	"context"
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
	Filter             *TableScanFilter
}

type TableScanFilter struct {
	Mode     FilterMode
	Includes TablePatterns
	Excludes TablePatterns
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

type FilterMode string

const (
	FilterModeExclude FilterMode = "exclude"
	FilterModeInclude FilterMode = "include"
)

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

func (a *Activities) getFilteredTables(allTables []TableIdentifier, filter *TableScanFilter) []TableIdentifier {
	if filter == nil {
		return allTables
	}

	var filteredTables []TableIdentifier
	switch filter.Mode {
	case FilterModeExclude:
		// Create lookup maps for quick checking
		excludedSchemas := makeStringSet(filter.Excludes.Schemas)
		excludedTables := makeTableSet(filter.Excludes.Tables)

		for _, table := range allTables {
			// Skip if schema is excluded or specific table is excluded
			if !excludedSchemas[table.Schema] && !excludedTables[table] {
				filteredTables = append(filteredTables, table)
			}
		}

	case FilterModeInclude:
		includedSchemas := makeStringSet(filter.Includes.Schemas)
		includedTables := makeTableSet(filter.Includes.Tables)

		for _, table := range allTables {
			// Include if schema is included or specific table is included
			if includedSchemas[table.Schema] || includedTables[table] {
				filteredTables = append(filteredTables, table)
			}
		}
	}

	return filteredTables
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
