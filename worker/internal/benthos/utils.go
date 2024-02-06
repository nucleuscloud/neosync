package neosync_benthos

import "fmt"

func BuildBenthosTable(schema, table string) string {
	if schema != "" {
		return fmt.Sprintf("%s.%s", schema, table)
	}
	return table
}

func BuildBenthosCacheKey(schema, table, col string) string {
	return fmt.Sprintf("%s.%s.%s", schema, table, col)
}
