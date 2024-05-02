package sqlmanager

import "fmt"

type ColumnInfo struct {
	OrdinalPosition        int32  // Specifies the sequence or order in which each column is defined within the table. Starts at 1 for the first column.
	ColumnDefault          string // Specifies the default value for a column, if any is set.
	IsNullable             string // Specifies if the column is nullable or not. Possible values are 'YES' for nullable and 'NO' for not nullable.
	DataType               string // Specifies the data type of the column, i.e., bool, varchar, int, etc.
	CharacterMaximumLength *int32 // Specifies the maximum allowable length of the column for character-based data types. For datatypes such as integers, boolean, dates etc. this is NULL.
	NumericPrecision       *int32 // Specifies the precision for numeric data types. It represents the TOTAL count of significant digits in the whole number, that is, the number of digits to BOTH sides of the decimal point. Null for non-numeric data types.
	NumericScale           *int32 // Specifies the scale of the column for numeric data types, specifically non-integers. It represents the number of digits to the RIGHT of the decimal point. Null for non-numeric data types and integers.
}

func GetUniqueSchemaColMappings(
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
		IsNullable:             row.IsNullable,
		DataType:               row.DataType,
		CharacterMaximumLength: Ptr(row.CharacterMaximumLength),
		NumericPrecision:       Ptr(row.NumericPrecision),
		NumericScale:           Ptr(row.NumericScale),
	}
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
