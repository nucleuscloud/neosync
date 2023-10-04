package dbschemas_postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type DatabaseSchema struct {
	TableSchema     string  `db:"table_schema,omitempty"`
	TableName       string  `db:"table_name,omitempty"`
	ColumnName      string  `db:"column_name,omitempty"`
	OrdinalPosition int     `db:"ordinal_position,omitempty"`
	ColumnDefault   *string `db:"column_default,omitempty"`
	IsNullable      string  `db:"is_nullable"`
	DataType        string  `db:"data_type,omitempty"`
}

func (d *DatabaseSchema) GetTableKey() string {
	return fmt.Sprintf("%s.%s", d.TableSchema, d.TableName)
}

// type DatabaseTableConstraints struct {
// 	Name       string `db:"name,omitempty"`
// 	Type       string `db:"contype,omitempty"`
// 	Definition string `db:"definition,omitempty"`
// }

func GetDatabaseSchemas(
	ctx context.Context,
	conn *pgx.Conn,
) ([]*DatabaseSchema, error) {
	rows, err := conn.Query(ctx, `
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
	`)
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

func isNoRows(err error) bool {
	return err != nil && err == pgx.ErrNoRows
}

type GetTableCreateStatementRequest struct {
	Table string
}

const (
	getTableCreateStatementSql = `--getTableCreateStatmentSql
	SELECT
    'CREATE TABLE IF NOT EXISTS ' || a.attrelid::regclass::TEXT || '(' ||
string_agg(
        a.attname || ' ' || pg_catalog.format_type(
            a.atttypid,
            a.atttypmod
        )||
    CASE
            WHEN
        (
                SELECT
                    substring(
                        pg_catalog.pg_get_expr(
                            d.adbin,
                            d.adrelid
                        ) FOR 128
                    )
                FROM
                    pg_catalog.pg_attrdef d
                WHERE
                    d.adrelid = a.attrelid
                    AND d.adnum = a.attnum
                    AND a.atthasdef
            ) IS NOT
NULL THEN
        ' DEFAULT ' || (
                SELECT
                    substring(
                        pg_catalog.pg_get_expr(
                            d.adbin,
                            d.adrelid
                        ) FOR 128
                    )
                FROM
                    pg_catalog.pg_attrdef d
                WHERE
                    d.adrelid = a.attrelid
                    AND d.adnum = a.attnum
                    AND a.atthasdef
            )
            ELSE
        ''
        END
||
    CASE
            WHEN a.attnotnull = TRUE THEN
        ' NOT NULL'
            ELSE
        ''
        END,
        E'\n,'
    ) || ');' AS create_stmt
FROM
    pg_catalog.pg_attribute a
JOIN pg_class ON
    a.attrelid = pg_class.oid
WHERE
    a.attrelid::regclass::varchar = $1
    AND a.attnum > 0
    AND NOT a.attisdropped
    AND pg_class.relkind = 'r'
GROUP BY
    a.attrelid;
	`
)

func GetTableCreateStatement(
	ctx context.Context,
	conn *pgx.Conn,
	req *GetTableCreateStatementRequest,
) (string, error) {
	// hack to fix tables in public schema
	table := req.Table
	if strings.HasPrefix(req.Table, "public.") {
		table = strings.TrimPrefix(req.Table, "public.")
	}
	row := conn.QueryRow(ctx, getTableCreateStatementSql, table)

	var createStmt string

	err := row.Scan(&createStmt)
	if err != nil {
		return "", err
	}
	return createStmt, nil
}

// type SchemaGen struct {
// 	schemas map[string]map[string]string
// }

// func GenerateCreateTableStatements(
// 	schemas []*DatabaseSchema,
// ) map[string]map[string]string {
// 	output := map[string]map[string]string{}

// 	tableMap := map[string][]*DatabaseSchema{}

// 	for _, schema := range schemas {
// 		output[schema.TableSchema] = map[string]string{}
// 		if _, ok := tableMap[schema.GetTableKey()]; ok {
// 			tableMap[schema.GetTableKey()] = append(tableMap[schema.GetTableKey()], schema)
// 		} else {
// 			tableMap[schema.GetTableKey()] = []*DatabaseSchema{schema}
// 		}
// 	}
// 	// for _, schema := range schemas {
// 	// }

// 	return output
// }

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
	conn *pgx.Conn,
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

func GetTableSeedQueue(
	schemas []*DatabaseSchema,
	constraints []*ForeignKeyConstraint,
) [][]string {
	tables := []string{}
	for _, schema := range schemas {
		tables = append(tables, buildTableKey(schema.TableSchema, schema.TableName))
	}

	td := GetPostgresTableDependencies(constraints)

	roots := []string{}
	for _, table := range tables {
		if _, ok := td[table]; !ok {
			roots = append(roots, table)
		}
	}

	output := append([][]string{}, roots)
	queuedTables := map[string]struct{}{}
	for _, root := range roots {
		queuedTables[root] = struct{}{}
	}

	for len(tables) != len(queuedTables) {
		nextRound := []string{}
		for table, deps := range td {
			if _, ok := queuedTables[table]; ok {
				continue
			}
			if ok := isTableReady(queuedTables, deps); ok {
				nextRound = append(nextRound, table)
			}
		}
		if len(nextRound) == 0 {
			fmt.Println("infinite loop!")
			break
		}
		output = append(output, nextRound)
		for _, table := range nextRound {
			queuedTables[table] = struct{}{}
		}
	}

	return output
}

func isTableReady(
	queue map[string]struct{},
	deps []string,
) bool {
	for _, dep := range deps {
		if _, ok := queue[dep]; !ok {
			return false
		}
	}
	return true
}

func UniqueSlice[T any](keyFn func(T) string, genSlices ...[]T) []T {
	uniqueSet := map[string]T{}

	for _, genSlice := range genSlices {
		for _, val := range genSlice {
			uniqueSet[keyFn(val)] = val
		}
	}

	result := make([]T, 0, len(uniqueSet))
	for _, val := range uniqueSet {
		result = append(result, val)
	}
	return result
}

func buildTableKey(
	schemaName string,
	tableName string,
) string {
	return fmt.Sprintf("%s.%s", schemaName, tableName)
}
