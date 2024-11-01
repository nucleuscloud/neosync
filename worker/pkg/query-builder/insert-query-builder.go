package querybuilder

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/doug-martin/goqu/v9"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	sqlserverutil "github.com/nucleuscloud/neosync/internal/sqlserver"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

func GetInsertBuilder(driver string, logger *slog.Logger) (InsertQueryBuilder, error) {
	switch driver {
	case sqlmanager_shared.PostgresDriver, "postgres":
		return &PostgresDriver{
			driver: driver,
			logger: logger,
		}, nil
	case sqlmanager_shared.MysqlDriver:
		return &MysqlDriver{
			driver: driver,
			logger: logger,
		}, nil
	case sqlmanager_shared.MssqlDriver:
		return &MssqlDriver{
			driver: driver,
			logger: logger,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

type InsertQueryBuilder interface {
	BuildInsertQuery(schema, table string, columns []string, rows [][]any, opts ...InsertOption) (string, []any, error)
}

type InsertOption func(*InsertOptions)
type InsertOptions struct {
	rawInsertMode       bool
	onConflictDoNothing bool
	columnDataTypes     []string
	columnDefaults      []*neosync_benthos.ColumnDefaultProperties
	prefix, suffix      *string
}

// WithRawInsertMode inserts data as is
func WithRawInsertMode() InsertOption {
	return func(opts *InsertOptions) {
		opts.rawInsertMode = true
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
	}
}

func WithColumnDataTypes(types []string) InsertOption {
	return func(o *InsertOptions) {
		o.columnDataTypes = types
	}
}

func WithColumnDefaults(defaults []*neosync_benthos.ColumnDefaultProperties) InsertOption {
	return func(o *InsertOptions) {
		o.columnDefaults = defaults
	}
}

type PostgresDriver struct {
	driver string
	logger *slog.Logger
}

func (d *PostgresDriver) BuildInsertQuery(schema, table string, columns []string, rows [][]any, opts ...InsertOption) (string, []any, error) {
	options := InsertOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	goquRows := d.getPostgresVals(rows, options.columnDataTypes, options.columnDefaults)
	query, args, err := BuildInsertQuery(d.driver, schema, table, columns, goquRows, &options.onConflictDoNothing)
	if err != nil {
		return "", nil, err
	}
	if d.shouldOverrideColumnDefault(options.columnDefaults) {
		query = sqlmanager_postgres.BuildPgInsertIdentityAlwaysSql(query)
	}
	return query, args, err
}

func (d *PostgresDriver) getPostgresVals(row [][]any, columnDataTypes []string, columnDefaults []*neosync_benthos.ColumnDefaultProperties) []goqu.Vals {
	gval := []goqu.Vals{}

	return gval
}

func (d *PostgresDriver) shouldOverrideColumnDefault(columnDefaults []*neosync_benthos.ColumnDefaultProperties) bool {
	for _, d := range columnDefaults {
		if !d.HasDefaultTransformer && d.NeedsOverride {
			return true
		}
	}
	return false
}

type MysqlDriver struct {
	driver string
	logger *slog.Logger
}

func (d *MysqlDriver) BuildInsertQuery(schema, table string, columns []string, rows [][]any, opts ...InsertOption) (string, []any, error) {
	options := InsertOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	goquRows := d.getMysqlVals(rows, options.columnDataTypes, options.columnDefaults)

	query, args, err := BuildInsertQuery(d.driver, schema, table, columns, goquRows, &options.onConflictDoNothing)
	if err != nil {
		return "", nil, err
	}

	if len(goquRows) == 0 {
		query = sqlserverutil.GeSqlServerDefaultValuesInsertSql(s.schema, s.table, len(rows))
	}

	if options.includePrefixSuffix {
		query = addPrefixSuffix(query, options.prefix, options.suffix)
	}
	return query, args, err
}

func (d *MysqlDriver) getMysqlVals(rows [][]any, columnDataTypes []string, columnDefaults []*neosync_benthos.ColumnDefaultProperties) []goqu.Vals {
	gval := []goqu.Vals{}

	return gval
}

type MssqlDriver struct {
	driver string
	logger *slog.Logger
}

func (d *MssqlDriver) BuildInsertQuery(schema, table string, columns []string, rows [][]any, opts ...InsertOption) (string, []any, error) {
	options := InsertOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	goquRows := d.getMssqlVals(rows, options.columnDataTypes, options.columnDefaults)
	query, args, err := BuildInsertQuery(d.driver, schema, table, columns, goquRows, &options.onConflictDoNothing)
	if err != nil {
		return "", nil, err
	}
	if options.includePrefixSuffix {
		query = addPrefixSuffix(query, options.prefix, options.suffix)
	}
	return query, args, err
}

func (d *MssqlDriver) getMssqlVals(row [][]any, columnDataTypes []string, columnDefaults []*neosync_benthos.ColumnDefaultProperties) []goqu.Vals {
	gval := []goqu.Vals{}

	return gval
}

func addPrefixSuffix(insertQuery string, prefix, suffix *string) string {
	var query string
	if prefix != nil && *prefix != "" {
		query = *prefix
	}

	query += strings.TrimSuffix(insertQuery, ";") + ";"

	if suffix != nil && *suffix != "" {
		query += *suffix
	}
	return query
}
