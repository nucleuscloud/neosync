package neosync_benthos_sql

import (
	"context"
	"encoding/binary"
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

	switch v := value.(type) {
	case nil:
		return v, nil
	case []byte:
		value, err := handlePgxByteSlice(v, datatype)
		if err != nil {
			return nil, fmt.Errorf("unable to handle byte slice: %w", err)
		}
		return value, nil
	default:
		if pgutil.IsJsonPgDataType(datatype) {
			bits, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("unable to marshal JSON: %w", err)
			}
			return bits, nil
		} else if gotypeutil.IsMultiDimensionalSlice(v) || gotypeutil.IsSliceOfMaps(v) {
			return goqu.Literal(pgutil.FormatPgArrayLiteral(v, datatype)), nil
		} else if gotypeutil.IsSlice(v) {
			return pq.Array(v), nil
		}
		return v, nil
	}
}

func getPgxNeosyncValue(root any) (value any, isNeosyncValue bool, err error) {
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

func handlePgxByteSlice(v []byte, datatype string) (any, error) {
	if pgutil.IsPgArrayColumnDataType(datatype) {
		// this handles the case where the array is in the form {1,2,3}
		if strings.HasPrefix(string(v), "{") {
			return string(v), nil
		}
		pgarray, err := processPgArrayFromJson(v, datatype)
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
	case "date":
		return convertDateForPostgres(v)
	case "timestamp with time zone":
		return convertTimestampWithTimezoneForPostgres(v), nil
	case "timestamp", "timestamp without time zone":
		return convertTimestampForPostgres(v)
	case "money", "uuid", "tsvector", "time with time zone":
		// Convert UUID []byte to string before inserting since postgres driver stores uuid bytes in different order
		return string(v), nil
	}
	return v, nil
}

func convertBitsToTime(bits []byte) (time.Time, error) {
	timeStr := string(bits)
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

func convertDateForPostgres(input []byte) (string, error) {
	return convertTimeForPostgres(input, time.DateOnly)
}

func convertTimestampForPostgres(input []byte) (string, error) {
	return convertTimeForPostgres(input, time.DateTime)
}

// pgtypes does not handle BC dates correctly
// convertTimeForPostgres handles BC dates properly
func convertTimeForPostgres(timebits []byte, layout string) (string, error) {
	if strings.HasPrefix(string(timebits), "-") {
		t, err := time.Parse("-2006-01-02T15:04:05Z", string(timebits))
		if err != nil {
			return "", err
		}
		// For negative years, add 1 to get correct BC year
		// year -1 is 2 BC, year -2 is 3 BC, etc.
		yearsToAdd := t.Year() + 1

		newT := time.Date(yearsToAdd, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
		return fmt.Sprintf("%s BC", newT.Format(layout)), nil
	}

	t, err := convertBitsToTime(timebits)
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
func convertTimestampWithTimezoneForPostgres(input []byte) string {
	// Remove the 'T'
	withoutT := strings.Replace(string(input), "T", " ", 1)

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

	return string(input)
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
