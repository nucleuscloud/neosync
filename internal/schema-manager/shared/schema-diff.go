package schemamanager_shared

import sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"

type ExistsInSource struct {
	Tables                   []*sqlmanager_shared.SchemaTable
	Columns                  []*sqlmanager_shared.DatabaseSchemaRow
	NonForeignKeyConstraints []*sqlmanager_shared.NonForeignKeyConstraint
	ForeignKeyConstraints    []*sqlmanager_shared.ForeignKeyConstraint
	Triggers                 []*sqlmanager_shared.TableTrigger
}

type ExistsInBoth struct {
	Tables []*sqlmanager_shared.SchemaTable
}

type ExistsInDestination struct {
	Columns                  []*sqlmanager_shared.DatabaseSchemaRow
	NonForeignKeyConstraints []*sqlmanager_shared.NonForeignKeyConstraint
	ForeignKeyConstraints    []*sqlmanager_shared.ForeignKeyConstraint
	Triggers                 []*sqlmanager_shared.TableTrigger
}

type SchemaDifferences struct {
	// Exists in source but not destination
	ExistsInSource *ExistsInSource
	// Exists in both source and destination
	ExistsInBoth *ExistsInBoth
	// Exists in destination but not source
	ExistsInDestination *ExistsInDestination
}

type DatabaseData struct {
	Columns                  map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow       // map of schema.table -> column name -> column info
	ForeignKeyConstraints    map[string]map[string]*sqlmanager_shared.ForeignKeyConstraint    // map of schema.table -> fingerprint -> foreign key constraint
	NonForeignKeyConstraints map[string]map[string]*sqlmanager_shared.NonForeignKeyConstraint // map of schema.table -> fingerprint -> non foreign key constraint
	Triggers                 map[string]map[string]*sqlmanager_shared.TableTrigger            // map of schema.table -> fingerprint -> trigger
}

type SchemaDifferencesBuilder struct {
	diff             *SchemaDifferences
	source           *DatabaseData
	destination      *DatabaseData
	jobmappingTables []*sqlmanager_shared.SchemaTable
}

// NewSchemaDifferencesBuilder initializes a new builder with empty slices.
func NewSchemaDifferencesBuilder(
	jobmappingTables []*sqlmanager_shared.SchemaTable,
	sourceData *DatabaseData,
	destData *DatabaseData,
) *SchemaDifferencesBuilder {
	return &SchemaDifferencesBuilder{
		jobmappingTables: jobmappingTables,
		source:           sourceData,
		destination:      destData,
		diff: &SchemaDifferences{
			ExistsInSource: &ExistsInSource{
				Tables:                   []*sqlmanager_shared.SchemaTable{},
				Columns:                  []*sqlmanager_shared.DatabaseSchemaRow{},
				NonForeignKeyConstraints: []*sqlmanager_shared.NonForeignKeyConstraint{},
				ForeignKeyConstraints:    []*sqlmanager_shared.ForeignKeyConstraint{},
			},
			ExistsInDestination: &ExistsInDestination{
				Columns:                  []*sqlmanager_shared.DatabaseSchemaRow{},
				NonForeignKeyConstraints: []*sqlmanager_shared.NonForeignKeyConstraint{},
				ForeignKeyConstraints:    []*sqlmanager_shared.ForeignKeyConstraint{},
			},
			ExistsInBoth: &ExistsInBoth{
				Tables: []*sqlmanager_shared.SchemaTable{},
			},
		},
	}
}

func (b *SchemaDifferencesBuilder) Build() *SchemaDifferences {
	b.buildTableColumnDifferences()
	b.buildTableForeignKeyConstraintDifferences()
	b.buildTableNonForeignKeyConstraintDifferences()
	b.buildTableTriggerDifferences()
	return b.diff
}

func (b *SchemaDifferencesBuilder) buildTableColumnDifferences() {
	for _, table := range b.jobmappingTables {
		sourceTable := b.source.Columns[table.String()]
		destTable := b.destination.Columns[table.String()]
		if len(sourceTable) > 0 && len(destTable) == 0 {
			b.diff.ExistsInSource.Tables = append(b.diff.ExistsInSource.Tables, table)
		} else if len(sourceTable) > 0 && len(destTable) > 0 {
			// table exists in both source and destination
			b.diff.ExistsInBoth.Tables = append(b.diff.ExistsInBoth.Tables, table)

			// column diff
			for _, column := range sourceTable {
				_, ok := destTable[column.ColumnName]
				if !ok {
					b.diff.ExistsInSource.Columns = append(b.diff.ExistsInSource.Columns, column)
				}
			}
			for _, column := range destTable {
				_, ok := sourceTable[column.ColumnName]
				if !ok {
					b.diff.ExistsInDestination.Columns = append(b.diff.ExistsInDestination.Columns, column)
				}
			}
		}
	}
}

func (b *SchemaDifferencesBuilder) buildTableForeignKeyConstraintDifferences() {
	for _, table := range b.diff.ExistsInBoth.Tables {
		srcTableFkConstraints := b.source.ForeignKeyConstraints[table.String()]
		dstTableFkConstraints := b.destination.ForeignKeyConstraints[table.String()]

		for _, fkConstraint := range srcTableFkConstraints {
			_, ok := dstTableFkConstraints[fkConstraint.Fingerprint]
			if !ok {
				b.diff.ExistsInSource.ForeignKeyConstraints = append(b.diff.ExistsInSource.ForeignKeyConstraints, fkConstraint)
			}
		}

		for _, fkConstraint := range dstTableFkConstraints {
			_, ok := srcTableFkConstraints[fkConstraint.Fingerprint]
			if !ok {
				b.diff.ExistsInDestination.ForeignKeyConstraints = append(b.diff.ExistsInDestination.ForeignKeyConstraints, fkConstraint)
			}
		}
	}
}

func (b *SchemaDifferencesBuilder) buildTableNonForeignKeyConstraintDifferences() {
	for _, table := range b.diff.ExistsInBoth.Tables {
		srcTableNonFkConstraints := b.source.NonForeignKeyConstraints[table.String()]
		dstTableNonFkConstraints := b.destination.NonForeignKeyConstraints[table.String()]

		for _, nonFkConstraint := range srcTableNonFkConstraints {
			_, ok := dstTableNonFkConstraints[nonFkConstraint.Fingerprint]
			if !ok {
				b.diff.ExistsInSource.NonForeignKeyConstraints = append(b.diff.ExistsInSource.NonForeignKeyConstraints, nonFkConstraint)
			}
		}

		for _, nonFkConstraint := range dstTableNonFkConstraints {
			_, ok := srcTableNonFkConstraints[nonFkConstraint.Fingerprint]
			if !ok {
				b.diff.ExistsInDestination.NonForeignKeyConstraints = append(b.diff.ExistsInDestination.NonForeignKeyConstraints, nonFkConstraint)
			}
		}
	}
}

func (b *SchemaDifferencesBuilder) buildTableTriggerDifferences() {
	for _, table := range b.diff.ExistsInBoth.Tables {
		srcTriggers := b.source.Triggers[table.String()]
		dstTriggers := b.destination.Triggers[table.String()]
		for _, trigger := range srcTriggers {
			_, ok := dstTriggers[trigger.Fingerprint]
			if !ok {
				b.diff.ExistsInSource.Triggers = append(b.diff.ExistsInSource.Triggers, trigger)
			}
		}

		for _, trigger := range dstTriggers {
			_, ok := srcTriggers[trigger.Fingerprint]
			if !ok {
				b.diff.ExistsInDestination.Triggers = append(b.diff.ExistsInDestination.Triggers, trigger)
			}
		}
	}
}
