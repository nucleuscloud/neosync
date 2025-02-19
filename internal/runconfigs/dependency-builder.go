package runconfigs

import (
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/backend/pkg/utils"
)

type dependencies struct {
	filteredDeps   map[string][]string                      // only include tables that are in tables arg list
	foreignKeys    map[string]map[string][]string           // map: table -> foreign key table -> foreign key column
	foreignKeyCols map[string]map[string]*ConstraintColumns // map: table -> foreign key table -> ConstraintColumns
}

func buildDependencies(
	dependencyMap map[string][]*sqlmanager_shared.ForeignConstraint,
	tableColumnsMap map[string][]string,
) *dependencies {
	deps := &dependencies{
		filteredDeps:   make(map[string][]string),
		foreignKeys:    make(map[string]map[string][]string),
		foreignKeyCols: make(map[string]map[string]*ConstraintColumns),
	}

	for table, constraints := range dependencyMap {
		processTableConstraints(deps, table, constraints, tableColumnsMap)
	}

	deduplicateFilteredDeps(deps)
	return deps
}

func processTableConstraints(
	deps *dependencies,
	table string,
	constraints []*sqlmanager_shared.ForeignConstraint,
	tableColumnsMap map[string][]string,
) {
	deps.foreignKeys[table] = make(map[string][]string)
	deps.foreignKeyCols[table] = make(map[string]*ConstraintColumns)

	for _, constraint := range constraints {
		processForeignKeyConstraint(deps, table, constraint, tableColumnsMap)
	}
}

func processForeignKeyConstraint(
	deps *dependencies,
	table string,
	constraint *sqlmanager_shared.ForeignConstraint,
	tableColumnsMap map[string][]string,
) {
	fkTable := constraint.ForeignKey.Table
	if !checkTableHasCols([]string{table, fkTable}, tableColumnsMap) {
		return
	}

	if _, exists := deps.foreignKeyCols[table][fkTable]; !exists {
		deps.foreignKeyCols[table][fkTable] = &ConstraintColumns{
			NullableColumns:    []string{},
			NonNullableColumns: []string{},
		}
	}

	for idx, col := range constraint.ForeignKey.Columns {
		updateForeignKeyMaps(deps, table, fkTable, col, constraint.NotNullable[idx])
	}

	deps.filteredDeps[table] = append(deps.filteredDeps[table], fkTable)
}

func updateForeignKeyMaps(deps *dependencies, table, fkTable, col string, notNullable bool) {
	if notNullable {
		deps.foreignKeyCols[table][fkTable].NonNullableColumns = append(deps.foreignKeyCols[table][fkTable].NonNullableColumns, col)
	} else {
		deps.foreignKeyCols[table][fkTable].NullableColumns = append(deps.foreignKeyCols[table][fkTable].NullableColumns, col)
	}
	deps.foreignKeys[table][fkTable] = append(deps.foreignKeys[table][fkTable], col)
}

func deduplicateFilteredDeps(deps *dependencies) {
	for table, filteredDeps := range deps.filteredDeps {
		deps.filteredDeps[table] = utils.DedupeSliceOrdered(filteredDeps)
	}
}
