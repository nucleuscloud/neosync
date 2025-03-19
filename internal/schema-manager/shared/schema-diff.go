package schemamanager_shared

import (
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
)

type ExistsInSource struct {
	Tables                   []*sqlmanager_shared.SchemaTable
	Columns                  []*sqlmanager_shared.TableColumn
	NonForeignKeyConstraints []*sqlmanager_shared.NonForeignKeyConstraint
	ForeignKeyConstraints    []*sqlmanager_shared.ForeignKeyConstraint
	Triggers                 []*sqlmanager_shared.TableTrigger
	Functions                []*sqlmanager_shared.DataType
}

type ExistsInBoth struct {
	Tables  []*sqlmanager_shared.SchemaTable
	Columns []*sqlmanager_shared.TableColumn // only columns that exist in both source and destination and have a fingerprint difference
}

type ExistsInDestination struct {
	Columns                  []*sqlmanager_shared.TableColumn
	NonForeignKeyConstraints []*sqlmanager_shared.NonForeignKeyConstraint
	ForeignKeyConstraints    []*sqlmanager_shared.ForeignKeyConstraint
	Triggers                 []*sqlmanager_shared.TableTrigger
	Functions                []*sqlmanager_shared.DataType
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
	Columns                  map[string]map[string]*sqlmanager_shared.TableColumn  // map of schema.table -> column name -> column info
	ForeignKeyConstraints    map[string]*sqlmanager_shared.ForeignKeyConstraint    // map of fingerprint -> foreign key constraint
	NonForeignKeyConstraints map[string]*sqlmanager_shared.NonForeignKeyConstraint // map of fingerprint -> non foreign key constraint
	Triggers                 map[string]*sqlmanager_shared.TableTrigger            // map of fingerprint -> trigger
	Functions                map[string]*sqlmanager_shared.DataType                // map of fingerprint -> function
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
				Columns:                  []*sqlmanager_shared.TableColumn{},
				NonForeignKeyConstraints: []*sqlmanager_shared.NonForeignKeyConstraint{},
				ForeignKeyConstraints:    []*sqlmanager_shared.ForeignKeyConstraint{},
				Triggers:                 []*sqlmanager_shared.TableTrigger{},
				Functions:                []*sqlmanager_shared.DataType{},
			},
			ExistsInDestination: &ExistsInDestination{
				Columns:                  []*sqlmanager_shared.TableColumn{},
				NonForeignKeyConstraints: []*sqlmanager_shared.NonForeignKeyConstraint{},
				ForeignKeyConstraints:    []*sqlmanager_shared.ForeignKeyConstraint{},
				Triggers:                 []*sqlmanager_shared.TableTrigger{},
				Functions:                []*sqlmanager_shared.DataType{},
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
	b.buildSchemaFunctionDifferences()
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
				_, ok := destTable[column.Name]
				if !ok {
					b.diff.ExistsInSource.Columns = append(b.diff.ExistsInSource.Columns, column)
				}
			}
			for _, column := range destTable {
				_, ok := sourceTable[column.Name]
				if !ok {
					b.diff.ExistsInDestination.Columns = append(b.diff.ExistsInDestination.Columns, column)
				}
			}

			// todo: handle column name changes
			for _, column := range sourceTable {
				destColumn, ok := destTable[column.Name]
				if !ok {
					continue
				}
				if column.Fingerprint != destColumn.Fingerprint {
					b.diff.ExistsInBoth.Columns = append(b.diff.ExistsInBoth.Columns, column)
				}
			}
		}
	}
}

func (b *SchemaDifferencesBuilder) buildTableForeignKeyConstraintDifferences() {
	existsInSource, existsInDestination := buildDifferencesByFingerprint(b.source.ForeignKeyConstraints, b.destination.ForeignKeyConstraints)
	b.diff.ExistsInSource.ForeignKeyConstraints = existsInSource
	b.diff.ExistsInDestination.ForeignKeyConstraints = existsInDestination
}

func (b *SchemaDifferencesBuilder) buildTableNonForeignKeyConstraintDifferences() {
	existsInSource, existsInDestination := buildDifferencesByFingerprint(b.source.NonForeignKeyConstraints, b.destination.NonForeignKeyConstraints)
	b.diff.ExistsInSource.NonForeignKeyConstraints = existsInSource
	b.diff.ExistsInDestination.NonForeignKeyConstraints = existsInDestination
}

func (b *SchemaDifferencesBuilder) buildTableTriggerDifferences() {
	existsInSource, existsInDestination := buildDifferencesByFingerprint(b.source.Triggers, b.destination.Triggers)
	b.diff.ExistsInSource.Triggers = existsInSource
	b.diff.ExistsInDestination.Triggers = existsInDestination
}

func (b *SchemaDifferencesBuilder) buildSchemaFunctionDifferences() {
	existsInSource, existsInDestination := buildDifferencesByFingerprint(b.source.Functions, b.destination.Functions)
	b.diff.ExistsInSource.Functions = existsInSource
	b.diff.ExistsInDestination.Functions = existsInDestination
}

// buildDifferencesForMap compares two maps keyed by fingerprint.
// It appends items in `src` that are not in `dest` to `existsInSource`,
// and items in `dest` not in `src` to `existsInDestination`.
func buildDifferencesByFingerprint[T any](
	src, dest map[string]*T,
) (existsInSource, existsInDestination []*T) {
	existsInSource = []*T{}
	existsInDestination = []*T{}
	// For each item in src, check if missing in dest
	for fingerprint, srcVal := range src {
		if _, ok := dest[fingerprint]; !ok {
			existsInSource = append(existsInSource, srcVal)
		}
	}

	// For each item in dest, check if missing in src
	for fingerprint, destVal := range dest {
		if _, ok := src[fingerprint]; !ok {
			existsInDestination = append(existsInDestination, destVal)
		}
	}
	return existsInSource, existsInDestination
}
