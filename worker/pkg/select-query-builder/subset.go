package selectquerybuilder

import (
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	runcfg "github.com/nucleuscloud/neosync/internal/runconfigs"
)

// BuildSelectQueryMap builds a map of SelectQuery objects for each RunConfig.
// It creates a QueryBuilder with the specified driver and global configuration options,
// then iterates through each RunConfig to build the appropriate SQL SELECT queries.
//
// Parameters:
//   - driver: The database driver name (e.g., "postgres", "mysql")
//   - runConfigs: A slice of RunConfig objects that define what data to select
//   - subsetByForeignKeyConstraints: Whether to respect foreign key constraints when subsetting data
//   - pageLimit: The maximum number of rows to return per page (for pagination query)
//
// Returns:
//   - A map where keys are RunConfig IDs and values are SelectQuery objects
func BuildSelectQueryMap(
	driver string,
	runConfigs []*runcfg.RunConfig,
	subsetByForeignKeyConstraints bool,
	pageLimit int,
) (map[string]*sqlmanager_shared.SelectQuery, error) {
	qb := NewSelectQueryBuilder("public", driver, subsetByForeignKeyConstraints, pageLimit)
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
