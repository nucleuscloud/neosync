package neosync_benthos_sql

import (
	"context"

	pgutil "github.com/nucleuscloud/neosync/internal/postgres"
	"github.com/warpstreamlabs/bento/public/service"
)

func jsonToSqlProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().Field(service.NewStringMapField("column_data_types"))
}

func RegisterJsonToSqlProcessor(env *service.Environment) error {
	return env.RegisterBatchProcessor(
		"json_to_sql",
		jsonToSqlProcessorConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			proc, err := newJsonToSqlProcessor(conf, mgr)
			if err != nil {
				return nil, err
			}
			return proc, nil
		})
}

type jsonToSqlProcessor struct {
	logger          *service.Logger
	columnDataTypes map[string]string // column name to datatype
}

func newJsonToSqlProcessor(conf *service.ParsedConfig, mgr *service.Resources) (*jsonToSqlProcessor, error) {
	columnDataTypes, err := conf.FieldStringMap("column_data_types")
	if err != nil {
		return nil, err
	}
	return &jsonToSqlProcessor{
		logger:          mgr.Logger(),
		columnDataTypes: columnDataTypes,
	}, nil
}

func (p *jsonToSqlProcessor) ProcessBatch(ctx context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
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

func (m *jsonToSqlProcessor) Close(context.Context) error {
	return nil
}

func (p *jsonToSqlProcessor) transform(path string, root any) any {
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
		// TODO move to pgx processor
		if pgutil.IsPgArrayColumnDataType(datatype) {
			pgarray, err := processPgArray(v, datatype)
			if err != nil {
				p.logger.Errorf("unable to process PG Array: %w", err)
				return v
			}
			return pgarray
		}
		switch datatype {
		case "bit":
			bit, err := convertStringToBit(string(v))
			if err != nil {
				p.logger.Errorf("unable to convert bit string to SQL bit []byte: %w", err)
				return v
			}
			return bit
		case "json", "jsonb":
			validJson, err := getValidJson(v)
			if err != nil {
				p.logger.Errorf("unable to get valid json: %w", err)
				return v
			}
			return validJson
		case "money", "uuid", "time with time zone", "timestamp with time zone":
			// Convert UUID []byte to string before inserting since postgres driver stores uuid bytes in different order
			return string(v)
		}
		return v
	default:
		return v
	}
}
