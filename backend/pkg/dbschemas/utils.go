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
	OrdinalPosition        int32
	ColumnDefault          string
	IsNullable             string
	DataType               string
	CharacterMaximumLength *int32
	NumericPrecision       *int32
	NumericScale           *int32
}

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
