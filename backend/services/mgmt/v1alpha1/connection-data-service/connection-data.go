package v1alpha1_connectiondataservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
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
	initStatementsResponse, err := connectiondatabuilder.GetInitStatements(ctx, req.Msg.GetOptions())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(initStatementsResponse), nil
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

	connectiondatabuilder, err := s.connectiondatabuilder.NewDataConnection(logger, dbconnectionResp.Msg.GetConnection())
	if err != nil {
		return nil, err
	}
	dbcols, err := connectiondatabuilder.GetTableSchema(ctx, req.Msg.GetTable().GetSchema(), req.Msg.GetTable().GetTable())
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
	tableConstraints, err := connectiondatabuilder.GetTableConstraints(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(tableConstraints), nil
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
	connectiondatabuilder, err := s.connectiondatabuilder.NewDataConnection(logger, connection.Msg.GetConnection())
	if err != nil {
		return nil, err
	}
	count, err := connectiondatabuilder.GetTableRowCount(ctx, req.Msg.Schema, req.Msg.Table, req.Msg.WhereClause)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetTableRowCountResponse{
		Count: count,
	}), nil
}
