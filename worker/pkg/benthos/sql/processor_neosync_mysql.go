package neosync_benthos_sql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	mysqlutil "github.com/nucleuscloud/neosync/internal/mysql"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/warpstreamlabs/bento/public/service"
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

func newNeosyncToMysqlProcessor(conf *service.ParsedConfig, mgr *service.Resources) (*neosyncToMysqlProcessor, error) {
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

func (p *neosyncToMysqlProcessor) ProcessBatch(ctx context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
	newBatch := make(service.MessageBatch, 0, len(batch))
	for _, msg := range batch {
		root, err := msg.AsStructuredMut()
		if err != nil {
			return nil, err
		}
		jsonF, _ := json.MarshalIndent(root, "", " ")
		fmt.Printf("\n\n root: %s \n\n", string(jsonF))
		newRoot, err := transformNeosyncToMysql(p.logger, root, p.columns, p.columnDataTypes, p.columnDefaultProperties)
		if err != nil {
			return nil, err
		}
		jsonF, _ = json.MarshalIndent(newRoot, "", " ")
		fmt.Printf("\n\n newRoot: %s \n\n", string(jsonF))
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
	logger *service.Logger,
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
			logger.Warn(err.Error())
		}
		newMap[col] = newVal
	}

	return newMap, nil
}

func getMysqlValue(value any, colDefaults *neosync_benthos.ColumnDefaultProperties, datatype string) (any, error) {
	if colDefaults != nil && colDefaults.HasDefaultTransformer {
		return goqu.Default(), nil
	}

	switch v := value.(type) {
	case nil:
		return v, nil
	case []byte:
		value, err := handleMysqlByteSlice(v, datatype)
		if err != nil {
			return nil, fmt.Errorf("unable to handle byte slice: %w", err)
		}
		return value, nil
	default:
		if mysqlutil.IsJsonDataType(datatype) {
			bits, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("unable to marshal JSON: %w", err)
			}
			return bits, nil
		}
		return v, nil
	}
}

func handleMysqlByteSlice(v []byte, datatype string) (any, error) {
	if datatype == "bit" {
		bit, err := convertStringToBit(string(v))
		if err != nil {
			return nil, fmt.Errorf("unable to convert bit string to SQL bit []byte: %w", err)
		}
		return bit, nil
	}
	return v, nil
}
