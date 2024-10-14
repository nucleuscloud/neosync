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
	IsNullable             string
	CharacterMaximumLength int
	NumericPrecision       int
	NumericScale           int
	OrdinalPosition        int
	GeneratedType          *string
	IdentityGeneration     *string
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
	}
	return -1, errors.ErrUnsupported
}

const (
	PostgresDriver = "pgx"
	MysqlDriver    = "mysql"
	MssqlDriver    = "sqlserver"
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
}

type ColumnInfo struct {
	OrdinalPosition        int     // Specifies the sequence or order in which each column is defined within the table. Starts at 1 for the first column.
	ColumnDefault          string  // Specifies the default value for a column, if any is set.
	IsNullable             bool    // Specifies if the column is nullable or not.
	DataType               string  // Specifies the data type of the column, i.e., bool, varchar, int, etc.
	CharacterMaximumLength *int    // Specifies the maximum allowable length of the column for character-based data types. For datatypes such as integers, boolean, dates etc. this is NULL.
	NumericPrecision       *int    // Specifies the precision for numeric data types. It represents the TOTAL count of significant digits in the whole number, that is, the number of digits to BOTH sides of the decimal point. Null for non-numeric data types.
	NumericScale           *int    // Specifies the scale of the column for numeric data types, specifically non-integers. It represents the number of digits to the RIGHT of the decimal point. Null for non-numeric data types and integers.
	IdentityGeneration     *string // Specifies the identity generation strategy for the column, if applicable.
}

type DataType struct {
	Schema     string
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
	for _, fn := range s.Functions {
		output = append(output, fn.Definition)
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
	return output
}

type InitSchemaStatements struct {
	Label      string
	Statements []string
}
