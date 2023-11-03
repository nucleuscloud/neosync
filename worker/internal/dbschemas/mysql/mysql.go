package dbschemas_mysql

import (
	"context"
	"database/sql"
	"fmt"

	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
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
	c.table_schema NOT IN('sys', 'performance_schema', 'mysql')
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
	c.table_schema = ? AND t.table_name = ?
	AND t.table_type = 'BASE TABLE'
	ORDER BY c.ordinal_position ASC;
	`
)

type DBTX interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func GetDatabaseSchemas(
	ctx context.Context,
	conn DBTX,
) ([]*DatabaseSchema, error) {
	rows, err := conn.QueryContext(ctx, getDatabaseSchemaSql)
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
	Schema            string `db:"db_schema"`
	Table             string `db:"table_name"`
	ConstraintName    string `db:"constraint_name"`
	ColumnName        string `db:"column_name"`
	ForeignSchemaName string `db:"foreign_schema_name"`
	ForeignTableName  string `db:"foreign_table_name"`
	ForeignColumnName string `db:"foreign_column_name"`
	UpdateRule        string `db:"update_rule"`
	DeleteRule        string `db:"delete_rule"`
}

const (
	getTableConstraintsSql = `-- name: GetTableConstraints
	SELECT
	kcu.constraint_name
	,
	kcu.table_schema AS db_schema
	,
	kcu.table_name as table_name
	,
	kcu.column_name as column_name
	,
	kcu.referenced_table_schema AS foreign_schema_name
	,
	kcu.referenced_table_name AS foreign_table_name
	,
	kcu.referenced_column_name AS foreign_column_name
	,
	rc.update_rule
	,
	rc.delete_rule
FROM information_schema.key_column_usage kcu
LEFT JOIN information_schema.referential_constraints rc
	ON
	kcu.constraint_name = rc.constraint_name
WHERE
	kcu.table_schema = ? AND kcu.table_name = ?;
`
)

func GetTableConstraints(
	ctx context.Context,
	conn DBTX,
	schema string,
	table string,
) ([]*DatabaseTableConstraint, error) {
	rows, err := conn.QueryContext(ctx, getTableConstraintsSql, schema, table)
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
			&o.ColumnName,
			&o.ForeignSchemaName,
			&o.ForeignTableName,
			&o.ForeignColumnName,
			&o.UpdateRule,
			&o.DeleteRule,
		)
		if err != nil {
			return nil, err
		}
		output = append(output, &o)
	}
	return output, nil
}

type DatabaseTableShowCreate struct {
	Table       string `db:"Table"`
	CreateTable string `db:"Create Table"`
}

func getShowTableCreate(
	ctx context.Context,
	conn DBTX,
	schema string,
	table string,
) (*DatabaseTableShowCreate, error) {
	getShowTableCreateSql := fmt.Sprintf(`SHOW CREATE TABLE %s.%s;`, schema, table)
	row := conn.QueryRowContext(ctx, getShowTableCreateSql)

	var output DatabaseTableShowCreate
	err := row.Scan(
		&output.Table,
		&output.CreateTable,
	)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func isNoRows(err error) bool {
	return err != nil && err == sql.ErrNoRows
}

type GetTableCreateStatementRequest struct {
	Schema string
	Table  string
}

func GetTableCreateStatement(
	ctx context.Context,
	conn DBTX,
	req *GetTableCreateStatementRequest,
) (string, error) {
	result, err := getShowTableCreate(ctx, conn, req.Schema, req.Table)
	if err != nil {
		return "", fmt.Errorf("unable to get table create statement: %w", err)
	}
	return result.CreateTable, nil
}

const (
	fkConstraintSql = `
	SELECT
	rc.constraint_name
	,
	kcu.table_schema AS schema_name
	,
	kcu.table_name as table_name
	,
	kcu.column_name as column_name
	,
	kcu.referenced_table_schema AS foreign_schema_name
	,
	kcu.referenced_table_name AS foreign_table_name
	,
	kcu.referenced_column_name AS foreign_column_name
FROM
	information_schema.referential_constraints rc
JOIN information_schema.key_column_usage kcu
	ON
	kcu.constraint_name = rc.constraint_name
WHERE
	kcu.table_schema = ?
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

	rows, err := conn.QueryContext(ctx, fkConstraintSql, tableSchema)
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
func GetMysqlTableDependencies(
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

func GetUniqueSchemaColMappings(
	dbschemas []*DatabaseSchema,
) map[string]map[string]struct{} {
	groupedSchemas := map[string]map[string]struct{}{} // ex: {public.users: { id: struct{}{}, created_at: struct{}{}}}
	for _, record := range dbschemas {
		key := neosync_benthos.BuildBenthosTable(record.TableSchema, record.TableName)
		if _, ok := groupedSchemas[key]; ok {
			groupedSchemas[key][record.ColumnName] = struct{}{}
		} else {
			groupedSchemas[key] = map[string]struct{}{
				record.ColumnName: {},
			}
		}
	}
	return groupedSchemas
}
