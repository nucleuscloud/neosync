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
	tableDependencies map[string][]*sqlmanager_shared.ForeignConstraint,
	runConfigs []*tabledependency.RunConfig,
	subsetByForeignKeyConstraints bool,
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.ColumnInfo,
) (map[string]map[tabledependency.RunType]string, error) {
	qb := NewQueryBuilderFromSchemaDefinition(groupedColumnInfo, tableDependencies, "public")

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

func NewQueryBuilderFromSchemaDefinition(
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.ColumnInfo,
	tableDependencies map[string][]*sqlmanager_shared.ForeignConstraint,
	defaultSchema string,
) *QueryBuilder {
	qb := NewQueryBuilder(defaultSchema)

	tables := map[string]*TableInfo{}
	for table, entry := range groupedColumnInfo {
		pieces := strings.SplitN(table, ".", 2)

		tableInfo := &TableInfo{
			Schema:  pieces[0],
			Name:    pieces[1],
			Columns: []string{},
		}
		for col := range entry {
			tableInfo.Columns = append(tableInfo.Columns, col)
		}
		tables[table] = tableInfo
		qb.AddTable(tableInfo)
	}

	for tableName, foreignKeys := range tableDependencies {
		tableInfo, ok := tables[tableName]
		if !ok {
			continue
		}

		for _, fk := range foreignKeys {
			pieces := strings.SplitN(fk.ForeignKey.Table, ".", 2)

			tableInfo.ForeignKeys = append(tableInfo.ForeignKeys, ForeignKey{
				Columns:          fk.Columns,
				ReferenceSchema:  pieces[0],
				ReferenceTable:   pieces[1],
				ReferenceColumns: fk.ForeignKey.Columns,
			})
		}

		qb.AddTable(tableInfo)
	}
	return qb
}

func splitTable(input string) (string, string) {
	pieces := strings.SplitN(input, ".", 2)
	return pieces[0], pieces[1]
}

func GetSubsetSelects(
	driver string,
	tableConstraintsMap ForeignKeyTableConstraints,
	runcfgs []*tabledependency.RunConfig,
	shouldSubsetByFkConstraints bool,
	columnInfo map[string]map[string]*sqlmanager_shared.ColumnInfo,
) (map[string]map[tabledependency.RunType]string, error) {

	return nil, nil
}

func buildSubsetSelects(
	driver string,
	tableConstraintsMap ForeignKeyTableConstraints,
	wheres map[string]string,
	tables []string,
) (any, error) {
	filteredConstraintsMap := filterFkConstraintsForSubsetting(tableConstraintsMap, tables, wheres)
	_ = filteredConstraintsMap
	return nil, nil
}

func filterFkConstraintsForSubsetting(
	tableConstraintsMap ForeignKeyTableConstraints,
	tables []string,
	wheres map[string]string,
) ForeignKeyTableConstraints {
	tablesToSubset := map[string]bool{}
	for _, table := range tables {
		tablesToSubset[table] = shouldSubsetTable(tableConstraintsMap, wheres, table)
	}

	filtered := ForeignKeyTableConstraints{}
	for _, table := range tables {
		filtered[table] = []*sqlmanager_shared.ForeignConstraint{}
		if constraints, ok := tableConstraintsMap[table]; ok {
			for _, fkdef := range constraints {
				if exists := tablesToSubset[fkdef.ForeignKey.Table]; exists {
					filtered[table] = append(filtered[table], fkdef)
				}
			}
		}
	}

	return filtered
}

func shouldSubsetTable(data ForeignKeyTableConstraints, whereClauses map[string]string, table string) bool {
	visited := make(map[string]bool)
	toVisit := []string{table}

	for len(toVisit) > 0 {
		currentTable := toVisit[len(toVisit)-1]
		toVisit = toVisit[:len(toVisit)-1]

		if _, exists := whereClauses[currentTable]; exists {
			return true
		}

		if visited[currentTable] {
			continue
		}
		visited[currentTable] = true

		if columns, exists := data[currentTable]; exists {
			for _, col := range columns {
				if col.ForeignKey.Table != "" && !visited[col.ForeignKey.Table] {
					toVisit = append(toVisit, col.ForeignKey.Table)
				}
			}
		}
	}

	return false
}

func buildAliasRefs(
	driver string,
	constraints ForeignKeyTableConstraints,
	wheres map[string]string,
) (any, error) {

	return nil, nil
}
