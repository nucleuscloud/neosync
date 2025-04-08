package querybuilder

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"

	// import the dialect
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlserver"
	"github.com/doug-martin/goqu/v9/exp"
)

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
		return goqu.Dialect(sqlmanager_shared.GoquPostgresDriver)
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

func BuildSampledSelectLimitQuery(
	driver, table string, limit uint,
) (string, error) {
	var randStmt string
	switch driver {
	case sqlmanager_shared.GoquPostgresDriver:
		randStmt = "RANDOM()"
	case sqlmanager_shared.MysqlDriver:
		randStmt = "RAND()"
	case sqlmanager_shared.MssqlDriver:
		randStmt = "NEWID()"
	}

	builder := getGoquDialect(driver)
	sqltable := goqu.I(table)
	sql, _, err := builder.
		From((sqltable)).
		Order(goqu.L(randStmt).Asc()).
		Limit(limit).
		ToSQL()
	if err != nil {
		return "", err
	}
	return sql, nil
}

func BuildInsertQuery(
	driver, schema, table string,
	records []goqu.Record,
	onConflictDoNothing *bool,
) (sql string, args []any, err error) {
	builder := getGoquDialect(driver)
	sqltable := goqu.S(schema).Table(table)
	insert := builder.Insert(sqltable).Prepared(true).Rows(records)
	// adds on conflict do nothing to insert query
	if *onConflictDoNothing {
		insert = insert.OnConflict(goqu.DoNothing())
	}

	query, args, err := insert.ToSQL()
	if err != nil {
		// check if it's a goqu encoding error and sanitize it
		if strings.Contains(err.Error(), "goqu_encode_error") {
			return "", nil, fmt.Errorf("goqu_encode_error: Unable to encode value")
		}
		return "", nil, err
	}
	return query, args, nil
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
		// check if it's a goqu encoding error and sanitize it
		if strings.Contains(err.Error(), "goqu_encode_error") {
			return "", fmt.Errorf("goqu_encode_error: Unable to encode value")
		}
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

func GetGoquDriverFromConnection(connection *mgmtv1alpha1.Connection) (string, error) {
	switch connection.GetConnectionConfig().GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return sqlmanager_shared.GoquPostgresDriver, nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return sqlmanager_shared.MysqlDriver, nil
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		return sqlmanager_shared.MssqlDriver, nil
	default:
		return "", fmt.Errorf("unsupported connection type: %T for goqu", connection.GetConnectionConfig().GetConfig())
	}
}
