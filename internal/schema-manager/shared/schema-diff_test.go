package schemamanager_shared

import (
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/require"
)

func Test_NewSchemaDifferencesBuilder(t *testing.T) {
	jobmappingTables := []*sqlmanager_shared.SchemaTable{
		{Schema: "public", Table: "users"},
	}
	sourceData := &DatabaseData{
		Columns:          map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{},
		TableConstraints: map[string]*sqlmanager_shared.AllTableConstraints{},
	}
	destData := &DatabaseData{
		Columns:          map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{},
		TableConstraints: map[string]*sqlmanager_shared.AllTableConstraints{},
	}

	builder := NewSchemaDifferencesBuilder(jobmappingTables, sourceData, destData)

	require.NotNil(t, builder)
	require.NotNil(t, builder.diff)
	require.NotNil(t, builder.diff.ExistsInSource)
	require.NotNil(t, builder.diff.ExistsInDestination)
	require.NotNil(t, builder.diff.ExistsInBoth)
	require.Empty(t, builder.diff.ExistsInSource.Tables)
	require.Empty(t, builder.diff.ExistsInSource.Columns)
	require.Empty(t, builder.diff.ExistsInSource.NonForeignKeyConstraints)
	require.Empty(t, builder.diff.ExistsInSource.ForeignKeyConstraints)
	require.Empty(t, builder.diff.ExistsInDestination.Columns)
	require.Empty(t, builder.diff.ExistsInDestination.NonForeignKeyConstraints)
	require.Empty(t, builder.diff.ExistsInDestination.ForeignKeyConstraints)
	require.Empty(t, builder.diff.ExistsInBoth.Tables)
}

func Test_Build_TableColumnDifferences(t *testing.T) {
	jobmappingTables := []*sqlmanager_shared.SchemaTable{
		{Schema: "public", Table: "users"},
	}

	sourceData := &DatabaseData{
		Columns: map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"public.users": {
				"id": {
					TableSchema: "public",
					TableName:   "users",
					ColumnName:  "id",
				},
				"name": {
					TableSchema: "public",
					TableName:   "users",
					ColumnName:  "name",
				},
			},
		},
		TableConstraints: map[string]*sqlmanager_shared.AllTableConstraints{},
	}

	destData := &DatabaseData{
		Columns: map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"public.users": {
				"id": {
					TableSchema: "public",
					TableName:   "users",
					ColumnName:  "id",
				},
				"email": {
					TableSchema: "public",
					TableName:   "users",
					ColumnName:  "email",
				},
			},
		},
		TableConstraints: map[string]*sqlmanager_shared.AllTableConstraints{},
	}

	builder := NewSchemaDifferencesBuilder(jobmappingTables, sourceData, destData)
	diff := builder.Build()

	require.NotNil(t, diff)
	require.Len(t, diff.ExistsInBoth.Tables, 1)
	require.Equal(t, "users", diff.ExistsInBoth.Tables[0].Table)
	require.Len(t, diff.ExistsInSource.Columns, 1)
	require.Equal(t, "name", diff.ExistsInSource.Columns[0].ColumnName)
	require.Len(t, diff.ExistsInDestination.Columns, 1)
	require.Equal(t, "email", diff.ExistsInDestination.Columns[0].ColumnName)
}

func Test_Build_TableConstraintDifferences(t *testing.T) {
	jobmappingTables := []*sqlmanager_shared.SchemaTable{
		{Schema: "public", Table: "users"},
	}

	sourceData := &DatabaseData{
		Columns: map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"public.users": {
				"id": {
					TableSchema: "public",
					TableName:   "users",
					ColumnName:  "id",
				},
			},
		},
		TableConstraints: map[string]*sqlmanager_shared.AllTableConstraints{
			"public.users": {
				NonForeignKeyConstraints: []*sqlmanager_shared.NonForeignKeyConstraint{
					{
						SchemaName:     "public",
						TableName:      "users",
						ConstraintName: "pk_users",
						ConstraintType: "PRIMARY KEY",
						Fingerprint:    "pk_users_fingerprint",
					},
				},
				ForeignKeyConstraints: []*sqlmanager_shared.ForeignKeyConstraint{
					{
						ReferencedSchema:  "public",
						ReferencedTable:   "roles",
						ReferencingSchema: "public",
						ReferencingTable:  "users",
						ConstraintName:    "fk_users_roles",
						ConstraintType:    "FOREIGN KEY",
						Fingerprint:       "fk_users_roles_fingerprint",
					},
				},
			},
		},
	}

	destData := &DatabaseData{
		Columns: map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"public.users": {
				"id": {
					TableSchema: "public",
					TableName:   "users",
					ColumnName:  "id",
				},
			},
		},
		TableConstraints: map[string]*sqlmanager_shared.AllTableConstraints{
			"public.users": {
				NonForeignKeyConstraints: []*sqlmanager_shared.NonForeignKeyConstraint{
					{
						SchemaName:     "public",
						TableName:      "users",
						ConstraintName: "unique_email",
						ConstraintType: "UNIQUE",
						Fingerprint:    "unique_email_fingerprint",
					},
				},
				ForeignKeyConstraints: []*sqlmanager_shared.ForeignKeyConstraint{
					{
						ReferencedSchema:  "public",
						ReferencedTable:   "teams",
						ReferencingSchema: "public",
						ReferencingTable:  "users",
						ConstraintName:    "fk_users_teams",
						ConstraintType:    "FOREIGN KEY",
						Fingerprint:       "fk_users_teams_fingerprint",
					},
				},
			},
		},
	}

	builder := NewSchemaDifferencesBuilder(jobmappingTables, sourceData, destData)
	diff := builder.Build()

	require.NotNil(t, diff)
	require.Len(t, diff.ExistsInSource.NonForeignKeyConstraints, 1)
	require.Equal(t, "pk_users", diff.ExistsInSource.NonForeignKeyConstraints[0].ConstraintName)
	require.Len(t, diff.ExistsInSource.ForeignKeyConstraints, 1)
	require.Equal(t, "fk_users_roles", diff.ExistsInSource.ForeignKeyConstraints[0].ConstraintName)
	require.Len(t, diff.ExistsInDestination.NonForeignKeyConstraints, 1)
	require.Equal(t, "unique_email", diff.ExistsInDestination.NonForeignKeyConstraints[0].ConstraintName)
	require.Len(t, diff.ExistsInDestination.ForeignKeyConstraints, 1)
	require.Equal(t, "fk_users_teams", diff.ExistsInDestination.ForeignKeyConstraints[0].ConstraintName)
}
