package querybuilder

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/doug-martin/goqu/v9"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
)

func GetInsertBuilder(
	logger *slog.Logger,
	driver, schema, table string,
	opts ...InsertOption,
) (InsertQueryBuilder, error) {
	options := &InsertOptions{}
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

type InsertOption func(*InsertOptions)
type InsertOptions struct {
	shouldOverrideColumnDefault bool
	onConflictDoNothing         bool
	onConflictDoUpdate          bool
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
		opts.onConflictDoNothing = true
		opts.onConflictDoUpdate = false
	}
}

// WithOnConflictDoUpdate adds on conflict do update to insert query
func WithOnConflictDoUpdate() InsertOption {
	return func(opts *InsertOptions) {
		opts.onConflictDoUpdate = true
		opts.onConflictDoNothing = false
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

	insertQuery, args, err := BuildInsertQuery(d.driver, d.schema, d.table, goquRows, &d.options.onConflictDoNothing)
	if err != nil {
		return "", nil, err
	}
	if d.options.shouldOverrideColumnDefault {
		insertQuery = sqlmanager_postgres.BuildPgInsertIdentityAlwaysSql(insertQuery)
	}
	return insertQuery, args, err
}

type MysqlDriver struct {
	driver        string
	logger        *slog.Logger
	schema, table string
	options       *InsertOptions
}

func (d *MysqlDriver) BuildInsertQuery(rows []map[string]any) (query string, queryargs []any, err error) {
	goquRows := toGoquRecords(rows)

	insertQuery, args, err := BuildInsertQuery(d.driver, d.schema, d.table, goquRows, &d.options.onConflictDoNothing)
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

	insertQuery, args, err := BuildInsertQuery(d.driver, d.schema, d.table, goquRows, &d.options.onConflictDoNothing)
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
