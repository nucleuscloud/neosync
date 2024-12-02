package querybuilder

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"

	// import the dialect
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlserver"
	"github.com/doug-martin/goqu/v9/exp"
)

const defaultStr = "DEFAULT"

type SubsetReferenceKey struct {
	Table         string
	Columns       []string
	OriginalTable *string
}
type SubsetColumnConstraint struct {
	Columns     []string
	NotNullable []bool
	ForeignKey  *SubsetReferenceKey
}

func getGoquDialect(driver string) goqu.DialectWrapper {
	if driver == sqlmanager_shared.PostgresDriver {
		return goqu.Dialect(sqlmanager_shared.DefaultPostgresDriver)
	}
	return goqu.Dialect(driver)
}

func BuildSelectQuery(
	driver, table string,
	columns []string,
	whereClause *string,
) (string, error) {
	builder := getGoquDialect(driver)
	sqltable := goqu.I(table)

	selectColumns := make([]any, len(columns))
	for i, col := range columns {
		selectColumns[i] = col
	}
	query := builder.From(sqltable).Select(selectColumns...)

	if whereClause != nil && *whereClause != "" {
		query = query.Where(goqu.L(*whereClause))
	}
	sql, _, err := query.ToSQL()
	if err != nil {
		return "", err
	}

	return formatSqlQuery(sql), nil
}

func formatSqlQuery(sql string) string {
	return fmt.Sprintf("%s;", sql)
}

func BuildSelectLimitQuery(
	driver, table string,
	limit uint,
) (string, error) {
	builder := getGoquDialect(driver)
	sqltable := goqu.I(table)
	sql, _, err := builder.From((sqltable)).Limit(limit).ToSQL()
	if err != nil {
		return "", err
	}
	return sql, nil
}

func BuildInsertQuery(
	driver, schema, table string,
	columns []string,
	values []goqu.Vals,
	onConflictDoNothing *bool,
) (sql string, args []any, err error) {
	builder := getGoquDialect(driver)
	sqltable := goqu.S(schema).Table(table)
	insertCols := make([]any, len(columns))
	for i, col := range columns {
		insertCols[i] = col
	}
	insert := builder.Insert(sqltable).Prepared(true).Cols(insertCols...)
	for _, row := range values {
		insert = insert.Vals(row)
	}
	// adds on conflict do nothing to insert query
	if *onConflictDoNothing {
		insert = insert.OnConflict(goqu.DoNothing())
	}

	query, args, err := insert.ToSQL()
	if err != nil {
		return "", nil, err
	}
	return query, args, nil
}

// BuildPreparedQuery creates a prepared statement query template
func BuildPreparedInsertQuery(
	driver, schema, table string,
	columns []string,
	rowCount int,
	onConflictDoNothing bool,
) (string, error) {
	if rowCount < 1 {
		rowCount = 1
	}

	builder := getGoquDialect(driver)
	sqltable := goqu.S(schema).Table(table)

	insertCols := make([]any, len(columns))
	for i, col := range columns {
		insertCols[i] = col
	}

	insert := builder.Insert(sqltable).
		Prepared(true).
		Cols(insertCols...)

	// Add placeholder rows based on rowCount
	for i := 0; i < rowCount; i++ {
		placeholderRow := make(goqu.Vals, len(columns))
		for j := range columns {
			placeholderRow[j] = nil
		}
		insert = insert.Vals(placeholderRow)
	}

	if onConflictDoNothing {
		insert = insert.OnConflict(goqu.DoNothing())
	}

	query, _, err := insert.ToSQL()
	if err != nil {
		return "", err
	}

	return query, nil
}

func BuildUpdateQuery(
	driver, schema, table string,
	insertColumns []string,
	whereColumns []string,
	columnValueMap map[string]any,
) (string, error) {
	builder := getGoquDialect(driver)
	sqltable := goqu.S(schema).Table(table)

	updateRecord := goqu.Record{}
	for _, col := range insertColumns {
		val := columnValueMap[col]
		updateRecord[col] = val
	}

	where := []exp.Expression{}
	for _, col := range whereColumns {
		val := columnValueMap[col]
		where = append(where, goqu.Ex{col: val})
	}

	update := builder.Update(sqltable).
		Set(updateRecord).
		Where(where...)

	query, _, err := update.ToSQL()
	if err != nil {
		return "", err
	}
	return query, nil
}

func BuildTruncateQuery(
	driver, table string,
) (string, error) {
	builder := getGoquDialect(driver)
	sqltable := goqu.I(table)
	truncate := builder.Truncate(sqltable)
	query, _, err := truncate.ToSQL()
	if err != nil {
		return "", err
	}
	return query, nil
}
