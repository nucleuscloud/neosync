package validate_schema

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
)

// ValidationReport holds differences found between the schema and job mappings.
type ValidationReport struct {
	MissingColumns []*mgmtv1alpha1.DatabaseColumn   // Columns job mappings that are not defined in schema.
	ExtraColumns   []*mgmtv1alpha1.DatabaseColumn   // Columns in schema that are not defined in job mappings but the table exists in job mappings.
	MissingTables  []*sqlmanager_shared.SchemaTable // Tables in job mappings that are not defined in schema.
	MissingSchemas []string                         // Schemas in job mappings that are not defined in schema.
}

// getColumnKey generates a unique key for a column.
func getColumnKey(row *mgmtv1alpha1.DatabaseColumn) string {
	return fmt.Sprintf("%s.%s.%s", row.Schema, row.Table, row.Column)
}

// getJobMappingKey generates a unique key for a job mapping.
func getJobMappingKey(mapping *mgmtv1alpha1.JobMapping) string {
	return fmt.Sprintf("%s.%s.%s", mapping.Schema, mapping.Table, mapping.Column)
}

// diffColumnsAgainstMappings compares schema columns against the job mapping columns.
// It returns two slices:
// - missing: columns defined in the job mappings but not found in the schema (excluding those whose schema or table are missing).
// - extra: columns present in the schema but not defined in the job mappings.
// This function takes into account missing schemas and missing tables (passed in as arguments) to avoid duplicate missing column errors.
func diffColumnsAgainstMappings(schema []*mgmtv1alpha1.DatabaseColumn, mappings []*mgmtv1alpha1.JobMapping, missingSchemaSet, missingTableSet map[string]bool) (missingCols, extraCols []*mgmtv1alpha1.DatabaseColumn) {
	extra := []*mgmtv1alpha1.DatabaseColumn{}
	missing := []*mgmtv1alpha1.DatabaseColumn{}

	// Build a set of tables from the job mappings.
	tablesInMappings := make(map[string]bool)
	for _, mappingItem := range mappings {
		tableKey := fmt.Sprintf("%s.%s", mappingItem.Schema, mappingItem.Table)
		tablesInMappings[tableKey] = true
	}

	// Build a lookup map for job mapping columns.
	mappingMap := make(map[string]*mgmtv1alpha1.JobMapping)
	for _, mappingItem := range mappings {
		mappingMap[getJobMappingKey(mappingItem)] = mappingItem
	}

	// Build a lookup map for schema columns; only include those from tables in the job mappings.
	schemaMap := make(map[string]*mgmtv1alpha1.DatabaseColumn)
	for _, col := range schema {
		tableKey := fmt.Sprintf("%s.%s", col.Schema, col.Table)
		if _, exists := tablesInMappings[tableKey]; !exists {
			continue
		}
		schemaMap[getColumnKey(col)] = col
	}

	// Determine missing columns from job mappings, skipping ones from missing schemas or missing tables.
	for key, mappingItem := range mappingMap {
		tableKey := fmt.Sprintf("%s.%s", mappingItem.Schema, mappingItem.Table)
		if missingSchemaSet[mappingItem.Schema] || missingTableSet[tableKey] {
			continue
		}
		if _, exists := schemaMap[key]; !exists {
			missing = append(missing, &mgmtv1alpha1.DatabaseColumn{
				Schema: mappingItem.Schema,
				Table:  mappingItem.Table,
				Column: mappingItem.Column,
			})
		}
	}

	// Extra columns: those in the schema but not defined in the job mappings.
	for key, col := range schemaMap {
		if _, exists := mappingMap[key]; !exists {
			extra = append(extra, col)
		}
	}

	return missing, extra
}

// diffTablesAgainstMappings compares tables from job mappings against the schema tables.
// It returns a list of tables (as SchemaTable objects) that are defined in the job mappings but missing in the schema,
// excluding those whose schemas are already missing.
// This function accepts missingSchemas as an argument.
func diffTablesAgainstMappings(schema []*mgmtv1alpha1.DatabaseColumn, mappings []*mgmtv1alpha1.JobMapping, missingSchemaSet map[string]bool) []*sqlmanager_shared.SchemaTable {
	missing := []*sqlmanager_shared.SchemaTable{}
	schemaTables := make(map[string]*sqlmanager_shared.SchemaTable)
	mappingTables := make(map[string]*sqlmanager_shared.SchemaTable)

	// Build a map for tables present in the schema.
	for _, col := range schema {
		st := &sqlmanager_shared.SchemaTable{Schema: col.Schema, Table: col.Table}
		schemaTables[st.String()] = st
	}

	// Build a map for tables defined in the job mappings.
	for _, mappingItem := range mappings {
		st := &sqlmanager_shared.SchemaTable{Schema: mappingItem.Schema, Table: mappingItem.Table}
		mappingTables[st.String()] = st
	}

	// Identify tables missing in the schema, excluding those whose schema is missing.
	for key, st := range mappingTables {
		if _, exists := schemaTables[key]; !exists {
			if !missingSchemaSet[st.Schema] {
				missing = append(missing, st)
			}
		}
	}

	return missing
}

// diffSchemasAgainstMappings compares the schemas used in job mappings with the schema.
// It returns a list of schemas that appear in the job mappings but not in the schema.
func diffSchemaAgainstMappings(schema []*mgmtv1alpha1.DatabaseColumn, mappings []*mgmtv1alpha1.JobMapping) []string {
	missing := []string{}
	schemaSchemas := make(map[string]bool)
	mappingSchemas := make(map[string]bool)

	// Build set of schemas from the schema.
	for _, col := range schema {
		schemaSchemas[col.Schema] = true
	}

	// Build set of schemas from the job mappings.
	for _, mappingItem := range mappings {
		mappingSchemas[mappingItem.Schema] = true
	}

	// Identify schemas from the job mappings that are missing in the schema.
	for schemaName := range mappingSchemas {
		if !schemaSchemas[schemaName] {
			missing = append(missing, schemaName)
		}
	}

	return missing
}

// ValidateSchemaAgainstJobMappings returns the differences between the schema and the job mappings.
// It produces a SchemaDiff that details columns, tables, and schemas in the schema that lack corresponding job mappings,
// as well as extra columns that are in the schema but not in the mappings.
func ValidateSchemaAgainstJobMappings(schema []*mgmtv1alpha1.DatabaseColumn, mappings []*mgmtv1alpha1.JobMapping) ValidationReport {
	missingSchemas := diffSchemaAgainstMappings(schema, mappings)
	missingSchemaSet := make(map[string]bool)
	for _, s := range missingSchemas {
		missingSchemaSet[s] = true
	}

	missingTables := diffTablesAgainstMappings(schema, mappings, missingSchemaSet)
	missingTableSet := make(map[string]bool)
	for _, mt := range missingTables {
		missingTableSet[mt.String()] = true
	}
	missingColumns, extraColumns := diffColumnsAgainstMappings(schema, mappings, missingSchemaSet, missingTableSet)

	return ValidationReport{
		MissingColumns: missingColumns,
		ExtraColumns:   extraColumns,
		MissingTables:  missingTables,
		MissingSchemas: missingSchemas,
	}
}
