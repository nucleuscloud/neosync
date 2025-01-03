package connectiondata

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	neosync_gcp "github.com/nucleuscloud/neosync/backend/internal/gcp"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
)

type GcpConnectionDataService struct {
	logger     *slog.Logger
	gcpmanager neosync_gcp.ManagerInterface
	connection *mgmtv1alpha1.Connection
	connconfig *mgmtv1alpha1.GcpCloudStorageConnectionConfig
	jobservice mgmtv1alpha1connect.JobServiceHandler
}

func NewGcpConnectionDataService(
	logger *slog.Logger,
	gcpmanager neosync_gcp.ManagerInterface,
	connection *mgmtv1alpha1.Connection,

	jobservice mgmtv1alpha1connect.JobServiceHandler,
) *GcpConnectionDataService {
	return &GcpConnectionDataService{
		logger:     logger,
		gcpmanager: gcpmanager,
		connection: connection,
		connconfig: connection.GetConnectionConfig().GetGcpCloudstorageConfig(),

		jobservice: jobservice,
	}
}

func (s *GcpConnectionDataService) StreamData(
	ctx context.Context,
	stream *connect.ServerStream[mgmtv1alpha1.GetConnectionDataStreamResponse],
	config *mgmtv1alpha1.ConnectionStreamConfig,
	schema, table string,
) error {
	gcpStreamCfg := config.GetGcpCloudstorageConfig()
	if gcpStreamCfg == nil {
		return nucleuserrors.NewBadRequest("must provide non-nil gcp cloud storage config in request")
	}
	gcpclient, err := s.gcpmanager.GetClient(ctx, s.logger)
	if err != nil {
		return fmt.Errorf("unable to init gcp storage client: %w", err)
	}

	var jobRunId string
	switch id := gcpStreamCfg.Id.(type) {
	case *mgmtv1alpha1.GcpCloudStorageStreamConfig_JobRunId:
		jobRunId = id.JobRunId
	case *mgmtv1alpha1.GcpCloudStorageStreamConfig_JobId:
		runId, err := s.getLatestJobRunFromGcs(ctx, gcpclient, id.JobId, s.connconfig.GetBucket(), s.connconfig.PathPrefix)
		if err != nil {
			return err
		}
		jobRunId = runId
	default:
		return nucleuserrors.NewNotImplemented(fmt.Sprintf("unsupported GCP Cloud Storage config id: %T", id))
	}

	onRecord := func(record map[string][]byte) error {
		var rowbytes bytes.Buffer
		enc := gob.NewEncoder(&rowbytes)
		if err := enc.Encode(record); err != nil {
			return fmt.Errorf("unable to encode gcp record using gob: %w", err)
		}
		return stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{RowBytes: rowbytes.Bytes()})
	}
	tablePath := neosync_gcp.GetWorkflowActivityDataPrefix(jobRunId, sqlmanager_shared.BuildTable(schema, table), s.connconfig.PathPrefix)
	err = gcpclient.GetRecordStreamFromPrefix(ctx, s.connconfig.GetBucket(), tablePath, onRecord)
	if err != nil {
		return fmt.Errorf("unable to finish sending record stream: %w", err)
	}
	return nil
}

func (s *GcpConnectionDataService) GetSchema(
	ctx context.Context,
	config *mgmtv1alpha1.ConnectionSchemaConfig,
) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	gcpCfg := config.GetGcpCloudstorageConfig()
	if gcpCfg == nil {
		return nil, nucleuserrors.NewBadRequest("must provide gcp cloud storage config")
	}

	gcpclient, err := s.gcpmanager.GetClient(ctx, s.logger)
	if err != nil {
		return nil, fmt.Errorf("unable to init gcp storage client: %w", err)
	}

	var jobRunId string
	switch id := gcpCfg.Id.(type) {
	case *mgmtv1alpha1.GcpCloudStorageSchemaConfig_JobRunId:
		jobRunId = id.JobRunId
	case *mgmtv1alpha1.GcpCloudStorageSchemaConfig_JobId:
		runId, err := s.getLatestJobRunFromGcs(ctx, gcpclient, id.JobId, s.connconfig.GetBucket(), s.connconfig.PathPrefix)
		if err != nil {
			return nil, err
		}
		jobRunId = runId
	default:
		return nil, nucleuserrors.NewNotImplemented(fmt.Sprintf("unsupported GCP Cloud Storage config id: %T", id))
	}

	schemas, err := gcpclient.GetDbSchemaFromPrefix(
		ctx,
		s.connconfig.GetBucket(), neosync_gcp.GetWorkflowActivityPrefix(jobRunId, s.connconfig.PathPrefix),
	)
	if err != nil {
		return nil, fmt.Errorf("uanble to retrieve db schema from gcs: %w", err)
	}
	return schemas, nil
}

func (s *GcpConnectionDataService) getLatestJobRunFromGcs(
	ctx context.Context,
	client neosync_gcp.ClientInterface,
	jobId string,
	bucket string,
	pathPrefix *string,
) (string, error) {
	// TODO: this should find lastest run from GCP not jobservice
	jobRunsResp, err := s.jobservice.GetJobRecentRuns(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRecentRunsRequest{
		JobId: jobId,
	}))
	if err != nil {
		return "", err
	}
	jobRuns := jobRunsResp.Msg.GetRecentRuns()
	for i := len(jobRuns) - 1; i >= 0; i-- {
		runId := jobRuns[i].GetJobRunId()
		prefix := neosync_gcp.GetWorkflowActivityPrefix(
			runId,
			pathPrefix,
		)
		ok, err := client.DoesPrefixContainTables(ctx, bucket, prefix)
		if err != nil {
			return "", fmt.Errorf("unable to check if prefix contains tables: %w", err)
		}
		if ok {
			return runId, nil
		}
	}
	return "", fmt.Errorf("unable to find latest job run for job: %s", jobId)
}

func (s *GcpConnectionDataService) GetInitStatements(
	ctx context.Context,
	options *mgmtv1alpha1.InitStatementOptions,
) (*mgmtv1alpha1.GetConnectionInitStatementsResponse, error) {
	return nil, errors.ErrUnsupported
}

func (s *GcpConnectionDataService) GetTableConstraints(
	ctx context.Context,
) (*mgmtv1alpha1.GetConnectionTableConstraintsResponse, error) {
	return nil, errors.ErrUnsupported
}

func (s *GcpConnectionDataService) GetTableSchema(ctx context.Context, schema, table string) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	return nil, errors.ErrUnsupported
}

func (s *GcpConnectionDataService) GetTableRowCount(ctx context.Context, schema, table string, whereClause *string) (int64, error) {
	return 0, errors.ErrUnsupported
}
