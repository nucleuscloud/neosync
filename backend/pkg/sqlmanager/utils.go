package sqlmanager

import "fmt"

func getUniqueSchemaColMappings(
	schemas []*DatabaseSchemaRow,
) map[string]map[string]*ColumnInfo {
	groupedSchemas := map[string]map[string]*ColumnInfo{} // ex: {public.users: { id: struct{}{}, created_at: struct{}{}}}
	for _, record := range schemas {
		key := BuildTable(record.TableSchema, record.TableName)
		if _, ok := groupedSchemas[key]; ok {
			groupedSchemas[key][record.ColumnName] = toColumnInfo(record)
		} else {
			groupedSchemas[key] = map[string]*ColumnInfo{
				record.ColumnName: toColumnInfo(record),
			}
		}
	}
	return groupedSchemas
}

func toColumnInfo(row *DatabaseSchemaRow) *ColumnInfo {
	return &ColumnInfo{
		OrdinalPosition:        int32(row.OrdinalPosition),
		ColumnDefault:          row.ColumnDefault,
		IsNullable:             convertNullableTextToBool(row.IsNullable),
		DataType:               row.DataType,
		CharacterMaximumLength: Ptr(row.CharacterMaximumLength),
		NumericPrecision:       Ptr(row.NumericPrecision),
		NumericScale:           Ptr(row.NumericScale),
	}
}

func convertNullableTextToBool(isNullableStr string) bool {
	return isNullableStr != "NO"
}

func Ptr[T any](val T) *T {
	return &val
}

func BuildTable(schema, table string) string {
	if schema != "" {
		return fmt.Sprintf("%s.%s", schema, table)
	}
	return table
}

func dedupeSlice(input []string) []string {
	set := map[string]any{}
	for _, i := range input {
		set[i] = struct{}{}
	}
	output := make([]string, 0, len(set))
	for key := range set {
		output = append(output, key)
	}
	return output
}
