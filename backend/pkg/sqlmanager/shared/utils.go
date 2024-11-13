package sqlmanager_shared

import (
	"fmt"
	"strings"
)

func GetUniqueSchemaColMappings(
	schemas []*DatabaseSchemaRow,
) map[string]map[string]*DatabaseSchemaRow {
	groupedSchemas := map[string]map[string]*DatabaseSchemaRow{} // ex: {public.users: { id: struct{}{}, created_at: struct{}{}}}
	for _, record := range schemas {
		key := BuildTable(record.TableSchema, record.TableName)
		if _, ok := groupedSchemas[key]; ok {
			groupedSchemas[key][record.ColumnName] = record
		} else {
			groupedSchemas[key] = map[string]*DatabaseSchemaRow{
				record.ColumnName: record,
			}
		}
	}
	return groupedSchemas
}

func ConvertNullableTextToBool(isNullableStr string) bool {
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

// Dedupes the input slice and ensures consistent ordering with the input. Returns a niew slice.
func DedupeSlice(input []string) []string {
	seen := make(map[string]struct{})
	output := []string{}

	for _, item := range input {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			output = append(output, item)
		}
	}

	return output
}

func SplitTableKey(key string) (schema, table string) {
	pieces := strings.Split(key, ".")
	if len(pieces) == 1 {
		return "public", pieces[0]
	}
	return pieces[0], pieces[1]
}
