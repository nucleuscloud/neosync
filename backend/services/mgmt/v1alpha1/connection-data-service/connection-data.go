package v1alpha1_connectiondataservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sqlmanager_mysql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mysql"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	neosyncgob "github.com/nucleuscloud/neosync/internal/gob"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	neosyncgob.RegisterGobTypes()
}

// GetConnectionDataStream streams data from a connection source (e.g. MySQL, Postgres, S3, etc)
// The data is first converted from its native format into Go types, then encoded using gob encoding
// before being streamed back to the client.
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

	connectiondatabuilder, err := s.connectiondatabuilder.NewDataConnection(logger, connection)
	if err != nil {
		return err
	}
	err = connectiondatabuilder.StreamData(ctx, stream, req.Msg.StreamConfig, req.Msg.Schema, req.Msg.Table)
	if err != nil {
		return err
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

	connectiondatabuilder, err := s.connectiondatabuilder.NewDataConnection(logger, connection)
	if err != nil {
		return nil, err
	}
	schemas, err := connectiondatabuilder.GetSchema(ctx, req.Msg.GetSchemaConfig())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
		Schemas: schemas,
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

	connectiondatabuilder, err := s.connectiondatabuilder.NewDataConnection(logger, connection.Msg.GetConnection())
	if err != nil {
		return nil, err
	}
	schemas, err := connectiondatabuilder.GetSchema(ctx, nil)
	if err != nil {
		return nil, err
	}

	schemaTableMap := map[string]*mgmtv1alpha1.DatabaseColumn{}
	for _, s := range schemas {
		schemaTableMap[sqlmanager_shared.BuildTable(s.Schema, s.Table)] = s
	}

	db, err := s.sqlmanager.NewSqlConnection(ctx, connectionmanager.NewUniqueSession(), connection.Msg.GetConnection(), logger)
	if err != nil {
		return nil, err
	}
	defer db.Db().Close()

	createStmtsMap := map[string]string{}
	truncateStmtsMap := map[string]string{}
	initSchemaStmts := []*mgmtv1alpha1.SchemaInitStatements{}
	if req.Msg.GetOptions().GetInitSchema() {
		tables := []*sqlmanager_shared.SchemaTable{}
		for k, v := range schemaTableMap {
			stmt, err := db.Db().GetCreateTableStatement(ctx, v.Schema, v.Table)
			if err != nil {
				return nil, err
			}
			createStmtsMap[k] = stmt
			tables = append(tables, &sqlmanager_shared.SchemaTable{Schema: v.Schema, Table: v.Table})
		}
		initBlocks, err := db.Db().GetSchemaInitStatements(ctx, tables)
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

func (s *Service) getConnectionTableSchema(ctx context.Context, connection *mgmtv1alpha1.Connection, schema, table string, logger *slog.Logger) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	conntimeout := uint32(5)
	switch connection.GetConnectionConfig().Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		conn, err := s.sqlConnector.NewDbFromConnectionConfig(connection.GetConnectionConfig(), logger, sqlconnect.WithConnectionTimeout(conntimeout))
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
		conn, err := s.sqlConnector.NewDbFromConnectionConfig(connection.GetConnectionConfig(), logger, sqlconnect.WithConnectionTimeout(conntimeout))
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

	connectiondatabuilder, err := s.connectiondatabuilder.NewDataConnection(logger, connection.Msg.GetConnection())
	if err != nil {
		return nil, err
	}
	schemaDbCols, err := connectiondatabuilder.GetSchema(ctx, nil)
	if err != nil {
		return nil, err
	}

	schemaMap := map[string]struct{}{}
	for _, s := range schemaDbCols {
		schemaMap[s.Schema] = struct{}{}
	}
	schemas := []string{}
	for s := range schemaMap {
		schemas = append(schemas, s)
	}

	switch connection.Msg.GetConnection().GetConnectionConfig().GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		db, err := s.sqlmanager.NewSqlConnection(ctx, connectionmanager.NewUniqueSession(), connection.Msg.GetConnection(), logger)
		if err != nil {
			return nil, err
		}
		defer db.Db().Close()
		tableConstraints, err := db.Db().GetTableConstraintsBySchema(ctx, schemas)
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

	switch connection.Msg.GetConnection().GetConnectionConfig().Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		db, err := s.sqlmanager.NewSqlConnection(ctx, connectionmanager.NewUniqueSession(), connection.Msg.GetConnection(), logger)
		if err != nil {
			return nil, err
		}
		defer db.Db().Close()

		count, err := db.Db().GetTableRowCount(ctx, req.Msg.Schema, req.Msg.Table, req.Msg.WhereClause)
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
