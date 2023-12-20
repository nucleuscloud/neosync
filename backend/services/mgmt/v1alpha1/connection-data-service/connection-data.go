package v1alpha1_connectiondataservice

import (
	"bufio"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	aws_s3 "github.com/nucleuscloud/neosync/backend/internal/aws/s3"
	aws_session "github.com/nucleuscloud/neosync/backend/internal/aws/session"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	conn_utils "github.com/nucleuscloud/neosync/backend/internal/utils/connections"
	dbschemas_mysql "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/mysql"
	dbschemas_postgres "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/postgres"
	"golang.org/x/sync/errgroup"
)

const (
	mysqlDriver    = "mysql"
	postgresDriver = "postgres"

	getPostgresTableSchemaSql = `-- name: GetPostgresTableSchema
	SELECT
	c.table_schema,
	c.table_name,
	c.column_name,
	c.data_type
	FROM
		information_schema.columns AS c
		JOIN information_schema.tables AS t ON c.table_schema = t.table_schema
			AND c.table_name = t.table_name
	WHERE
		c.table_schema NOT IN ('pg_catalog', 'information_schema')
		AND t.table_type = 'BASE TABLE';
`

	getMysqlTableSchemaSql = `-- name: GetMysqlTableSchema
	SELECT
	c.table_schema,
	c.table_name,
	c.column_name,
	c.data_type
	FROM
		information_schema.columns AS c
		JOIN information_schema.tables AS t ON c.table_schema = t.table_schema
			AND c.table_name = t.table_name
	WHERE
		c.table_schema NOT IN ('sys', 'performance_schema', 'mysql')
		AND t.table_type = 'BASE TABLE';
`
)

type DatabaseSchema struct {
	TableSchema string `db:"table_schema,omitempty"`
	TableName   string `db:"table_name,omitempty"`
	ColumnName  string `db:"column_name,omitempty"`
	DataType    string `db:"data_type,omitempty"`
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

	switch config := req.Msg.StreamConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionStreamConfig_MysqlConfig, *mgmtv1alpha1.ConnectionStreamConfig_PgConfig:
		// check that schema and table are valid
		schemas, err := s.getConnectionSchema(ctx, connection, &SchemaOpts{})
		if err != nil {
			return err
		}

		if !isValidSchema(req.Msg.Schema, schemas) || !isValidTable(req.Msg.Table, schemas) {
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

	switch config := req.Msg.SchemaConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionSchemaConfig_MysqlConfig:
		connCfg := connection.ConnectionConfig
		connDetails, err := s.getConnectionDetails(connCfg)
		if err != nil {
			return nil, err
		}

		conn, err := s.sqlConnector.Open(connDetails.ConnectionDriver, connDetails.ConnectionString)
		if err != nil {
			logger.Error("unable to connect", err)
			return nil, err
		}
		defer func() {
			if err := conn.Close(); err != nil {
				logger.Error(fmt.Errorf("failed to close sql connection: %w", err).Error())
			}
		}()
		cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancel()

		dbSchema, err := getDatabaseSchema(cctx, conn, getMysqlTableSchemaSql)
		if err != nil {
			return nil, err
		}

		return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
			Schemas: ToDatabaseColumn(dbSchema),
		}), nil

	case *mgmtv1alpha1.ConnectionSchemaConfig_PgConfig:
		connCfg := connection.ConnectionConfig
		connDetails, err := s.getConnectionDetails(connCfg)
		if err != nil {
			return nil, err
		}

		conn, err := s.sqlConnector.Open(connDetails.ConnectionDriver, connDetails.ConnectionString)
		if err != nil {
			logger.Error("unable to connect", err)
			return nil, err
		}
		defer func() {
			if err := conn.Close(); err != nil {
				logger.Error(fmt.Errorf("failed to close sql connection: %w", err).Error())
			}
		}()
		cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancel()

		dbSchema, err := getDatabaseSchema(cctx, conn, getPostgresTableSchemaSql)
		if err != nil {
			return nil, err
		}

		return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
			Schemas: ToDatabaseColumn(dbSchema),
		}), nil

	case *mgmtv1alpha1.ConnectionSchemaConfig_AwsS3Config:
		var jobRunId string
		switch id := config.AwsS3Config.Id.(type) {
		case *mgmtv1alpha1.AwsS3SchemaConfig_JobRunId:
			jobRunId = id.JobRunId
		case *mgmtv1alpha1.AwsS3SchemaConfig_JobId:
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

		schema := []*mgmtv1alpha1.DatabaseColumn{}
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
				schema = append(schema, &mgmtv1alpha1.DatabaseColumn{
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
		return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
			Schemas: schema,
		}), nil

	default:
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}
}

func getDatabaseSchema(ctx context.Context, conn *sql.DB, query string) ([]DatabaseSchema, error) {
	rows, err := conn.QueryContext(ctx, query)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	}
	if err != nil && nucleusdb.IsNoRows(err) {
		return []DatabaseSchema{}, nil
	}

	output := []DatabaseSchema{}
	for rows.Next() {
		var o DatabaseSchema
		err := rows.Scan(
			&o.TableSchema,
			&o.TableName,
			&o.ColumnName,
			&o.DataType,
		)
		if err != nil {
			return nil, err
		}
		output = append(output, o)
	}
	return output, nil
}

func ToDatabaseColumn(
	input []DatabaseSchema,
) []*mgmtv1alpha1.DatabaseColumn {
	columns := []*mgmtv1alpha1.DatabaseColumn{}
	for _, col := range input {
		columns = append(columns, &mgmtv1alpha1.DatabaseColumn{
			Schema:   col.TableSchema,
			Table:    col.TableName,
			Column:   col.ColumnName,
			DataType: col.DataType,
		})
	}
	return columns
}

func (s *Service) GetConnectionForeignConstraints(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionForeignConstraintsRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionForeignConstraintsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionId", req.Msg.ConnectionId)
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

	connDetails, err := s.getConnectionDetails(connection.Msg.Connection.ConnectionConfig)
	if err != nil {
		return nil, err
	}

	schemaResp, err := s.getConnectionSchema(ctx, connection.Msg.Connection, &SchemaOpts{})
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

	var td map[string][]string
	switch connDetails.ConnectionDriver {
	case postgresDriver:
		pgquerier := pg_queries.New()
		pool, err := pgxpool.New(ctx, connDetails.ConnectionString)
		if err != nil {
			return nil, err
		}
		cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancel()
		allConstraints, err := getAllPostgresFkConstraints(pgquerier, cctx, pool, schemas)
		if err != nil {
			return nil, err
		}
		td = dbschemas_postgres.GetPostgresTableDependencies(allConstraints)
	case mysqlDriver:
		mysqlquerier := mysql_queries.New()
		conn, err := s.sqlConnector.Open(connDetails.ConnectionDriver, connDetails.ConnectionString)
		if err != nil {
			logger.Error("unable to connect", err)
			return nil, err
		}
		defer func() {
			if err := conn.Close(); err != nil {
				logger.Error(fmt.Errorf("failed to close connection: %w", err).Error())
			}
		}()
		cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancel()
		allConstraints, err := getAllMysqlFkConstraints(mysqlquerier, cctx, conn, schemas)
		if err != nil {
			return nil, err
		}
		td = dbschemas_mysql.GetMysqlTableDependencies(allConstraints)
	default:
		return nil, errors.New("unsupported fk connection")
	}

	constraints := map[string]*mgmtv1alpha1.ForeignConstraintTables{}
	for key, tables := range td {
		constraints[key] = &mgmtv1alpha1.ForeignConstraintTables{
			Tables: tables,
		}
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionForeignConstraintsResponse{
		TableConstraints: constraints,
	}), nil
}

func (s *Service) GetConnectionInitStatements(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionInitStatementsRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionInitStatementsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionId", req.Msg.ConnectionId)
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

	connDetails, err := s.getConnectionDetails(connection.Msg.Connection.ConnectionConfig)
	if err != nil {
		return nil, err
	}

	schemaResp, err := s.getConnectionSchema(ctx, connection.Msg.Connection, &SchemaOpts{})
	if err != nil {
		return nil, err
	}

	schemaTableMap := map[string]*mgmtv1alpha1.DatabaseColumn{}
	for _, s := range schemaResp {
		schemaTableMap[fmt.Sprintf("%s.%s", s.Schema, s.Table)] = s
	}

	statementsMap := map[string]string{}
	switch connDetails.ConnectionDriver {
	case postgresDriver:
		pgquerier := pg_queries.New()
		pool, err := pgxpool.New(ctx, connDetails.ConnectionString)
		if err != nil {
			return nil, err
		}
		cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancel()
		for k, v := range schemaTableMap {
			statements := []string{}
			if req.Msg.Options.InitSchema {
				stmt, err := dbschemas_postgres.GetTableCreateStatement(cctx, pool, pgquerier, v.Schema, v.Table)
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
			statementsMap[k] = strings.Join(statements, "\n")
		}
	case mysqlDriver:
		conn, err := s.sqlConnector.Open(connDetails.ConnectionDriver, connDetails.ConnectionString)
		if err != nil {
			logger.Error("unable to connect", err)
			return nil, err
		}
		defer func() {
			if err := conn.Close(); err != nil {
				logger.Error(fmt.Errorf("failed to close connection: %w", err).Error())
			}
		}()
		cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancel()
		for k, v := range schemaTableMap {
			statements := []string{}
			if req.Msg.Options.InitSchema {
				stmt, err := dbschemas_mysql.GetTableCreateStatement(cctx, conn, &dbschemas_mysql.GetTableCreateStatementRequest{
					Schema: v.Schema,
					Table:  v.Table,
				})
				if err != nil {
					return nil, err
				}
				statements = append(statements, stmt)
			}
			if req.Msg.Options.TruncateBeforeInsert {
				statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s.%s;", v.Schema, v.Table))
			}
			statementsMap[k] = strings.Join(statements, "\n")
		}
	default:
		return nil, errors.New("unsupported connection config")
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionInitStatementsResponse{
		TableInitStatements: statementsMap,
	}), nil
}

func getAllPostgresFkConstraints(
	pgquerier pg_queries.Querier,
	ctx context.Context,
	conn pg_queries.DBTX,
	uniqueSchemas []string,
) ([]*pg_queries.GetForeignKeyConstraintsRow, error) {
	holder := make([][]*pg_queries.GetForeignKeyConstraintsRow, len(uniqueSchemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range uniqueSchemas {
		idx := idx
		schema := uniqueSchemas[idx]
		errgrp.Go(func() error {
			constraints, err := pgquerier.GetForeignKeyConstraints(errctx, conn, schema)
			if err != nil {
				return err
			}
			holder[idx] = constraints
			return nil
		})
	}

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	output := []*pg_queries.GetForeignKeyConstraintsRow{}
	for _, schemas := range holder {
		output = append(output, schemas...)
	}
	return output, nil
}

func getAllMysqlFkConstraints(
	mysqlquerier mysql_queries.Querier,
	ctx context.Context,
	conn *sql.DB,
	schemas []string,
) ([]*mysql_queries.GetForeignKeyConstraintsRow, error) {
	holder := make([][]*mysql_queries.GetForeignKeyConstraintsRow, len(schemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range schemas {
		idx := idx
		schema := schemas[idx]
		errgrp.Go(func() error {
			constraints, err := mysqlquerier.GetForeignKeyConstraints(errctx, conn, schema)
			if err != nil {
				return err
			}
			holder[idx] = constraints
			return nil
		})
	}

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	output := []*mysql_queries.GetForeignKeyConstraintsRow{}
	for _, schemas := range holder {
		output = append(output, schemas...)
	}
	return output, nil
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

type SchemaOpts struct {
	JobId    *string
	JobRunId *string
}

func (s *Service) getConnectionSchema(ctx context.Context, connection *mgmtv1alpha1.Connection, opts *SchemaOpts) ([]*mgmtv1alpha1.DatabaseColumn, error) {
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
