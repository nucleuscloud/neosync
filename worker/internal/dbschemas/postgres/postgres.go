package dbschemas_postgres

import (
	"context"
	"fmt"

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

type DatabaseTableConstraints struct {
	Name       string `db:"name,omitempty"`
	Type       string `db:"contype,omitempty"`
	Definition string `db:"definition,omitempty"`
}

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
    'CREATE TABLE ' || a.attrelid::regclass::TEXT || '(' ||
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
	row := conn.QueryRow(ctx, getTableCreateStatementSql, req.Table)

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
