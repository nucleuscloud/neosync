package sqlmanager_shared

import (
	"errors"
)

const (
	DisableForeignKeyChecks = "SET FOREIGN_KEY_CHECKS = 0;"
)

type DatabaseSchemaRow struct {
	TableSchema            string
	TableName              string
	ColumnName             string
	DataType               string
	ColumnDefault          string
	ColumnDefaultType      *string
	IsNullable             bool
	CharacterMaximumLength int
	NumericPrecision       int
	NumericScale           int
	OrdinalPosition        int
	GeneratedType          *string
	IdentityGeneration     *string
	IdentitySeed           *int
	IdentityIncrement      *int
}

func (d *DatabaseSchemaRow) NullableString() string {
	if d.IsNullable {
		return "YES"
	}
	return "NO"
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
	Schema      string
	Table       string
	TriggerName string
	Definition  string
}

type TableInitStatement struct {
	CreateTableStatement string
	AlterTableStatements []*AlterTableStatement
	IndexStatements      []string
}

type AlterTableStatement struct {
	Statement      string
	ConstraintType ConstraintType
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
	Schema     string
	Name       string
	Definition string
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
