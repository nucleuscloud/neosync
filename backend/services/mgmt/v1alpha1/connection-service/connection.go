package v1alpha1_connectionservice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	conn_utils "github.com/nucleuscloud/neosync/backend/internal/utils/connections"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"golang.org/x/sync/errgroup"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	dbschemas_mysql "github.com/nucleuscloud/neosync/backend/internal/dbschemas/mysql"
	dbschemas_postgres "github.com/nucleuscloud/neosync/backend/internal/dbschemas/postgres"
)

const (
	mysqlDriver    = "mysql"
	postgresDriver = "postgres"
)

func (s *Service) CheckConnectionConfig(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest],
) (*connect.Response[mgmtv1alpha1.CheckConnectionConfigResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	connDetails, err := s.getConnectionDetails(req.Msg.ConnectionConfig)
	if err != nil {
		return nil, err
	}
	conn, err := s.sqlConnector.Open(connDetails.ConnectionDriver, connDetails.ConnectionString)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error(fmt.Errorf("failed to close mysql connection: %w", err).Error())
		}
	}()
	cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer cancel()
	err = conn.PingContext(cctx)
	if err != nil {
		msg := err.Error()
		return connect.NewResponse(&mgmtv1alpha1.CheckConnectionConfigResponse{
			IsConnected:     false,
			ConnectionError: &msg,
		}), nil
	}
	return connect.NewResponse(&mgmtv1alpha1.CheckConnectionConfigResponse{
		IsConnected:     true,
		ConnectionError: nil,
	}), nil
}

func (s *Service) IsConnectionNameAvailable(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.IsConnectionNameAvailableRequest],
) (*connect.Response[mgmtv1alpha1.IsConnectionNameAvailableResponse], error) {
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	count, err := s.db.Q.IsConnectionNameAvailable(ctx, s.db.Db, db_queries.IsConnectionNameAvailableParams{
		AccountId:      *accountUuid,
		ConnectionName: req.Msg.ConnectionName,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.IsConnectionNameAvailableResponse{
		IsAvailable: count == 0,
	}), nil
}

func (s *Service) GetConnections(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionsRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionsResponse], error) {
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	connections, err := s.db.Q.GetConnectionsByAccount(ctx, s.db.Db, *accountUuid)
	if err != nil {
		return nil, err
	}

	dtoConns := []*mgmtv1alpha1.Connection{}
	for idx := range connections {
		connection := connections[idx]
		dtoConns = append(dtoConns, dtomaps.ToConnectionDto(&connection))
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionsResponse{
		Connections: dtoConns,
	}), nil
}

func (s *Service) GetConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
	idUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	connection, err := s.db.Q.GetConnectionById(ctx, s.db.Db, idUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find connection by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(connection.AccountID))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: dtomaps.ToConnectionDto(&connection),
	}), nil
}

func (s *Service) CreateConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateConnectionRequest],
) (*connect.Response[mgmtv1alpha1.CreateConnectionResponse], error) {
	cc := &pg_models.ConnectionConfig{}
	if err := cc.FromDto(req.Msg.ConnectionConfig); err != nil {
		return nil, err
	}

	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	connection, err := s.db.Q.CreateConnection(ctx, s.db.Db, db_queries.CreateConnectionParams{
		AccountID:        *accountUuid,
		Name:             req.Msg.Name,
		ConnectionConfig: cc,
		CreatedByID:      *userUuid,
		UpdatedByID:      *userUuid,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.CreateConnectionResponse{
		Connection: dtomaps.ToConnectionDto(&connection),
	}), nil
}

func (s *Service) UpdateConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateConnectionRequest],
) (*connect.Response[mgmtv1alpha1.UpdateConnectionResponse], error) {
	connectionUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	connection, err := s.db.Q.GetConnectionById(ctx, s.db.Db, connectionUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find connection by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(connection.AccountID))
	if err != nil {
		return nil, err
	}

	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	cc := &pg_models.ConnectionConfig{}
	if err := cc.FromDto(req.Msg.ConnectionConfig); err != nil {
		return nil, err
	}

	connection, err = s.db.Q.UpdateConnection(ctx, s.db.Db, db_queries.UpdateConnectionParams{
		ID:               connection.ID,
		ConnectionConfig: cc,
		UpdatedByID:      *userUuid,
		Name:             req.Msg.Name,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.UpdateConnectionResponse{
		Connection: dtomaps.ToConnectionDto(&connection),
	}), nil
}

func (s *Service) DeleteConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteConnectionRequest],
) (*connect.Response[mgmtv1alpha1.DeleteConnectionResponse], error) {
	idUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	connection, err := s.db.Q.GetConnectionById(ctx, s.db.Db, idUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteConnectionResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(connection.AccountID))
	if err != nil {
		return nil, err
	}

	err = s.db.Q.RemoveConnectionById(ctx, s.db.Db, connection.ID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.DeleteConnectionResponse{}), nil
}

func (s *Service) CheckSqlQuery(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CheckSqlQueryRequest],
) (*connect.Response[mgmtv1alpha1.CheckSqlQueryResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionId", req.Msg.Id)
	connection, err := s.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{Id: req.Msg.Id}))
	if err != nil {
		return nil, err
	}

	connDetails, err := s.getConnectionDetails(connection.Msg.Connection.ConnectionConfig)
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
			logger.Error(fmt.Errorf("failed to close connection: %w", err).Error())
		}
	}()
	tx, err := conn.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer nucleusdb.HandleSqlRollback(tx, logger)

	_, err = tx.PrepareContext(ctx, req.Msg.Query)
	var errorMsg *string
	if err != nil {
		msg := err.Error()
		errorMsg = &msg
	}
	return connect.NewResponse(&mgmtv1alpha1.CheckSqlQueryResponse{
		IsValid:      err == nil,
		ErorrMessage: errorMsg,
	}), nil
}

func (s *Service) GetConnectionDataStream(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionDataStreamRequest],
	stream *connect.ServerStream[mgmtv1alpha1.GetConnectionDataStreamResponse],
) error {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionId", req.Msg.SourceConnectionId)
	sourceConn, err := s.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.SourceConnectionId,
	}))
	if err != nil {
		return err
	}
	_, err = s.verifyUserInAccount(ctx, sourceConn.Msg.Connection.AccountId)
	if err != nil {
		return err
	}

	connCfg := sourceConn.Msg.Connection.ConnectionConfig
	connDetails, err := s.getConnectionDetails(connCfg)
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
	query := fmt.Sprintf("SELECT * FROM %s.%s LIMIT 1;", req.Msg.Schema, req.Msg.Table)
	r, err := conn.QueryContext(ctx, query)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return err
	}

	columnNames, err := r.Columns()
	if err != nil {
		return err
	}

	columnTypes, err := r.ColumnTypes()
	if err != nil {
		return err
	}

	selectQuery := fmt.Sprintf("SELECT %s FROM %s.%s", strings.Join(columnNames, ", "), req.Msg.Schema, req.Msg.Table)
	rows, err := conn.QueryContext(ctx, selectQuery)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return err
	}

	for rows.Next() {
		columnPointers := make([]interface{}, len(columnNames))
		for i := range columnNames {
			columnPointers[i] = new(interface{})
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return err
		}

		row := map[string]*mgmtv1alpha1.Value{}

		for i, name := range columnNames {
			val := *columnPointers[i].(*interface{})
			value := &mgmtv1alpha1.Value{}

			if val == nil {
				value.Kind = &mgmtv1alpha1.Value_NullValue{}
				row[name] = value
				continue
			}

			// Get the PostgreSQL data type of the current column
			dbType := columnTypes[i].DatabaseTypeName()
			fmt.Println(name)
			fmt.Println(dbType)
			fmt.Println(val)

			switch strings.ToLower(dbType) {
			case "text", "varchar", "char", "citext", "json", "jsonb", "uuid":
				value.Kind = &mgmtv1alpha1.Value_StringValue{StringValue: fmt.Sprintf("%v", val)}
			case "bpchar": // Handling BPCHAR separately
				byteSlice, ok := val.([]uint8)
				if !ok {
					// Handle the error if val is not a byte slice
					value.Kind = &mgmtv1alpha1.Value_NullValue{}
					continue
				}
				strValue := string(byteSlice)
				value.Kind = &mgmtv1alpha1.Value_StringValue{StringValue: strValue}
			case "int", "int2", "int4", "int8", "serial", "serial2", "serial4", "serial8":
				value.Kind = &mgmtv1alpha1.Value_NumberValue{NumberValue: float64(val.(int64))}
			case "float4", "float8", "decimal":
				value.Kind = &mgmtv1alpha1.Value_NumberValue{NumberValue: float64(val.(int64))}
			case "numeric":
				// Convert the byte slice to string first
				byteSlice, ok := val.([]uint8)
				if !ok {
					// Handle the error if val is not a byte slice
					// For example, set value.Kind to a null value or an error value
					value.Kind = &mgmtv1alpha1.Value_NullValue{}
					continue
				}
				// Parse the string to a float64
				strValue := string(byteSlice)
				floatValue, err := strconv.ParseFloat(strValue, 64)
				if err != nil {
					// Handle the error if the string cannot be parsed to float64
					// For example, set value.Kind to a null value or an error value
					value.Kind = &mgmtv1alpha1.Value_NullValue{}
					continue
				}
				value.Kind = &mgmtv1alpha1.Value_NumberValue{NumberValue: floatValue}
			case "date", "timestamp": // Handling date and timestamp types
				timeVal, ok := val.(time.Time)
				if !ok {
					value.Kind = &mgmtv1alpha1.Value_NullValue{}
					continue
				}
				strValue := timeVal.Format(time.RFC3339) // You can change the format as needed
				value.Kind = &mgmtv1alpha1.Value_StringValue{StringValue: strValue}
			case "bool":
				value.Kind = &mgmtv1alpha1.Value_BoolValue{BoolValue: val.(bool)}
			default:
				value.Kind = &mgmtv1alpha1.Value_NullValue{}
			}
			row[name] = value
		}

		if err := stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{Row: row}); err != nil {
			return err
		}
	}
	return nil

}

func (s *Service) GetConnectionForeignConstraints(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionForeignConstraintsRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionForeignConstraintsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionId", req.Msg.ConnectionId)
	connection, err := s.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{Id: req.Msg.ConnectionId}))
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

	schemaResp, err := s.GetConnectionSchema(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{Id: req.Msg.ConnectionId}))
	if err != nil {
		return nil, err
	}

	schemaMap := map[string]struct{}{}
	for _, s := range schemaResp.Msg.GetSchemas() {
		schemaMap[s.Schema] = struct{}{}
	}
	schemas := []string{}
	for s := range schemaMap {
		schemas = append(schemas, s)
	}

	var td map[string][]string
	switch connDetails.ConnectionDriver {
	case "postgres":
		pgquerier := pg_queries.New()
		pool, err := pgxpool.New(ctx, connDetails.ConnectionString)
		if err != nil {
			return nil, err
		}
		allConstraints, err := getAllPostgresFkConstraints(pgquerier, ctx, pool, schemas)
		if err != nil {
			return nil, err
		}
		td = dbschemas_postgres.GetPostgresTableDependencies(allConstraints)
	case "mysql":
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
		allConstraints, err := getAllMysqlFkConstraints(mysqlquerier, ctx, conn, schemas)
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

func getAllPostgresFkConstraints(
	pgquerier pg_queries.Querier,
	ctx context.Context,
	conn pg_queries.DBTX,
	// conn *sql.DB,
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

// func mapDatabaseTypeToValueFieldType(dbType string) mgmtv1alpha1.isValue_Kind {
// 	// Map PostgreSQL types to corresponding Value field types
// 	switch dbType {
// 	case "text", "varchar", "char", "citext", "json", "jsonb", "uuid":
// 		return &mgmtv1alpha1.Value_StringValue{}
// 	case "int", "int2", "int4", "int8", "serial", "serial2", "serial4", "serial8":
// 		return mgmtv1alpha1.Value_NUMBER_VALUE
// 	case "float4", "float8", "numeric", "decimal":
// 		return mgmtv1alpha1.Value_NUMBER_VALUE
// 	case "bool":
// 		return mgmtv1alpha1.Value_BOOL_VALUE
// 	case "struct", "hstore":
// 		return mgmtv1alpha1.Value_STRUCT_VALUE
// 	// Add more cases for other PostgreSQL types as needed
// 	default:
// 		return mgmtv1alpha1.Value_NULL_VALUE
// 	}
// }

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
