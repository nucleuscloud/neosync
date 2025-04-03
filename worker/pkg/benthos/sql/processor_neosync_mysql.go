package neosync_benthos_sql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/redpanda-data/benthos/v4/public/service"
)

func neosyncToMysqlProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringListField("columns")).
		Field(service.NewStringMapField("column_data_types")).
		Field(service.NewAnyMapField("column_default_properties"))
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
	logger                  *service.Logger
	columns                 []string
	columnDataTypes         map[string]string
	columnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties
}

func newNeosyncToMysqlProcessor(
	conf *service.ParsedConfig,
	mgr *service.Resources,
) (*neosyncToMysqlProcessor, error) {
	columns, err := conf.FieldStringList("columns")
	if err != nil {
		return nil, err
	}

	columnDataTypes, err := conf.FieldStringMap("column_data_types")
	if err != nil {
		return nil, err
	}

	columnDefaultPropertiesConfig, err := conf.FieldAnyMap("column_default_properties")
	if err != nil {
		return nil, err
	}

	columnDefaultProperties, err := getColumnDefaultProperties(columnDefaultPropertiesConfig)
	if err != nil {
		return nil, err
	}

	return &neosyncToMysqlProcessor{
		logger:                  mgr.Logger(),
		columns:                 columns,
		columnDataTypes:         columnDataTypes,
		columnDefaultProperties: columnDefaultProperties,
	}, nil
}

func (p *neosyncToMysqlProcessor) ProcessBatch(
	ctx context.Context,
	batch service.MessageBatch,
) ([]service.MessageBatch, error) {
	newBatch := make(service.MessageBatch, 0, len(batch))
	for _, msg := range batch {
		root, err := msg.AsStructuredMut()
		if err != nil {
			return nil, err
		}
		newRoot, err := transformNeosyncToMysql(
			root,
			p.columns,
			p.columnDataTypes,
			p.columnDefaultProperties,
		)
		if err != nil {
			return nil, err
		}
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

func transformNeosyncToMysql(
	root any,
	columns []string,
	columnDataTypes map[string]string,
	columnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties,
) (map[string]any, error) {
	rootMap, ok := root.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("root value must be a map[string]any")
	}

	newMap := make(map[string]any)
	for col, val := range rootMap {
		// Skip values that aren't in the column list to handle circular references
		if !isColumnInList(col, columns) {
			continue
		}
		colDefaults := columnDefaultProperties[col]
		datatype := columnDataTypes[col]
		newVal, err := getMysqlValue(val, colDefaults, datatype)
		if err != nil {
			return nil, fmt.Errorf("failed to get MySQL value for column %s: %w", col, err)
		}
		newMap[col] = newVal
	}

	return newMap, nil
}

func getMysqlValue(
	value any,
	colDefaults *neosync_benthos.ColumnDefaultProperties,
	datatype string,
) (any, error) {
	if colDefaults != nil && colDefaults.HasDefaultTransformer {
		return goqu.Default(), nil
	}

	if value == nil {
		return nil, nil
	}

	value, isNeosyncValue, err := getMysqlNeosyncValue(value)
	if err != nil {
		return nil, fmt.Errorf("unable to get MySQL value from neosync value: %w", err)
	}
	if isNeosyncValue {
		return value, nil
	}

	switch datatype {
	case "json":
		if v, ok := value.([]byte); ok {
			validJson, err := getValidJson(v)
			if err != nil {
				return nil, fmt.Errorf("unable to get valid json: %w", err)
			}
			return validJson, nil
		}
		if value == "null" {
			return value, nil
		}
		bits, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal mysql json to bits: %w", err)
		}
		return bits, nil
	default:
		return value, nil
	}
}

func getMysqlNeosyncValue(root any) (value any, isNeosyncValue bool, err error) {
	if valuer, ok := root.(neosynctypes.NeosyncMysqlValuer); ok {
		value, err := valuer.ValueMysql()
		if err != nil {
			return nil, false, fmt.Errorf(
				"unable to get MYSQL value from NeosyncMysqlValuer: %w",
				err,
			)
		}
		return value, true, nil
	}
	return root, false, nil
}
