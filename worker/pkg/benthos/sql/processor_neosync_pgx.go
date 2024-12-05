package neosync_benthos_sql

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/doug-martin/goqu/v9"
	"github.com/lib/pq"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	pgutil "github.com/nucleuscloud/neosync/internal/postgres"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/warpstreamlabs/bento/public/service"
)

const defaultStr = "DEFAULT"

func neosyncToPgxProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringMapField("column_data_types")).
		Field(service.NewAnyMapField("column_default_properties"))
}

func RegisterNeosyncToPgxProcessor(env *service.Environment) error {
	return env.RegisterBatchProcessor(
		"neosync_to_pgx",
		neosyncToPgxProcessorConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			proc, err := newNeosyncToPgxProcessor(conf, mgr)
			if err != nil {
				return nil, err
			}
			return proc, nil
		})
}

type neosyncToPgxProcessor struct {
	logger                  *service.Logger
	columnDataTypes         map[string]string
	columnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties
}

func newNeosyncToPgxProcessor(conf *service.ParsedConfig, mgr *service.Resources) (*neosyncToPgxProcessor, error) {
	columnDataTypes, err := conf.FieldStringMap("column_data_types")
	if err != nil {
		return nil, err
	}

	columnDefaultPropertiesConfig, err := conf.FieldAnyMap("column_default_properties")
	if err != nil {
		return nil, err
	}

	columnDefaultProperties := map[string]*neosync_benthos.ColumnDefaultProperties{}
	for key, properties := range columnDefaultPropertiesConfig {
		props, err := properties.FieldAny()
		if err != nil {
			return nil, err
		}
		jsonData, err := json.Marshal(props)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal properties for key %s: %w", key, err)
		}

		var colDefaults neosync_benthos.ColumnDefaultProperties
		if err := json.Unmarshal(jsonData, &colDefaults); err != nil {
			return nil, fmt.Errorf("failed to unmarshal properties for key %s: %w", key, err)
		}

		columnDefaultProperties[key] = &colDefaults
	}

	return &neosyncToPgxProcessor{
		logger:                  mgr.Logger(),
		columnDataTypes:         columnDataTypes,
		columnDefaultProperties: columnDefaultProperties,
	}, nil
}

func (p *neosyncToPgxProcessor) ProcessBatch(ctx context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
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

func (m *neosyncToPgxProcessor) Close(context.Context) error {
	return nil
}

func (p *neosyncToPgxProcessor) transform(path string, root any) any {
	value, isNeosyncValue, err := p.getNeosyncValue(root)
	if err != nil {
		p.logger.Warn(err.Error())
	}
	if isNeosyncValue {
		return value
	}

	colDefaults := p.columnDefaultProperties[path]
	if colDefaults != nil && colDefaults.HasDefaultTransformer {
		return goqu.Literal(defaultStr)
	}

	datatype := p.columnDataTypes[path]
	if pgutil.IsJsonPgDataType(datatype) {
		bits, err := json.Marshal(root)
		if err != nil {
			p.logger.Errorf("unable to marshal JSON", "error", err.Error())
			return root
		}
		return bits
	}

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
		value, err := p.handleByteSlice(v, datatype)
		if err != nil {
			p.logger.Errorf("unable to handle byte slice: %w", err)
			return v
		}
		return value
	default:
		if gotypeutil.IsMultiDimensionalSlice(v) || gotypeutil.IsSliceOfMaps(v) {
			return goqu.Literal(pgutil.FormatPgArrayLiteral(v, datatype))
		} else if gotypeutil.IsSlice(v) {
			return pq.Array(v)
		}
		return v
	}
}

func (p *neosyncToPgxProcessor) getNeosyncValue(root any) (value any, isNeosyncValue bool, err error) {
	if valuer, ok := root.(neosynctypes.NeosyncPgxValuer); ok {
		value, err := valuer.ValuePgx()
		if err != nil {
			return nil, false, fmt.Errorf("unable to get PGX value from NeosyncPgxValuer: %w", err)
		}
		if gotypeutil.IsSlice(value) {
			return pq.Array(value), true, nil
		}
		return value, true, nil
	}
	return root, false, nil
}

func (p *neosyncToPgxProcessor) handleByteSlice(v []byte, datatype string) (any, error) {
	if pgutil.IsPgArrayColumnDataType(datatype) {
		pgarray, err := processPgArray(v, datatype)
		if err != nil {
			return nil, fmt.Errorf("unable to process PG Array: %w", err)
		}
		return pgarray, nil
	}
	switch datatype {
	case "bit":
		bit, err := convertStringToBit(string(v))
		if err != nil {
			return nil, fmt.Errorf("unable to convert bit string to SQL bit []byte: %w", err)
		}
		return bit, nil
	case "json", "jsonb":
		validJson, err := getValidJson(v)
		if err != nil {
			return nil, fmt.Errorf("unable to get valid json: %w", err)
		}
		return validJson, nil
	case "money", "uuid", "time with time zone", "timestamp with time zone":
		// Convert UUID []byte to string before inserting since postgres driver stores uuid bytes in different order
		return string(v), nil
	}
	return v, nil
}

func processPgArray(bits []byte, datatype string) (any, error) {
	var pgarray []any
	err := json.Unmarshal(bits, &pgarray)
	if err != nil {
		return nil, err
	}
	switch datatype {
	case "json[]", "jsonb[]":
		jsonArray, err := stringifyJsonArray(pgarray)
		if err != nil {
			return nil, err
		}
		return pq.Array(jsonArray), nil
	default:
		return pq.Array(pgarray), nil
	}
}

// handles case where json strings are not quoted
func getValidJson(jsonData []byte) ([]byte, error) {
	isValidJson := json.Valid(jsonData)
	if isValidJson {
		return jsonData, nil
	}

	quotedData, err := json.Marshal(string(jsonData))
	if err != nil {
		return nil, err
	}
	return quotedData, nil
}

func stringifyJsonArray(pgarray []any) ([]string, error) {
	jsonArray := make([]string, len(pgarray))
	for i, item := range pgarray {
		bytes, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		jsonArray[i] = string(bytes)
	}
	return jsonArray, nil
}

func convertStringToBit(bitString string) ([]byte, error) {
	val, err := strconv.ParseUint(bitString, 2, len(bitString))
	if err != nil {
		return nil, err
	}

	// Always allocate 8 bytes for PutUint64
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, val)

	// Calculate actual needed bytes and return only those
	neededBytes := (len(bitString) + 7) / 8
	return bytes[len(bytes)-neededBytes:], nil
}
