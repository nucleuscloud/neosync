package querybuilder2

import (
	"strings"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
)

type ForeignKeyTableConstraints = map[string][]*sqlmanager_shared.ForeignConstraint

// returns map of schema.table -> select query
func BuildSelectQueryMap(
	driver string,
	tableDependencies map[string]*TableConstraints,
	runConfigs []*tabledependency.RunConfig,
	subsetByForeignKeyConstraints bool,
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.ColumnInfo,
) (map[string]map[tabledependency.RunType]string, error) {
	qb := NewQueryBuilderFromSchemaDefinition(groupedColumnInfo, tableDependencies, "public", driver)

	for _, cfg := range runConfigs {
		if cfg.RunType != tabledependency.RunTypeInsert || cfg.WhereClause == nil || *cfg.WhereClause == "" {
			continue
		}
		schema, table := splitTable(cfg.Table)
		qb.AddWhereCondition(schema, table, *cfg.WhereClause)
	}

	querymap := map[string]map[tabledependency.RunType]string{}
	for _, cfg := range runConfigs {
		if _, ok := querymap[cfg.Table]; !ok {
			querymap[cfg.Table] = map[tabledependency.RunType]string{}
		}
		schema, table := splitTable(cfg.Table)
		query, _, err := qb.BuildQuery(schema, table)
		if err != nil {
			return nil, err
		}
		querymap[cfg.Table][cfg.RunType] = query
	}

	return querymap, nil
}

type TableConstraints struct {
	ForeignKeys []*sqlmanager_shared.ForeignConstraint
	PrimaryKeys []*sqlmanager_shared.PrimaryKey
}

func NewQueryBuilderFromSchemaDefinition(
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.ColumnInfo,
	tableDependencies map[string]*TableConstraints,
	defaultSchema string,
	driver string,
) *QueryBuilder {
	qb := NewQueryBuilder(defaultSchema, driver)

	for table, columns := range groupedColumnInfo {
		schema, tableName := splitTable(table)
		tableInfo := &TableInfo{
			Schema:  schema,
			Name:    tableName,
			Columns: make([]string, 0, len(columns)),
		}
		for col := range columns {
			tableInfo.Columns = append(tableInfo.Columns, col)
		}
		qb.AddTable(tableInfo)
	}

	for tableName, constraints := range tableDependencies {
		schema, table := splitTable(tableName)
		tableInfo := qb.tables[qb.getTableKey(schema, table)]
		if tableInfo == nil {
			tableInfo = &TableInfo{
				Schema:  schema,
				Name:    table,
				Columns: []string{},
			}
			for _, pk := range constraints.PrimaryKeys {
				tableInfo.Columns = append(tableInfo.Columns, pk.Columns...)
				tableInfo.PrimaryKeys = append(tableInfo.PrimaryKeys, pk.Columns...)
			}
			qb.AddTable(tableInfo)
		}

		for _, fk := range constraints.ForeignKeys {
			refSchema, refTable := splitTable(fk.ForeignKey.Table)
			tableInfo.ForeignKeys = append(tableInfo.ForeignKeys, ForeignKey{
				Columns:          fk.Columns,
				ReferenceSchema:  refSchema,
				ReferenceTable:   refTable,
				ReferenceColumns: fk.ForeignKey.Columns,
			})
			tableInfo.Columns = append(tableInfo.Columns, fk.Columns...)
		}
		tableInfo.Columns = uniqueStrings(tableInfo.Columns)
	}

	return qb
}

func splitTable(fullTableName string) (string, string) {
	parts := strings.SplitN(fullTableName, ".", 2)
	if len(parts) == 1 {
		return "", parts[0]
	}
	return parts[0], parts[1]
}

func uniqueStrings(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	result := make([]string, 0, len(input))
	for _, v := range input {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
