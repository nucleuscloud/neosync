package neosync_benthos_json

import (
	"context"
	"time"

	"github.com/nucleuscloud/neosync/internal/sqlscanners"
	"github.com/warpstreamlabs/bento/public/service"
)

func sqlToJsonProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec()
}

func RegisterSqlToJsonProcessor(env *service.Environment) error {
	return env.RegisterBatchProcessor(
		"sql_to_json",
		sqlToJsonProcessorConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			proc := newMysqlToJsonProcessor(conf, mgr)
			return proc, nil
		})
}

type sqlToJsonProcessor struct {
	logger *service.Logger
}

func newMysqlToJsonProcessor(_ *service.ParsedConfig, mgr *service.Resources) *sqlToJsonProcessor {
	return &sqlToJsonProcessor{
		logger: mgr.Logger(),
	}
}

func (m *sqlToJsonProcessor) ProcessBatch(ctx context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
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

func (m *sqlToJsonProcessor) Close(context.Context) error {
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
		return v.Format(time.DateTime)
	case []uint8:
		return string(v)
	case *sqlscanners.BitString:
		return v.String()
	default:
		return v
	}
}
