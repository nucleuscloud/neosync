package input

import (
	"context"
	"net/http"
	"sync"

	"connectrpc.com/connect"
	"github.com/benthosdev/benthos/v4/public/service"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
)

var neosyncConnectionDataConfigSpec = service.NewConfigSpec().
	Summary("Creates an input that generates garbage.").
	Field(service.NewStringField("api_key").Optional()).
	Field(service.NewStringField("api_url")).
	Field(service.NewStringField("connection_id")).
	Field(service.NewStringField("schema")).
	Field(service.NewStringField("table"))

func newNeosyncConnectionDataInput(conf *service.ParsedConfig) (service.Input, error) {
	var apiKey *string
	if conf.Contains("api_key") {
		apiKeyStr, err := conf.FieldString("api_key")
		if err != nil {
			return nil, err
		}
		apiKey = &apiKeyStr
	}

	apiUrl, err := conf.FieldString("api_url")
	if err != nil {
		return nil, err
	}

	connectionId, err := conf.FieldString("connection_id")
	if err != nil {
		return nil, err
	}

	schema, err := conf.FieldString("schema")
	if err != nil {
		return nil, err
	}
	table, err := conf.FieldString("table")
	if err != nil {
		return nil, err
	}
	return service.AutoRetryNacks(&neosyncInput{
		apiKey:       apiKey,
		apiUrl:       apiUrl,
		connectionId: connectionId,
		schema:       schema,
		table:        table,
	}), nil
}

func init() {
	err := service.RegisterInput(
		"neosync_connection_data", neosyncConnectionDataConfigSpec,
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Input, error) {
			return newNeosyncConnectionDataInput(conf)
		})
	if err != nil {
		panic(err)
	}
}

//------------------------------------------------------------------------------

type neosyncInput struct {
	apiKey *string
	apiUrl string

	connectionId string
	schema       string
	table        string

	neosyncConnectApi mgmtv1alpha1connect.ConnectionServiceClient

	recvMut sync.Mutex

	resp *connect.ServerStreamForClient[mgmtv1alpha1.GetConnectionDataStreamResponse]
}

func (g *neosyncInput) Connect(ctx context.Context) error {
	g.neosyncConnectApi = mgmtv1alpha1connect.NewConnectionServiceClient(
		http.DefaultClient,
		g.apiUrl,
		connect.WithInterceptors(auth_interceptor.NewInterceptor(g.apiKey != nil, auth.AuthHeader, auth.GetAuthHeaderTokenFn(g.apiKey))),
	)

	resp, err := g.neosyncConnectApi.GetConnectionDataStream(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionDataStreamRequest{
		SourceConnectionId: g.connectionId,
		Schema:             g.schema,
		Table:              g.table,
	}))
	if err != nil {
		return err
	}
	g.resp = resp
	return nil
}

func (g *neosyncInput) Read(ctx context.Context) (*service.Message, service.AckFunc, error) {
	g.recvMut.Lock()
	defer g.recvMut.Unlock()

	if g.neosyncConnectApi == nil && g.resp == nil {
		return nil, nil, service.ErrNotConnected
	}
	if g.resp == nil {
		return nil, nil, service.ErrEndOfInput
	}

	ok := g.resp.Receive()
	if !ok {
		err := g.resp.Err()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, service.ErrEndOfInput
	}
	valuesMap := map[string]any{}
	row := g.resp.Msg().Row

	for col, value := range row {
		switch value.Kind.(type) {
		case *mgmtv1alpha1.Value_StringValue:
			valuesMap[col] = value.GetStringValue()
		case *mgmtv1alpha1.Value_NumberValue:
			valuesMap[col] = value.GetNumberValue()
		case *mgmtv1alpha1.Value_BoolValue:
			valuesMap[col] = value.GetBoolValue()
		case *mgmtv1alpha1.Value_NullValue:
			valuesMap[col] = value.GetNullValue()
		case *mgmtv1alpha1.Value_ListValue:
			valuesMap[col] = value.GetListValue()
		default:
		}
	}

	msg := service.NewMessage(nil)
	msg.SetStructuredMut(valuesMap)
	return msg, func(ctx context.Context, err error) error {
		// Nacks are retried automatically when we use service.AutoRetryNacks
		return nil
	}, nil
}

func (g *neosyncInput) Close(ctx context.Context) error {
	// close client
	// todo: prob need mutex
	if g.resp != nil {
		err := g.resp.Close()
		if err != nil {
			return err
		}
		g.resp = nil
	}

	g.neosyncConnectApi = nil // idk if this really matters
	return nil
}
