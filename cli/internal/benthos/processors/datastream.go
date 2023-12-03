package processors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/service"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"google.golang.org/protobuf/proto"
)

func init() {
	// Config spec is empty for now as we don't have any dynamic fields.
	configSpec := service.NewConfigSpec()

	constructor := func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
		return newDataStreamProcessor(mgr.Logger(), mgr.Metrics()), nil
	}

	err := service.RegisterProcessor("datastream", configSpec, constructor)
	if err != nil {
		panic(err)
	}
}

//------------------------------------------------------------------------------

type dataSreamProcessor struct {
	logger *service.Logger
}

func newDataStreamProcessor(logger *service.Logger, metrics *service.Metrics) *dataSreamProcessor {
	// The logger and metrics components will already be labelled with the
	// identifier of this component within a config.
	return &dataSreamProcessor{
		logger: logger,
		// countPalindromes: metrics.NewCounter("palindromes"),
	}
}

func (r *dataSreamProcessor) Process(ctx context.Context, m *service.Message) (service.MessageBatch, error) {
	bytesContent, err := m.AsBytes()
	if err != nil {
		return nil, err
	}
	fmt.Println(string(bytesContent))

	resp := &mgmtv1alpha1.GetConnectionDataStreamResponse{}

	if err := proto.Unmarshal(bytesContent, resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf message '%v': %w", resp, err)
	}
	valuesMap := map[string]any{}
	for col, value := range resp.Row {
		fmt.Println(col)
		switch value.Kind.(type) {
		case *mgmtv1alpha1.Value_StringValue:
			valuesMap[col] = value.GetStringValue()
		case *mgmtv1alpha1.Value_NumberValue:
			valuesMap[col] = value.GetNumberValue()
		case *mgmtv1alpha1.Value_BoolValue:
			valuesMap[col] = value.GetBoolValue()
		default:
			fmt.Println("default")

			//TODO
			// Handle other types or set a default value if necessary
		}
	}
	jsonData, err := json.Marshal(valuesMap)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(jsonData))
	m.SetBytes(jsonData)
	return []*service.Message{m}, nil
}

func (r *dataSreamProcessor) Close(ctx context.Context) error {
	return nil
}
