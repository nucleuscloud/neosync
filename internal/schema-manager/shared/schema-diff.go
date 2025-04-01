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

type CompositeDiff struct {
	Composite                *sqlmanager_shared.CompositeDataType
	ChangedAttributeDatatype map[string]string
	NewAttributes            map[string]string
	RemovedAttributes        []string
	ChangedAttributeName     map[string]string
}

type DomainDiff struct {
	Domain             *sqlmanager_shared.DomainDataType
	IsNullDifferent    bool
	IsDefaultDifferent bool

	// constraints
	NewConstraints     map[string]string
	RemovedConstraints []string
}

type ColumnAction string

const (
	// datatype
	SetDatatype ColumnAction = "SET_DATATYPE"

	// default
	SetDefault  ColumnAction = "SET_DEFAULT"
	DropDefault ColumnAction = "DROP_DEFAULT"

	// not null
	SetNotNull  ColumnAction = "SET_NOT_NULL"
	DropNotNull ColumnAction = "DROP_NOT_NULL"

	// identity
	SetIdentity  ColumnAction = "SET_IDENTITY"
	DropIdentity ColumnAction = "DROP_IDENTITY"
)

type ColumnRename struct {
	OldName string
}

type ColumnDiff struct {
	Column       *sqlmanager_shared.TableColumn
	RenameColumn *ColumnRename
	// Actions represents what needs to be updated on the column
	Actions []ColumnAction
}

type Different struct {
	Columns                  []*ColumnDiff
	NonForeignKeyConstraints []*sqlmanager_shared.NonForeignKeyConstraint
	ForeignKeyConstraints    []*sqlmanager_shared.ForeignKeyConstraint
	Triggers                 []*sqlmanager_shared.TableTrigger
	Functions                []*sqlmanager_shared.DataType
	Enums                    []*EnumDiff
	Composites               []*CompositeDiff
	Domains                  []*DomainDiff
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
					Columns:                  []*ColumnDiff{},
					NonForeignKeyConstraints: []*sqlmanager_shared.NonForeignKeyConstraint{},
					ForeignKeyConstraints:    []*sqlmanager_shared.ForeignKeyConstraint{},
					Triggers:                 []*sqlmanager_shared.TableTrigger{},
					Functions:                []*sqlmanager_shared.DataType{},
					Enums:                    []*EnumDiff{},
					Composites:               []*CompositeDiff{},
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
	b.buildTableCompositeDifferences()
	b.buildTableDomainDifferences()
	return b.diff
}

func (b *SchemaDifferencesBuilder) buildTableColumnDifferences() {
	for _, table := range b.jobmappingTables {
		sourceTableCols := b.source.Columns[table.String()]
		destTableCols := b.destination.Columns[table.String()]
		if len(sourceTableCols) > 0 && len(destTableCols) == 0 {
			b.diff.ExistsInSource.Tables = append(b.diff.ExistsInSource.Tables, table)
		} else if len(sourceTableCols) > 0 && len(destTableCols) > 0 {
			// table exists in both source and destination
			b.diff.ExistsInBoth.Tables = append(b.diff.ExistsInBoth.Tables, table)

			// column diff
			for _, srcColumn := range sourceTableCols {
				destColumn := findColumn(destTableCols, srcColumn)
				if destColumn == nil {
					b.diff.ExistsInSource.Columns = append(b.diff.ExistsInSource.Columns, srcColumn)
				} else if srcColumn.Fingerprint != destColumn.Fingerprint {
					// column differences
					actions := []ColumnAction{}
					if srcColumn.DataType != destColumn.DataType {
						actions = append(actions, SetDatatype)
					}
					if srcColumn.ColumnDefault != destColumn.ColumnDefault {
						defaultAction := DropDefault
						if srcColumn.ColumnDefault != "" {
							defaultAction = SetDefault
						}
						actions = append(actions, defaultAction)
					}

					switch {
					case srcColumn.IdentityGeneration == nil && destColumn.IdentityGeneration != nil:
						actions = append(actions, DropIdentity)
					case srcColumn.IdentityGeneration != nil && destColumn.IdentityGeneration == nil:
						actions = append(actions, SetIdentity)
					case *srcColumn.IdentityGeneration != *destColumn.IdentityGeneration && *srcColumn.IdentityGeneration != "":
						actions = append(actions, SetIdentity)
					}

					if srcColumn.IsNullable != destColumn.IsNullable {
						nullableAction := SetNotNull
						if srcColumn.IsNullable {
							nullableAction = DropNotNull
						}
						actions = append(actions, nullableAction)
					}

					var renameColumn *ColumnRename
					if srcColumn.Name != destColumn.Name {
						renameColumn = &ColumnRename{
							OldName: destColumn.Name,
						}
					}

					if len(actions) > 0 || renameColumn != nil {
						b.diff.ExistsInBoth.Different.Columns = append(b.diff.ExistsInBoth.Different.Columns, &ColumnDiff{
							Column:       srcColumn,
							Actions:      actions,
							RenameColumn: renameColumn,
						})
					}
				}
			}

			for _, column := range destTableCols {
				sourceColumn := findColumn(sourceTableCols, column)
				if sourceColumn == nil {
					b.diff.ExistsInDestination.Columns = append(b.diff.ExistsInDestination.Columns, column)
				}
			}
		}
	}
}

func findColumn(columns map[string]*sqlmanager_shared.TableColumn, column *sqlmanager_shared.TableColumn) *sqlmanager_shared.TableColumn {
	// perfect match
	for _, c := range columns {
		if c.Schema != column.Schema || c.Table != column.Table {
			continue
		}
		if c.Fingerprint == column.Fingerprint {
			return c
		}
		if c.Name == column.Name && c.OrdinalPosition == column.OrdinalPosition {
			return c
		}
	}

	// name match
	for _, c := range columns {
		if c.Schema != column.Schema || c.Table != column.Table {
			continue
		}
		if c.Name == column.Name {
			return c
		}
	}

	// ordinal match
	for _, c := range columns {
		if c.Schema != column.Schema || c.Table != column.Table {
			continue
		}
		if c.OrdinalPosition == column.OrdinalPosition {
			return c
		}
	}
	return nil
}
func (b *SchemaDifferencesBuilder) buildTableForeignKeyConstraintDifferences() {
	existsInSource, existsInBoth, existsInDestination := buildDifferencesByFingerprint(
		b.source.ForeignKeyConstraints,
		b.destination.ForeignKeyConstraints,
	)
	b.diff.ExistsInSource.ForeignKeyConstraints = existsInSource
	b.diff.ExistsInBoth.Different.ForeignKeyConstraints = existsInBoth
	b.diff.ExistsInDestination.ForeignKeyConstraints = existsInDestination
}

func (b *SchemaDifferencesBuilder) buildTableNonForeignKeyConstraintDifferences() {
	existsInSource, existsInBoth, existsInDestination := buildDifferencesByFingerprint(
		b.source.NonForeignKeyConstraints,
		b.destination.NonForeignKeyConstraints,
	)
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

func (b *SchemaDifferencesBuilder) buildTableCompositeDifferences() {
	for key, srcComposite := range b.source.Composites {
		if destComposite, ok := b.destination.Composites[key]; ok {
			changedAttributesDatatype := map[string]string{}
			changedAttributesName := map[string]string{}
			newAttributes := map[string]string{}
			removedAttributes := []string{}

			srcAttributes := map[int]*sqlmanager_shared.CompositeAttribute{}
			for _, attr := range srcComposite.Attributes {
				srcAttributes[attr.Id] = attr
			}
			destAttributes := map[int]*sqlmanager_shared.CompositeAttribute{}
			for _, attr := range destComposite.Attributes {
				destAttributes[attr.Id] = attr
			}

			// The id here is unique to the attribute and doesn't get reused when deleted.
			for id, attr := range srcAttributes {
				if destAttr, ok := destAttributes[id]; ok {
					if attr.Datatype != destAttr.Datatype {
						changedAttributesDatatype[destAttr.Name] = attr.Datatype
					}
					if attr.Name != destAttr.Name {
						changedAttributesName[destAttr.Name] = attr.Name
					}
				} else {
					newAttributes[attr.Name] = attr.Datatype
				}
			}

			for id, attr := range destAttributes {
				if _, ok := srcAttributes[id]; !ok {
					removedAttributes = append(removedAttributes, attr.Name)
				}
			}

			b.diff.ExistsInBoth.Different.Composites = append(b.diff.ExistsInBoth.Different.Composites, &CompositeDiff{
				Composite:                srcComposite,
				ChangedAttributeDatatype: changedAttributesDatatype,
				NewAttributes:            newAttributes,
				RemovedAttributes:        removedAttributes,
				ChangedAttributeName:     changedAttributesName,
			})
		} else {
			b.diff.ExistsInSource.Composites = append(b.diff.ExistsInSource.Composites, srcComposite)
		}
	}

	for key, destComposite := range b.destination.Composites {
		if _, ok := b.source.Composites[key]; !ok {
			b.diff.ExistsInDestination.Composites = append(b.diff.ExistsInDestination.Composites, destComposite)
		}
	}
}

func (b *SchemaDifferencesBuilder) buildTableDomainDifferences() {
	for key, srcDomain := range b.source.Domains {
		if destDomain, ok := b.destination.Domains[key]; ok {
			domain := &DomainDiff{
				Domain:             srcDomain,
				IsNullDifferent:    srcDomain.IsNullable != destDomain.IsNullable,
				IsDefaultDifferent: srcDomain.Default != destDomain.Default,
				NewConstraints:     map[string]string{},
				RemovedConstraints: []string{},
			}

			srcConstraints := map[string]*sqlmanager_shared.DomainConstraint{}
			for _, constraint := range srcDomain.Constraints {
				srcConstraints[constraint.Name] = constraint
			}
			destConstraints := map[string]*sqlmanager_shared.DomainConstraint{}
			for _, constraint := range destDomain.Constraints {
				destConstraints[constraint.Name] = constraint
			}

			for _, constraint := range srcConstraints {
				if destConstraint, ok := destConstraints[constraint.Name]; ok {
					if constraint.Definition != destConstraint.Definition {
						domain.RemovedConstraints = append(domain.RemovedConstraints, constraint.Name)
					}
				} else {
					domain.NewConstraints[constraint.Name] = constraint.Definition
				}
			}

			for _, constraint := range destConstraints {
				if _, ok := srcConstraints[constraint.Name]; !ok {
					domain.NewConstraints[constraint.Name] = constraint.Definition
				}
			}

			b.diff.ExistsInBoth.Different.Domains = append(b.diff.ExistsInBoth.Different.Domains, domain)
		} else {
			b.diff.ExistsInSource.Domains = append(b.diff.ExistsInSource.Domains, srcDomain)
		}
	}

	for key, destDomain := range b.destination.Domains {
		if _, ok := b.source.Domains[key]; !ok {
			b.diff.ExistsInDestination.Domains = append(b.diff.ExistsInDestination.Domains, destDomain)
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
