package querybuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
)

func GetInsertBuilder(
	logger *slog.Logger,
	driver, schema, table string,
	opts ...InsertOption,
) (InsertQueryBuilder, error) {
	options := &InsertOptions{
		conflictConfig: &conflictConfig{},
	}
	for _, opt := range opts {
		opt(options)
	}

	switch driver {
	case sqlmanager_shared.PostgresDriver:
		return &PostgresDriver{
			driver:  driver,
			logger:  logger,
			schema:  schema,
			table:   table,
			options: options,
		}, nil
	case sqlmanager_shared.MysqlDriver:
		return &MysqlDriver{
			driver:  driver,
			logger:  logger,
			schema:  schema,
			table:   table,
			options: options,
		}, nil
	case sqlmanager_shared.MssqlDriver:
		return &MssqlDriver{
			driver:  driver,
			logger:  logger,
			schema:  schema,
			table:   table,
			options: options,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

// InsertQueryBuilder provides an interface for building SQL insert queries across different database drivers.
type InsertQueryBuilder interface {
	// BuildInsertQuery generates a complete SQL insert statement for multiple rows of data.
	BuildInsertQuery(rows []map[string]any) (query string, args []any, err error)
}

type onConflictDoUpdateConfig struct {
	conflictColumns []string
}

type onConflictDoNothingConfig struct{}

type conflictConfig struct {
	onConflictDoNothing *onConflictDoNothingConfig
	onConflictDoUpdate  *onConflictDoUpdateConfig
}

type InsertOption func(*InsertOptions)
type InsertOptions struct {
	shouldOverrideColumnDefault bool
	conflictConfig              *conflictConfig
	prefix, suffix              *string
}

func WithShouldOverrideColumnDefault() InsertOption {
	return func(opts *InsertOptions) {
		opts.shouldOverrideColumnDefault = true
	}
}

// WithPrefix adds prefix to insert query
func WithPrefix(prefix *string) InsertOption {
	return func(opts *InsertOptions) {
		opts.prefix = prefix
	}
}

// WithSuffix adds suffix to insert query
func WithSuffix(suffix *string) InsertOption {
	return func(opts *InsertOptions) {
		opts.suffix = suffix
	}
}

// WithOnConflictDoNothing adds on conflict do nothing to insert query
func WithOnConflictDoNothing() InsertOption {
	return func(opts *InsertOptions) {
		opts.conflictConfig = &conflictConfig{
			onConflictDoNothing: &onConflictDoNothingConfig{},
			onConflictDoUpdate:  nil,
		}
	}
}

// WithOnConflictDoUpdate adds on conflict do update to insert query
func WithOnConflictDoUpdate(conflictColumns []string) InsertOption {
	return func(opts *InsertOptions) {
		opts.conflictConfig = &conflictConfig{
			onConflictDoUpdate: &onConflictDoUpdateConfig{
				conflictColumns: conflictColumns,
			},
			onConflictDoNothing: nil,
		}
	}
}

type PostgresDriver struct {
	driver        string
	logger        *slog.Logger
	schema, table string
	options       *InsertOptions
}

func (d *PostgresDriver) BuildInsertQuery(rows []map[string]any) (query string, queryargs []any, err error) {
	goquRows := toGoquRecords(rows)

	if d.options.conflictConfig.onConflictDoUpdate != nil {
		if len(rows) == 0 {
			return "", []any{}, errors.New("no rows to insert")
		}
		if len(d.options.conflictConfig.onConflictDoUpdate.conflictColumns) == 0 {
			d.logger.Warn("no conflict columns specified for on conflict do update, defaulting to on conflict do nothing")
			onConflictDoNothing := true
			insertQuery, args, err := BuildInsertQuery(d.driver, d.schema, d.table, goquRows, &onConflictDoNothing)
			if err != nil {
				return "", nil, fmt.Errorf("failed to build insert query on conflict do nothing fallback: %w", err)
			}
			if d.options.shouldOverrideColumnDefault {
				insertQuery = sqlmanager_postgres.BuildPgInsertIdentityAlwaysSql(insertQuery)
			}
			return insertQuery, args, nil
		}

		columns := make([]string, 0, len(rows[0]))
		for col := range rows[0] {
			columns = append(columns, col)
		}
		return d.buildInsertOnConflictDoUpdateQuery(goquRows, d.options.conflictConfig.onConflictDoUpdate.conflictColumns, columns)
	}

	onConflictDoNothing := d.options.conflictConfig.onConflictDoNothing != nil
	insertQuery, args, err := BuildInsertQuery(d.driver, d.schema, d.table, goquRows, &onConflictDoNothing)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build insert query: %w", err)
	}
	if d.options.shouldOverrideColumnDefault {
		insertQuery = sqlmanager_postgres.BuildPgInsertIdentityAlwaysSql(insertQuery)
	}
	return insertQuery, args, nil
}

func (d *PostgresDriver) buildInsertOnConflictDoUpdateQuery(
	records []goqu.Record,
	conflictColumns []string,
	updateColumns []string,
) (sql string, args []any, err error) {
	builder := getGoquDialect(sqlmanager_shared.GoquPostgresDriver)
	sqltable := goqu.S(d.schema).Table(d.table)
	insert := builder.Insert(sqltable).Prepared(true).Rows(records)

	updateRecord := goqu.Record{}
	for _, col := range updateColumns {
		if !slices.Contains(conflictColumns, col) {
			updateRecord[col] = goqu.L(fmt.Sprintf("EXCLUDED.%q", col))
		}
	}
	targetColumns := strings.Join(conflictColumns, ", ")
	insert = insert.OnConflict(goqu.DoUpdate(targetColumns, updateRecord))

	query, args, err := insert.ToSQL()
	if err != nil {
		return "", nil, err
	}
	return query, args, nil
}

type MysqlDriver struct {
	driver        string
	logger        *slog.Logger
	schema, table string
	options       *InsertOptions
}

func (d *MysqlDriver) BuildInsertQuery(rows []map[string]any) (query string, queryargs []any, err error) {
	goquRows := toGoquRecords(rows)

	if d.options.conflictConfig.onConflictDoUpdate != nil {
		if len(rows) == 0 {
			return "", []any{}, errors.New("no rows to insert")
		}

		columns := make([]string, 0, len(rows[0]))
		for col := range rows[0] {
			columns = append(columns, col)
		}
		insertQuery, args, err := d.buildMysqlInsertOnConflictDoUpdateQuery(goquRows, columns)
		if err != nil {
			return "", nil, err
		}
		return insertQuery, args, nil
	}

	onConflictDoNothing := d.options.conflictConfig.onConflictDoNothing != nil
	insertQuery, args, err := BuildInsertQuery(d.driver, d.schema, d.table, goquRows, &onConflictDoNothing)
	if err != nil {
		return "", nil, err
	}

	if d.options.prefix != nil && *d.options.prefix != "" {
		insertQuery = addPrefix(insertQuery, *d.options.prefix)
	}
	if d.options.suffix != nil && *d.options.suffix != "" {
		insertQuery = addSuffix(insertQuery, *d.options.suffix)
	}
	return insertQuery, args, err
}

func (d *MysqlDriver) buildMysqlInsertOnConflictDoUpdateQuery(
	records []goqu.Record,
	updateColumns []string,
) (sql string, args []any, err error) {
	builder := getGoquDialect(sqlmanager_shared.MysqlDriver)
	sqltable := goqu.S(d.schema).Table(d.table)
	insert := builder.Insert(sqltable).As("new").Prepared(true).Rows(records)
	if d.table == "departments" {
		jsonF, _ := json.MarshalIndent(records, "", " ")
		fmt.Printf("\n\n %s \n\n", string(jsonF))
		fmt.Println(records[0]["dept_label"], reflect.TypeOf(records[0]["dept_label"]))
	}

	updateRecord := goqu.Record{}
	for _, col := range updateColumns {
		updateRecord[col] = exp.NewIdentifierExpression("", "new", col)
	}
	targetColumn := "" // mysql does not support target column
	insert = insert.OnConflict(goqu.DoUpdate(targetColumn, updateRecord))

	query, args, err := insert.ToSQL()
	if err != nil {
		return "", nil, err
	}
	return strings.Replace(query, "INSERT IGNORE INTO", "INSERT INTO", 1), args, nil
}

type MssqlDriver struct {
	driver        string
	logger        *slog.Logger
	schema, table string
	options       *InsertOptions
}

func (d *MssqlDriver) BuildInsertQuery(rows []map[string]any) (query string, queryargs []any, err error) {
	if len(rows) == 0 || areAllRowsEmpty(rows) {
		return getSqlServerDefaultValuesInsertSql(d.schema, d.table, len(rows)), []any{}, nil
	}

	goquRows := toGoquRecords(rows)

	onConflictDoNothing := d.options.conflictConfig.onConflictDoNothing != nil
	insertQuery, args, err := BuildInsertQuery(d.driver, d.schema, d.table, goquRows, &onConflictDoNothing)
	if err != nil {
		return "", nil, err
	}

	if d.options.prefix != nil && *d.options.prefix != "" {
		insertQuery = addPrefix(insertQuery, *d.options.prefix)
	}
	if d.options.suffix != nil && *d.options.suffix != "" {
		insertQuery = addSuffix(insertQuery, *d.options.suffix)
	}

	return insertQuery, args, err
}

func areAllRowsEmpty(rows []map[string]any) bool {
	for _, row := range rows {
		if len(row) > 0 {
			return false
		}
	}
	return true
}

func addPrefix(insertQuery, prefix string) string {
	return prefix + strings.TrimSuffix(insertQuery, ";") + ";"
}

func addSuffix(insertQuery, suffix string) string {
	return strings.TrimSuffix(insertQuery, ";") + ";" + suffix
}

func toGoquRecords(rows []map[string]any) []goqu.Record {
	records := []goqu.Record{}
	for _, row := range rows {
		records = append(records, goqu.Record(row))
	}
	return records
}

func getSqlServerDefaultValuesInsertSql(schema, table string, rowCount int) string {
	var sqlStr string
	for i := 0; i < rowCount; i++ {
		sqlStr += fmt.Sprintf("INSERT INTO %q.%q DEFAULT VALUES;", schema, table)
	}
	return sqlStr
}
