package v1alpha1_connectiondataservice

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	aws_s3 "github.com/nucleuscloud/neosync/backend/internal/aws/s3"
	aws_session "github.com/nucleuscloud/neosync/backend/internal/aws/session"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	conn_utils "github.com/nucleuscloud/neosync/backend/internal/utils/connections"
)

const (
	mysqlDriver    = "mysql"
	postgresDriver = "postgres"
)

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

	switch config := req.Msg.StreamConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionStreamConfig_MysqlConfig, *mgmtv1alpha1.ConnectionStreamConfig_PgConfig:
		// check that schema and table are valid
		schemaResp, err := s.connectionService.GetConnectionSchema(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{Id: req.Msg.ConnectionId}))
		if err != nil {
			return err
		}

		if !isValidSchema(req.Msg.Schema, schemaResp.Msg.Schemas) || !isValidTable(req.Msg.Table, schemaResp.Msg.Schemas) {
			return nucleuserrors.NewBadRequest("must provide valid schema and table")
		}

		connDetails, err := s.getConnectionDetails(connection.ConnectionConfig)
		if err != nil {
			return err
		}

		conn, err := s.sqlConnector.Open(connDetails.ConnectionDriver, connDetails.ConnectionString)
		if err != nil {
			logger.Error("unable to connect", err)
			return err
		}
		defer func() {
			if err := conn.Close(); err != nil {
				logger.Error(fmt.Errorf("failed to close sql connection: %w", err).Error())
			}
		}()

		// used to get column names
		query := fmt.Sprintf("SELECT * FROM %s.%s LIMIT 1;", req.Msg.Schema, req.Msg.Table) //nolint
		r, err := conn.QueryContext(ctx, query)
		if err != nil && !nucleusdb.IsNoRows(err) {
			return err
		}

		columnNames, err := r.Columns()
		if err != nil {
			return err
		}

		selectQuery := fmt.Sprintf("SELECT %s FROM %s.%s", strings.Join(columnNames, ", "), req.Msg.Schema, req.Msg.Table) //nolint
		rows, err := conn.QueryContext(ctx, selectQuery)
		if err != nil && !nucleusdb.IsNoRows(err) {
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

	case *mgmtv1alpha1.ConnectionStreamConfig_AwsS3Config:
		var jobRunId string
		switch id := config.AwsS3Config.Id.(type) {
		case *mgmtv1alpha1.AwsS3StreamConfig_JobRunId:
			jobRunId = id.JobRunId
		case *mgmtv1alpha1.AwsS3StreamConfig_JobId:
			// get latest job run id and compare to bucket
			jobResp, err := s.jobService.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
				Id: id.JobId,
			}))
			if err != nil {
				return err
			}
			job := jobResp.Msg.Job
			if job.AccountId != connection.AccountId {
				return nucleuserrors.NewForbidden("must provide valid job id")
			}
		default:
			return nucleuserrors.NewInternalError("unsupported AWS S3 config id")
		}

		awsS3Config := connection.ConnectionConfig.GetAwsS3Config()
		if awsS3Config == nil {
			return nucleuserrors.NewInternalError("AWS S3 connection config missing")
		}

		sess, err := aws_session.NewSession(awsS3Config)
		if err != nil {
			logger.Error("unable to create AWS session")
			return err
		}
		logger.Info("created AWS session")

		svc := s3.New(sess)
		tableName := fmt.Sprintf("%s.%s", req.Msg.Schema, req.Msg.Table)
		path := fmt.Sprintf("workflows/%s/activities/%s/data", jobRunId, tableName)
		var pageToken *string
		for {
			var output *s3.ListObjectsV2Output
			output, err = svc.ListObjectsV2(&s3.ListObjectsV2Input{
				Bucket:            aws.String(awsS3Config.Bucket),
				Prefix:            aws.String(path),
				ContinuationToken: pageToken,
			})
			if err != nil && !aws_s3.IsNotFound(err) {
				return err
			}
			if err != nil || *output.KeyCount == 0 {
				break
			}
			for _, item := range output.Contents {
				result, err := svc.GetObject(&s3.GetObjectInput{
					Bucket: aws.String(awsS3Config.Bucket),
					Key:    aws.String(*item.Key),
				})
				if err != nil {
					return fmt.Errorf("error getting object from S3: %w", err)
				}

				gzr, err := gzip.NewReader(result.Body)
				if err != nil {
					result.Body.Close()
					return fmt.Errorf("error creating gzip reader: %w", err)
				}

				scanner := bufio.NewScanner(gzr)
				for scanner.Scan() {
					line := scanner.Bytes()
					var data map[string]interface{} // nolint
					err = json.Unmarshal(line, &data)
					if err != nil {
						result.Body.Close()
						gzr.Close()
						return err
					}

					rowMap := make(map[string][]byte)
					for key, value := range data {
						byteValue, err := json.Marshal(value)
						if err != nil {
							result.Body.Close()
							gzr.Close()
							return err
						}
						rowMap[key] = byteValue
					}
					if err := stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{Row: rowMap}); err != nil {
						result.Body.Close()
						gzr.Close()
						return err
					}

				}
				if err := scanner.Err(); err != nil {
					result.Body.Close()
					gzr.Close()
					return err
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

	default:
		return nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}

	return nil

}

func (s *Service) GetConnectionDataSchema(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionDataSchemaRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionDataSchemaResponse], error) {
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

	switch config := req.Msg.SchemaConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionSchemaConfig_MysqlConfig, *mgmtv1alpha1.ConnectionSchemaConfig_PgConfig:
		return connect.NewResponse(&mgmtv1alpha1.GetConnectionDataSchemaResponse{}), nil
	case *mgmtv1alpha1.ConnectionSchemaConfig_AwsS3Config:
		var jobRunId string
		switch id := config.AwsS3Config.Id.(type) {
		case *mgmtv1alpha1.AwsS3StreamConfig_JobRunId:
			jobRunId = id.JobRunId
		case *mgmtv1alpha1.AwsS3StreamConfig_JobId:
			// get latest job run id and compare to bucket
		default:
			return nil, nucleuserrors.NewInternalError("unsupported AWS S3 config id")
		}

		awsS3Config := connection.ConnectionConfig.GetAwsS3Config()
		if awsS3Config == nil {
			return nil, nucleuserrors.NewInternalError("AWS S3 connection config missing")
		}

		sess, err := aws_session.NewSession(awsS3Config)
		if err != nil {
			logger.Error("unable to create AWS session")
			return nil, err
		}
		logger.Info("created AWS session")

		svc := s3.New(sess)
		path := fmt.Sprintf("workflows/%s/activities/", jobRunId)

		schema := []*mgmtv1alpha1.Column{}
		var pageToken *string
		for {
			var output *s3.ListObjectsV2Output
			output, err = svc.ListObjectsV2(&s3.ListObjectsV2Input{
				Bucket:            aws.String(awsS3Config.Bucket),
				Prefix:            aws.String(path),
				Delimiter:         aws.String("/"),
				ContinuationToken: pageToken,
			})
			if err != nil && !aws_s3.IsNotFound(err) {
				return nil, err
			}
			if err != nil || *output.KeyCount == 0 {
				break
			}
			for _, item := range output.CommonPrefixes {
				folders := strings.Split(*item.Prefix, "activities")
				tableFolder := strings.ReplaceAll(folders[len(folders)-1], "/", "")
				pieces := strings.Split(tableFolder, ".")
				schema = append(schema, &mgmtv1alpha1.Column{
					Schema: pieces[0],
					Table:  pieces[1],
				})
			}
			if *output.IsTruncated {
				pageToken = output.NextContinuationToken
				continue
			}
			break
		}
		return connect.NewResponse(&mgmtv1alpha1.GetConnectionDataSchemaResponse{
			Schemas: schema,
		}), nil

	default:
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}
}

type connectionDetails struct {
	ConnectionString string
	ConnectionDriver string
}

func (s *Service) getConnectionDetails(c *mgmtv1alpha1.ConnectionConfig) (*connectionDetails, error) {
	switch config := c.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		var connectionString *string
		switch connectionConfig := config.PgConfig.ConnectionConfig.(type) {
		case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
			connStr := conn_utils.GetPostgresUrl(&conn_utils.PostgresConnectConfig{
				Host:     connectionConfig.Connection.Host,
				Port:     connectionConfig.Connection.Port,
				Database: connectionConfig.Connection.Name,
				User:     connectionConfig.Connection.User,
				Pass:     connectionConfig.Connection.Pass,
				SslMode:  connectionConfig.Connection.SslMode,
			})
			connectionString = &connStr
		case *mgmtv1alpha1.PostgresConnectionConfig_Url:
			connectionString = &connectionConfig.Url
		default:
			return nil, nucleuserrors.NewBadRequest("must provide valid postgres connection")
		}
		return &connectionDetails{ConnectionString: *connectionString, ConnectionDriver: postgresDriver}, nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		var connectionString *string
		switch connectionConfig := config.MysqlConfig.ConnectionConfig.(type) {
		case *mgmtv1alpha1.MysqlConnectionConfig_Connection:
			connStr := conn_utils.GetMysqlUrl(&conn_utils.MysqlConnectConfig{
				Host:     connectionConfig.Connection.Host,
				Port:     connectionConfig.Connection.Port,
				Database: connectionConfig.Connection.Name,
				Username: connectionConfig.Connection.User,
				Password: connectionConfig.Connection.Pass,
				Protocol: connectionConfig.Connection.Protocol,
			})
			connectionString = &connStr
		case *mgmtv1alpha1.MysqlConnectionConfig_Url:
			connectionString = &connectionConfig.Url
		default:
			return nil, nucleuserrors.NewBadRequest("must provide valid mysql connection")
		}
		return &connectionDetails{ConnectionString: *connectionString, ConnectionDriver: mysqlDriver}, nil
	default:
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}
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
