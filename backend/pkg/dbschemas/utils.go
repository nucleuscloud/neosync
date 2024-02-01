package dbschemas_utils

import (
	"fmt"
)

type ForeignKey struct {
	Table  string
	Column string
}
type ForeignConstraint struct {
	Column     string
	IsNullable bool
	ForeignKey *ForeignKey
}
type TableConstraints struct {
	Constraints []*ForeignConstraint
}
type TableDependency = map[string]*TableConstraints

type ColumnInfo struct {
	OrdinalPosition        int32  // Specifies the sequence or order in which each column is defined within the table. Starts at 1 for the first column.
	ColumnDefault          string // Specifies the default value for a column, if any is set.
	IsNullable             string // Specifies if the column is nullable or not. Possible values are 'YES' for nullable and 'NO' for not nullable.
	DataType               string // Specifies the data type of the column, i.e., bool, varchar, int, etc.
	CharacterMaximumLength *int32 // Specifies the maximum allowable length of the column for character-based data types. For datatypes such as integers, boolean, dates etc. this is NULL.
	NumericPrecision       *int32 // Specifies the precision for numeric data types. It represents the TOTAL count of significant digits in the whole number, that is, the number of digits to BOTH sides of the decimal point. Null for non-numeric data types.
	NumericScale           *int32 // Specifies the scale of the column for numeric data types, specifically non-integers. It represents the number of digits to the RIGHT of the decimal point. Null for non-numeric data types and integers.
}

/* More on NumericPrecision and NumericScale including examples

SMALLINT: 2 bytes, range from -32,768 to 32,767.
INTEGER: 4 bytes, range from -2,147,483,648 to 2,147,483,647.
BIGINT: 8 bytes, range from -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807.

Value: 123.45
Numeric Precision: 5 (total number of digits)
Numeric Scale: 2 (number of digits after the decimal point)
Explanation: The number has a total of 5 significant digits, with 2 digits following the decimal point.

Value: 0.00123
Numeric Precision: 3
Numeric Scale: 5
Explanation: The number has 3 significant digits, all after the decimal point, which itself has 5 places.

Value: 12345 (as a DECIMAL or NUMERIC)
Numeric Precision: 5
Numeric Scale: 0
Explanation: There are 5 total digits, and no digits after the decimal point.


Value: 10000.00
Numeric Precision: 7
Numeric Scale: 2
Explanation: The number has 7 significant digits in total, with 2 digits after the decimal point.

Value: 0.000001
Numeric Precision: 1
Numeric Scale: 6
Explanation: There is 1 significant digit after the decimal point, which itself has 6 places.

Value: 123456789.987654321
Numeric Precision: 18
Numeric Scale: 9
Explanation: The number has 18 significant digits with 9 of them after the decimal point.
*/

func BuildTable(schema, table string) string {
	if schema != "" {
		return fmt.Sprintf("%s.%s", schema, table)
	}
	return table
}

func BuildDependsOnSlice(constraintsMap map[string]*TableConstraints) map[string][]string {
	dependsOn := map[string][]string{}

	for tableName, constraints := range constraintsMap {
		dependsOnMap := map[string]struct{}{}
		for _, c := range constraints.Constraints {
			dependsOnMap[c.ForeignKey.Table] = struct{}{}
		}
		uniqueDependsOn := []string{}
		for t := range dependsOnMap {
			uniqueDependsOn = append(uniqueDependsOn, t)
		}
		dependsOn[tableName] = uniqueDependsOn
	}
	return dependsOn
}

func ConvertIsNullableToBool(isNullableStr string) bool {
	return isNullableStr != "NO"
}
