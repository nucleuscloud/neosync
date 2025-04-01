package sqlmanager_shared

import (
	"errors"
)

const (
	DisableForeignKeyChecks           = "SET FOREIGN_KEY_CHECKS = 0;"
	ExtensionsLabel                   = "extensions"
	SchemasLabel                      = "schemas"
	CreateTablesLabel                 = "create table"
	AddColumnsLabel                   = "add columns"
	DropColumnsLabel                  = "drop columns"
	DropTriggersLabel                 = "drop triggers"
	DropFunctionsLabel                = "drop functions"
	DropNonForeignKeyConstraintsLabel = "drop non-foreign key constraints"
	DropForeignKeyConstraintsLabel    = "drop foreign key constraints"
	UpdateColumnsLabel                = "update columns"
	UpdateFunctionsLabel              = "update functions"
	DropDatatypesLabel                = "drop datatypes"
	UpdateDatatypesLabel              = "update datatypes"
)

type DatabaseSchemaRow struct {
	TableSchema            string
	TableName              string
	ColumnName             string
	DataType               string
	MysqlColumnType        string // will only be populated for mysql. Same as the DataType but includes length, etc. varchar(255), enum('a', 'b'), etc.
	ColumnDefault          string
	ColumnDefaultType      *string
	IsNullable             bool
	CharacterMaximumLength int
	NumericPrecision       int
	NumericScale           int
	OrdinalPosition        int
	GeneratedType          *string
	GeneratedExpression    *string
	IdentityGeneration     *string
	IdentitySeed           *int
	IdentityIncrement      *int
	// UpdateAllowed indicates whether updates are permitted for this column.
	// generated columns are an example of columns that do not allow updates.
	UpdateAllowed bool
	Comment       *string
}

func (d *DatabaseSchemaRow) NullableString() string {
	if d.IsNullable {
		return "YES"
	}
	return "NO"
}

type DatabaseSchemaNameRow struct {
	SchemaName string
}

type DatabaseTableRow struct {
	SchemaName string
	TableName  string
}

// Fingerprinter is implemented by any type that exposes a Fingerprint via GetFingerprint().
type Fingerprinter interface {
	GetFingerprint() string
}

type TableColumn struct {
	Fingerprint         string
	Schema              string
	Table               string
	Name                string
	DataType            string
	IsNullable          bool
	ColumnDefault       string
	ColumnDefaultType   *string
	GeneratedType       *string
	GeneratedExpression *string
	IdentityGeneration  *string
	SequenceDefinition  *string
	OrdinalPosition     int
	Comment             *string
}

func (c *TableColumn) GetFingerprint() string {
	return c.Fingerprint
}

type ForeignKeyConstraintsRow struct {
	ConstraintName    string
	SchemaName        string
	TableName         string
	ColumnName        string
	IsNullable        bool
	ForeignSchemaName string
	ForeignTableName  string
	ForeignColumnName string
}

type PrimaryKey struct {
	Schema  string
	Table   string
	Columns []string
}

type SchemaTable struct {
	Schema string
	Table  string
}

func (s SchemaTable) String() string {
	return BuildTable(s.Schema, s.Table)
}

type TableTrigger struct {
	Fingerprint   string
	Schema        string
	Table         string
	TriggerSchema *string
	TriggerName   string
	Definition    string
}

func (t *TableTrigger) GetFingerprint() string {
	return t.Fingerprint
}

type DomainConstraint struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}

type DomainDataType struct {
	Fingerprint string
	Schema      string
	Name        string
	IsNullable  bool
	Default     string
	Constraints []*DomainConstraint
}

func (d *DomainDataType) GetFingerprint() string {
	return d.Fingerprint
}

type EnumDataType struct {
	Fingerprint string
	Schema      string
	Name        string
	Values      []string
}

func (e *EnumDataType) GetFingerprint() string {
	return e.Fingerprint
}

type CompositeDataType struct {
	Fingerprint string
	Schema      string
	Name        string
	Attributes  []*CompositeAttribute
}

type CompositeAttribute struct {
	Name     string `json:"name"`
	Datatype string `json:"datatype"`
	Id       int    `json:"id"` // helps track when attributes are added or removed
}

func (c *CompositeDataType) GetFingerprint() string {
	return c.Fingerprint
}

type AllTableDataTypes struct {
	Functions  []*DataType
	Domains    []*DomainDataType
	Enums      []*EnumDataType
	Composites []*CompositeDataType
}

type TableInitStatement struct {
	CreateTableStatement string
	AlterTableStatements []*AlterTableStatement
	IndexStatements      []string
	PartitionStatements  []string
}

type AlterTableStatement struct {
	Statement      string
	ConstraintType ConstraintType
}

type ForeignKeyConstraint struct {
	Fingerprint        string
	ConstraintName     string
	ConstraintType     string
	ReferencingSchema  string
	ReferencingTable   string
	ReferencingColumns []string
	ReferencedSchema   string
	ReferencedTable    string
	ReferencedColumns  []string
	NotNullable        []bool
	UpdateRule         *string
	DeleteRule         *string
}

func (f *ForeignKeyConstraint) GetFingerprint() string {
	return f.Fingerprint
}

type NonForeignKeyConstraint struct {
	Fingerprint    string
	ConstraintName string
	ConstraintType string
	SchemaName     string
	TableName      string
	Columns        []string
	Definition     string
}

func (n *NonForeignKeyConstraint) GetFingerprint() string {
	return n.Fingerprint
}

type AllTableConstraints struct {
	ForeignKeyConstraints    []*ForeignKeyConstraint
	NonForeignKeyConstraints []*NonForeignKeyConstraint
}

type ConstraintType int

const (
	PrimaryConstraintType ConstraintType = iota
	ForeignConstraintType
	UniqueConstraintType
	CheckConstraintType
	ExclusionConstraintType
)

func ToConstraintType(constraintType string) (ConstraintType, error) {
	switch constraintType {
	case "p":
		return PrimaryConstraintType, nil
	case "u":
		return UniqueConstraintType, nil
	case "f":
		return ForeignConstraintType, nil
	case "c":
		return CheckConstraintType, nil
	case "x":
		return ExclusionConstraintType, nil
	}
	return -1, errors.ErrUnsupported
}

const (
	PostgresDriver     = "pgx"
	GoquPostgresDriver = "postgres"
	MysqlDriver        = "mysql"
	MssqlDriver        = "sqlserver"
)

type BatchExecOpts struct {
	Prefix *string // this string will be added to the start of each statement
}

type ForeignKey struct {
	Table   string
	Columns []string
}
type ForeignConstraint struct {
	Columns     []string
	NotNullable []bool
	ForeignKey  *ForeignKey
}

type TableConstraints struct {
	ForeignKeyConstraints map[string][]*ForeignConstraint
	PrimaryKeyConstraints map[string][]string
	UniqueConstraints     map[string][][]string
	UniqueIndexes         map[string][][]string
}

type DataType struct {
	Fingerprint string
	Schema      string
	Name        string
	Definition  string
}

func (d *DataType) GetFingerprint() string {
	return d.Fingerprint
}

type ExtensionDataType struct {
	Name       string
	Definition string
}

// These are all items that live at the schema level, but are used by tables
type SchemaTableDataTypeResponse struct {
	// Custom Sequences not tied to the SERIAL data type
	Sequences []*DataType

	// SQL Functions
	Functions []*DataType

	// actual Data Types
	Composites []*DataType
	Enums      []*DataType
	Domains    []*DataType
}

func (s *SchemaTableDataTypeResponse) GetStatements() []string {
	output := []string{}

	if s == nil {
		return output
	}

	for _, seq := range s.Sequences {
		output = append(output, seq.Definition)
	}
	for _, comp := range s.Composites {
		output = append(output, comp.Definition)
	}
	for _, enumeration := range s.Enums {
		output = append(output, enumeration.Definition)
	}
	for _, domain := range s.Domains {
		output = append(output, domain.Definition)
	}
	for _, fn := range s.Functions {
		output = append(output, fn.Definition)
	}
	return output
}

type InitSchemaStatements struct {
	Label      string
	Statements []string
}

type SelectQuery struct {
	// Query is the query used to get all data
	Query string
	// PageQuery is the query used to get a page of data based on a unique identifier like a primary key in the WHERE clause
	PageQuery string
	PageLimit int

	// If true, this query could return rows that violate foreign key constraints
	IsNotForeignKeySafeSubset bool
}
