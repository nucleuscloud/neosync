package neosync_benthos_sql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/warpstreamlabs/bento/public/service"
)

func neosyncToMssqlProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringListField("columns")).
		Field(service.NewStringMapField("column_data_types")).
		Field(service.NewAnyMapField("column_default_properties"))
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
	logger                  *service.Logger
	columns                 []string
	columnDataTypes         map[string]string
	columnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties
}

func newNeosyncToMssqlProcessor(conf *service.ParsedConfig, mgr *service.Resources) (*neosyncToMssqlProcessor, error) {
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

	return &neosyncToMssqlProcessor{
		logger:                  mgr.Logger(),
		columns:                 columns,
		columnDataTypes:         columnDataTypes,
		columnDefaultProperties: columnDefaultProperties,
	}, nil
}

func (p *neosyncToMssqlProcessor) ProcessBatch(ctx context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
	newBatch := make(service.MessageBatch, 0, len(batch))
	for _, msg := range batch {
		root, err := msg.AsStructuredMut()
		if err != nil {
			return nil, err
		}
		newRoot, err := transformNeosyncToMssql(p.logger, root, p.columns, p.columnDefaultProperties)
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

func (m *neosyncToMssqlProcessor) Close(context.Context) error {
	return nil
}

func transformNeosyncToMssql(
	logger *service.Logger,
	root any,
	columns []string,
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
		// sqlserver doesn't support default values. must be removed
		if colDefaults != nil && colDefaults.HasDefaultTransformer {
			continue
		}

		newVal, err := getMssqlValue(val)
		if err != nil {
			logger.Warn(err.Error())
		}
		newMap[col] = newVal
	}

	return newMap, nil
}

func getMssqlValue(value any) (any, error) {
	value, isNeosyncValue, err := getMssqlNeosyncValue(value)
	if err != nil {
		return nil, err
	}
	if isNeosyncValue {
		return value, nil
	}
	if gotypeutil.IsMap(value) {
		bits, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal go map to json bits: %w", err)
		}
		return bits, nil
	}

	return value, nil
}

func getMssqlNeosyncValue(root any) (value any, isNeosyncValue bool, err error) {
	if valuer, ok := root.(neosynctypes.NeosyncMssqlValuer); ok {
		value, err := valuer.ValueMssql()
		if err != nil {
			return nil, false, fmt.Errorf("unable to get MSSQL value from NeosyncMssqlValuer: %w", err)
		}
		return value, true, nil
	}
	return root, false, nil
}
