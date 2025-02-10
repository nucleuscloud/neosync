package validate_schema

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/assert"
)

// Start Generation Here
func TestValidateSchemaAgainstJobMappings_NoDiff(t *testing.T) {
	schema := []*mgmtv1alpha1.DatabaseColumn{
		{
			Schema: "public",
			Table:  "users",
			Column: "id",
		},
		{
			Schema: "public",
			Table:  "users",
			Column: "name",
		},
	}
	mappings := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "public",
			Table:  "users",
			Column: "id",
		},
		{
			Schema: "public",
			Table:  "users",
			Column: "name",
		},
	}

	diff := ValidateSchemaAgainstJobMappings(schema, mappings)

	assert.Len(t, diff.MissingColumns, 0, "expected no missing columns")
	assert.Len(t, diff.ExtraColumns, 0, "expected no extra columns")
	assert.Len(t, diff.MissingTables, 0, "expected no missing tables")
	assert.Len(t, diff.MissingSchemas, 0, "expected no missing schemas")
}

func TestValidateSchemaAgainstJobMappings_MissingColumn(t *testing.T) {
	// Schema has only the "id" column, but mappings expect both "id" and "name"
	schema := []*mgmtv1alpha1.DatabaseColumn{
		{
			Schema: "public",
			Table:  "users",
			Column: "id",
		},
	}
	mappings := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "public",
			Table:  "users",
			Column: "id",
		},
		{
			Schema: "public",
			Table:  "users",
			Column: "name",
		},
	}

	diff := ValidateSchemaAgainstJobMappings(schema, mappings)

	expectedMissing := []*mgmtv1alpha1.DatabaseColumn{
		{
			Schema: "public",
			Table:  "users",
			Column: "name",
		},
	}
	assert.ElementsMatch(t, expectedMissing, diff.MissingColumns, "expected missing column 'name'")
	assert.Len(t, diff.ExtraColumns, 0, "expected no extra columns")
	assert.Len(t, diff.MissingTables, 0, "expected no missing tables")
	assert.Len(t, diff.MissingSchemas, 0, "expected no missing schemas")
}

func TestValidateSchemaAgainstJobMappings_ExtraColumn(t *testing.T) {
	// Schema has an extra column "email" that is not referenced in the mappings
	schema := []*mgmtv1alpha1.DatabaseColumn{
		{
			Schema: "public",
			Table:  "users",
			Column: "id",
		},
		{
			Schema: "public",
			Table:  "users",
			Column: "email",
		},
	}
	mappings := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "public",
			Table:  "users",
			Column: "id",
		},
	}

	diff := ValidateSchemaAgainstJobMappings(schema, mappings)

	expectedExtra := []*mgmtv1alpha1.DatabaseColumn{
		{
			Schema: "public",
			Table:  "users",
			Column: "email",
		},
	}
	assert.ElementsMatch(t, expectedExtra, diff.ExtraColumns, "expected extra column 'email'")
	assert.Len(t, diff.MissingColumns, 0, "expected no missing columns")
	assert.Len(t, diff.MissingTables, 0, "expected no missing tables")
	assert.Len(t, diff.MissingSchemas, 0, "expected no missing schemas")
}

func TestValidateSchemaAgainstJobMappings_MissingTable(t *testing.T) {
	// Schema contains table "users" but mapping references a non-existent table "orders"
	schema := []*mgmtv1alpha1.DatabaseColumn{
		{
			Schema: "public",
			Table:  "users",
			Column: "id",
		},
		{
			Schema: "public",
			Table:  "users",
			Column: "name",
		},
	}
	mappings := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "public",
			Table:  "orders",
			Column: "id",
		},
	}

	diff := ValidateSchemaAgainstJobMappings(schema, mappings)

	assert.Len(t, diff.MissingTables, 1, "expected one missing table")
	missingTable := diff.MissingTables[0]
	assert.Equal(t, "public", missingTable.Schema, "expected schema 'public' for missing table")
	assert.Equal(t, "orders", missingTable.Table, "expected table 'orders' as missing")
	// MissingColumns should be empty because missing table mappings are omitted from column diffing.
	assert.Len(t, diff.MissingColumns, 0, "expected no missing columns for a missing table")
	assert.Len(t, diff.ExtraColumns, 0, "expected no extra columns")
	assert.Len(t, diff.MissingSchemas, 0, "expected no missing schemas")
}

func TestValidateSchemaAgainstJobMappings_MissingSchema(t *testing.T) {
	// Mapping uses a schema "other" that is not present in the actual schema data
	schema := []*mgmtv1alpha1.DatabaseColumn{
		{
			Schema: "public",
			Table:  "users",
			Column: "id",
		},
	}
	mappings := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "other",
			Table:  "users",
			Column: "id",
		},
	}

	diff := ValidateSchemaAgainstJobMappings(schema, mappings)

	assert.Len(t, diff.MissingSchemas, 1, "expected one missing schema")
	assert.Contains(t, diff.MissingSchemas, "other", "expected missing schema 'other'")
	// No missing columns or missing tables should be reported when the entire schema is missing.
	assert.Len(t, diff.MissingColumns, 0, "expected no missing columns when schema is missing")
	assert.Len(t, diff.ExtraColumns, 0, "expected no extra columns")
	assert.Len(t, diff.MissingTables, 0, "expected no missing tables when schema is missing")
}
