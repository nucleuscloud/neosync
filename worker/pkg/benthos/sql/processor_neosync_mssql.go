package neosync_benthos_sql

import (
	"context"

	"github.com/warpstreamlabs/bento/public/service"
)

func neosyncToMssqlProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().Field(service.NewStringMapField("column_data_types"))
}

func RegisterNeosyncToMssqlProcessor(env *service.Environment) error {
	return env.RegisterBatchProcessor(
		"neosync_to_mssql",
		neosyncToMssqlProcessorConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			proc, err := newNeosyncToMssqlProcessor(conf, mgr)
			if err != nil {
				return nil, err
			}
			return proc, nil
		})
}

type neosyncToMssqlProcessor struct {
	logger          *service.Logger
	columnDataTypes map[string]string
}

func newNeosyncToMssqlProcessor(conf *service.ParsedConfig, mgr *service.Resources) (*neosyncToMssqlProcessor, error) {
	columnDataTypes, err := conf.FieldStringMap("column_data_types")
	if err != nil {
		return nil, err
	}
	return &neosyncToMssqlProcessor{
		logger:          mgr.Logger(),
		columnDataTypes: columnDataTypes,
	}, nil
}

func (p *neosyncToMssqlProcessor) ProcessBatch(ctx context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
	newBatch := make(service.MessageBatch, 0, len(batch))
	for _, msg := range batch {
		root, err := msg.AsStructuredMut()
		if err != nil {
			return nil, err
		}
		newRoot := p.transform(root)
		newMsg := msg.Copy()
		newMsg.SetStructured(newRoot)
		newBatch = append(newBatch, newMsg)
	}

	if len(newBatch) == 0 {
		return nil, nil
	}
	return []service.MessageBatch{newBatch}, nil
}

func (m *neosyncToMssqlProcessor) Close(context.Context) error {
	return nil
}

func (p *neosyncToMssqlProcessor) transform(root any) any {
	switch v := root.(type) {
	case map[string]any:
		newMap := make(map[string]any)
		for k, v2 := range v {
			newValue := p.transform(v2)
			newMap[k] = newValue
		}
		return newMap
	case nil:
		return v
	default:
		return v
	}
}
