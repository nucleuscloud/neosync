package neosync_benthos_sql

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/lib/pq"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	pgutil "github.com/nucleuscloud/neosync/internal/postgres"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/warpstreamlabs/bento/public/service"
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

func newNeosyncToPgxProcessor(conf *service.ParsedConfig, mgr *service.Resources) (*neosyncToPgxProcessor, error) {
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

func (p *neosyncToPgxProcessor) ProcessBatch(ctx context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
	newBatch := make(service.MessageBatch, 0, len(batch))
	for _, msg := range batch {
		root, err := msg.AsStructuredMut()
		if err != nil {
			return nil, err
		}
		newRoot, err := transformNeosyncToPgx(p.logger, root, p.columns, p.columnDataTypes, p.columnDefaultProperties)
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
		newVal, err := getPgxValue(val, colDefaults, datatype)
		if err != nil {
			logger.Warn(err.Error())
		}

		newMap[col] = newVal
	}

	return newMap, nil
}

func getPgxValue(value any, colDefaults *neosync_benthos.ColumnDefaultProperties, datatype string) (any, error) {
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
	case pgutil.IsJsonPgDataType(datatype):
		bits, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal JSON: %w", err)
		}
		return bits, nil
	case pgutil.IsPgArrayColumnDataType(datatype):
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
			return goqu.Literal(pgutil.FormatPgArrayLiteral(value, datatype)), nil
		} else if gotypeutil.IsSlice(value) {
			return pq.Array(value), nil
		}
		return value, nil
	// case datatype == "bytea":
	// 	if b64String, ok := value.(string); ok {
	// 		bytes, err := base64.StdEncoding.DecodeString(b64String)
	// 		if err != nil {
	// 			return nil, fmt.Errorf("unable to decode base64 string: %w", err)
	// 		}
	// 		return bytes, nil
	// 	}
	// 	return value, nil
	case datatype == "date":
		return convertDateForPostgres(value)
	case datatype == "timestamp with time zone":
		return convertTimestampWithTimezoneForPostgres(value), nil
	case datatype == "timestamp" || datatype == "timestamp without time zone":
		return convertTimestampForPostgres(value)
	case datatype == "money" || datatype == "uuid" || datatype == "tsvector" || datatype == "time with time zone":
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

func convertBitsToTime(input any) (time.Time, error) {
	var timeStr string
	switch v := input.(type) {
	case []byte:
		timeStr = string(v)
	case string:
		timeStr = v
	default:
		return time.Time{}, fmt.Errorf("unsupported type for time conversion: %T", input)
	}

	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// Try parsing as DateTime format if not RFC3339 format
		t, err = time.Parse(time.DateTime, timeStr)
		if err != nil {
			return time.Time{}, err
		}
	}

	return t, nil
}

func convertDateForPostgres(input any) (string, error) {
	return convertTimeForPostgres(input, time.DateOnly)
}

func convertTimestampForPostgres(input any) (string, error) {
	return convertTimeForPostgres(input, time.DateTime)
}

// pgtypes does not handle BC dates correctly
// convertTimeForPostgres handles BC dates properly
func convertTimeForPostgres(input any, layout string) (string, error) {
	var timeStr string
	switch v := input.(type) {
	case []byte:
		timeStr = string(v)
	case string:
		timeStr = v
	default:
		return "", fmt.Errorf("unsupported type for time conversion: %T", input)
	}

	if strings.HasPrefix(timeStr, "-") {
		t, err := time.Parse("-2006-01-02T15:04:05Z", timeStr)
		if err != nil {
			return "", err
		}
		// For negative years, add 1 to get correct BC year
		// year -1 is 2 BC, year -2 is 3 BC, etc.
		yearsToAdd := t.Year() + 1

		newT := time.Date(yearsToAdd, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
		return fmt.Sprintf("%s BC", newT.Format(layout)), nil
	}

	t, err := convertBitsToTime(timeStr)
	if err != nil {
		return "", err
	}
	// Handle BC dates year 0
	// year 0 is 1 BC,
	if t.Year() == 0 {
		newT := time.Date(1, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
		return fmt.Sprintf("%s BC", newT.Format(layout)), nil
	}
	return t.Format(layout), nil
}

// pgtype.Timestamptz does not support BC dates, so we need to reformat them
func convertTimestampWithTimezoneForPostgres(input any) string {
	var timeStr string
	switch v := input.(type) {
	case []byte:
		timeStr = string(v)
	case string:
		timeStr = v
	default:
		return fmt.Sprintf("%v", input) // Fallback to string representation
	}

	// Remove the 'T'
	withoutT := strings.Replace(timeStr, "T", " ", 1)

	// Handle year 0000 case (should become 0001 BC)
	if strings.HasPrefix(withoutT, "0000") {
		rest := withoutT[4:]
		return "0001" + rest[:len(rest)-6] + rest[len(rest)-6:] + " BC"
	}

	// If starts with '-', remove it and add BC before timezone
	if strings.HasPrefix(withoutT, "-") {
		yearNum, _ := strconv.Atoi(withoutT[1:5])
		yearNum++ // Add 1 to convert from Go BC year to Postgres BC year
		year := fmt.Sprintf("%04d", yearNum)
		rest := withoutT[5:]
		return year + rest[:len(rest)-6] + rest[len(rest)-6:] + " BC"
	}

	return timeStr
}

// this expects the bits to be in the form [1,2,3]
func processPgArrayFromJson(bits []byte, datatype string) (any, error) {
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

func isColumnInList(column string, columns []string) bool {
	return slices.Contains(columns, column)
}

func getColumnDefaultProperties(columnDefaultPropertiesConfig map[string]*service.ParsedConfig) (map[string]*neosync_benthos.ColumnDefaultProperties, error) {
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
