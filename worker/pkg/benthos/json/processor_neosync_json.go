package neosync_benthos_json

import (
	"context"
	"time"

	"github.com/redpanda-data/benthos/v4/public/service"
)

func neosyncToJsonProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec()
}

func RegisterNeosyncToJsonProcessor(env *service.Environment) error {
	return env.RegisterBatchProcessor(
		"neosync_to_json",
		neosyncToJsonProcessorConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			proc := newNeosyncToJsonProcessor(conf, mgr)
			return proc, nil
		})
}

type neosyncToJsonProcessor struct {
	logger *service.Logger
}

func newNeosyncToJsonProcessor(
	_ *service.ParsedConfig,
	mgr *service.Resources,
) *neosyncToJsonProcessor {
	return &neosyncToJsonProcessor{
		logger: mgr.Logger(),
	}
}

func (m *neosyncToJsonProcessor) ProcessBatch(
	ctx context.Context,
	batch service.MessageBatch,
) ([]service.MessageBatch, error) {
	newBatch := make(service.MessageBatch, 0, len(batch))
	for _, msg := range batch {
		root, err := msg.AsStructuredMut()
		if err != nil {
			return nil, err
		}
		newRoot := transform(root)
		newMsg := msg.Copy()
		newMsg.SetStructured(newRoot)
		newBatch = append(newBatch, newMsg)
	}

	if len(newBatch) == 0 {
		return nil, nil
	}
	return []service.MessageBatch{newBatch}, nil
}

func (m *neosyncToJsonProcessor) Close(context.Context) error {
	return nil
}

func transform(root any) any {
	switch v := root.(type) {
	case map[string]any:
		newMap := make(map[string]any)
		for k, v2 := range v {
			newValue := transform(v2)
			newMap[k] = newValue
		}
		return newMap
	case []any:
		newSlice := make([]any, len(v))
		for i, v2 := range v {
			newSlice[i] = transform(v2)
		}
		return newSlice
	case time.Time:
		return v.Format(time.RFC3339)
	case []uint8:
		return string(v)
	default:
		return v
	}
}
