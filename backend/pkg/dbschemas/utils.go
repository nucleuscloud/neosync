package dbschemas_utils

import "fmt"

func BuildTable(schema, table string) string {
	if schema != "" {
		return fmt.Sprintf("%s.%s", schema, table)
	}
	return table
}
