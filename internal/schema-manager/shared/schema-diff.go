package schemamanager_shared

import sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"

type ExistsInSource struct {
	Tables                   []*sqlmanager_shared.SchemaTable
	Columns                  []*sqlmanager_shared.DatabaseSchemaRow
	NonForeignKeyConstraints []*sqlmanager_shared.NonForeignKeyConstraint
	ForeignKeyConstraints    []*sqlmanager_shared.ForeignKeyConstraint
}

type ExistsInBoth struct {
	Tables []*sqlmanager_shared.SchemaTable
}

type ExistsInDestination struct {
	Columns                  []*sqlmanager_shared.DatabaseSchemaRow
	NonForeignKeyConstraints []*sqlmanager_shared.NonForeignKeyConstraint
	ForeignKeyConstraints    []*sqlmanager_shared.ForeignKeyConstraint
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
	Columns          map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow
	TableConstraints map[string]*sqlmanager_shared.AllTableConstraints
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
	b.buildTableConstraintDifferences()
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

func (b *SchemaDifferencesBuilder) buildTableConstraintDifferences() {
	for _, table := range b.diff.ExistsInBoth.Tables {
		srcTableConstraints, hasSrcConstraints := b.source.TableConstraints[table.String()]
		dstTableConstraints, hasDstConstraints := b.destination.TableConstraints[table.String()]

		// if there's nothing in source but something in dest => all in dest need to be dropped
		if !hasSrcConstraints && hasDstConstraints {
			b.diff.ExistsInDestination.NonForeignKeyConstraints = append(b.diff.ExistsInDestination.NonForeignKeyConstraints, dstTableConstraints.NonForeignKeyConstraints...)
			b.diff.ExistsInDestination.ForeignKeyConstraints = append(b.diff.ExistsInDestination.ForeignKeyConstraints, dstTableConstraints.ForeignKeyConstraints...)
		} else if hasSrcConstraints && !hasDstConstraints {
			// if there's constraints in source but none in dest => all in source need to be created
			b.diff.ExistsInSource.NonForeignKeyConstraints = append(b.diff.ExistsInSource.NonForeignKeyConstraints, srcTableConstraints.NonForeignKeyConstraints...)
			b.diff.ExistsInSource.ForeignKeyConstraints = append(b.diff.ExistsInSource.ForeignKeyConstraints, srcTableConstraints.ForeignKeyConstraints...)
		} else if hasSrcConstraints && hasDstConstraints {
			// if there's constraints in both source and destination compare them
			b.buildNonFkDifferences(srcTableConstraints, dstTableConstraints)
			b.buildTableFkDifferences(srcTableConstraints, dstTableConstraints)
		}
	}
}

func (b *SchemaDifferencesBuilder) buildTableFkDifferences(srcTableConstraints, dstTableConstraints *sqlmanager_shared.AllTableConstraints) {
	srcFkMap := make(map[string]*sqlmanager_shared.ForeignKeyConstraint)
	for _, c := range srcTableConstraints.ForeignKeyConstraints {
		srcFkMap[c.Fingerprint] = c
	}

	dstFkMap := make(map[string]*sqlmanager_shared.ForeignKeyConstraint)
	for _, c := range dstTableConstraints.ForeignKeyConstraints {
		dstFkMap[c.Fingerprint] = c
	}

	// in source but not in destination
	for fingerprint, cObj := range srcFkMap {
		if _, ok := dstFkMap[fingerprint]; !ok {
			b.diff.ExistsInSource.ForeignKeyConstraints = append(b.diff.ExistsInSource.ForeignKeyConstraints, cObj)
		}
	}
	// in destination but not in source
	for fingerprint, cObj := range dstFkMap {
		if _, ok := srcFkMap[fingerprint]; !ok {
			b.diff.ExistsInDestination.ForeignKeyConstraints = append(b.diff.ExistsInDestination.ForeignKeyConstraints, cObj)
		}
	}
}

func (b *SchemaDifferencesBuilder) buildNonFkDifferences(srcTableConstraints, dstTableConstraints *sqlmanager_shared.AllTableConstraints) {
	srcNonFkMap := make(map[string]*sqlmanager_shared.NonForeignKeyConstraint)
	for _, c := range srcTableConstraints.NonForeignKeyConstraints {
		srcNonFkMap[c.Fingerprint] = c
	}

	dstNonFkMap := make(map[string]*sqlmanager_shared.NonForeignKeyConstraint)
	for _, c := range dstTableConstraints.NonForeignKeyConstraints {
		dstNonFkMap[c.Fingerprint] = c
	}

	// in source but not in destination
	for fingerprint, cObj := range srcNonFkMap {
		if _, ok := dstNonFkMap[fingerprint]; !ok {
			b.diff.ExistsInSource.NonForeignKeyConstraints = append(b.diff.ExistsInSource.NonForeignKeyConstraints, cObj)
		}
	}
	// in destination but not in source
	for fingerprint, cObj := range dstNonFkMap {
		if _, ok := srcNonFkMap[fingerprint]; !ok {
			b.diff.ExistsInDestination.NonForeignKeyConstraints = append(b.diff.ExistsInDestination.NonForeignKeyConstraints, cObj)
		}
	}
}
