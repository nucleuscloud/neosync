package querybuilder2

import (
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
	qb := NewQueryBuilder("public", driver, subsetByForeignKeyConstraints, pageLimit)
	querymap := map[string]*sqlmanager_shared.SelectQuery{}
	for _, cfg := range runConfigs {
		query, _, pageQuery, isNotForeignKeySafe, err := qb.BuildQuery(cfg)
		if err != nil {
			return nil, err
		}

		querymap[cfg.Id()] = &sqlmanager_shared.SelectQuery{
			Query:                     query,
			PageQuery:                 pageQuery,
			PageLimit:                 pageLimit,
			IsNotForeignKeySafeSubset: isNotForeignKeySafe,
		}
	}

	return querymap, nil
}
