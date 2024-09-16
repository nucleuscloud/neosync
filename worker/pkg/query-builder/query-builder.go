package querybuilder

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/lib/pq"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"

	// import the dialect
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
	gotypeutil "github.com/nucleuscloud/neosync/internal/gotypeutil"
	pgutil "github.com/nucleuscloud/neosync/internal/postgres"
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

func BuildSelectQuery(
	driver, table string,
	columns []string,
	whereClause *string,
) (string, error) {
	builder := goqu.Dialect(driver)
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
	builder := goqu.Dialect(driver)
	sqltable := goqu.I(table)
	sql, _, err := builder.From((sqltable)).Limit(limit).ToSQL()
	if err != nil {
		return "", err
	}
	return sql, nil
}

func getGoquVals(driver string, row []any, columnDataTypes []string) goqu.Vals {
	if driver == sqlmanager_shared.PostgresDriver {
		return getPgGoquVals(row, columnDataTypes)
	}
	gval := goqu.Vals{}
	for _, a := range row {
		if isDefault(a) {
			gval = append(gval, goqu.Literal(defaultStr))
		} else {
			gval = append(gval, a)
		}
	}
	return gval
}

func getPgGoquVals(row []any, columnDataTypes []string) goqu.Vals {
	gval := goqu.Vals{}
	for i, a := range row {
		var colDataType string
		if i < len(columnDataTypes) {
			colDataType = columnDataTypes[i]
		}
		if gotypeutil.IsMap(a) {
			bits, err := gotypeutil.MapToJson(a)
			if err != nil {
				gval = append(gval, a)
				continue
			}
			gval = append(gval, bits)
		} else if gotypeutil.IsMultiDimensionalSlice(a) || gotypeutil.IsSliceOfMaps(a) {
			gval = append(gval, goqu.Literal(pgutil.FormatPgArrayLiteral(a, colDataType)))
		} else if gotypeutil.IsSlice(a) {
			s, err := gotypeutil.ParseSlice(a)
			if err != nil {
				gval = append(gval, a)
				continue
			}
			gval = append(gval, pq.Array(s))
		} else if isDefault(a) {
			gval = append(gval, goqu.Literal(defaultStr))
		} else {
			gval = append(gval, a)
		}
	}
	return gval
}

func isDefault(val any) bool {
	valStr, isString := val.(string)
	if !isString {
		return false
	}
	return strings.EqualFold(valStr, defaultStr)
}

func BuildInsertQuery(
	driver, schema, table string,
	columns []string,
	columnDataTypes []string,
	values [][]any,
	onConflictDoNothing *bool,
) (string, error) {
	builder := goqu.Dialect(driver)
	sqltable := goqu.S(schema).Table(table)
	insertCols := make([]any, len(columns))
	for i, col := range columns {
		insertCols[i] = col
	}
	insert := builder.Insert(sqltable).Cols(insertCols...)
	for _, row := range values {
		gval := getGoquVals(driver, row, columnDataTypes)
		insert = insert.Vals(gval)
	}
	// adds on conflict do nothing to insert query
	if *onConflictDoNothing {
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
	builder := goqu.Dialect(driver)
	sqltable := goqu.S(schema).Table(table)

	updateRecord := goqu.Record{}
	for _, col := range insertColumns {
		val := columnValueMap[col]
		if isDefault(val) {
			updateRecord[col] = goqu.L(defaultStr)
		} else {
			updateRecord[col] = val
		}
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
	builder := goqu.Dialect(driver)
	sqltable := goqu.I(table)
	truncate := builder.Truncate(sqltable)
	query, _, err := truncate.ToSQL()
	if err != nil {
		return "", err
	}
	return query, nil
}
