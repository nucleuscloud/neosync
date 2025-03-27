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
	Enums                    []*sqlmanager_shared.EnumDataType
	Domains                  []*sqlmanager_shared.DomainDataType
	Composites               []*sqlmanager_shared.CompositeDataType
}

type EnumDiff struct {
	Enum          *sqlmanager_shared.EnumDataType
	NewValues     []string
	ChangedValues map[string]string
}

type Different struct {
	Columns                  []*sqlmanager_shared.TableColumn
	NonForeignKeyConstraints []*sqlmanager_shared.NonForeignKeyConstraint
	ForeignKeyConstraints    []*sqlmanager_shared.ForeignKeyConstraint
	Triggers                 []*sqlmanager_shared.TableTrigger
	Functions                []*sqlmanager_shared.DataType
	Enums                    []*EnumDiff
}
type ExistsInBoth struct {
	Tables []*sqlmanager_shared.SchemaTable

	// exists in both source and destination but have a fingerprint difference
	Different *Different
}

type ExistsInDestination struct {
	Columns                  []*sqlmanager_shared.TableColumn
	NonForeignKeyConstraints []*sqlmanager_shared.NonForeignKeyConstraint
	ForeignKeyConstraints    []*sqlmanager_shared.ForeignKeyConstraint
	Triggers                 []*sqlmanager_shared.TableTrigger
	Functions                []*sqlmanager_shared.DataType
	Enums                    []*sqlmanager_shared.EnumDataType
	Domains                  []*sqlmanager_shared.DomainDataType
	Composites               []*sqlmanager_shared.CompositeDataType
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
	ForeignKeyConstraints    map[string]*sqlmanager_shared.ForeignKeyConstraint    // map of key -> foreign key constraint
	NonForeignKeyConstraints map[string]*sqlmanager_shared.NonForeignKeyConstraint // map of key -> non foreign key constraint
	Triggers                 map[string]*sqlmanager_shared.TableTrigger            // map of key -> trigger
	Functions                map[string]*sqlmanager_shared.DataType                // map of key -> function
	Domains                  map[string]*sqlmanager_shared.DomainDataType          // map of key -> domain
	Enums                    map[string]*sqlmanager_shared.EnumDataType            // map of key -> enum
	Composites               map[string]*sqlmanager_shared.CompositeDataType       // map of key -> composite
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
				Different: &Different{
					Columns:                  []*sqlmanager_shared.TableColumn{},
					NonForeignKeyConstraints: []*sqlmanager_shared.NonForeignKeyConstraint{},
					ForeignKeyConstraints:    []*sqlmanager_shared.ForeignKeyConstraint{},
					Triggers:                 []*sqlmanager_shared.TableTrigger{},
					Functions:                []*sqlmanager_shared.DataType{},
				},
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
	b.buildTableEnumDifferences()
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

			for _, column := range sourceTable {
				destColumn, ok := destTable[column.Name]
				if !ok {
					continue
				}
				if column.Fingerprint != destColumn.Fingerprint {
					b.diff.ExistsInBoth.Different.Columns = append(b.diff.ExistsInBoth.Different.Columns, column)
				}
			}
		}
	}
}

func (b *SchemaDifferencesBuilder) buildTableForeignKeyConstraintDifferences() {
	existsInSource, existsInBoth, existsInDestination := buildDifferencesByFingerprint(b.source.ForeignKeyConstraints, b.destination.ForeignKeyConstraints)
	b.diff.ExistsInSource.ForeignKeyConstraints = existsInSource
	b.diff.ExistsInBoth.Different.ForeignKeyConstraints = existsInBoth
	b.diff.ExistsInDestination.ForeignKeyConstraints = existsInDestination
}

func (b *SchemaDifferencesBuilder) buildTableNonForeignKeyConstraintDifferences() {
	existsInSource, existsInBoth, existsInDestination := buildDifferencesByFingerprint(b.source.NonForeignKeyConstraints, b.destination.NonForeignKeyConstraints)
	b.diff.ExistsInSource.NonForeignKeyConstraints = existsInSource
	b.diff.ExistsInBoth.Different.NonForeignKeyConstraints = existsInBoth
	b.diff.ExistsInDestination.NonForeignKeyConstraints = existsInDestination
}

func (b *SchemaDifferencesBuilder) buildTableTriggerDifferences() {
	existsInSource, existsInBoth, existsInDestination := buildDifferencesByFingerprint(b.source.Triggers, b.destination.Triggers)
	b.diff.ExistsInSource.Triggers = existsInSource
	b.diff.ExistsInBoth.Different.Triggers = existsInBoth
	b.diff.ExistsInDestination.Triggers = existsInDestination
}

func (b *SchemaDifferencesBuilder) buildSchemaFunctionDifferences() {
	existsInSource, existsInBoth, existsInDestination := buildDifferencesByFingerprint(b.source.Functions, b.destination.Functions)
	b.diff.ExistsInSource.Functions = existsInSource
	b.diff.ExistsInBoth.Different.Functions = existsInBoth
	b.diff.ExistsInDestination.Functions = existsInDestination
}

func (b *SchemaDifferencesBuilder) buildTableEnumDifferences() {
	for key, srcEnum := range b.source.Enums {
		if destEnum, ok := b.destination.Enums[key]; ok {
			newValues := []string{}
			changedValues := map[string]string{}
			for idx, srcValue := range srcEnum.Values {
				if idx >= len(destEnum.Values) {
					newValues = append(newValues, srcValue)
				} else if srcValue != destEnum.Values[idx] {
					changedValues[destEnum.Values[idx]] = srcValue
				}
			}
			b.diff.ExistsInBoth.Different.Enums = append(b.diff.ExistsInBoth.Different.Enums, &EnumDiff{
				Enum:          srcEnum,
				NewValues:     newValues,
				ChangedValues: changedValues,
			})
		} else {
			b.diff.ExistsInSource.Enums = append(b.diff.ExistsInSource.Enums, srcEnum)
		}
	}

	for key, destEnum := range b.destination.Enums {
		if _, ok := b.source.Enums[key]; !ok {
			b.diff.ExistsInDestination.Enums = append(b.diff.ExistsInDestination.Enums, destEnum)
		}
	}
}

type FingerprintedType interface {
	GetFingerprint() string
}

// buildDifferencesForMap compares two maps keyed by an identifier.
// It appends items in `src` that are not in `dest` to `existsInSource`,
// and items in `dest` not in `src` to `existsInDestination`.
func buildDifferencesByFingerprint[T FingerprintedType](
	src, dest map[string]T,
) (existsInSource, existsInBoth, existsInDestination []T) {
	inSource := []T{}
	inDestination := []T{}
	inBoth := []T{}
	for key, srcVal := range src {
		if _, ok := dest[key]; !ok {
			inSource = append(inSource, srcVal)
		} else if dest[key].GetFingerprint() != srcVal.GetFingerprint() {
			inBoth = append(inBoth, srcVal)
		}
	}

	for key, destVal := range dest {
		if _, ok := src[key]; !ok {
			inDestination = append(inDestination, destVal)
		}
	}
	return inSource, inBoth, inDestination
}
