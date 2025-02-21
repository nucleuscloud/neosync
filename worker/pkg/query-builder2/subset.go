package querybuilder2

import (
	"fmt"
	"strings"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	runcfg "github.com/nucleuscloud/neosync/internal/runconfigs"
)

// returns map of schema.table -> select query
func BuildSelectQueryMap(
	driver string,
	runConfigs []*runcfg.RunConfig,
	subsetByForeignKeyConstraints bool,
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	pageLimit int,
) (map[string]*sqlmanager_shared.SelectQuery, error) {
	tableDependencies := map[string]*TableConstraints{}
	for _, rc := range runConfigs {
		if rc.RunType() != runcfg.RunTypeInsert {
			continue
		}
		td, ok := tableDependencies[rc.Table()]
		if !ok {
			td = &TableConstraints{
				PrimaryKeys: []*sqlmanager_shared.PrimaryKey{},
				ForeignKeys: []*sqlmanager_shared.ForeignConstraint{},
			}
			tableDependencies[rc.Table()] = td
		}

		for _, fk := range rc.ForeignKeys() {
			td.ForeignKeys = append(td.ForeignKeys, &sqlmanager_shared.ForeignConstraint{
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   fk.ReferenceTable,
					Columns: fk.ReferenceColumns,
				},
				NotNullable: fk.NotNullable,
				Columns:     fk.Columns,
			})
		}
		td.PrimaryKeys = append(td.PrimaryKeys, &sqlmanager_shared.PrimaryKey{
			Columns: rc.PrimaryKeys(),
		})
	}
	qb := NewQueryBuilderFromSchemaDefinition(groupedColumnInfo, tableDependencies, "public", driver, subsetByForeignKeyConstraints, pageLimit)

	for _, cfg := range runConfigs {
		if cfg.RunType() != runcfg.RunTypeInsert {
			continue
		}
		// add order by to query builder
		schema, table := splitTable(cfg.Table())
		if len(cfg.OrderByColumns()) > 0 {
			qb.AddOrderBy(schema, table, cfg.OrderByColumns())
		}
		// add where clause to query builder
		if cfg.WhereClause() != nil && *cfg.WhereClause() != "" {
			qualifiedWhereCaluse, err := qb.qualifyWhereCondition(nil, table, *cfg.WhereClause())
			if err != nil {
				return nil, err
			}
			qb.AddWhereCondition(schema, table, qualifiedWhereCaluse)
		}
	}

	querymap := map[string]*sqlmanager_shared.SelectQuery{}
	for _, cfg := range runConfigs {
		schema, table := splitTable(cfg.Table())
		query, _, pageQuery, isNotForeignKeySafe, err := qb.BuildQuery(schema, table)
		if err != nil {
			return nil, err
		}
		fmt.Println()
		fmt.Println(cfg.Id())
		fmt.Println(cfg.RunType())
		fmt.Println(cfg.InsertColumns())
		fmt.Println(query)
		fmt.Println()
		querymap[cfg.Id()] = &sqlmanager_shared.SelectQuery{
			Query:                     query,
			PageQuery:                 pageQuery,
			PageLimit:                 pageLimit,
			IsNotForeignKeySafeSubset: isNotForeignKeySafe,
		}
	}

	return querymap, nil
}

type TableConstraints struct {
	ForeignKeys []*sqlmanager_shared.ForeignConstraint
	PrimaryKeys []*sqlmanager_shared.PrimaryKey
}

func NewQueryBuilderFromSchemaDefinition(
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	tableDependencies map[string]*TableConstraints,
	defaultSchema string,
	driver string,
	subsetByForeignKeyConstraints bool,
	pageLimit int,
) *QueryBuilder {
	qb := NewQueryBuilder(defaultSchema, driver, subsetByForeignKeyConstraints, groupedColumnInfo, pageLimit)

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
			tableInfo.PrimaryKeys = uniqueStrings(tableInfo.PrimaryKeys)
			qb.AddTable(tableInfo)
		}

		for _, fk := range constraints.ForeignKeys {
			refSchema, refTable := splitTable(fk.ForeignKey.Table)
			tableInfo.ForeignKeys = append(tableInfo.ForeignKeys, &ForeignKey{
				Columns:          fk.Columns,
				NotNullable:      fk.NotNullable,
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

func splitTable(fullTableName string) (schema, table string) {
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
