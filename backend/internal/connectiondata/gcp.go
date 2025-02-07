package connectiondata

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	neosync_gcp "github.com/nucleuscloud/neosync/backend/internal/gcp"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
)

type GcpConnectionDataService struct {
	logger     *slog.Logger
	gcpmanager neosync_gcp.ManagerInterface
	connection *mgmtv1alpha1.Connection
	connconfig *mgmtv1alpha1.GcpCloudStorageConnectionConfig
}

func NewGcpConnectionDataService(
	logger *slog.Logger,
	gcpmanager neosync_gcp.ManagerInterface,
	connection *mgmtv1alpha1.Connection,
) *GcpConnectionDataService {
	return &GcpConnectionDataService{
		logger:     logger,
		gcpmanager: gcpmanager,
		connection: connection,
		connconfig: connection.GetConnectionConfig().GetGcpCloudstorageConfig(),
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
	// Build a base prefix for listing job run directories.
	var gcsPathPieces []string
	if pathPrefix != nil && *pathPrefix != "" {
		trimmed := strings.Trim(*pathPrefix, "/")
		gcsPathPieces = append(gcsPathPieces, trimmed)
	}
	gcsPathPieces = append(gcsPathPieces, "workflows", jobId)
	basePrefix := strings.Join(gcsPathPieces, "/")

	prefixes, err := client.ListObjectPrefixes(ctx, bucket, basePrefix, "/")
	if err != nil {
		return "", fmt.Errorf("unable to list job run directories from GCS: %w", err)
	}

	// Extract run IDs from the directory names.
	runIDs := make([]string, 0, len(prefixes))
	for _, cp := range prefixes {
		trimmedPrefix := strings.TrimSuffix(cp, "/")
		parts := strings.Split(trimmedPrefix, "/")
		if len(parts) > 0 {
			runIDs = append(runIDs, parts[len(parts)-1])
		}
	}

	if len(runIDs) == 0 {
		return "", fmt.Errorf("unable to find any job runs for job: %s", jobId)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(runIDs)))

	for _, runId := range runIDs {
		prefix := neosync_gcp.GetWorkflowActivityPrefix(runId, pathPrefix)
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
