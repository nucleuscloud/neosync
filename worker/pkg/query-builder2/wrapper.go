package querybuilder2

import (
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
)

// QueryMapBuilderWrapper implements the SelectQueryMapBuilder interface
type QueryMapBuilderWrapper struct{}

// BuildSelectQueryMap wraps the original BuildSelectQueryMap function
func (w *QueryMapBuilderWrapper) BuildSelectQueryMap(
	driver string,
	tableFkConstraints map[string][]*sqlmanager_shared.ForeignConstraint,
	runConfigs []*tabledependency.RunConfig,
	subsetByForeignKeyConstraints bool,
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
) (map[string]map[tabledependency.RunType]string, error) {
	return BuildSelectQueryMap(
		driver,
		tableFkConstraints,
		runConfigs,
		subsetByForeignKeyConstraints,
		groupedColumnInfo,
	)
}
