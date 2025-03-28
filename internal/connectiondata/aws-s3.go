package connectiondata

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	aws_manager "github.com/nucleuscloud/neosync/internal/aws"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
)

type AwsS3ConnectionDataService struct {
	logger              *slog.Logger
	awsmanager          aws_manager.NeosyncAwsManagerClient
	neosynctyperegistry neosynctypes.NeosyncTypeRegistry
	connection          *mgmtv1alpha1.Connection
	connconfig          *mgmtv1alpha1.AwsS3ConnectionConfig
}

func NewAwsS3ConnectionDataService(
	logger *slog.Logger,
	awsmanager aws_manager.NeosyncAwsManagerClient,
	neosynctyperegistry neosynctypes.NeosyncTypeRegistry,
	connection *mgmtv1alpha1.Connection,
) *AwsS3ConnectionDataService {
	return &AwsS3ConnectionDataService{
		logger:              logger,
		awsmanager:          awsmanager,
		neosynctyperegistry: neosynctyperegistry,
		connection:          connection,
		connconfig:          connection.GetConnectionConfig().GetAwsS3Config(),
	}
}

func (s *AwsS3ConnectionDataService) GetAllTables(ctx context.Context) ([]TableIdentifier, error) {
	return nil, errors.ErrUnsupported
}

func (s *AwsS3ConnectionDataService) GetAllSchemas(ctx context.Context) ([]string, error) {
	return nil, errors.ErrUnsupported
}

func (s *AwsS3ConnectionDataService) SampleData(
	ctx context.Context,
	stream SampleDataStream,
	schema, table string,
	numRows uint,
) error {
	return errors.ErrUnsupported
}

func (s *AwsS3ConnectionDataService) StreamData(
	ctx context.Context,
	stream *connect.ServerStream[mgmtv1alpha1.GetConnectionDataStreamResponse],
	config *mgmtv1alpha1.ConnectionStreamConfig,
	schema, table string,
) error {
	awsS3StreamCfg := config.GetAwsS3Config()
	if awsS3StreamCfg == nil {
		return nucleuserrors.NewBadRequest("jobId or jobRunId required for AWS S3 connections")
	}

	s3Client, err := s.awsmanager.NewS3Client(ctx, s.connconfig)
	if err != nil {
		return fmt.Errorf("unable to create AWS S3 client: %w", err)
	}
	logger := s.logger
	logger.Debug("created AWS S3 client")

	s3pathpieces := []string{}
	if s.connconfig != nil && s.connconfig.GetPathPrefix() != "" {
		s3pathpieces = append(s3pathpieces, strings.Trim(s.connconfig.GetPathPrefix(), "/"))
	}

	var jobRunId string
	switch id := awsS3StreamCfg.Id.(type) {
	case *mgmtv1alpha1.AwsS3StreamConfig_JobRunId:
		jobRunId = id.JobRunId
	case *mgmtv1alpha1.AwsS3StreamConfig_JobId:
		logger = logger.With("jobId", id.JobId)
		runId, err := s.getLastestJobRunFromAwsS3(ctx, s3Client, id.JobId, s.connconfig.Bucket, s.connconfig.Region, s3pathpieces)
		if err != nil {
			return err
		}
		logger.Debug(fmt.Sprintf("found run id for job in s3: %s", runId))
		jobRunId = runId
	default:
		return nucleuserrors.NewInternalError("unsupported AWS S3 config id")
	}
	logger = logger.With("runId", jobRunId)

	tableName := sqlmanager_shared.BuildTable(schema, table)
	s3pathpieces = append(
		s3pathpieces,
		"workflows",
		jobRunId,
		"activities",
		tableName,
		"data",
	)
	path := strings.Join(s3pathpieces, "/")
	var pageToken *string
	for {
		output, err := s.awsmanager.ListObjectsV2(
			ctx,
			s3Client,
			s.connconfig.Region,
			&s3.ListObjectsV2Input{
				Bucket:            aws.String(s.connconfig.Bucket),
				Prefix:            aws.String(path),
				ContinuationToken: pageToken,
			},
		)
		if err != nil {
			return err
		}
		if output == nil {
			logger.Debug(fmt.Sprintf("0 files found for path: %s", path))
			break
		}
		for _, item := range output.Contents {
			result, err := s.awsmanager.GetObject(
				ctx,
				s3Client,
				s.connconfig.Region,
				&s3.GetObjectInput{
					Bucket: aws.String(s.connconfig.Bucket),
					Key:    aws.String(*item.Key),
				},
			)
			if err != nil {
				return err
			}

			gzr, err := gzip.NewReader(result.Body)
			if err != nil {
				result.Body.Close()
				return fmt.Errorf("error creating gzip reader: %w", err)
			}

			decoder := json.NewDecoder(gzr)
			for {
				var rowData map[string]any
				// Decode the next JSON object
				err = decoder.Decode(&rowData)
				if err != nil && err == io.EOF {
					break // End of file, stop the loop
				} else if err != nil {
					result.Body.Close()
					gzr.Close()
					return err
				}

				for k, v := range rowData {
					newVal, err := s.neosynctyperegistry.Unmarshal(v)
					if err != nil {
						return fmt.Errorf(
							"unable to unmarshal row value using neosync type registry: %w",
							err,
						)
					}
					rowData[k] = newVal
				}

				// Encode the row data using gob
				var rowbytes bytes.Buffer
				enc := gob.NewEncoder(&rowbytes)
				if err := enc.Encode(rowData); err != nil {
					result.Body.Close()
					gzr.Close()
					return fmt.Errorf("unable to encode S3 row data using gob: %w", err)
				}

				if err := stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{RowBytes: rowbytes.Bytes()}); err != nil {
					result.Body.Close()
					gzr.Close()
					return err
				}
			}
			result.Body.Close()
			gzr.Close()
		}
		if *output.IsTruncated {
			pageToken = output.NextContinuationToken
			continue
		}
		break
	}
	return nil
}

func (s *AwsS3ConnectionDataService) GetSchema(
	ctx context.Context,
	config *mgmtv1alpha1.ConnectionSchemaConfig,
) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	awsCfg := config.GetAwsS3Config()
	if config == nil {
		return nil, nucleuserrors.NewBadRequest("jobId or jobRunId required for AWS S3 connections")
	}

	s3Client, err := s.awsmanager.NewS3Client(ctx, s.connconfig)
	if err != nil {
		return nil, err
	}
	s.logger.Debug("created S3 AWS session")

	s3pathpieces := []string{}
	if s.connconfig != nil && s.connconfig.GetPathPrefix() != "" {
		s3pathpieces = append(s3pathpieces, strings.Trim(s.connconfig.GetPathPrefix(), "/"))
	}

	var jobRunId string
	switch id := awsCfg.Id.(type) {
	case *mgmtv1alpha1.AwsS3SchemaConfig_JobRunId:
		jobRunId = id.JobRunId
	case *mgmtv1alpha1.AwsS3SchemaConfig_JobId:
		runId, err := s.getLastestJobRunFromAwsS3(ctx, s3Client, id.JobId, s.connconfig.Bucket, s.connconfig.Region, s3pathpieces)
		if err != nil {
			return nil, err
		}
		jobRunId = runId
	default:
		return nil, nucleuserrors.NewInternalError("unsupported AWS S3 config id")
	}

	s3pathpieces = append(
		s3pathpieces,
		"workflows",
		jobRunId,
		"activities/",
	)
	path := strings.Join(s3pathpieces, "/")

	schemas := []*mgmtv1alpha1.DatabaseColumn{}
	var pageToken *string
	for {
		output, err := s.awsmanager.ListObjectsV2(
			ctx,
			s3Client,
			s.connconfig.Region,
			&s3.ListObjectsV2Input{
				Bucket:            aws.String(s.connconfig.Bucket),
				Prefix:            aws.String(path),
				Delimiter:         aws.String("/"),
				ContinuationToken: pageToken,
			},
		)
		if err != nil {
			return nil, err
		}
		if output == nil {
			break
		}
		for _, cp := range output.CommonPrefixes {
			folders := strings.Split(*cp.Prefix, "activities")
			tableFolder := strings.ReplaceAll(folders[len(folders)-1], "/", "")
			schemaTableList := strings.Split(tableFolder, ".")

			filePath := fmt.Sprintf(
				"%s%s/data",
				path,
				sqlmanager_shared.BuildTable(schemaTableList[0], schemaTableList[1]),
			)
			out, err := s.awsmanager.ListObjectsV2(
				ctx,
				s3Client,
				s.connconfig.Region,
				&s3.ListObjectsV2Input{
					Bucket:  aws.String(s.connconfig.Bucket),
					Prefix:  aws.String(filePath),
					MaxKeys: aws.Int32(1),
				},
			)
			if err != nil {
				return nil, err
			}
			if out == nil {
				s.logger.Warn(
					fmt.Sprintf(
						"AWS S3 table folder missing data folder: %s, continuing..",
						tableFolder,
					),
				)
				continue
			}
			item := out.Contents[0]
			result, err := s.awsmanager.GetObject(
				ctx,
				s3Client,
				s.connconfig.Region,
				&s3.GetObjectInput{
					Bucket: aws.String(s.connconfig.Bucket),
					Key:    aws.String(*item.Key),
				},
			)
			if err != nil {
				return nil, err
			}
			if result.ContentLength == nil || *result.ContentLength == 0 {
				s.logger.Warn(
					fmt.Sprintf(
						"empty AWS S3 data folder for table: %s, continuing...",
						tableFolder,
					),
				)
				continue
			}

			gzr, err := gzip.NewReader(result.Body)
			if err != nil {
				result.Body.Close()
				return nil, fmt.Errorf("error creating gzip reader: %w", err)
			}

			decoder := json.NewDecoder(gzr)
			for {
				var data map[string]any
				// Decode the next JSON object
				err = decoder.Decode(&data)
				if err != nil && err == io.EOF {
					break // End of file, stop the loop
				} else if err != nil {
					result.Body.Close()
					gzr.Close()
					return nil, err
				}

				for key := range data {
					schemas = append(schemas, &mgmtv1alpha1.DatabaseColumn{
						Schema: schemaTableList[0],
						Table:  schemaTableList[1],
						Column: key,
					})
				}
				break // Only care about first record
			}
			result.Body.Close()
			gzr.Close()
		}
		if *output.IsTruncated {
			pageToken = output.NextContinuationToken
			continue
		}
		break
	}
	return schemas, nil
}

// returns the first job run id for a given job that is in S3
func (s *AwsS3ConnectionDataService) getLastestJobRunFromAwsS3(
	ctx context.Context,
	s3Client *s3.Client,
	jobId,
	bucket string,
	region *string,
	s3pathpieces []string,
) (string, error) {
	pieces := []string{}
	pieces = append(pieces, s3pathpieces...)
	pieces = append(pieces, "workflows", jobId)
	path := strings.Join(pieces, "/")

	var continuationToken *string
	done := false
	commonPrefixes := []string{}
	for !done {
		output, err := s.awsmanager.ListObjectsV2(ctx, s3Client, region, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(path),
			Delimiter:         aws.String("/"),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return "", fmt.Errorf("unable to list job run directories from s3: %w", err)
		}
		if output == nil {
			break
		}
		continuationToken = output.NextContinuationToken
		done = !*output.IsTruncated
		for _, cp := range output.CommonPrefixes {
			commonPrefixes = append(commonPrefixes, *cp.Prefix)
		}
	}

	s.logger.Debug(fmt.Sprintf("found %d common prefixes for job in s3", len(commonPrefixes)))

	runIDs := make([]string, 0, len(commonPrefixes))
	for _, prefix := range commonPrefixes {
		parts := strings.Split(strings.TrimSuffix(prefix, "/"), "/")
		if len(parts) >= 2 {
			runID := parts[len(parts)-1]
			runIDs = append(runIDs, runID)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(runIDs)))

	if len(runIDs) == 0 {
		return "", nucleuserrors.NewNotFound(
			fmt.Sprintf(
				"unable to find latest job run for job in s3 after processing common prefixes: %s",
				jobId,
			),
		)
	}
	s.logger.Debug(fmt.Sprintf("found %d run ids for job in s3", len(runIDs)))
	return runIDs[0], nil
}

func (s *AwsS3ConnectionDataService) GetInitStatements(
	ctx context.Context,
	options *mgmtv1alpha1.InitStatementOptions,
) (*mgmtv1alpha1.GetConnectionInitStatementsResponse, error) {
	return nil, errors.ErrUnsupported
}

func (s *AwsS3ConnectionDataService) GetTableConstraints(
	ctx context.Context,
) (*mgmtv1alpha1.GetConnectionTableConstraintsResponse, error) {
	return nil, errors.ErrUnsupported
}

func (s *AwsS3ConnectionDataService) GetTableSchema(
	ctx context.Context,
	schema, table string,
) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	return nil, errors.ErrUnsupported
}

func (s *AwsS3ConnectionDataService) GetTableRowCount(
	ctx context.Context,
	schema, table string,
	whereClause *string,
) (int64, error) {
	return 0, errors.ErrUnsupported
}
