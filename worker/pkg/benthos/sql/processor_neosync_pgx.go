package neosync_benthos_sql

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/lib/pq"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/redpanda-data/benthos/v4/public/service"
)

func neosyncToPgxProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringListField("columns")).
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
	columns                 []string
	columnDataTypes         map[string]string
	columnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties
}

func newNeosyncToPgxProcessor(
	conf *service.ParsedConfig,
	mgr *service.Resources,
) (*neosyncToPgxProcessor, error) {
	columnDataTypes, err := conf.FieldStringMap("column_data_types")
	if err != nil {
		return nil, err
	}

	columns, err := conf.FieldStringList("columns")
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

	return &neosyncToPgxProcessor{
		logger:                  mgr.Logger(),
		columns:                 columns,
		columnDataTypes:         columnDataTypes,
		columnDefaultProperties: columnDefaultProperties,
	}, nil
}

func (p *neosyncToPgxProcessor) ProcessBatch(
	ctx context.Context,
	batch service.MessageBatch,
) ([]service.MessageBatch, error) {
	newBatch := make(service.MessageBatch, 0, len(batch))
	for _, msg := range batch {
		root, err := msg.AsStructuredMut()
		if err != nil {
			return nil, err
		}
		newRoot, err := transformNeosyncToPgx(
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

func (m *neosyncToPgxProcessor) Close(context.Context) error {
	return nil
}
func transformNeosyncToPgx(
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
		newVal, err := getPgxValue(val, colDefaults, datatype)
		if err != nil {
			return nil, fmt.Errorf("failed to get PGX value for column %s: %w", col, err)
		}
		newMap[col] = newVal
	}

	return newMap, nil
}

func getPgxValue(
	value any,
	colDefaults *neosync_benthos.ColumnDefaultProperties,
	datatype string,
) (any, error) {
	value, isNeosyncValue, err := getPgxNeosyncValue(value)
	if err != nil {
		return nil, err
	}
	if isNeosyncValue {
		return value, nil
	}

	if colDefaults != nil && colDefaults.HasDefaultTransformer {
		return goqu.Default(), nil
	}

	if value == nil {
		return nil, nil
	}

	switch {
	case strings.EqualFold(datatype, "json") || strings.EqualFold(datatype, "jsonb"):
		if value == "null" {
			return value, nil
		}
		bits, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal postgres json to bits: %w", err)
		}
		return bits, nil
	case strings.HasSuffix(datatype, "[]"):
		if byteSlice, ok := value.([]byte); ok {
			// this handles the case where the array is in the form {1,2,3}
			if strings.HasPrefix(string(byteSlice), "{") {
				return string(byteSlice), nil
			}
			pgarray, err := processPgArrayFromJson(byteSlice, datatype)
			if err != nil {
				return nil, fmt.Errorf("unable to process PG Array: %w", err)
			}
			return pgarray, nil
		} else if gotypeutil.IsMultiDimensionalSlice(value) || gotypeutil.IsSliceOfMaps(value) {
			return goqu.Literal(formatPgArrayLiteral(value, datatype)), nil
		} else if gotypeutil.IsSlice(value) {
			return pq.Array(value), nil
		}
		return value, nil
	case datatype == "money" || datatype == "uuid" || datatype == "tsvector":
		if byteSlice, ok := value.([]byte); ok {
			// Convert UUID []byte to string before inserting since postgres driver stores uuid bytes in different order
			return string(byteSlice), nil
		}
		return value, nil
	default:
		return value, nil
	}
}

func getPgxNeosyncValue(root any) (value any, isNeosyncValue bool, err error) {
	if valuer, ok := root.(neosynctypes.NeosyncPgxValuer); ok {
		value, err := valuer.ValuePgx()
		if err != nil {
			return nil, false, fmt.Errorf("unable to get PGX value from NeosyncPgxValuer: %w", err)
		}
		return value, true, nil
	}
	return root, false, nil
}

// this expects the bits to be in the form [1,2,3]
func processPgArrayFromJson(bits []byte, datatype string) (any, error) {
	var pgarray []any
	err := json.Unmarshal(bits, &pgarray)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal postgres array from bits: %w", err)
	}
	switch datatype {
	case "json[]", "jsonb[]":
		jsonArray, err := stringifyJsonArray(pgarray)
		if err != nil {
			return nil, fmt.Errorf("unable to stringify postgres array: %w", err)
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
		return nil, fmt.Errorf("unable to marshal postgres json string to bits: %w", err)
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

func isColumnInList(column string, columns []string) bool {
	return slices.Contains(columns, column)
}

func getColumnDefaultProperties(
	columnDefaultPropertiesConfig map[string]*service.ParsedConfig,
) (map[string]*neosync_benthos.ColumnDefaultProperties, error) {
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
	return columnDefaultProperties, nil
}

// returns string in this form ARRAY[[a,b],[c,d]]
func formatPgArrayLiteral(arr any, castType string) string {
	arrayLiteral := "ARRAY" + formatArrayLiteral(arr)
	if castType == "" {
		return arrayLiteral
	}

	return arrayLiteral + "::" + castType
}

func formatArrayLiteral(arr any) string {
	v := reflect.ValueOf(arr)

	if v.Kind() == reflect.Slice {
		result := "["
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				result += ","
			}
			result += formatArrayLiteral(v.Index(i).Interface())
		}
		result += "]"
		return result
	}

	switch val := arr.(type) {
	case map[string]any:
		return formatMapLiteral(val)
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "''"))
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatMapLiteral(m map[string]any) string {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("%v", m)
	}

	return fmt.Sprintf("'%s'", string(jsonBytes))
}
