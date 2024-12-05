package neosync_benthos_sql

import (
	"context"
	"fmt"

	"github.com/warpstreamlabs/bento/public/service"
)

func neosyncToMysqlProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().Field(service.NewStringMapField("column_data_types"))
}

func RegisterNeosyncToMysqlProcessor(env *service.Environment) error {
	return env.RegisterBatchProcessor(
		"neosync_to_mysql",
		neosyncToMysqlProcessorConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			proc, err := newNeosyncToMysqlProcessor(conf, mgr)
			if err != nil {
				return nil, err
			}
			return proc, nil
		})
}

type neosyncToMysqlProcessor struct {
	logger          *service.Logger
	columnDataTypes map[string]string
}

func newNeosyncToMysqlProcessor(conf *service.ParsedConfig, mgr *service.Resources) (*neosyncToMysqlProcessor, error) {
	columnDataTypes, err := conf.FieldStringMap("column_data_types")
	if err != nil {
		return nil, err
	}
	return &neosyncToMysqlProcessor{
		logger:          mgr.Logger(),
		columnDataTypes: columnDataTypes,
	}, nil
}

func (p *neosyncToMysqlProcessor) ProcessBatch(ctx context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
	newBatch := make(service.MessageBatch, 0, len(batch))
	for _, msg := range batch {
		root, err := msg.AsStructuredMut()
		if err != nil {
			return nil, err
		}
		newRoot := p.transform("", root)
		newMsg := msg.Copy()
		newMsg.SetStructured(newRoot)
		newBatch = append(newBatch, newMsg)
	}

	if len(newBatch) == 0 {
		return nil, nil
	}
	return []service.MessageBatch{newBatch}, nil
}

func (m *neosyncToMysqlProcessor) Close(context.Context) error {
	return nil
}

func (p *neosyncToMysqlProcessor) transform(path string, root any) any {
	switch v := root.(type) {
	case map[string]any:
		newMap := make(map[string]any)
		for k, v2 := range v {
			newValue := p.transform(k, v2)
			newMap[k] = newValue
		}
		return newMap
	case nil:
		return v
	case []byte:
		datatype, ok := p.columnDataTypes[path]
		if !ok {
			return v
		}
		value, err := p.handleByteSlice(v, datatype)
		if err != nil {
			p.logger.Errorf("unable to handle byte slice: %w", err)
			return v
		}
		return value
	default:
		return v
	}
}

func (p *neosyncToMysqlProcessor) handleByteSlice(v []byte, datatype string) (any, error) {
	switch datatype {
	case "bit":
		bit, err := convertStringToBit(string(v))
		if err != nil {
			return nil, fmt.Errorf("unable to convert bit string to SQL bit []byte: %w", err)
		}
		return bit, nil
	}
	return v, nil
}
