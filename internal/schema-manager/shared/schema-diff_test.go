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
	sourceTriggerNames := []string{
		diff.ExistsInSource.Triggers[0].TriggerName,
		diff.ExistsInSource.Triggers[1].TriggerName,
	}
	require.Contains(t, sourceTriggerNames, "users_trigger")
	require.Contains(t, sourceTriggerNames, "insert_users_trigger")

	require.Len(t, diff.ExistsInDestination.Triggers, 2)
	destTriggerNames := []string{
		diff.ExistsInDestination.Triggers[0].TriggerName,
		diff.ExistsInDestination.Triggers[1].TriggerName,
	}
	require.Contains(t, destTriggerNames, "users_trigger")
	require.Contains(t, destTriggerNames, "delete_users_trigger")
}

func Test_buildTableCompositeDifferences(t *testing.T) {
	jobmappingTables := []*sqlmanager_shared.SchemaTable{
		{Schema: "public", Table: "users"},
	}

	sourceData := &DatabaseData{
		Composites: map[string]*sqlmanager_shared.CompositeDataType{
			"public.address": {
				Schema: "public",
				Name:   "address",
				Attributes: []*sqlmanager_shared.CompositeAttribute{
					{Id: 1, Name: "street", Datatype: "text"},      // changed datatype
					{Id: 2, Name: "street_name", Datatype: "text"}, // changed name
					{Id: 5, Name: "state", Datatype: "text"},       // new attribute
				},
			},
			"public.contact": { // exists in source but not dest
				Schema: "public",
				Name:   "contact",
				Attributes: []*sqlmanager_shared.CompositeAttribute{
					{Id: 1, Name: "email", Datatype: "text"},
					{Id: 2, Name: "phone", Datatype: "text"},
				},
			},
		},
	}

	destData := &DatabaseData{
		Composites: map[string]*sqlmanager_shared.CompositeDataType{
			"public.address": {
				Schema: "public",
				Name:   "address",
				Attributes: []*sqlmanager_shared.CompositeAttribute{
					{Id: 1, Name: "street", Datatype: "varchar(255)"}, // changed datatype
					{Id: 2, Name: "city", Datatype: "text"},           // changed name
					{Id: 3, Name: "country", Datatype: "text"},        // removed
					{Id: 4, Name: "postal_code", Datatype: "text"},    // removed
				},
			},
			"public.person": { // exists in dest but not source
				Schema: "public",
				Name:   "person",
				Attributes: []*sqlmanager_shared.CompositeAttribute{
					{Id: 1, Name: "first_name", Datatype: "text"},
					{Id: 2, Name: "last_name", Datatype: "text"},
				},
			},
		},
	}

	builder := NewSchemaDifferencesBuilder(jobmappingTables, sourceData, destData)
	diff := builder.Build()

	require.NotNil(t, diff)

	// Check composites that exist in source but not in destination
	require.Len(t, diff.ExistsInSource.Composites, 1)
	require.Equal(t, "contact", diff.ExistsInSource.Composites[0].Name)

	// Check composites that exist in destination but not in source
	require.Len(t, diff.ExistsInDestination.Composites, 1)
	require.Equal(t, "person", diff.ExistsInDestination.Composites[0].Name)

	// Check composites that exist in both but have differences
	require.Len(t, diff.ExistsInBoth.Different.Composites, 1)
	compositeDiff := diff.ExistsInBoth.Different.Composites[0]

	// Check changed attribute datatypes
	require.Len(t, compositeDiff.ChangedAttributeDatatype, 1)
	require.Equal(t, "text", compositeDiff.ChangedAttributeDatatype["street"])

	// Check changed attribute names
	require.Len(t, compositeDiff.ChangedAttributeName, 1)
	require.Equal(t, "street_name", compositeDiff.ChangedAttributeName["city"])

	// Check new attributes
	require.Len(t, compositeDiff.NewAttributes, 1)
	require.Equal(t, "text", compositeDiff.NewAttributes["state"])

	// Check removed attributes
	require.Len(t, compositeDiff.RemovedAttributes, 2)
	require.Contains(t, compositeDiff.RemovedAttributes, "country")
	require.Contains(t, compositeDiff.RemovedAttributes, "postal_code")
}

func TestBuildTableEnumDifferences(t *testing.T) {
	jobmappingTables := []*sqlmanager_shared.SchemaTable{}

	sourceData := &DatabaseData{
		Enums: map[string]*sqlmanager_shared.EnumDataType{
			"public.status": { // exists in both but different
				Schema: "public",
				Name:   "status",
				Values: []string{"active", "inactive", "pending", "cancelled"},
			},
			"public.priority": { // exists in source only
				Schema: "public",
				Name:   "priority",
				Values: []string{"low", "medium", "high"},
			},
		},
	}

	destData := &DatabaseData{
		Enums: map[string]*sqlmanager_shared.EnumDataType{
			"public.status": { // exists in both but different
				Schema: "public",
				Name:   "status",
				Values: []string{"active", "disabled", "pending"},
			},
			"public.level": { // exists in dest only
				Schema: "public",
				Name:   "level",
				Values: []string{"beginner", "intermediate", "expert"},
			},
		},
	}

	builder := NewSchemaDifferencesBuilder(jobmappingTables, sourceData, destData)
	diff := builder.Build()

	require.NotNil(t, diff)

	// Check enums that exist in source but not in destination
	require.Len(t, diff.ExistsInSource.Enums, 1)
	require.Equal(t, "priority", diff.ExistsInSource.Enums[0].Name)

	// Check enums that exist in destination but not in source
	require.Len(t, diff.ExistsInDestination.Enums, 1)
	require.Equal(t, "level", diff.ExistsInDestination.Enums[0].Name)

	// Check enums that exist in both but have differences
	require.Len(t, diff.ExistsInBoth.Different.Enums, 1)
	enumDiff := diff.ExistsInBoth.Different.Enums[0]

	// Check new values
	require.Len(t, enumDiff.NewValues, 1)
	require.Equal(t, "cancelled", enumDiff.NewValues[0])

	// Check changed values
	require.Len(t, enumDiff.ChangedValues, 1)
	require.Equal(t, "inactive", enumDiff.ChangedValues["disabled"])
}
