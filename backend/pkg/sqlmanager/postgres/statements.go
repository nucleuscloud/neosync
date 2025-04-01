package sqlmanager_postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	schemamanager_shared "github.com/nucleuscloud/neosync/internal/schema-manager/shared"
)

// Finds any schemas referenced in datatypes that don't exist in tables and returns the statements to create them
func getSchemaCreationStatementsFromDataTypes(
	tables []*sqlmanager_shared.SchemaTable,
	datatypes *sqlmanager_shared.SchemaTableDataTypeResponse,
) []string {
	schemaStmts := []string{}
	schemaSet := map[string]struct{}{}
	for _, table := range tables {
		schemaSet[table.Schema] = struct{}{}
	}

	// Check each datatype schema against the table schemas
	for _, composite := range datatypes.Composites {
		if _, exists := schemaSet[composite.Schema]; !exists {
			schemaStmts = append(
				schemaStmts,
				fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q;", composite.Schema),
			)
			schemaSet[composite.Schema] = struct{}{}
		}
	}

	for _, enum := range datatypes.Enums {
		if _, exists := schemaSet[enum.Schema]; !exists {
			schemaStmts = append(
				schemaStmts,
				fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q;", enum.Schema),
			)
			schemaSet[enum.Schema] = struct{}{}
		}
	}

	for _, domain := range datatypes.Domains {
		if _, exists := schemaSet[domain.Schema]; !exists {
			schemaStmts = append(
				schemaStmts,
				fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q;", domain.Schema),
			)
			schemaSet[domain.Schema] = struct{}{}
		}
	}
	return schemaStmts
}

func wrapPgIdempotentIndex(
	schema,
	constraintname,
	alterStatement string,
) string {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relkind in ('i', 'I')
		AND c.relname = '%s'
		AND n.nspname = '%s'
	) THEN
		%s
	END IF;
END $$;
`, constraintname, schema, addSuffixIfNotExist(alterStatement, ";"))
	return strings.TrimSpace(stmt)
}

func wrapPgIdempotentConstraint(
	schema, table,
	constraintName,
	alterStatement string,
) string {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = '%s'
		AND connamespace = (SELECT oid FROM pg_namespace WHERE nspname = '%s')
		AND conrelid = (
			SELECT oid
			FROM pg_class
			WHERE relname = '%s'
			AND relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = '%s')
		)
	) THEN
		%s
	END IF;
END $$;
	`, constraintName, schema, table, schema, addSuffixIfNotExist(alterStatement, ";"))
	return strings.TrimSpace(stmt)
}

func wrapPgIdempotentSequence(
	schema,
	sequenceName,
	createStatement string,
) string {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relkind = 'S'
        AND c.relname = '%s'
        AND n.nspname = '%s'
    ) THEN
        %s
    END IF;
END $$;
`, sequenceName, schema, addSuffixIfNotExist(createStatement, ";"))
	return strings.TrimSpace(stmt)
}

func wrapPgIdempotentTrigger(
	schema,
	tableName,
	triggerName,
	createStatement string,
) string {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger t
        JOIN pg_class c ON c.oid = t.tgrelid
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE t.tgname = '%s'
        AND c.relname = '%s'
        AND n.nspname = '%s'
    ) THEN
        %s
    END IF;
END $$;
`, triggerName, tableName, schema, addSuffixIfNotExist(createStatement, ";"))
	return strings.TrimSpace(stmt)
}

func wrapPgIdempotentFunction(
	schema,
	functionName,
	functionSignature,
	createStatement string,
) string {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_proc p
        JOIN pg_namespace n ON n.oid = p.pronamespace
        WHERE p.proname = '%s'
        AND n.nspname = '%s'
        AND pg_catalog.pg_get_function_identity_arguments(p.oid) = '%s'
    ) THEN
        %s
    END IF;
END $$;
`, functionName, schema, functionSignature, addSuffixIfNotExist(createStatement, ";"))
	return strings.TrimSpace(stmt)
}

func wrapPgIdempotentDataType(
	schema,
	dataTypeName,
	createStatement string,
) string {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_type t
        JOIN pg_namespace n ON n.oid = t.typnamespace
        WHERE t.typname = '%s'
        AND n.nspname = '%s'
    ) THEN
        %s
    END IF;
END $$;
`, dataTypeName, schema, addSuffixIfNotExist(createStatement, ";"))
	return strings.TrimSpace(stmt)
}

func wrapPgIdempotentExtension(
	schema sql.NullString,
	extensionName,
	version string,
) string {
	if schema.Valid && strings.EqualFold(schema.String, "public") {
		return fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS %q VERSION %q;`, extensionName, version)
	}
	return fmt.Sprintf(
		`CREATE EXTENSION IF NOT EXISTS %q VERSION %q SCHEMA %q;`,
		extensionName,
		version,
		schema.String,
	)
}

//nolint:unparam
func addSuffixIfNotExist(input, suffix string) string {
	if !strings.HasSuffix(input, suffix) {
		return fmt.Sprintf("%s%s", input, suffix)
	}
	return input
}

func buildAlterStatementByForeignKeyConstraint(
	constraint *pg_queries.GetForeignKeyConstraintsBySchemasRow,
) (string, error) {
	if constraint == nil {
		return "", errors.New("unable to build alter statement as constraint is nil")
	}
	return fmt.Sprintf(
		"ALTER TABLE %q.%q ADD CONSTRAINT %q FOREIGN KEY (%s) REFERENCES %q.%q (%s);",
		constraint.ReferencingSchema,
		constraint.ReferencingTable,
		constraint.ConstraintName,
		strings.Join(EscapePgColumns(constraint.ReferencingColumns), ", "),
		constraint.ReferencedSchema,
		constraint.ReferencedTable,
		strings.Join(EscapePgColumns(constraint.ReferencedColumns), ", "),
	), nil
}

func buildAlterStatementByConstraint(
	constraint *pg_queries.GetNonForeignKeyTableConstraintsBySchemaRow,
) (string, error) {
	if constraint == nil {
		return "", errors.New("unable to build alter statement as constraint is nil")
	}
	return fmt.Sprintf(
		"ALTER TABLE %q.%q ADD CONSTRAINT %q %s;",
		constraint.SchemaName,
		constraint.TableName,
		constraint.ConstraintName,
		constraint.ConstraintDefinition,
	), nil
}

func BuildAddColumnStatement(column *sqlmanager_shared.TableColumn) string {
	col := buildTableCol(&buildTableColRequest{
		ColumnName:         column.Name,
		ColumnDefault:      column.ColumnDefault,
		DataType:           column.DataType,
		IsNullable:         column.IsNullable,
		GeneratedType:      *column.GeneratedType,
		SequenceDefinition: column.SequenceDefinition,
		// IsSerial:           column.IsSerial,
	})
	return fmt.Sprintf("ALTER TABLE %q.%q ADD COLUMN %s;", column.Schema, column.Table, col)
}

func BuildAlterColumnStatement(column *schemamanager_shared.ColumnDiff) []string {
	statements := []string{}
	pieces := []string{}

	base := fmt.Sprintf("ALTER COLUMN %q", column.Column.Name)
	for _, action := range column.Actions {
		switch action {
		case schemamanager_shared.SetDatatype:
			pieces = append(pieces, fmt.Sprintf("%s TYPE %s USING %q::%s", base, column.Column.DataType, column.Column.Name, column.Column.DataType))
		case schemamanager_shared.DropNotNull:
			pieces = append(pieces, fmt.Sprintf("%s DROP NOT NULL", base))
		case schemamanager_shared.SetNotNull:
			pieces = append(pieces, fmt.Sprintf("%s SET NOT NULL", base))
		case schemamanager_shared.DropDefault:
			pieces = append(pieces, fmt.Sprintf("%s DROP DEFAULT", base))
		case schemamanager_shared.SetDefault:
			if column.Column.GeneratedType != nil && *column.Column.GeneratedType == "s" {
				// pieces = append(pieces, fmt.Sprintf("%s SET GENERATED ALWAYS AS (%s) STORED", base, column.Column.ColumnDefault))
				// need to drop then recreate
				statements = append(statements, BuildDropColumnStatement(column.Column.Schema, column.Column.Table, column.Column.Name))
				statements = append(statements, BuildAddColumnStatement(column.Column))
			} else {
				pieces = append(pieces, fmt.Sprintf("%s SET DEFAULT %s", base, column.Column.ColumnDefault))
			}
		case schemamanager_shared.DropIdentity:
			pieces = append(pieces, fmt.Sprintf("%s DROP IDENTITY IF EXISTS", base))
		}
	}

	if column.RenameColumn != nil {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %q.%q RENAME COLUMN %q TO %q;", column.Column.Schema, column.Column.Table, column.RenameColumn.OldName, column.Column.Name))
	}

	if len(pieces) > 0 {
		alterStatement := fmt.Sprintf("ALTER TABLE %q.%q %s;", column.Column.Schema, column.Column.Table, strings.Join(pieces, ", "))
		statements = append(statements, alterStatement)
	}

	return statements
}

func BuildDropColumnStatement(schema, table, column string) string {
	// cascade is used to drop the column and all the constraints, views, and indexes that depend on it
	return fmt.Sprintf("ALTER TABLE %q.%q DROP COLUMN IF EXISTS %q CASCADE;", schema, table, column)
}

func BuildDropConstraintStatement(schema, table, constraintName string) string {
	// cascade is used to drop the constraint and any dependent objects (other constraints, indexes, triggers, etc)
	return fmt.Sprintf(
		"ALTER TABLE %q.%q DROP CONSTRAINT IF EXISTS %q CASCADE;",
		schema,
		table,
		constraintName,
	)
}

func BuildDropTriggerStatement(schema, table, triggerName string) string {
	return fmt.Sprintf("DROP TRIGGER IF EXISTS %q ON %q.%q;", triggerName, schema, table)
}

func BuildDropFunctionStatement(schema, functionName string) string {
	return fmt.Sprintf("DROP FUNCTION IF EXISTS %q.%q;", schema, functionName)
}

func BuildUpdateFunctionStatement(schema, functionName, createStatement string) string {
	if strings.Contains(strings.ToUpper(createStatement), "CREATE FUNCTION") &&
		!strings.Contains(strings.ToUpper(createStatement), "CREATE OR REPLACE FUNCTION") {
		createStatement = strings.Replace(strings.ToUpper(createStatement), "CREATE FUNCTION", "CREATE OR REPLACE FUNCTION", 1)
	}
	return createStatement
}

func BuildDropDatatypesStatement(schema, enumName string) string {
	return fmt.Sprintf("DROP TYPE IF EXISTS %q.%q;", schema, enumName)
}

func BuildUpdateEnumStatements(schema, enumName string, newValues []string, changedValues map[string]string) []string {
	statements := []string{}
	for _, value := range newValues {
		statements = append(statements, fmt.Sprintf("ALTER TYPE %q.%q ADD VALUE IF NOT EXISTS '%s';", schema, enumName, value))
	}
	for value, newVal := range changedValues {
		statements = append(statements, fmt.Sprintf("ALTER TYPE %q.%q RENAME VALUE '%s' TO '%s';", schema, enumName, value, newVal))
	}
	return statements
}

func BuildUpdateCompositeStatements(
	schema, compositeName string,
	changedAttributesDatatype, changedAttributesName, newAttributes map[string]string,
	removedAttributes []string,
) []string {
	statements := []string{}
	for attribute, newDatatype := range changedAttributesDatatype {
		statements = append(
			statements,
			fmt.Sprintf("ALTER TYPE %q.%q ALTER ATTRIBUTE %q SET DATA TYPE '%s';", schema, compositeName, attribute, newDatatype),
		)
	}
	for oldName, newName := range changedAttributesName {
		statements = append(
			statements,
			fmt.Sprintf("ALTER TYPE %q.%q RENAME ATTRIBUTE %q TO  %q;", schema, compositeName, oldName, newName),
		)
	}
	for attribute, datatype := range newAttributes {
		statements = append(statements, fmt.Sprintf("ALTER TYPE %q.%q ADD ATTRIBUTE %q %s;", schema, compositeName, attribute, datatype))
	}
	for _, attribute := range removedAttributes {
		statements = append(statements, fmt.Sprintf("ALTER TYPE %q.%q DROP ATTRIBUTE IF EXISTS %q;", schema, compositeName, attribute))
	}
	return statements
}

func BuildDropDomainStatement(schema, domainName string) string {
	return fmt.Sprintf("DROP DOMAIN IF EXISTS %q.%q;", schema, domainName)
}

func BuildDomainConstraintStatements(schema, domainName string, newConstraints map[string]string, removedConstraints []string) []string {
	statements := []string{}
	for constraint, definition := range newConstraints {
		statements = append(statements, fmt.Sprintf("ALTER DOMAIN %q.%q ADD CONSTRAINT %q %s;", schema, domainName, constraint, definition))
	}
	for _, constraint := range removedConstraints {
		statements = append(statements, fmt.Sprintf("ALTER DOMAIN %q.%q DROP CONSTRAINT IF EXISTS %q;", schema, domainName, constraint))
	}
	return statements
}

func BuildUpdateDomainDefaultStatement(schema, domainName, defaultString string) string {
	return fmt.Sprintf("ALTER DOMAIN %q.%q SET DEFAULT %s;", schema, domainName, defaultString)
}

func BuildDropDomainDefaultStatement(schema, domainName string) string {
	return fmt.Sprintf("ALTER DOMAIN %q.%q DROP DEFAULT;", schema, domainName)
}

func BuildUpdateDomainNotNullStatement(schema, domainName string, isNullable bool) string {
	nullString := "NOT NULL"
	if isNullable {
		nullString = "NULL"
	}
	return fmt.Sprintf("ALTER DOMAIN %q.%q SET %s;", schema, domainName, nullString)
}

type buildTableColRequest struct {
	ColumnName    string
	ColumnDefault string
	DataType      string
	IsNullable    bool
	GeneratedType string
	// IsSerial           bool
	SequenceDefinition *string
	Sequence           *SequenceConfiguration
}

type SequenceConfiguration struct {
	IncrementBy int64
	MinValue    int64
	MaxValue    int64
	StartValue  int64
	CacheValue  int64
	CycleOption bool
}

func (s *SequenceConfiguration) ToGeneratedDefaultIdentity() string {
	return fmt.Sprintf("GENERATED BY DEFAULT AS IDENTITY ( %s )", s.identitySequenceConfiguration())
}
func (s *SequenceConfiguration) ToGeneratedAlwaysIdentity() string {
	return fmt.Sprintf("GENERATED ALWAYS AS IDENTITY ( %s )", s.identitySequenceConfiguration())
}

func (s *SequenceConfiguration) identitySequenceConfiguration() string {
	return fmt.Sprintf("INCREMENT BY %d MINVALUE %d MAXVALUE %d START %d CACHE %d %s",
		s.IncrementBy, s.MinValue, s.MaxValue, s.StartValue, s.CacheValue, s.toCycelText(),
	)
}

func (s *SequenceConfiguration) toCycelText() string {
	if s.CycleOption {
		return "CYCLE"
	}
	return "NO CYCLE"
}

func BuildSequencOwnerStatement(seq *pg_queries.GetSequencesOwnedByTablesRow) string {
	return fmt.Sprintf("ALTER SEQUENCE %q.%q OWNED BY %q.%q.%q;", seq.SequenceSchema, seq.SequenceName, seq.TableSchema, seq.TableName, seq.ColumnName)
}

func buildTableCol(record *buildTableColRequest) string {
	pieces := []string{
		EscapePgColumn(record.ColumnName),
		record.DataType,
		buildNullableText(record.IsNullable),
	}

	// if record.IsSerial {
	// 	switch record.DataType {
	// 	case "smallint":
	// 		pieces[1] = "SMALLSERIAL"
	// 	case "bigint":
	// 		pieces[1] = "BIGSERIAL"
	// 	default:
	// 		pieces[1] = "SERIAL"
	// 	}
	// } else
	if record.SequenceDefinition != nil && *record.SequenceDefinition != "" {
		pieces = append(pieces, *record.SequenceDefinition)
	} else if record.ColumnDefault != "" {
		if record.GeneratedType == "s" {
			pieces = append(pieces, fmt.Sprintf("GENERATED ALWAYS AS (%s) STORED", record.ColumnDefault))
		} else if record.ColumnDefault != "NULL" {
			pieces = append(pieces, "DEFAULT", record.ColumnDefault)
		}
	}
	return strings.Join(pieces, " ")
}

func buildSequenceDefinition(identityType string, seqConfig *SequenceConfiguration) string {
	var seqStr string
	switch identityType {
	case "d":
		seqStr = seqConfig.ToGeneratedDefaultIdentity()
	case "a":
		seqStr = seqConfig.ToGeneratedAlwaysIdentity()
	}
	return seqStr
}

func BuildUpdateCommentStatement(schema, table, column string, comment *string) string {
	if comment == nil || *comment == "" {
		return fmt.Sprintf("COMMENT ON COLUMN %q.%q.%q IS NULL;", schema, table, column)
	}
	return fmt.Sprintf(
		"COMMENT ON COLUMN %q.%q.%q IS '%s';",
		schema,
		table,
		column,
		strings.ReplaceAll(*comment, "'", "''"),
	)
}

func buildNullableText(isNullable bool) string {
	if isNullable {
		return "NULL"
	}
	return "NOT NULL"
}

func getGoquDialect() goqu.DialectWrapper {
	return goqu.Dialect(sqlmanager_shared.GoquPostgresDriver)
}

func BuildPgTruncateStatement(
	tables []*sqlmanager_shared.SchemaTable,
) (string, error) {
	builder := getGoquDialect()
	gTables := []any{}
	for _, t := range tables {
		gTables = append(gTables, goqu.S(t.Schema).Table(t.Table))
	}
	stmt, _, err := builder.From(gTables...).Truncate().Identity("RESTART").ToSQL()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s;", stmt), nil
}

func BuildPgTruncateCascadeStatement(
	schema string,
	table string,
) (string, error) {
	builder := getGoquDialect()
	sqltable := goqu.S(schema).Table(table)
	stmt, _, err := builder.From(sqltable).Truncate().Cascade().Identity("RESTART").ToSQL()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s;", stmt), nil
}

func EscapePgColumns(cols []string) []string {
	outcols := make([]string, len(cols))
	for idx := range cols {
		outcols[idx] = EscapePgColumn(cols[idx])
	}
	return outcols
}

func EscapePgColumn(col string) string {
	return fmt.Sprintf("%q", col)
}

func BuildPgIdentityColumnResetCurrentSql(
	schema, table, column string,
) string {
	return fmt.Sprintf(
		"SELECT setval(pg_get_serial_sequence('%q.%q', '%s'), COALESCE((SELECT MAX(%q) FROM %q.%q), 1));",
		schema,
		table,
		column,
		column,
		schema,
		table,
	)
}

func BuildPgInsertIdentityAlwaysSql(
	insertQuery string,
) string {
	sqlSplit := strings.Split(insertQuery, ") VALUES (")
	return sqlSplit[0] + ") OVERRIDING SYSTEM VALUE VALUES(" + sqlSplit[1]
}

func BuildPgResetSequenceSql(schema, sequenceName string) string {
	return fmt.Sprintf("ALTER SEQUENCE %q.%q RESTART;", schema, sequenceName)
}

func GetPostgresColumnOverrideAndResetProperties(
	columnInfo *sqlmanager_shared.DatabaseSchemaRow,
) (needsOverride, needsReset bool) {
	needsOverride = false
	needsReset = false

	// check if the column is an idenitity type
	if columnInfo.IdentityGeneration != nil && *columnInfo.IdentityGeneration != "" {
		switch *columnInfo.IdentityGeneration {
		case "a": // ALWAYS
			needsOverride = true
			needsReset = true
		case "d": // DEFAULT
			needsReset = true
		}
		return
	}

	// check if column default is sequence
	if columnInfo.ColumnDefault != "" &&
		gotypeutil.CaseInsensitiveContains(columnInfo.ColumnDefault, "nextVal") {
		needsReset = true
		return
	}

	return
}
