package schemamanager_shared

import sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"

const (
	BatchSizeConst = 20
)

type InitSchemaError struct {
	Statement string
	Error     string
}

type Missing struct {
	Tables  []*sqlmanager_shared.SchemaTable
	Columns []*sqlmanager_shared.DatabaseSchemaRow
}

type ExistsInBoth struct {
	Tables []*sqlmanager_shared.SchemaTable
}

type SchemaDifferences struct {
	Missing      *Missing
	ExistsInBoth *ExistsInBoth
}

// filtered by tables found in job mappings
func GetFilteredForeignToPrimaryTableMap(td map[string][]*sqlmanager_shared.ForeignConstraint, uniqueTables map[string]struct{}) map[string][]string {
	dpMap := map[string][]string{}
	for table := range uniqueTables {
		_, dpOk := dpMap[table]
		if !dpOk {
			dpMap[table] = []string{}
		}
		constraints, ok := td[table]
		if !ok {
			continue
		}
		for _, dep := range constraints {
			_, ok := uniqueTables[dep.ForeignKey.Table]
			// only add to map if dependency is an included table
			if ok {
				dpMap[table] = append(dpMap[table], dep.ForeignKey.Table)
			}
		}
	}
	return dpMap
}
