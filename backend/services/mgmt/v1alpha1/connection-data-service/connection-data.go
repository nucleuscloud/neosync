package v1alpha1_connectiondataservice

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	neosync_gcp "github.com/nucleuscloud/neosync/backend/internal/gcp"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sqlmanager_mysql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mysql"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	neosync_dynamodb "github.com/nucleuscloud/neosync/internal/dynamodb"
	querybuilder "github.com/nucleuscloud/neosync/worker/pkg/query-builder"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/structpb"
)

type DatabaseSchema struct {
	TableSchema string `db:"table_schema,omitempty"`
	TableName   string `db:"table_name,omitempty"`
	ColumnName  string `db:"column_name,omitempty"`
	DataType    string `db:"data_type,omitempty"`
}

type DateScanner struct {
	val *time.Time
}

func (ds *DateScanner) Scan(input any) error {
	if input == nil {
		return nil
	}

	switch input := input.(type) {
	case time.Time:
		*ds.val = input
		return nil
	default:
		return fmt.Errorf("unable to scan type %T into DateScanner", input)
	}
}

func (s *Service) GetConnectionDataStream(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionDataStreamRequest],
	stream *connect.ServerStream[mgmtv1alpha1.GetConnectionDataStreamResponse],
) error {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionId", req.Msg.ConnectionId)
	connResp, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.ConnectionId,
	}))
	if err != nil {
		return err
	}
	connection := connResp.Msg.Connection
	_, err = s.verifyUserInAccount(ctx, connection.AccountId)
	if err != nil {
		return err
	}

	connectionTimeout := uint32(5)

	switch config := connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		err := s.areSchemaAndTableValid(ctx, connection, req.Msg.Schema, req.Msg.Table)
		if err != nil {
			return err
		}

		conn, err := s.sqlConnector.NewDbFromConnectionConfig(connection.ConnectionConfig, &connectionTimeout, logger, sqlconnect.WithMysqlParseTimeDisabled())
		if err != nil {
			return err
		}
		defer conn.Close()
		db, err := conn.Open()
		if err != nil {
			return err
		}

		table := sqlmanager_shared.BuildTable(req.Msg.Schema, req.Msg.Table)
		// used to get column names
		query, err := querybuilder.BuildSelectLimitQuery("mysql", table, 1)
		if err != nil {
			return err
		}
		r, err := db.QueryContext(ctx, query)
		if err != nil && !neosyncdb.IsNoRows(err) {
			return err
		}

		columnNames, err := r.Columns()
		if err != nil {
			return err
		}

		selectQuery, err := querybuilder.BuildSelectQuery("mysql", table, columnNames, nil)
		if err != nil {
			return err
		}
		rows, err := db.QueryContext(ctx, selectQuery)
		if err != nil && !neosyncdb.IsNoRows(err) {
			return err
		}

		for rows.Next() {
			values := make([][]byte, len(columnNames))
			valuesWrapped := make([]any, 0, len(columnNames))
			for i := range values {
				valuesWrapped = append(valuesWrapped, &values[i])
			}
			if err := rows.Scan(valuesWrapped...); err != nil {
				return err
			}
			row := map[string][]byte{}
			for i, v := range values {
				col := columnNames[i]
				row[col] = v
			}

			if err := stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{Row: row}); err != nil {
				return err
			}
		}

	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		err := s.areSchemaAndTableValid(ctx, connection, req.Msg.Schema, req.Msg.Table)
		if err != nil {
			return err
		}

		conn, err := s.sqlConnector.NewDbFromConnectionConfig(connection.GetConnectionConfig(), &connectionTimeout, logger, sqlconnect.WithDefaultPostgresDriver())
		if err != nil {
			return err
		}
		db, err := conn.Open()
		if err != nil {
			return err
		}
		defer conn.Close()

		table := sqlmanager_shared.BuildTable(req.Msg.Schema, req.Msg.Table)
		// used to get column names
		query, err := querybuilder.BuildSelectLimitQuery("postgres", table, 1)
		if err != nil {
			return err
		}
		r, err := db.QueryContext(ctx, query)
		if err != nil && !neosyncdb.IsNoRows(err) {
			return err
		}
		defer r.Close()

		columnNames, err := r.Columns()
		if err != nil {
			return err
		}

		selectQuery, err := querybuilder.BuildSelectQuery("postgres", table, columnNames, nil)
		if err != nil {
			return err
		}
		rows, err := db.QueryContext(ctx, selectQuery)
		if err != nil && !neosyncdb.IsNoRows(err) {
			return err
		}
		defer rows.Close()

		// todo: this is probably way fucking broken now
		for rows.Next() {
			values := make([][]byte, len(columnNames))
			valuesWrapped := make([]any, 0, len(columnNames))
			for i := range values {
				valuesWrapped = append(valuesWrapped, &values[i])
			}
			if err := rows.Scan(valuesWrapped...); err != nil {
				return err
			}
			row := map[string][]byte{}
			for i, v := range values {
				col := columnNames[i]
				row[col] = v
			}

			if err := stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{Row: row}); err != nil {
				return err
			}
		}
		return nil

	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		awsS3StreamCfg := req.Msg.StreamConfig.GetAwsS3Config()
		if awsS3StreamCfg == nil {
			return nucleuserrors.NewBadRequest("jobId or jobRunId required for AWS S3 connections")
		}

		awsS3Config := config.AwsS3Config
		s3Client, err := s.awsManager.NewS3Client(ctx, awsS3Config)
		if err != nil {
			return fmt.Errorf("unable to create AWS S3 client: %w", err)
		}
		logger.Debug("created AWS S3 client")

		connAwsConfig := connection.ConnectionConfig.GetAwsS3Config()
		s3pathpieces := []string{}
		if connAwsConfig != nil && connAwsConfig.GetPathPrefix() != "" {
			s3pathpieces = append(s3pathpieces, strings.Trim(connAwsConfig.GetPathPrefix(), "/"))
		}

		var jobRunId string
		switch id := awsS3StreamCfg.Id.(type) {
		case *mgmtv1alpha1.AwsS3StreamConfig_JobRunId:
			jobRunId = id.JobRunId
		case *mgmtv1alpha1.AwsS3StreamConfig_JobId:
			logger = logger.With("jobId", id.JobId)
			runId, err := s.getLastestJobRunFromAwsS3(ctx, logger, s3Client, id.JobId, awsS3Config.Bucket, awsS3Config.Region, s3pathpieces)
			if err != nil {
				return err
			}
			logger.Debug(fmt.Sprintf("found run id for job in s3: %s", runId))
			jobRunId = runId
		default:
			return nucleuserrors.NewInternalError("unsupported AWS S3 config id")
		}
		logger = logger.With("runId", jobRunId)

		tableName := sqlmanager_shared.BuildTable(req.Msg.Schema, req.Msg.Table)
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
			output, err := s.awsManager.ListObjectsV2(ctx, s3Client, awsS3Config.Region, &s3.ListObjectsV2Input{
				Bucket:            aws.String(awsS3Config.Bucket),
				Prefix:            aws.String(path),
				ContinuationToken: pageToken,
			})
			if err != nil {
				return err
			}
			if output == nil {
				logger.Debug(fmt.Sprintf("0 files found for path: %s", path))
				break
			}
			for _, item := range output.Contents {
				result, err := s.awsManager.GetObject(ctx, s3Client, awsS3Config.Region, &s3.GetObjectInput{
					Bucket: aws.String(awsS3Config.Bucket),
					Key:    aws.String(*item.Key),
				})
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
					var data map[string]any

					// Decode the next JSON object
					err = decoder.Decode(&data)
					if err != nil && err == io.EOF {
						break // End of file, stop the loop
					} else if err != nil {
						result.Body.Close()
						gzr.Close()
						return err
					}
					rowMap := make(map[string][]byte)
					for key, value := range data {
						var byteValue []byte
						switch v := value.(type) {
						case string:
							// try converting string directly to []byte
							// prevents quoted strings
							byteValue = []byte(v)
						default:
							// if not a string use JSON encoding
							byteValue, err = json.Marshal(v)
							if err != nil {
								result.Body.Close()
								gzr.Close()
								return err
							}
							if string(byteValue) == "null" {
								byteValue = nil
							}
						}
						rowMap[key] = byteValue
					}
					if err := stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{Row: rowMap}); err != nil {
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

	case *mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig:
		gcpStreamCfg := req.Msg.GetStreamConfig().GetGcpCloudstorageConfig()
		if gcpStreamCfg == nil {
			return nucleuserrors.NewBadRequest("must provide non-nil gcp cloud storage config in request")
		}
		gcpclient, err := s.gcpmanager.GetClient(ctx, logger)
		if err != nil {
			return fmt.Errorf("unable to init gcp storage client: %w", err)
		}
		gcpConfig := config.GcpCloudstorageConfig

		var jobRunId string
		switch id := gcpStreamCfg.Id.(type) {
		case *mgmtv1alpha1.GcpCloudStorageStreamConfig_JobRunId:
			jobRunId = id.JobRunId
		case *mgmtv1alpha1.GcpCloudStorageStreamConfig_JobId:
			runId, err := s.getLatestJobRunFromGcs(ctx, gcpclient, id.JobId, gcpConfig.GetBucket(), gcpConfig.PathPrefix)
			if err != nil {
				return err
			}
			jobRunId = runId
		default:
			return nucleuserrors.NewNotImplemented(fmt.Sprintf("unsupported GCP Cloud Storage config id: %T", id))
		}

		onRecord := func(record map[string][]byte) error {
			return stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{Row: record})
		}
		tablePath := neosync_gcp.GetWorkflowActivityDataPrefix(jobRunId, sqlmanager_shared.BuildTable(req.Msg.Schema, req.Msg.Table), gcpConfig.PathPrefix)
		err = gcpclient.GetRecordStreamFromPrefix(ctx, gcpConfig.GetBucket(), tablePath, onRecord)
		if err != nil {
			return fmt.Errorf("unable to finish sending record stream: %w", err)
		}
	case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
		dynamoclient, err := s.awsManager.NewDynamoDbClient(ctx, config.DynamodbConfig)
		if err != nil {
			return fmt.Errorf("unable to create dynamodb client from connection: %w", err)
		}
		var lastEvaluatedKey map[string]dynamotypes.AttributeValue

		for {
			output, err := dynamoclient.ScanTable(ctx, req.Msg.Table, lastEvaluatedKey)
			if err != nil {
				return fmt.Errorf("failed to scan table %s: %w", req.Msg.Table, err)
			}

			for _, item := range output.Items {
				row := make(map[string][]byte)

				itemBits, err := neosync_dynamodb.ConvertMapToJSONBytes(item)
				if err != nil {
					return err
				}
				row["item"] = itemBits
				if err := stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{Row: row}); err != nil {
					return fmt.Errorf("failed to send stream response: %w", err)
				}
			}

			lastEvaluatedKey = output.LastEvaluatedKey
			if lastEvaluatedKey == nil {
				break
			}
		}

	default:
		return nucleuserrors.NewNotImplemented(fmt.Sprintf("this connection config is not currently supported: %T", config))
	}
	return nil
}

func (s *Service) GetConnectionSchemaMaps(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionSchemaMapsRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionSchemaMapsResponse], error) {
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(3)

	responses := make([]*mgmtv1alpha1.GetConnectionSchemaMapResponse, len(req.Msg.GetRequests()))
	connectionIds := make([]string, len(req.Msg.GetRequests()))

	for idx, mapReq := range req.Msg.GetRequests() {
		idx := idx
		mapReq := mapReq
		connectionIds[idx] = mapReq.GetConnectionId()

		errgrp.Go(func() error {
			resp, err := s.GetConnectionSchemaMap(errctx, connect.NewRequest(mapReq))
			if err != nil {
				return err
			}
			responses[idx] = &mgmtv1alpha1.GetConnectionSchemaMapResponse{
				SchemaMap: resp.Msg.GetSchemaMap(),
			}
			return nil
		})
	}

	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaMapsResponse{
		Responses:     responses,
		ConnectionIds: connectionIds,
	}), nil
}

func (s *Service) GetConnectionSchemaMap(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionSchemaMapRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionSchemaMapResponse], error) {
	schemaResp, err := s.GetConnectionSchema(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{
		ConnectionId: req.Msg.GetConnectionId(),
		SchemaConfig: req.Msg.GetSchemaConfig(),
	}))
	if err != nil {
		return nil, err
	}
	outputMap := map[string]*mgmtv1alpha1.GetConnectionSchemaResponse{}
	for _, dbcol := range schemaResp.Msg.GetSchemas() {
		schematableKey := sqlmanager_shared.SchemaTable{Schema: dbcol.Schema, Table: dbcol.Table}.String()
		resp, ok := outputMap[schematableKey]
		if !ok {
			resp = &mgmtv1alpha1.GetConnectionSchemaResponse{}
		}
		resp.Schemas = append(resp.Schemas, dbcol)
		outputMap[schematableKey] = resp
	}
	return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaMapResponse{
		SchemaMap: outputMap,
	}), nil
}

func (s *Service) GetConnectionSchema(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionSchemaResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionId", req.Msg.ConnectionId)
	connResp, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.ConnectionId,
	}))
	if err != nil {
		return nil, err
	}
	connection := connResp.Msg.Connection
	_, err = s.verifyUserInAccount(ctx, connection.AccountId)
	if err != nil {
		return nil, err
	}

	switch config := connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		connectionTimeout := 5
		db, err := s.sqlmanager.NewSqlDb(ctx, logger, connection, &connectionTimeout)
		if err != nil {
			return nil, err
		}
		defer db.Db.Close()

		dbschema, err := db.Db.GetDatabaseSchema(ctx)
		if err != nil {
			return nil, err
		}
		schemas := []*mgmtv1alpha1.DatabaseColumn{}
		for _, col := range dbschema {
			col := col
			var defaultColumn *string
			if col.ColumnDefault != "" {
				defaultColumn = &col.ColumnDefault
			}

			schemas = append(schemas, &mgmtv1alpha1.DatabaseColumn{
				Schema:             col.TableSchema,
				Table:              col.TableName,
				Column:             col.ColumnName,
				DataType:           col.DataType,
				IsNullable:         col.NullableString(),
				ColumnDefault:      defaultColumn,
				GeneratedType:      col.GeneratedType,
				IdentityGeneration: col.IdentityGeneration,
			})
		}

		return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
			Schemas: schemas,
		}), nil

	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		db, err := s.mongoconnector.NewFromConnectionConfig(connection.GetConnectionConfig(), logger)
		if err != nil {
			return nil, err
		}
		mongoclient, err := db.Open(ctx)
		if err != nil {
			return nil, err
		}
		defer db.Close(ctx)
		dbnames, err := mongoclient.ListDatabaseNames(ctx, bson.D{})
		if err != nil {
			return nil, err
		}
		schemas := []*mgmtv1alpha1.DatabaseColumn{}
		for _, dbname := range dbnames {
			collectionNames, err := mongoclient.Database(dbname).ListCollectionNames(ctx, bson.D{})
			if err != nil {
				return nil, err
			}
			for _, collectionName := range collectionNames {
				schemas = append(schemas, &mgmtv1alpha1.DatabaseColumn{
					Schema: dbname,
					Table:  collectionName,
				})
			}
		}
		return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
			Schemas: schemas,
		}), nil
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		awsCfg := req.Msg.SchemaConfig.GetAwsS3Config()
		if awsCfg == nil {
			return nil, nucleuserrors.NewBadRequest("jobId or jobRunId required for AWS S3 connections")
		}

		awsS3Config := config.AwsS3Config
		s3Client, err := s.awsManager.NewS3Client(ctx, awsS3Config)
		if err != nil {
			return nil, err
		}
		logger.Debug("created S3 AWS session")

		connAwsConfig := connection.ConnectionConfig.GetAwsS3Config()
		s3pathpieces := []string{}
		if connAwsConfig != nil && connAwsConfig.GetPathPrefix() != "" {
			s3pathpieces = append(s3pathpieces, strings.Trim(connAwsConfig.GetPathPrefix(), "/"))
		}

		var jobRunId string
		switch id := awsCfg.Id.(type) {
		case *mgmtv1alpha1.AwsS3SchemaConfig_JobRunId:
			jobRunId = id.JobRunId
		case *mgmtv1alpha1.AwsS3SchemaConfig_JobId:
			runId, err := s.getLastestJobRunFromAwsS3(ctx, logger, s3Client, id.JobId, awsS3Config.Bucket, awsS3Config.Region, s3pathpieces)
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
			output, err := s.awsManager.ListObjectsV2(ctx, s3Client, awsS3Config.Region, &s3.ListObjectsV2Input{
				Bucket:            aws.String(awsS3Config.Bucket),
				Prefix:            aws.String(path),
				Delimiter:         aws.String("/"),
				ContinuationToken: pageToken,
			})
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

				filePath := fmt.Sprintf("%s%s/data", path, sqlmanager_shared.BuildTable(schemaTableList[0], schemaTableList[1]))
				out, err := s.awsManager.ListObjectsV2(ctx, s3Client, awsS3Config.Region, &s3.ListObjectsV2Input{
					Bucket:  aws.String(awsS3Config.Bucket),
					Prefix:  aws.String(filePath),
					MaxKeys: aws.Int32(1),
				})
				if err != nil {
					return nil, err
				}
				if out == nil {
					logger.Warn(fmt.Sprintf("AWS S3 table folder missing data folder: %s, continuing..", tableFolder))
					continue
				}
				item := out.Contents[0]
				result, err := s.awsManager.GetObject(ctx, s3Client, awsS3Config.Region, &s3.GetObjectInput{
					Bucket: aws.String(awsS3Config.Bucket),
					Key:    aws.String(*item.Key),
				})
				if err != nil {
					return nil, err
				}
				if result.ContentLength == nil || *result.ContentLength == 0 {
					logger.Warn(fmt.Sprintf("empty AWS S3 data folder for table: %s, continuing...", tableFolder))
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
		return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
			Schemas: schemas,
		}), nil
	case *mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig:
		gcpCfg := req.Msg.GetSchemaConfig().GetGcpCloudstorageConfig()
		if gcpCfg == nil {
			return nil, nucleuserrors.NewBadRequest("must provide gcp cloud storage config")
		}

		gcpclient, err := s.gcpmanager.GetClient(ctx, logger)
		if err != nil {
			return nil, fmt.Errorf("unable to init gcp storage client: %w", err)
		}
		gcpConfig := config.GcpCloudstorageConfig

		var jobRunId string
		switch id := gcpCfg.Id.(type) {
		case *mgmtv1alpha1.GcpCloudStorageSchemaConfig_JobRunId:
			jobRunId = id.JobRunId
		case *mgmtv1alpha1.GcpCloudStorageSchemaConfig_JobId:
			runId, err := s.getLatestJobRunFromGcs(ctx, gcpclient, id.JobId, gcpConfig.GetBucket(), gcpConfig.PathPrefix)
			if err != nil {
				return nil, err
			}
			jobRunId = runId
		default:
			return nil, nucleuserrors.NewNotImplemented(fmt.Sprintf("unsupported GCP Cloud Storage config id: %T", id))
		}

		schemas, err := gcpclient.GetDbSchemaFromPrefix(
			ctx,
			gcpConfig.GetBucket(), neosync_gcp.GetWorkflowActivityPrefix(jobRunId, gcpConfig.PathPrefix),
		)
		if err != nil {
			return nil, fmt.Errorf("uanble to retrieve db schema from gcs: %w", err)
		}
		return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
			Schemas: schemas,
		}), nil
	case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
		dynclient, err := s.awsManager.NewDynamoDbClient(ctx, config.DynamodbConfig)
		if err != nil {
			return nil, fmt.Errorf("unable to create dynamodb client from connection: %w", err)
		}
		tableNames, err := dynclient.ListAllTables(ctx, &dynamodb.ListTablesInput{})
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve dynamodb tables: %w", err)
		}
		schemas := []*mgmtv1alpha1.DatabaseColumn{}
		for _, tableName := range tableNames {
			schemas = append(schemas, &mgmtv1alpha1.DatabaseColumn{
				Schema: "dynamodb",
				Table:  tableName,
			})
		}
		return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
			Schemas: schemas,
		}), nil
	default:
		return nil, nucleuserrors.NewNotImplemented(fmt.Sprintf("this connection config is not currently supported: %T", config))
	}
}

func (s *Service) GetConnectionForeignConstraints(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionForeignConstraintsRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionForeignConstraintsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	connection, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.ConnectionId,
	}))
	if err != nil {
		return nil, err
	}

	_, err = s.verifyUserInAccount(ctx, connection.Msg.Connection.AccountId)
	if err != nil {
		return nil, err
	}

	schemaResp, err := s.getConnectionSchema(ctx, connection.Msg.Connection, &schemaOpts{})
	if err != nil {
		return nil, err
	}

	schemaMap := map[string]struct{}{}
	for _, s := range schemaResp {
		schemaMap[s.Schema] = struct{}{}
	}
	schemas := []string{}
	for s := range schemaMap {
		schemas = append(schemas, s)
	}

	connectionTimeout := 5
	db, err := s.sqlmanager.NewSqlDb(ctx, logger, connection.Msg.GetConnection(), &connectionTimeout)
	if err != nil {
		return nil, err
	}
	defer db.Db.Close()
	constraints, err := db.Db.GetTableConstraintsBySchema(ctx, schemas)
	if err != nil {
		return nil, err
	}

	tableConstraints := map[string]*mgmtv1alpha1.ForeignConstraintTables{}
	for tableName, d := range constraints.ForeignKeyConstraints {
		tableConstraints[tableName] = &mgmtv1alpha1.ForeignConstraintTables{
			Constraints: []*mgmtv1alpha1.ForeignConstraint{},
		}
		for _, constraint := range d {
			for idx, col := range constraint.Columns {
				tableConstraints[tableName].Constraints = append(tableConstraints[tableName].Constraints, &mgmtv1alpha1.ForeignConstraint{
					Column: col, IsNullable: !constraint.NotNullable[idx], ForeignKey: &mgmtv1alpha1.ForeignKey{
						Table:  constraint.ForeignKey.Table,
						Column: constraint.ForeignKey.Columns[idx],
					},
				})
			}
		}
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionForeignConstraintsResponse{
		TableConstraints: tableConstraints,
	}), nil
}

func (s *Service) GetConnectionPrimaryConstraints(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionPrimaryConstraintsRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionPrimaryConstraintsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	connection, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.ConnectionId,
	}))
	if err != nil {
		return nil, err
	}

	_, err = s.verifyUserInAccount(ctx, connection.Msg.Connection.AccountId)
	if err != nil {
		return nil, err
	}

	schemaResp, err := s.getConnectionSchema(ctx, connection.Msg.Connection, &schemaOpts{})
	if err != nil {
		return nil, err
	}

	schemaMap := map[string]struct{}{}
	for _, s := range schemaResp {
		schemaMap[s.Schema] = struct{}{}
	}
	schemas := []string{}
	for s := range schemaMap {
		schemas = append(schemas, s)
	}

	connectionTimeout := 5
	db, err := s.sqlmanager.NewSqlDb(ctx, logger, connection.Msg.GetConnection(), &connectionTimeout)
	if err != nil {
		return nil, err
	}
	defer db.Db.Close()

	constraints, err := db.Db.GetTableConstraintsBySchema(ctx, schemas)
	if err != nil {
		return nil, err
	}

	tableConstraints := map[string]*mgmtv1alpha1.PrimaryConstraint{}
	for tableName, cols := range constraints.PrimaryKeyConstraints {
		tableConstraints[tableName] = &mgmtv1alpha1.PrimaryConstraint{
			Columns: cols,
		}
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionPrimaryConstraintsResponse{
		TableConstraints: tableConstraints,
	}), nil
}

func (s *Service) GetConnectionInitStatements(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionInitStatementsRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionInitStatementsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	connection, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.ConnectionId,
	}))
	if err != nil {
		return nil, err
	}

	_, err = s.verifyUserInAccount(ctx, connection.Msg.Connection.AccountId)
	if err != nil {
		return nil, err
	}

	schemaResp, err := s.getConnectionSchema(ctx, connection.Msg.Connection, &schemaOpts{})
	if err != nil {
		return nil, err
	}

	schemaTableMap := map[string]*mgmtv1alpha1.DatabaseColumn{}
	for _, s := range schemaResp {
		schemaTableMap[sqlmanager_shared.BuildTable(s.Schema, s.Table)] = s
	}

	connectionTimeout := 5
	db, err := s.sqlmanager.NewSqlDb(ctx, logger, connection.Msg.GetConnection(), &connectionTimeout)
	if err != nil {
		return nil, err
	}
	defer db.Db.Close()

	createStmtsMap := map[string]string{}
	truncateStmtsMap := map[string]string{}
	initSchemaStmts := []*mgmtv1alpha1.SchemaInitStatements{}
	if req.Msg.GetOptions().GetInitSchema() {
		tables := []*sqlmanager_shared.SchemaTable{}
		for k, v := range schemaTableMap {
			stmt, err := db.Db.GetCreateTableStatement(ctx, v.Schema, v.Table)
			if err != nil {
				return nil, err
			}
			createStmtsMap[k] = stmt
			tables = append(tables, &sqlmanager_shared.SchemaTable{Schema: v.Schema, Table: v.Table})
		}
		initBlocks, err := db.Db.GetSchemaInitStatements(ctx, tables)
		if err != nil {
			return nil, err
		}
		for _, b := range initBlocks {
			initSchemaStmts = append(initSchemaStmts, &mgmtv1alpha1.SchemaInitStatements{
				Label:      b.Label,
				Statements: b.Statements,
			})
		}
	}

	switch connection.Msg.Connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		if req.Msg.GetOptions().GetTruncateBeforeInsert() {
			for k, v := range schemaTableMap {
				stmt, err := sqlmanager_mysql.BuildMysqlTruncateStatement(v.Schema, v.Table)
				if err != nil {
					return nil, err
				}
				truncateStmtsMap[k] = stmt
			}
		}

	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		if req.Msg.GetOptions().GetTruncateCascade() {
			for k, v := range schemaTableMap {
				stmt, err := sqlmanager_postgres.BuildPgTruncateCascadeStatement(v.Schema, v.Table)
				if err != nil {
					return nil, err
				}
				truncateStmtsMap[k] = stmt
			}
		} else if req.Msg.GetOptions().GetTruncateBeforeInsert() {
			return nil, nucleuserrors.NewNotImplemented("postgres truncate unsupported. table foreig keys required to build truncate statement.")
		}

	default:
		return nil, errors.New("unsupported connection config")
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionInitStatementsResponse{
		TableInitStatements:     createStmtsMap,
		TableTruncateStatements: truncateStmtsMap,
		SchemaInitStatements:    initSchemaStmts,
	}), nil
}

type schemaOpts struct {
	JobId    *string
	JobRunId *string
}

func (s *Service) getConnectionSchema(ctx context.Context, connection *mgmtv1alpha1.Connection, opts *schemaOpts) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	schemaReq := &mgmtv1alpha1.GetConnectionSchemaRequest{
		ConnectionId: connection.Id,
	}
	switch connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		schemaReq.SchemaConfig = &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresSchemaConfig{},
			},
		}
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		schemaReq.SchemaConfig = &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlSchemaConfig{},
			},
		}
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		var cfg *mgmtv1alpha1.AwsS3SchemaConfig
		if opts.JobRunId != nil && *opts.JobRunId != "" {
			cfg = &mgmtv1alpha1.AwsS3SchemaConfig{Id: &mgmtv1alpha1.AwsS3SchemaConfig_JobRunId{JobRunId: *opts.JobRunId}}
		} else if opts.JobId != nil && *opts.JobId != "" {
			cfg = &mgmtv1alpha1.AwsS3SchemaConfig{Id: &mgmtv1alpha1.AwsS3SchemaConfig_JobId{JobId: *opts.JobId}}
		}
		schemaReq.SchemaConfig = &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_AwsS3Config{
				AwsS3Config: cfg,
			},
		}
	case *mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig:
		var cfg *mgmtv1alpha1.GcpCloudStorageSchemaConfig
		if opts.JobRunId != nil && *opts.JobRunId != "" {
			cfg = &mgmtv1alpha1.GcpCloudStorageSchemaConfig{Id: &mgmtv1alpha1.GcpCloudStorageSchemaConfig_JobRunId{JobRunId: *opts.JobRunId}}
		} else if opts.JobId != nil && *opts.JobId != "" {
			cfg = &mgmtv1alpha1.GcpCloudStorageSchemaConfig{Id: &mgmtv1alpha1.GcpCloudStorageSchemaConfig_JobId{JobId: *opts.JobId}}
		}
		schemaReq.SchemaConfig = &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_GcpCloudstorageConfig{
				GcpCloudstorageConfig: cfg,
			},
		}
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		schemaReq.SchemaConfig = &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_MongoConfig{
				MongoConfig: &mgmtv1alpha1.MongoSchemaConfig{},
			},
		}
	case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
		schemaReq.SchemaConfig = &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_DynamodbConfig{},
		}
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		schemaReq.SchemaConfig = &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_MssqlConfig{
				MssqlConfig: &mgmtv1alpha1.MssqlSchemaConfig{},
			},
		}
	default:
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}
	schemaResp, err := s.GetConnectionSchema(ctx, connect.NewRequest(schemaReq))
	if err != nil {
		return nil, err
	}
	return schemaResp.Msg.GetSchemas(), nil
}

func (s *Service) getConnectionTableSchema(ctx context.Context, connection *mgmtv1alpha1.Connection, schema, table string, logger *slog.Logger) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	conntimeout := uint32(5)
	switch connection.GetConnectionConfig().Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		conn, err := s.sqlConnector.NewDbFromConnectionConfig(connection.GetConnectionConfig(), &conntimeout, logger)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		db, err := conn.Open()
		if err != nil {
			return nil, err
		}
		schematable := sqlmanager_shared.SchemaTable{Schema: schema, Table: table}
		dbschema, err := s.pgquerier.GetDatabaseTableSchemasBySchemasAndTables(ctx, db, []string{schematable.String()})
		if err != nil {
			return nil, err
		}
		schemas := []*mgmtv1alpha1.DatabaseColumn{}
		for _, col := range dbschema {
			schemas = append(schemas, &mgmtv1alpha1.DatabaseColumn{
				Schema:     col.SchemaName,
				Table:      col.TableName,
				Column:     col.ColumnName,
				DataType:   col.DataType,
				IsNullable: col.IsNullable,
			})
		}
		return schemas, nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		conn, err := s.sqlConnector.NewDbFromConnectionConfig(connection.GetConnectionConfig(), &conntimeout, logger)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		db, err := conn.Open()
		if err != nil {
			return nil, err
		}
		dbschema, err := s.mysqlquerier.GetDatabaseSchema(ctx, db)
		if err != nil {
			return nil, err
		}
		schemas := []*mgmtv1alpha1.DatabaseColumn{}
		for _, col := range dbschema {
			if col.TableSchema != schema || col.TableName != table {
				continue
			}
			schemas = append(schemas, &mgmtv1alpha1.DatabaseColumn{
				Schema:     col.TableSchema,
				Table:      col.TableName,
				Column:     col.ColumnName,
				DataType:   col.DataType,
				IsNullable: col.IsNullable,
			})
		}
		return schemas, nil
	default:
		return nil, nucleuserrors.NewBadRequest("this connection config is not currently supported")
	}
}

func (s *Service) getLatestJobRunFromGcs(
	ctx context.Context,
	client neosync_gcp.ClientInterface,
	jobId string,
	bucket string,
	pathPrefix *string,
) (string, error) {
	jobRunsResp, err := s.jobService.GetJobRecentRuns(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRecentRunsRequest{
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

// returns the first job run id for a given job that is in S3
func (s *Service) getLastestJobRunFromAwsS3(
	ctx context.Context,
	logger *slog.Logger,
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
		output, err := s.awsManager.ListObjectsV2(ctx, s3Client, region, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(path),
			Delimiter:         aws.String("/"),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return "", fmt.Errorf("unable to list job run directories from s3: %w", err)
		}
		continuationToken = output.NextContinuationToken
		done = !*output.IsTruncated
		for _, cp := range output.CommonPrefixes {
			commonPrefixes = append(commonPrefixes, *cp.Prefix)
		}
	}

	logger.Debug(fmt.Sprintf("found %d common prefixes for job in s3", len(commonPrefixes)))

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
		return "", nucleuserrors.NewNotFound(fmt.Sprintf("unable to find latest job run for job in s3 after processing common prefixes: %s", jobId))
	}
	logger.Debug(fmt.Sprintf("found %d run ids for job in s3", len(runIDs)))
	return runIDs[0], nil
}

func (s *Service) areSchemaAndTableValid(ctx context.Context, connection *mgmtv1alpha1.Connection, schema, table string) error {
	schemas, err := s.getConnectionSchema(ctx, connection, &schemaOpts{})
	if err != nil {
		return err
	}

	if !isValidSchema(schema, schemas) || !isValidTable(table, schemas) {
		return nucleuserrors.NewBadRequest("must provide valid schema and table")
	}
	return nil
}

func isValidTable(table string, columns []*mgmtv1alpha1.DatabaseColumn) bool {
	for _, c := range columns {
		if c.Table == table {
			return true
		}
	}
	return false
}

func isValidSchema(schema string, columns []*mgmtv1alpha1.DatabaseColumn) bool {
	for _, c := range columns {
		if c.Schema == schema {
			return true
		}
	}
	return false
}

func (s *Service) GetConnectionUniqueConstraints(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionUniqueConstraintsRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionUniqueConstraintsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	connection, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.ConnectionId,
	}))
	if err != nil {
		return nil, err
	}

	_, err = s.verifyUserInAccount(ctx, connection.Msg.Connection.AccountId)
	if err != nil {
		return nil, err
	}

	schemaResp, err := s.getConnectionSchema(ctx, connection.Msg.Connection, &schemaOpts{})
	if err != nil {
		return nil, err
	}

	schemaMap := map[string]struct{}{}
	for _, s := range schemaResp {
		schemaMap[s.Schema] = struct{}{}
	}
	schemas := []string{}
	for s := range schemaMap {
		schemas = append(schemas, s)
	}

	connectionTimeout := 5
	db, err := s.sqlmanager.NewSqlDb(ctx, logger, connection.Msg.GetConnection(), &connectionTimeout)
	if err != nil {
		return nil, err
	}
	defer db.Db.Close()

	constraints, err := db.Db.GetTableConstraintsBySchema(ctx, schemas)
	if err != nil {
		return nil, err
	}

	tableConstraints := map[string]*mgmtv1alpha1.UniqueConstraint{}
	for tableName, uc := range constraints.UniqueConstraints {
		columns := []string{}
		for _, c := range uc {
			columns = append(columns, c...)
		}
		tableConstraints[tableName] = &mgmtv1alpha1.UniqueConstraint{
			// TODO: this doesn't fully represent unique constraints
			Columns: columns,
		}
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionUniqueConstraintsResponse{
		TableConstraints: tableConstraints,
	}), nil
}

type completionResponse struct {
	Data []map[string]any `json:"data"`
}

func (s *Service) GetAiGeneratedData(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAiGeneratedDataRequest],
) (*connect.Response[mgmtv1alpha1.GetAiGeneratedDataResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	_ = logger
	aiconnectionResp, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.GetAiConnectionId(),
	}))
	if err != nil {
		return nil, err
	}
	aiconnection := aiconnectionResp.Msg.GetConnection()
	_, err = s.verifyUserInAccount(ctx, aiconnection.GetAccountId())
	if err != nil {
		return nil, err
	}

	dbconnectionResp, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.GetDataConnectionId(),
	}))
	if err != nil {
		return nil, err
	}
	dbcols, err := s.getConnectionTableSchema(ctx, dbconnectionResp.Msg.GetConnection(), req.Msg.GetTable().GetSchema(), req.Msg.GetTable().GetTable(), logger)
	if err != nil {
		return nil, err
	}

	columns := make([]string, 0, len(dbcols))
	for _, dbcol := range dbcols {
		columns = append(columns, fmt.Sprintf("%s is %s", dbcol.Column, dbcol.DataType))
	}

	openaiconfig := aiconnection.GetConnectionConfig().GetOpenaiConfig()
	if openaiconfig == nil {
		return nil, nucleuserrors.NewBadRequest("connection must be a valid openai connection")
	}

	client, err := azopenai.NewClientForOpenAI(openaiconfig.GetApiUrl(), azcore.NewKeyCredential(openaiconfig.GetApiKey()), &azopenai.ClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to init openai client: %w", err)
	}

	conversation := []azopenai.ChatRequestMessageClassification{
		&azopenai.ChatRequestSystemMessage{
			Content: azopenai.NewChatRequestSystemMessageContent(fmt.Sprintf("You generate data in JSON format. Generate %d records in a json array located on the data key", req.Msg.GetCount())),
		},
		&azopenai.ChatRequestUserMessage{
			Content: azopenai.NewChatRequestUserMessageContent(fmt.Sprintf("%s\n%s", req.Msg.GetUserPrompt(), fmt.Sprintf("Each record looks like this: %s", strings.Join(columns, ",")))),
		},
	}

	chatResp, err := client.GetChatCompletions(ctx, azopenai.ChatCompletionsOptions{
		Temperature:      ptr(float32(1.0)),
		DeploymentName:   ptr(req.Msg.GetModelName()),
		TopP:             ptr(float32(1.0)),
		FrequencyPenalty: ptr(float32(0)),
		N:                ptr(int32(1)),
		ResponseFormat:   &azopenai.ChatCompletionsJSONResponseFormat{},
		Messages:         conversation,
	}, &azopenai.GetChatCompletionsOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get chat completions: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return nil, errors.New("received no choices back from openai")
	}
	choice := chatResp.Choices[0]

	if *choice.FinishReason == azopenai.CompletionsFinishReasonTokenLimitReached {
		return nil, errors.New("completion limit reached")
	}

	var dataResponse completionResponse
	err = json.Unmarshal([]byte(*choice.Message.Content), &dataResponse)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal openai message content into expected response: %w", err)
	}

	dtoRecords := []*structpb.Struct{}
	for _, record := range dataResponse.Data {
		dto, err := structpb.NewStruct(record)
		if err != nil {
			return nil, fmt.Errorf("unable to convert response data to dto struct: %w", err)
		}
		dtoRecords = append(dtoRecords, dto)
	}

	return connect.NewResponse(&mgmtv1alpha1.GetAiGeneratedDataResponse{Records: dtoRecords}), nil
}

func ptr[T any](val T) *T {
	return &val
}

func (s *Service) GetConnectionTableConstraints(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionTableConstraintsRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionTableConstraintsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	connection, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.ConnectionId,
	}))
	if err != nil {
		return nil, err
	}

	_, err = s.verifyUserInAccount(ctx, connection.Msg.Connection.AccountId)
	if err != nil {
		return nil, err
	}

	schemaResp, err := s.getConnectionSchema(ctx, connection.Msg.Connection, &schemaOpts{})
	if err != nil {
		return nil, err
	}

	schemaMap := map[string]struct{}{}
	for _, s := range schemaResp {
		schemaMap[s.Schema] = struct{}{}
	}
	schemas := []string{}
	for s := range schemaMap {
		schemas = append(schemas, s)
	}

	switch connection.Msg.GetConnection().GetConnectionConfig().GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		connectionTimeout := 5
		db, err := s.sqlmanager.NewSqlDb(ctx, logger, connection.Msg.GetConnection(), &connectionTimeout)
		if err != nil {
			return nil, err
		}
		defer db.Db.Close()
		tableConstraints, err := db.Db.GetTableConstraintsBySchema(ctx, schemas)
		if err != nil {
			return nil, err
		}

		fkConstraintsMap := map[string]*mgmtv1alpha1.ForeignConstraintTables{}
		for tableName, d := range tableConstraints.ForeignKeyConstraints {
			fkConstraintsMap[tableName] = &mgmtv1alpha1.ForeignConstraintTables{
				Constraints: []*mgmtv1alpha1.ForeignConstraint{},
			}
			for _, constraint := range d {
				fkConstraintsMap[tableName].Constraints = append(fkConstraintsMap[tableName].Constraints, &mgmtv1alpha1.ForeignConstraint{
					Columns: constraint.Columns, NotNullable: constraint.NotNullable, ForeignKey: &mgmtv1alpha1.ForeignKey{
						Table:   constraint.ForeignKey.Table,
						Columns: constraint.ForeignKey.Columns,
					},
				})
			}
		}

		pkConstraintsMap := map[string]*mgmtv1alpha1.PrimaryConstraint{}
		for table, pks := range tableConstraints.PrimaryKeyConstraints {
			pkConstraintsMap[table] = &mgmtv1alpha1.PrimaryConstraint{
				Columns: pks,
			}
		}

		uniqueConstraintsMap := map[string]*mgmtv1alpha1.UniqueConstraints{}
		for table, uniqueConstraints := range tableConstraints.UniqueConstraints {
			uniqueConstraintsMap[table] = &mgmtv1alpha1.UniqueConstraints{
				Constraints: []*mgmtv1alpha1.UniqueConstraint{},
			}
			for _, uc := range uniqueConstraints {
				uniqueConstraintsMap[table].Constraints = append(uniqueConstraintsMap[table].Constraints, &mgmtv1alpha1.UniqueConstraint{
					Columns: uc,
				})
			}
		}

		return connect.NewResponse(&mgmtv1alpha1.GetConnectionTableConstraintsResponse{
			ForeignKeyConstraints: fkConstraintsMap,
			PrimaryKeyConstraints: pkConstraintsMap,
			UniqueConstraints:     uniqueConstraintsMap,
		}), nil
	}
	return connect.NewResponse(&mgmtv1alpha1.GetConnectionTableConstraintsResponse{}), nil
}

func (s *Service) GetTableRowCount(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetTableRowCountRequest],
) (*connect.Response[mgmtv1alpha1.GetTableRowCountResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	connection, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.ConnectionId,
	}))
	if err != nil {
		return nil, err
	}

	_, err = s.verifyUserInAccount(ctx, connection.Msg.Connection.AccountId)
	if err != nil {
		return nil, err
	}

	switch connection.Msg.GetConnection().GetConnectionConfig().Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		connectionTimeout := 5
		db, err := s.sqlmanager.NewSqlDb(ctx, logger, connection.Msg.GetConnection(), &connectionTimeout)
		if err != nil {
			return nil, err
		}
		defer db.Db.Close()

		count, err := db.Db.GetTableRowCount(ctx, req.Msg.Schema, req.Msg.Table, req.Msg.WhereClause)
		if err != nil {
			return nil, err
		}

		return connect.NewResponse(&mgmtv1alpha1.GetTableRowCountResponse{
			Count: count,
		}), nil
	default:
		return nil, fmt.Errorf("unsupported connection type when retrieving table row count %T", connection.Msg.GetConnection().GetConnectionConfig().Config)
	}
}
