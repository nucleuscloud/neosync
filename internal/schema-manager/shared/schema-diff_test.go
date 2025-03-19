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
		Columns:                  map[string]map[string]*sqlmanager_shared.TableColumn{},
		ForeignKeyConstraints:    map[string]*sqlmanager_shared.ForeignKeyConstraint{},
		NonForeignKeyConstraints: map[string]*sqlmanager_shared.NonForeignKeyConstraint{},
		Triggers:                 map[string]*sqlmanager_shared.TableTrigger{},
	}
	destData := &DatabaseData{
		Columns:                  map[string]map[string]*sqlmanager_shared.TableColumn{},
		ForeignKeyConstraints:    map[string]*sqlmanager_shared.ForeignKeyConstraint{},
		NonForeignKeyConstraints: map[string]*sqlmanager_shared.NonForeignKeyConstraint{},
		Triggers:                 map[string]*sqlmanager_shared.TableTrigger{},
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
		Columns: map[string]map[string]*sqlmanager_shared.TableColumn{
			"public.users": {
				"id": {
					Name:   "id",
					Schema: "public",
					Table:  "users",
				},
				"name": {
					Name:   "name",
					Schema: "public",
					Table:  "users",
				},
			},
		},
		ForeignKeyConstraints:    map[string]*sqlmanager_shared.ForeignKeyConstraint{},
		NonForeignKeyConstraints: map[string]*sqlmanager_shared.NonForeignKeyConstraint{},
		Triggers:                 map[string]*sqlmanager_shared.TableTrigger{},
	}

	destData := &DatabaseData{
		Columns: map[string]map[string]*sqlmanager_shared.TableColumn{
			"public.users": {
				"id": {
					Name:   "id",
					Schema: "public",
					Table:  "users",
				},
				"email": {
					Name:   "email",
					Schema: "public",
					Table:  "users",
				},
			},
		},
		ForeignKeyConstraints:    map[string]*sqlmanager_shared.ForeignKeyConstraint{},
		NonForeignKeyConstraints: map[string]*sqlmanager_shared.NonForeignKeyConstraint{},
		Triggers:                 map[string]*sqlmanager_shared.TableTrigger{},
	}

	builder := NewSchemaDifferencesBuilder(jobmappingTables, sourceData, destData)
	diff := builder.Build()

	require.NotNil(t, diff)
	require.Len(t, diff.ExistsInBoth.Tables, 1)
	require.Equal(t, "users", diff.ExistsInBoth.Tables[0].Table)
	require.Len(t, diff.ExistsInSource.Columns, 1)
	require.Equal(t, "name", diff.ExistsInSource.Columns[0].Name)
	require.Len(t, diff.ExistsInDestination.Columns, 1)
	require.Equal(t, "email", diff.ExistsInDestination.Columns[0].Name)
}

func Test_Build_TableConstraintDifferences(t *testing.T) {
	jobmappingTables := []*sqlmanager_shared.SchemaTable{
		{Schema: "public", Table: "users"},
	}

	sourceData := &DatabaseData{
		Columns: map[string]map[string]*sqlmanager_shared.TableColumn{
			"public.users": {
				"id": {
					Name:   "id",
					Schema: "public",
					Table:  "users",
				},
			},
		},
		NonForeignKeyConstraints: map[string]*sqlmanager_shared.NonForeignKeyConstraint{
			"pk_users_fingerprint": {
				SchemaName:     "public",
				TableName:      "users",
				ConstraintName: "pk_users",
				ConstraintType: "PRIMARY KEY",
				Fingerprint:    "pk_users_fingerprint",
			},
		},
		ForeignKeyConstraints: map[string]*sqlmanager_shared.ForeignKeyConstraint{
			"fk_users_roles_fingerprint": {
				ReferencedSchema:  "public",
				ReferencedTable:   "roles",
				ReferencingSchema: "public",
				ReferencingTable:  "users",
				ConstraintName:    "fk_users_roles",
				ConstraintType:    "FOREIGN KEY",
				Fingerprint:       "fk_users_roles_fingerprint",
			},
		},
		Triggers: map[string]*sqlmanager_shared.TableTrigger{},
	}

	destData := &DatabaseData{
		Columns: map[string]map[string]*sqlmanager_shared.TableColumn{
			"public.users": {
				"id": {
					Name:   "id",
					Schema: "public",
					Table:  "users",
				},
			},
		},
		NonForeignKeyConstraints: map[string]*sqlmanager_shared.NonForeignKeyConstraint{
			"unique_email_fingerprint": {
				SchemaName:     "public",
				TableName:      "users",
				ConstraintName: "unique_email",
				ConstraintType: "UNIQUE",
				Fingerprint:    "unique_email_fingerprint",
			},
		},
		ForeignKeyConstraints: map[string]*sqlmanager_shared.ForeignKeyConstraint{
			"fk_users_teams_fingerprint": {
				ReferencedSchema:  "public",
				ReferencedTable:   "teams",
				ReferencingSchema: "public",
				ReferencingTable:  "users",
				ConstraintName:    "fk_users_teams",
				ConstraintType:    "FOREIGN KEY",
				Fingerprint:       "fk_users_teams_fingerprint",
			},
		},
		Triggers: map[string]*sqlmanager_shared.TableTrigger{},
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

func Test_Build_TableTriggerDifferences(t *testing.T) {
	jobmappingTables := []*sqlmanager_shared.SchemaTable{
		{Schema: "public", Table: "users"},
	}

	sourceData := &DatabaseData{
		Columns: map[string]map[string]*sqlmanager_shared.TableColumn{
			"public.users": {
				"id": {
					Name:   "id",
					Schema: "public",
					Table:  "users",
				},
			},
		},
		ForeignKeyConstraints:    map[string]*sqlmanager_shared.ForeignKeyConstraint{},
		NonForeignKeyConstraints: map[string]*sqlmanager_shared.NonForeignKeyConstraint{},
		Triggers: map[string]*sqlmanager_shared.TableTrigger{
			"users_trigger_fingerprint": {
				Schema:      "public",
				Table:       "users",
				TriggerName: "users_trigger",
				Definition:  "CREATE TRIGGER users_trigger...",
				Fingerprint: "users_trigger_fingerprint",
			},
			"insert_users_trigger_fingerprint": {
				Schema:      "public",
				Table:       "users",
				TriggerName: "insert_users_trigger",
				Definition:  "CREATE TRIGGER insert_users_trigger...",
				Fingerprint: "insert_users_trigger_fingerprint",
			},
		},
	}

	destData := &DatabaseData{
		Columns: map[string]map[string]*sqlmanager_shared.TableColumn{
			"public.users": {
				"id": {
					Name:   "id",
					Schema: "public",
					Table:  "users",
				},
			},
		},
		ForeignKeyConstraints:    map[string]*sqlmanager_shared.ForeignKeyConstraint{},
		NonForeignKeyConstraints: map[string]*sqlmanager_shared.NonForeignKeyConstraint{},
		Triggers: map[string]*sqlmanager_shared.TableTrigger{
			"update_users_trigger_fingerprint": {
				Schema:      "public",
				Table:       "users",
				TriggerName: "users_trigger",
				Definition:  "CREATE TRIGGER update_users_trigger...",
				Fingerprint: "update_users_trigger_fingerprint",
			},
			"delete_users_trigger_fingerprint": {
				Schema:      "public",
				Table:       "users",
				TriggerName: "delete_users_trigger",
				Definition:  "CREATE TRIGGER delete_users_trigger...",
				Fingerprint: "delete_users_trigger_fingerprint",
			},
		},
	}

	builder := NewSchemaDifferencesBuilder(jobmappingTables, sourceData, destData)
	diff := builder.Build()

	require.NotNil(t, diff)
	require.Len(t, diff.ExistsInSource.Triggers, 2)
	require.Equal(t, "users_trigger", diff.ExistsInSource.Triggers[0].TriggerName)
	require.Equal(t, "insert_users_trigger", diff.ExistsInSource.Triggers[1].TriggerName)
	require.Len(t, diff.ExistsInDestination.Triggers, 2)
	require.Equal(t, "users_trigger", diff.ExistsInDestination.Triggers[0].TriggerName)
	require.Equal(t, "delete_users_trigger", diff.ExistsInDestination.Triggers[1].TriggerName)
}
