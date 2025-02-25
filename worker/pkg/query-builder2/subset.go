package querybuilder2

import (
	"encoding/json"
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
	pageLimit int,
) (map[string]*sqlmanager_shared.SelectQuery, error) {
	fmt.Println("Building select query map")
	qb := NewQueryBuilder("public", driver, subsetByForeignKeyConstraints, pageLimit)
	for _, cfg := range runConfigs {
		// if cfg.RunType() != runcfg.RunTypeInsert {
		// 	continue
		// }
		// add order by to query builder
		schema, table := splitTable(cfg.Table())

		tableInfo := &TableInfo{
			Id:             cfg.Id(),
			Schema:         schema,
			Name:           table,
			Columns:        cfg.SelectColumns(),
			PrimaryKeys:    cfg.PrimaryKeys(),
			SubsetPaths:    cfg.SubsetPaths(),
			OrderByColumns: cfg.OrderByColumns(),
		}
		qb.AddTable(tableInfo)
		// jsonF, _ := json.MarshalIndent(tableInfo, "", " ")
		// fmt.Printf("\n\n %s \n\n", string(jsonF))

		// if len(cfg.OrderByColumns()) > 0 {
		// 	qb.AddOrderBy(schema, table, cfg.OrderByColumns())
		// }
		// add where clause to query builder
		if cfg.RunType() == runcfg.RunTypeInsert && cfg.WhereClause() != nil && *cfg.WhereClause() != "" {
			qualifiedWhereCaluse, err := qb.qualifyWhereCondition(nil, table, *cfg.WhereClause())
			if err != nil {
				return nil, err
			}
			qb.AddWhereCondition(schema, table, qualifiedWhereCaluse)
		}
	}

	jsonF, _ := json.MarshalIndent(qb.whereConditions, "", " ")
	fmt.Printf("\n\nwhereConditions:  %s \n\n", string(jsonF))

	querymap := map[string]*sqlmanager_shared.SelectQuery{}
	for _, cfg := range runConfigs {
		query, _, pageQuery, isNotForeignKeySafe, err := qb.BuildQuery(cfg.Id())
		if err != nil {
			return nil, err
		}
		// fmt.Println()
		// fmt.Println(cfg.Id())
		// fmt.Println(cfg.RunType())
		// fmt.Println(cfg.InsertColumns())
		// fmt.Println(query)
		// fmt.Println()
		querymap[cfg.Id()] = &sqlmanager_shared.SelectQuery{
			Query:                     query,
			PageQuery:                 pageQuery,
			PageLimit:                 pageLimit,
			IsNotForeignKeySafeSubset: isNotForeignKeySafe,
		}
	}

	return querymap, nil
}

func splitTable(fullTableName string) (schema, table string) {
	parts := strings.SplitN(fullTableName, ".", 2)
	if len(parts) == 1 {
		return "", parts[0]
	}
	return parts[0], parts[1]
}
