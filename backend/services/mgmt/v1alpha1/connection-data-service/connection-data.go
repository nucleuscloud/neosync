package v1alpha1_connectiondataservice

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofrs/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	dbschemas_mysql "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/mysql"
	dbschemas_postgres "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/postgres"
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

type UUIDScanner struct {
	val *uuid.UUID
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

		conn, err := s.sqlConnector.NewDbFromConnectionConfig(connection.ConnectionConfig, &connectionTimeout, logger)
		if err != nil {
			return err
		}
		defer conn.Close()
		db, err := conn.Open()
		if err != nil {
			return err
		}

		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		// used to get column names
		query := fmt.Sprintf("SELECT * FROM %s.%s LIMIT 1;", req.Msg.Schema, req.Msg.Table)
		r, err := db.QueryContext(cctx, query)
		if err != nil && !nucleusdb.IsNoRows(err) {
			return err
		}

		columnNames, err := r.Columns()
		if err != nil {
			return err
		}

		selectQuery := fmt.Sprintf("SELECT %s FROM %s.%s;", strings.Join(columnNames, ", "), req.Msg.Schema, req.Msg.Table)
		rows, err := db.QueryContext(cctx, selectQuery)
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

	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		err := s.areSchemaAndTableValid(ctx, connection, req.Msg.Schema, req.Msg.Table)
		if err != nil {
			return err
		}

		conn, err := s.sqlConnector.NewPgPoolFromConnectionConfig(config.PgConfig, &connectionTimeout, logger)
		if err != nil {
			return err
		}
		db, err := conn.Open(ctx)
		if err != nil {
			return err
		}
		defer conn.Close()

		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		// used to get column names
		query := fmt.Sprintf("SELECT * FROM %s.%s LIMIT 1;", req.Msg.Schema, req.Msg.Table)
		r, err := db.Query(cctx, query)
		if err != nil && !nucleusdb.IsNoRows(err) {
			return err
		}
		defer r.Close()

		columnNames := []string{}
		for _, col := range r.FieldDescriptions() {
			columnNames = append(columnNames, col.Name)
		}

		selectQuery := fmt.Sprintf("SELECT %s FROM %s.%s;", strings.Join(columnNames, ", "), req.Msg.Schema, req.Msg.Table)
		rows, err := db.Query(cctx, selectQuery)
		if err != nil && !nucleusdb.IsNoRows(err) {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			values := make([][]byte, len(columnNames))
			valuesWrapped := make([]any, 0, len(columnNames))

			for i, col := range r.FieldDescriptions() {
				if col.DataTypeOID == 1082 { // OID for date
					var t time.Time
					ds := DateScanner{val: &t}
					valuesWrapped = append(valuesWrapped, &ds)
				} else {
					valuesWrapped = append(valuesWrapped, &values[i])
				}
			}

			if err := rows.Scan(valuesWrapped...); err != nil {
				return err
			}
			row := map[string][]byte{}
			for i, v := range values {
				col := columnNames[i]
				if r.FieldDescriptions()[i].DataTypeOID == 1082 { // OID for date
					// Convert time.Time value to []byte
					if ds, ok := valuesWrapped[i].(*DateScanner); ok && ds.val != nil {
						row[col] = []byte(ds.val.Format(time.RFC3339))
					} else {
						row[col] = nil
					}
				} else if r.FieldDescriptions()[i].DataTypeOID == 2950 { // OID for UUID
					// Convert the byte slice to a uuid.UUID type
					uuidValue, err := uuid.FromBytes(v)
					if err == nil {
						row[col] = []byte(uuidValue.String())
					} else {
						row[col] = nil
					}
				} else {
					row[col] = v
				}
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
			logger.Error("unable to create AWS S3 client")
			return err
		}
		logger.Info("created AWS S3 client")

		var jobRunId string
		switch id := awsS3StreamCfg.Id.(type) {
		case *mgmtv1alpha1.AwsS3StreamConfig_JobRunId:
			jobRunId = id.JobRunId
		case *mgmtv1alpha1.AwsS3StreamConfig_JobId:
			runId, err := s.getLastestJobRunFromAwsS3(ctx, logger, s3Client, id.JobId, awsS3Config.Bucket, awsS3Config.Region)
			if err != nil {
				return err
			}
			jobRunId = runId
		default:
			return nucleuserrors.NewInternalError("unsupported AWS S3 config id")
		}

		tableName := fmt.Sprintf("%s.%s", req.Msg.Schema, req.Msg.Table)
		path := fmt.Sprintf("workflows/%s/activities/%s/data", jobRunId, tableName)
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
				logger.Info(fmt.Sprintf("0 files found for path: %s", path))
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

				scanner := bufio.NewScanner(gzr)
				for scanner.Scan() {
					line := scanner.Bytes()
					var data map[string]any
					err = json.Unmarshal(line, &data)
					if err != nil {
						result.Body.Close()
						gzr.Close()
						return err
					}

					rowMap := make(map[string][]byte)
					for key, value := range data {
						var byteValue []byte
						if str, ok := value.(string); ok {
							// try converting string directly to []byte
							// prevents quoted strings
							byteValue = []byte(str)
						} else {
							// if not a string use JSON encoding
							byteValue, err = json.Marshal(value)
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

	connectionTimeout := uint32(5)

	switch config := connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		conn, err := s.sqlConnector.NewDbFromConnectionConfig(connection.ConnectionConfig, &connectionTimeout, logger)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		db, err := conn.Open()
		if err != nil {
			return nil, err
		}

		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		dbschema, err := s.mysqlquerier.GetDatabaseSchema(cctx, db)
		if err != nil && !nucleusdb.IsNoRows(err) {
			return nil, err
		} else if err != nil && nucleusdb.IsNoRows(err) {
			return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
				Schemas: []*mgmtv1alpha1.DatabaseColumn{},
			}), nil
		}

		schemas := []*mgmtv1alpha1.DatabaseColumn{}
		for _, col := range dbschema {
			schemas = append(schemas, &mgmtv1alpha1.DatabaseColumn{
				Schema:     col.TableSchema,
				Table:      col.TableName,
				Column:     col.ColumnName,
				DataType:   col.DataType,
				IsNullable: col.IsNullable,
			})
		}

		return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
			Schemas: schemas,
		}), nil

	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		conn, err := s.sqlConnector.NewPgPoolFromConnectionConfig(config.PgConfig, &connectionTimeout, logger)
		if err != nil {
			return nil, err
		}
		db, err := conn.Open(ctx)
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		dbschema, err := s.pgquerier.GetDatabaseSchema(cctx, db)
		if err != nil && !nucleusdb.IsNoRows(err) {
			return nil, err
		} else if err != nil && nucleusdb.IsNoRows(err) {
			return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
				Schemas: []*mgmtv1alpha1.DatabaseColumn{},
			}), nil
		}

		schemas := []*mgmtv1alpha1.DatabaseColumn{}
		for _, col := range dbschema {
			schemas = append(schemas, &mgmtv1alpha1.DatabaseColumn{
				Schema:     col.TableSchema,
				Table:      col.TableName,
				Column:     col.ColumnName,
				DataType:   col.DataType,
				IsNullable: col.IsNullable,
			})
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
		s3Client, err := s.awsManager.NewS3Client(ctx, config.AwsS3Config)
		if err != nil {
			return nil, err
		}
		logger.Info("created S3 AWS session")

		var jobRunId string
		switch id := awsCfg.Id.(type) {
		case *mgmtv1alpha1.AwsS3SchemaConfig_JobRunId:
			jobRunId = id.JobRunId
		case *mgmtv1alpha1.AwsS3SchemaConfig_JobId:
			runId, err := s.getLastestJobRunFromAwsS3(ctx, logger, s3Client, id.JobId, awsS3Config.Bucket, awsS3Config.Region)
			if err != nil {
				return nil, err
			}
			jobRunId = runId
		default:
			return nil, nucleuserrors.NewInternalError("unsupported AWS S3 config id")
		}

		path := fmt.Sprintf("workflows/%s/activities/", jobRunId)

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

				filePath := fmt.Sprintf("%s%s.%s/data", path, schemaTableList[0], schemaTableList[1])
				out, err := s.awsManager.ListObjectsV2(ctx, s3Client, awsS3Config.Region, &s3.ListObjectsV2Input{
					Bucket:  aws.String(awsS3Config.Bucket),
					Prefix:  aws.String(filePath),
					MaxKeys: aws.Int32(1),
				})
				if err != nil {
					return nil, err
				}
				if out == nil {
					break
				}
				item := out.Contents[0]
				result, err := s.awsManager.GetObject(ctx, s3Client, awsS3Config.Region, &s3.GetObjectInput{
					Bucket: aws.String(awsS3Config.Bucket),
					Key:    aws.String(*item.Key),
				})
				if err != nil {
					return nil, err
				}

				gzr, err := gzip.NewReader(result.Body)
				if err != nil {
					result.Body.Close()
					return nil, fmt.Errorf("error creating gzip reader: %w", err)
				}

				scanner := bufio.NewScanner(gzr)
				if scanner.Scan() {
					line := scanner.Bytes()
					var data map[string]any
					err = json.Unmarshal(line, &data)
					if err != nil {
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
				}
				if err := scanner.Err(); err != nil {
					result.Body.Close()
					gzr.Close()
					return nil, err
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

	default:
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
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

	connectionTimeout := uint32(5)

	var td dbschemas.TableDependency
	switch config := connection.Msg.Connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		conn, err := s.sqlConnector.NewDbFromConnectionConfig(connection.Msg.Connection.ConnectionConfig, &connectionTimeout, logger)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		db, err := conn.Open()
		if err != nil {
			return nil, err
		}

		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		allConstraints, err := dbschemas_mysql.GetAllMysqlFkConstraints(s.mysqlquerier, cctx, db, schemas)
		if err != nil && !nucleusdb.IsNoRows(err) {
			return nil, err
		} else if err != nil && nucleusdb.IsNoRows(err) {
			return connect.NewResponse(&mgmtv1alpha1.GetConnectionForeignConstraintsResponse{
				TableConstraints: map[string]*mgmtv1alpha1.ForeignConstraintTables{},
			}), nil
		}
		td = dbschemas_mysql.GetMysqlTableDependencies(allConstraints)

	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		conn, err := s.sqlConnector.NewPgPoolFromConnectionConfig(config.PgConfig, &connectionTimeout, logger)
		if err != nil {
			return nil, err
		}
		db, err := conn.Open(ctx)
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		allConstraints, err := dbschemas_postgres.GetAllPostgresFkConstraints(s.pgquerier, cctx, db, schemas)
		if err != nil && !nucleusdb.IsNoRows(err) {
			return nil, err
		} else if err != nil && nucleusdb.IsNoRows(err) {
			return connect.NewResponse(&mgmtv1alpha1.GetConnectionForeignConstraintsResponse{
				TableConstraints: map[string]*mgmtv1alpha1.ForeignConstraintTables{},
			}), nil
		}
		td = dbschemas_postgres.GetPostgresTableDependencies(allConstraints)

	default:
		return nil, errors.New("unsupported fk connection")
	}

	tableConstraints := map[string]*mgmtv1alpha1.ForeignConstraintTables{}
	for tableName, d := range td {
		tableConstraints[tableName] = &mgmtv1alpha1.ForeignConstraintTables{
			Constraints: []*mgmtv1alpha1.ForeignConstraint{},
		}
		for _, c := range d.Constraints {
			tableConstraints[tableName].Constraints = append(tableConstraints[tableName].Constraints, &mgmtv1alpha1.ForeignConstraint{
				Column: c.Column, IsNullable: c.IsNullable, ForeignKey: &mgmtv1alpha1.ForeignKey{
					Table:  c.ForeignKey.Table,
					Column: c.ForeignKey.Column,
				},
			})
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

	connectionTimeout := uint32(5)

	var pc map[string][]string
	switch config := connection.Msg.Connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		conn, err := s.sqlConnector.NewDbFromConnectionConfig(connection.Msg.Connection.ConnectionConfig, &connectionTimeout, logger)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		db, err := conn.Open()
		if err != nil {
			return nil, err
		}

		cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancel()

		allConstraints, err := dbschemas_mysql.GetAllMysqlPkConstraints(s.mysqlquerier, cctx, db, schemas)
		if err != nil && !nucleusdb.IsNoRows(err) {
			return nil, err
		} else if err != nil && nucleusdb.IsNoRows(err) {
			return connect.NewResponse(&mgmtv1alpha1.GetConnectionPrimaryConstraintsResponse{
				TableConstraints: map[string]*mgmtv1alpha1.PrimaryConstraint{},
			}), nil
		}
		pc = dbschemas_mysql.GetMysqlTablePrimaryKeys(allConstraints)

	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		conn, err := s.sqlConnector.NewPgPoolFromConnectionConfig(config.PgConfig, &connectionTimeout, logger)
		if err != nil {
			return nil, err
		}
		db, err := conn.Open(ctx)
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		allConstraints, err := dbschemas_postgres.GetAllPostgresPkConstraints(s.pgquerier, cctx, db, schemas)
		if err != nil && !nucleusdb.IsNoRows(err) {
			return nil, err
		} else if err != nil && nucleusdb.IsNoRows(err) {
			return connect.NewResponse(&mgmtv1alpha1.GetConnectionPrimaryConstraintsResponse{
				TableConstraints: map[string]*mgmtv1alpha1.PrimaryConstraint{},
			}), nil
		}
		pc = dbschemas_postgres.GetPostgresTablePrimaryKeys(allConstraints)

	default:
		return nil, errors.New("unsupported fk connection")
	}

	tableConstraints := map[string]*mgmtv1alpha1.PrimaryConstraint{}
	for tableName, cols := range pc {
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
		schemaTableMap[fmt.Sprintf("%s.%s", s.Schema, s.Table)] = s
	}

	connectionTimeout := uint32(5)

	createStmtsMap := map[string]string{}
	truncateStmtsMap := map[string]string{}
	switch config := connection.Msg.Connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		conn, err := s.sqlConnector.NewDbFromConnectionConfig(connection.Msg.Connection.ConnectionConfig, &connectionTimeout, logger)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		db, err := conn.Open()
		if err != nil {
			return nil, err
		}

		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if req.Msg.Options.InitSchema {
			for k, v := range schemaTableMap {
				stmt, err := dbschemas_mysql.GetTableCreateStatement(cctx, db, &dbschemas_mysql.GetTableCreateStatementRequest{
					Schema: v.Schema,
					Table:  v.Table,
				})
				if err != nil {
					return nil, err
				}
				createStmtsMap[k] = stmt
			}
		}

		if req.Msg.Options.TruncateBeforeInsert {
			for k, v := range schemaTableMap {
				truncateStmtsMap[k] = dbschemas_mysql.BuildTruncateStatement(v.Schema, v.Table)
			}
		}

	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		conn, err := s.sqlConnector.NewPgPoolFromConnectionConfig(config.PgConfig, &connectionTimeout, logger)
		if err != nil {
			return nil, err
		}
		db, err := conn.Open(ctx)
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if req.Msg.Options.InitSchema {
			for k, v := range schemaTableMap {
				stmt, err := dbschemas_postgres.GetTableCreateStatement(cctx, db, s.pgquerier, v.Schema, v.Table)
				if err != nil {
					return nil, err
				}
				createStmtsMap[k] = stmt
			}
		}

		if req.Msg.Options.TruncateCascade {
			for k, v := range schemaTableMap {
				truncateStmtsMap[k] = dbschemas_postgres.BuildTruncateCascadeStatement(v.Schema, v.Table)
			}
		} else if req.Msg.Options.TruncateBeforeInsert {
			return nil, nucleuserrors.NewNotImplemented("postgres truncate unsupported. table foreig keys required to build truncate statement.")
		}

		for k, v := range schemaTableMap {
			statements := []string{}
			if req.Msg.Options.InitSchema {
				stmt, err := dbschemas_postgres.GetTableCreateStatement(cctx, db, s.pgquerier, v.Schema, v.Table)
				if err != nil {
					return nil, err
				}
				statements = append(statements, stmt)
			}
			if req.Msg.Options.TruncateBeforeInsert {
				if req.Msg.Options.TruncateCascade {
					statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s.%s CASCADE;", v.Schema, v.Table))
				} else {
					statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s.%s;", v.Schema, v.Table))
				}
			}
			createStmtsMap[k] = strings.Join(statements, "\n")
		}
	default:
		return nil, errors.New("unsupported connection config")
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionInitStatementsResponse{
		TableInitStatements:     createStmtsMap,
		TableTruncateStatements: truncateStmtsMap,
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

	default:
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}
	schemaResp, err := s.GetConnectionSchema(ctx, connect.NewRequest(schemaReq))
	if err != nil {
		return nil, err
	}
	return schemaResp.Msg.GetSchemas(), nil
}

// returns the first job run id for a given job that is in S3
func (s *Service) getLastestJobRunFromAwsS3(
	ctx context.Context,
	logger *slog.Logger,
	s3Client *s3.Client,
	jobId, bucket string,
	region *string,
) (string, error) {
	jobRunsResp, err := s.jobService.GetJobRecentRuns(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRecentRunsRequest{
		JobId: jobId,
	}))
	if err != nil {
		return "", err
	}
	jobRuns := jobRunsResp.Msg.GetRecentRuns()

	for i := len(jobRuns) - 1; i >= 0; i-- {
		runId := jobRuns[i].JobRunId
		path := fmt.Sprintf("workflows/%s/activities/", runId)
		output, err := s.awsManager.ListObjectsV2(ctx, s3Client, region, &s3.ListObjectsV2Input{
			Bucket:    aws.String(bucket),
			Prefix:    aws.String(path),
			Delimiter: aws.String("/"),
		})
		if err != nil {
			return "", err
		}
		if output == nil {
			continue
		}
		if *output.KeyCount > 0 {
			logger.Info(fmt.Sprintf("found latest job run: %s", runId))
			return runId, nil
		}
	}
	return "", nucleuserrors.NewInternalError(fmt.Sprintf("unable to find latest job run for job: %s", jobId))
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
