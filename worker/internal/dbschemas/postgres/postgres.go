package dbschemas_postgres

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

type DatabaseSchema struct {
	TableSchema     string  `db:"table_schema"`
	TableName       string  `db:"table_name"`
	ColumnName      string  `db:"column_name"`
	OrdinalPosition int     `db:"ordinal_position"`
	ColumnDefault   *string `db:"column_default,omitempty"`
	IsNullable      string  `db:"is_nullable"`
	DataType        string  `db:"data_type"`
}

func (d *DatabaseSchema) GetTableKey() string {
	return fmt.Sprintf("%s.%s", d.TableSchema, d.TableName)
}

const (
	getDatabaseSchemaSql = `-- name: GetDatabaseSchema
SELECT
	c.table_schema,
	c.table_name,
	c.column_name,
	c.ordinal_position,
	c.column_default,
	c.is_nullable,
	c.data_type
FROM
	information_schema.columns AS c
	JOIN information_schema.tables AS t ON c.table_schema = t.table_schema
		AND c.table_name = t.table_name
WHERE
	c.table_schema NOT IN('pg_catalog', 'information_schema')
	AND t.table_type = 'BASE TABLE';
	`
	getDatabaseTableSchemaSql = `-- name: GetDatabaseTableSchema
SELECT
	c.table_schema,
	c.table_name,
	c.column_name,
	c.ordinal_position,
	c.column_default,
	c.is_nullable,
	c.data_type
FROM
	information_schema.columns AS c
	JOIN information_schema.tables AS t ON c.table_schema = t.table_schema
		AND c.table_name = t.table_name
WHERE
	c.table_schema = $1 AND t.table_name = $2
	AND t.table_type = 'BASE TABLE';
	`
)

func GetDatabaseSchemas(
	ctx context.Context,
	conn DBTX,
) ([]*DatabaseSchema, error) {
	rows, err := conn.Query(ctx, getDatabaseSchemaSql)
	if err != nil && !isNoRows(err) {
		return nil, err
	} else if err != nil && isNoRows(err) {
		return []*DatabaseSchema{}, nil
	}

	output := []*DatabaseSchema{}
	for rows.Next() {
		var o DatabaseSchema
		err := rows.Scan(
			&o.TableSchema,
			&o.TableName,
			&o.ColumnName,
			&o.OrdinalPosition,
			&o.ColumnDefault,
			&o.IsNullable,
			&o.DataType,
		)
		if err != nil {
			return nil, err
		}
		output = append(output, &o)
	}
	return output, nil
}

func getDatabaseTableSchema(
	ctx context.Context,
	conn DBTX,
	schema string,
	table string,
) ([]*DatabaseSchema, error) {
	rows, err := conn.Query(ctx, getDatabaseTableSchemaSql, schema, table)
	if err != nil && !isNoRows(err) {
		return nil, err
	} else if err != nil && isNoRows(err) {
		return []*DatabaseSchema{}, nil
	}

	output := []*DatabaseSchema{}
	for rows.Next() {
		var o DatabaseSchema
		err := rows.Scan(
			&o.TableSchema,
			&o.TableName,
			&o.ColumnName,
			&o.OrdinalPosition,
			&o.ColumnDefault,
			&o.IsNullable,
			&o.DataType,
		)
		if err != nil {
			return nil, err
		}
		output = append(output, &o)
	}
	return output, nil
}

type DatabaseTableConstraint struct {
	Schema               string `db:"db_schema"`
	Table                string `db:"table_name"`
	ConstraintName       string `db:"constraint_name"`
	ConstraintDefinition string `db:"constraint_definition"`
}

const (
	getTableConstraintsSql = `-- name: GetTableConstraints
SELECT
    nsp.nspname AS db_schema,
    rel.relname AS table_name,
    con.conname AS constraint_name,
    pg_get_constraintdef(con.oid) AS constraint_definition
FROM
    pg_catalog.pg_constraint con
INNER JOIN pg_catalog.pg_class rel
                       ON
    rel.oid = con.conrelid
INNER JOIN pg_catalog.pg_namespace nsp
                       ON
    nsp.oid = connamespace
WHERE
    nsp.nspname = $1 AND rel.relname = $2;
`
)

func GetTableConstraints(
	ctx context.Context,
	conn DBTX,
	schema string,
	table string,
) ([]*DatabaseTableConstraint, error) {
	rows, err := conn.Query(ctx, getTableConstraintsSql, schema, table)
	if err != nil && !isNoRows(err) {
		return nil, err
	} else if err != nil && isNoRows(err) {
		return []*DatabaseTableConstraint{}, nil
	}

	output := []*DatabaseTableConstraint{}
	for rows.Next() {
		var o DatabaseTableConstraint
		err := rows.Scan(
			&o.Schema,
			&o.Table,
			&o.ConstraintName,
			&o.ConstraintDefinition,
		)
		if err != nil {
			return nil, err
		}
		output = append(output, &o)
	}
	return output, nil
}

func isNoRows(err error) bool {
	return err != nil && err == pgx.ErrNoRows
}

type GetTableCreateStatementRequest struct {
	Schema string
	Table  string
}

type DBTX interface {
	// Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row

	// Begin(ctx context.Context) (pgx.Tx, error)
	// BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)

	// Ping(ctx context.Context) error

	// CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

func GetTableCreateStatement(
	ctx context.Context,
	conn DBTX,
	req *GetTableCreateStatementRequest,
) (string, error) {
	errgrp, errctx := errgroup.WithContext(ctx)

	var tableSchemas []*DatabaseSchema
	errgrp.Go(func() error {
		result, err := getDatabaseTableSchema(errctx, conn, req.Schema, req.Table)
		if err != nil {
			return fmt.Errorf("unable to generate database table schema: %w", err)
		}
		tableSchemas = result
		return nil
	})
	var tableConstraints []*DatabaseTableConstraint
	errgrp.Go(func() error {
		result, err := GetTableConstraints(errctx, conn, req.Schema, req.Table)
		if err != nil {
			return fmt.Errorf("unable to generate table constraints: %w", err)
		}
		tableConstraints = result
		return nil
	})
	if err := errgrp.Wait(); err != nil {
		return "", err
	}

	return generateCreateTableStatement(
		req.Schema,
		req.Table,
		tableSchemas,
		tableConstraints,
	), nil
}

// This assumes that the schemas and constraints as for a single table, not an entire db schema
func generateCreateTableStatement(
	schema string,
	table string,
	tableSchemas []*DatabaseSchema,
	tableConstraints []*DatabaseTableConstraint,
) string {
	// ensures the columns are built in the correct order
	sort.Slice(tableSchemas, func(i, j int) bool {
		return tableSchemas[i].OrdinalPosition < tableSchemas[j].OrdinalPosition
	})
	columns := make([]string, len(tableSchemas))
	for idx := range tableSchemas {
		record := tableSchemas[idx]
		columns[idx] = buildTableCol(record)
	}
	constraints := make([]string, len(tableConstraints))
	for idx := range tableConstraints {
		constraint := tableConstraints[idx]
		constraints[idx] = fmt.Sprintf("CONSTRAINT %s %s", constraint.ConstraintName, constraint.ConstraintDefinition)
	}
	tableDefs := append(columns, constraints...) //nolint
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (%s);`, schema, table, strings.Join(tableDefs, ", "))
}

func buildTableCol(record *DatabaseSchema) string {
	pieces := []string{record.ColumnName, record.DataType, buildNullableText(record)}
	if record.ColumnDefault != nil && *record.ColumnDefault != "" {
		pieces = append(pieces, "DEFAULT", *record.ColumnDefault)
	}
	return strings.Join(pieces, " ")
}
func buildNullableText(record *DatabaseSchema) string {
	if record.IsNullable == "NO" {
		return "NOT NULL"
	}
	return "NULL"
}

const (
	fkConstraintSql = `--getForeignKeyConstraints
SELECT
    rc.constraint_name
    ,
    kcu.table_schema AS schema_name
    ,
    kcu.table_name
    ,
    kcu.column_name
    ,
    kcu2.table_schema AS foreign_schema_name
    ,
    kcu2.table_name AS foreign_table_name
    ,
    kcu2.column_name AS foreign_column_name
FROM
    information_schema.referential_constraints rc
JOIN information_schema.key_column_usage kcu
    ON
    kcu.constraint_name = rc.constraint_name
JOIN information_schema.key_column_usage kcu2
    ON
    kcu2.ordinal_position = kcu.position_in_unique_constraint
    AND kcu2.constraint_name = rc.unique_constraint_name
WHERE
    kcu.table_schema = $1
ORDER BY
    rc.constraint_name,
    kcu.ordinal_position;
	`
)

type ForeignKeyConstraint struct {
	ConstraintName    string `db:"constraint_name"`
	SchemaName        string `db:"schema_name"`
	TableName         string `db:"table_name"`
	ColumnName        string `db:"column_name"`
	ForeignSchemaName string `db:"foreign_schema_name"`
	ForeignTableName  string `db:"foreign_table_name"`
	ForeignColumnName string `db:"foreign_column_name"`
}

func GetForeignKeyConstraints(
	ctx context.Context,
	conn DBTX,
	tableSchema string,
) ([]*ForeignKeyConstraint, error) {

	rows, err := conn.Query(ctx, fkConstraintSql, tableSchema)
	if err != nil && !isNoRows(err) {
		return nil, err
	} else if err != nil && isNoRows(err) {
		return []*ForeignKeyConstraint{}, nil
	}

	output := []*ForeignKeyConstraint{}
	for rows.Next() {
		var o ForeignKeyConstraint
		err := rows.Scan(
			&o.ConstraintName,
			&o.SchemaName,
			&o.TableName,
			&o.ColumnName,
			&o.ForeignSchemaName,
			&o.ForeignTableName,
			&o.ForeignColumnName,
		)
		if err != nil {
			return nil, err
		}
		output = append(output, &o)
	}
	return output, nil
}

type TableDependency = map[string][]string

// Key is schema.table value is list of tables that key depends on
func GetPostgresTableDependencies(
	constraints []*ForeignKeyConstraint,
) TableDependency {
	tdmap := map[string][]string{}
	for _, constraint := range constraints {
		tdmap[buildTableKey(constraint.SchemaName, constraint.TableName)] = []string{}
	}

	for _, constraint := range constraints {
		key := buildTableKey(constraint.SchemaName, constraint.TableName)
		tdmap[key] = append(tdmap[key], buildTableKey(constraint.ForeignSchemaName, constraint.ForeignTableName))
	}

	for k, v := range tdmap {
		tdmap[k] = UniqueSlice[string](func(val string) string { return val }, v)
	}
	return tdmap
}

func UniqueSlice[T any](keyFn func(T) string, genSlices ...[]T) []T {
	seen := map[string]struct{}{}
	output := []T{}

	for genIdx := range genSlices {
		for idx := range genSlices[genIdx] {
			val := genSlices[genIdx][idx]
			key := keyFn(val)
			if _, ok := seen[key]; !ok {
				output = append(output, val)
				seen[key] = struct{}{}
			}
		}
	}
	return output
}

func buildTableKey(
	schemaName string,
	tableName string,
) string {
	return fmt.Sprintf("%s.%s", schemaName, tableName)
}
