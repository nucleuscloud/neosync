package schemamanager_shared

import sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"

const (
	BatchSizeConst = 20
)

type InitSchemaError struct {
	Statement string
	Error     string
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

func GetUniqueSchemaColMappings(
	schemas []*sqlmanager_shared.TableColumn,
) map[string]map[string]*sqlmanager_shared.TableColumn {
	groupedSchemas := map[string]map[string]*sqlmanager_shared.TableColumn{} // ex: {public.users: { id: struct{}{}, created_at: struct{}{}}}
	for _, record := range schemas {
		key := sqlmanager_shared.SchemaTable{Schema: record.Schema, Table: record.Table}.String()
		if _, ok := groupedSchemas[key]; ok {
			groupedSchemas[key][record.Name] = record
		} else {
			groupedSchemas[key] = map[string]*sqlmanager_shared.TableColumn{
				record.Name: record,
			}
		}
	}
	return groupedSchemas
}
