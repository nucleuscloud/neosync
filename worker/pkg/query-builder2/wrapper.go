package querybuilder2

import (
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	rc "github.com/nucleuscloud/neosync/internal/runconfigs"
)

// QueryMapBuilderWrapper implements the SelectQueryMapBuilder interface
type QueryMapBuilderWrapper struct{}

// BuildSelectQueryMap wraps the original BuildSelectQueryMap function
func (w *QueryMapBuilderWrapper) BuildSelectQueryMap(
	driver string,
	runConfigs []*rc.RunConfig,
	subsetByForeignKeyConstraints bool,
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	pageLimit int,
) (map[string]*sqlmanager_shared.SelectQuery, error) {
	return BuildSelectQueryMap(
		driver,
		runConfigs,
		subsetByForeignKeyConstraints,
		groupedColumnInfo,
		pageLimit,
	)
}
