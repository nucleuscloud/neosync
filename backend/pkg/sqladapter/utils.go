package sqladapter

import dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"

func GetUniqueSchemaColMappings(
	schemas []*DatabaseSchemaRow,
) map[string]map[string]*dbschemas.ColumnInfo {
	groupedSchemas := map[string]map[string]*dbschemas.ColumnInfo{} // ex: {public.users: { id: struct{}{}, created_at: struct{}{}}}
	for _, record := range schemas {
		key := dbschemas.BuildTable(record.TableSchema, record.TableName)
		if _, ok := groupedSchemas[key]; ok {
			groupedSchemas[key][record.ColumnName] = toColumnInfo(record)
		} else {
			groupedSchemas[key] = map[string]*dbschemas.ColumnInfo{
				record.ColumnName: toColumnInfo(record),
			}
		}
	}
	return groupedSchemas
}

func toColumnInfo(row *DatabaseSchemaRow) *dbschemas.ColumnInfo {
	var colDefault string
	if row.ColumnDefault != nil {
		val, ok := row.ColumnDefault.(string)
		if ok {
			colDefault = val
		}
	}
	return &dbschemas.ColumnInfo{
		OrdinalPosition:        int32(row.OrdinalPosition),
		ColumnDefault:          colDefault,
		IsNullable:             row.IsNullable,
		DataType:               row.DataType,
		CharacterMaximumLength: ptr(row.CharacterMaximumLength),
		NumericPrecision:       ptr(row.NumericPrecision),
		NumericScale:           ptr(row.NumericScale),
	}
}

func ptr[T any](val T) *T {
	return &val
}

// Key is schema.table value is list of tables that key depends on
func GetDbTableDependencies(
	constraints []*ForeignKeyConstraintsRow,
) dbschemas.TableDependency {
	tableConstraints := map[string]*dbschemas.TableConstraints{}
	for _, c := range constraints {
		tableName := dbschemas.BuildTable(c.SchemaName, c.TableName)

		constraint, ok := tableConstraints[tableName]
		if !ok {
			tableConstraints[tableName] = &dbschemas.TableConstraints{
				Constraints: []*dbschemas.ForeignConstraint{
					{Column: c.ColumnName, IsNullable: dbschemas.ConvertIsNullableToBool(c.IsNullable), ForeignKey: &dbschemas.ForeignKey{
						Table:  dbschemas.BuildTable(c.ForeignSchemaName, c.ForeignTableName),
						Column: c.ForeignColumnName,
					}},
				},
			}
		} else {
			constraint.Constraints = append(constraint.Constraints, &dbschemas.ForeignConstraint{
				Column: c.ColumnName, IsNullable: dbschemas.ConvertIsNullableToBool(c.IsNullable), ForeignKey: &dbschemas.ForeignKey{
					Table:  dbschemas.BuildTable(c.ForeignSchemaName, c.ForeignTableName),
					Column: c.ForeignColumnName,
				},
			})
		}
	}
	return tableConstraints
}

func GetTablePrimaryKeysMap(
	primaryKeyConstraints []*PrimaryKeyConstraintsRow,
) map[string][]string {
	pkConstraintMap := map[string][]*PrimaryKeyConstraintsRow{}
	for _, c := range primaryKeyConstraints {
		_, ok := pkConstraintMap[c.ConstraintName]
		if ok {
			pkConstraintMap[c.ConstraintName] = append(pkConstraintMap[c.ConstraintName], c)
		} else {
			pkConstraintMap[c.ConstraintName] = []*PrimaryKeyConstraintsRow{c}
		}
	}
	pkMap := map[string][]string{}
	for _, constraints := range pkConstraintMap {
		for _, c := range constraints {
			key := dbschemas.BuildTable(c.SchemaName, c.TableName)
			_, ok := pkMap[key]
			if ok {
				pkMap[key] = append(pkMap[key], c.ColumnName)
			} else {
				pkMap[key] = []string{c.ColumnName}
			}
		}
	}
	return pkMap
}
