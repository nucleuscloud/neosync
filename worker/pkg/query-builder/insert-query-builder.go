package querybuilder

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/lib/pq"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	gotypeutil "github.com/nucleuscloud/neosync/internal/gotypeutil"
	mysqlutil "github.com/nucleuscloud/neosync/internal/mysql"
	pgutil "github.com/nucleuscloud/neosync/internal/postgres"
	sqlserverutil "github.com/nucleuscloud/neosync/internal/sqlserver"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

func GetInsertBuilder(
	logger *slog.Logger,
	driver, schema, table string,
	columns []string,
	opts ...InsertOption,
) (InsertQueryBuilder, error) {
	options := &InsertOptions{
		columnDefaults:  []*neosync_benthos.ColumnDefaultProperties{},
		columnDataTypes: []string{},
	}
	for _, opt := range opts {
		opt(options)
	}

	switch driver {
	case sqlmanager_shared.PostgresDriver, sqlmanager_shared.DefaultPostgresDriver:
		return &PostgresDriver{
			driver:  driver,
			logger:  logger,
			schema:  schema,
			table:   table,
			columns: columns,
			options: options,
		}, nil
	case sqlmanager_shared.MysqlDriver:
		return &MysqlDriver{
			driver:  driver,
			logger:  logger,
			schema:  schema,
			table:   table,
			columns: columns,
			options: options,
		}, nil
	case sqlmanager_shared.MssqlDriver:
		return &MssqlDriver{
			driver:  driver,
			logger:  logger,
			schema:  schema,
			table:   table,
			columns: columns,
			options: options,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

// InsertQueryBuilder provides an interface for building SQL insert queries across different database drivers.
type InsertQueryBuilder interface {
	// BuildInsertQuery generates a complete SQL insert statement for multiple rows of data.
	BuildInsertQuery(rows [][]any) (query string, args []any, err error)

	// BuildPreparedInsertQuerySingleRow generates a prepared SQL insert statement for a single row.
	BuildPreparedInsertQuerySingleRow() (query string, err error)
	// BuildPreparedInsertArgs processes the input rows and returns properly formatted arguments for use with a prepared statement
	BuildPreparedInsertArgs(rows [][]any) [][]any
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

// WithColumnDataTypes adds column datatypes
func WithColumnDataTypes(types []string) InsertOption {
	return func(opts *InsertOptions) {
		opts.columnDataTypes = types
	}
}

// WithColumnDefaults adds ColumnDefaultProperties
func WithColumnDefaults(defaults []*neosync_benthos.ColumnDefaultProperties) InsertOption {
	return func(opts *InsertOptions) {
		opts.columnDefaults = defaults
	}
}

type PostgresDriver struct {
	driver        string
	logger        *slog.Logger
	schema, table string
	columns       []string
	options       *InsertOptions
}

func (d *PostgresDriver) BuildInsertQuery(rows [][]any) (query string, queryargs []any, err error) {
	var goquRows []exp.Vals
	if d.options.rawInsertMode {
		goquRows = toGoquVals(updateDefaultVals(rows, d.options.columnDefaults))
	} else {
		goquRows = toGoquVals(getPostgresVals(d.logger, rows, d.options.columnDataTypes, d.options.columnDefaults))
	}

	insertQuery, args, err := BuildInsertQuery(d.driver, d.schema, d.table, d.columns, goquRows, &d.options.onConflictDoNothing)
	if err != nil {
		return "", nil, err
	}
	if d.shouldOverrideColumnDefault(d.options.columnDefaults) {
		insertQuery = sqlmanager_postgres.BuildPgInsertIdentityAlwaysSql(insertQuery)
	}
	return insertQuery, args, err
}

func (d *PostgresDriver) BuildPreparedInsertQuerySingleRow() (string, error) {
	query, err := BuildPreparedInsertQuery(d.driver, d.schema, d.table, d.columns, 1, d.options.onConflictDoNothing)
	if err != nil {
		return "", err
	}

	if d.shouldOverrideColumnDefault(d.options.columnDefaults) {
		query = sqlmanager_postgres.BuildPgInsertIdentityAlwaysSql(query)
	}
	return query, err
}

func (d *PostgresDriver) BuildPreparedInsertArgs(rows [][]any) [][]any {
	if d.options.rawInsertMode {
		return rows
	}
	return getPostgresVals(d.logger, rows, d.options.columnDataTypes, d.options.columnDefaults)
}

// TODO move this logic to PGX processor
func getPostgresVals(logger *slog.Logger, rows [][]any, columnDataTypes []string, columnDefaultProperties []*neosync_benthos.ColumnDefaultProperties) [][]any {
	newVals := [][]any{}
	for _, row := range rows {
		newRow := []any{}
		for i, a := range row {
			var colDataType string
			if i < len(columnDataTypes) {
				colDataType = columnDataTypes[i]
			}
			var colDefaults *neosync_benthos.ColumnDefaultProperties
			if i < len(columnDefaultProperties) {
				colDefaults = columnDefaultProperties[i]
			}
			if pgutil.IsJsonPgDataType(colDataType) {
				bits, err := json.Marshal(a)
				if err != nil {
					logger.Error("unable to marshal JSON", "error", err.Error())
					newRow = append(newRow, a)
					continue
				}
				newRow = append(newRow, bits)
			} else if gotypeutil.IsMultiDimensionalSlice(a) || gotypeutil.IsSliceOfMaps(a) {
				newRow = append(newRow, goqu.Literal(pgutil.FormatPgArrayLiteral(a, colDataType)))
			} else if gotypeutil.IsSlice(a) {
				newRow = append(newRow, pq.Array(a))
			} else if colDefaults != nil && colDefaults.HasDefaultTransformer {
				newRow = append(newRow, goqu.Literal(defaultStr))
			} else {
				newRow = append(newRow, a)
			}
		}
		newVals = append(newVals, newRow)
	}
	return newVals
}

func (d *PostgresDriver) shouldOverrideColumnDefault(columnDefaults []*neosync_benthos.ColumnDefaultProperties) bool {
	for _, cd := range columnDefaults {
		if cd != nil && !cd.HasDefaultTransformer && cd.NeedsOverride {
			return true
		}
	}
	return false
}

type MysqlDriver struct {
	driver        string
	logger        *slog.Logger
	schema, table string
	columns       []string
	options       *InsertOptions
}

func (d *MysqlDriver) BuildInsertQuery(rows [][]any) (query string, queryargs []any, err error) {
	var goquRows []exp.Vals
	if d.options.rawInsertMode {
		goquRows = toGoquVals(updateDefaultVals(rows, d.options.columnDefaults))
	} else {
		goquRows = toGoquVals(getMysqlVals(d.logger, rows, d.options.columnDataTypes, d.options.columnDefaults))
	}

	insertQuery, args, err := BuildInsertQuery(d.driver, d.schema, d.table, d.columns, goquRows, &d.options.onConflictDoNothing)
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

func (d *MysqlDriver) BuildPreparedInsertQuerySingleRow() (string, error) {
	query, err := BuildPreparedInsertQuery(d.driver, d.schema, d.table, d.columns, 1, d.options.onConflictDoNothing)
	if err != nil {
		return "", err
	}

	if d.options.prefix != nil && *d.options.prefix != "" {
		query = addPrefix(query, *d.options.prefix)
	}
	if d.options.suffix != nil && *d.options.suffix != "" {
		query = addSuffix(query, *d.options.suffix)
	}
	return query, err
}

func (d *MysqlDriver) BuildPreparedInsertArgs(rows [][]any) [][]any {
	if d.options.rawInsertMode {
		return rows
	}
	return getMysqlVals(d.logger, rows, d.options.columnDataTypes, d.options.columnDefaults)
}

func getMysqlVals(logger *slog.Logger, rows [][]any, columnDataTypes []string, columnDefaultProperties []*neosync_benthos.ColumnDefaultProperties) [][]any {
	newVals := [][]any{}
	for _, row := range rows {
		newRow := []any{}
		for idx, a := range row {
			var colDataType string
			if idx < len(columnDataTypes) {
				colDataType = columnDataTypes[idx]
			}
			var colDefaults *neosync_benthos.ColumnDefaultProperties
			if idx < len(columnDefaultProperties) {
				colDefaults = columnDefaultProperties[idx]
			}
			if colDefaults != nil && colDefaults.HasDefaultTransformer {
				newRow = append(newRow, goqu.Literal(defaultStr))
			} else if mysqlutil.IsJsonDataType(colDataType) {
				bits, err := json.Marshal(a)
				if err != nil {
					logger.Error("unable to marshal JSON", "error", err.Error())
					newRow = append(newRow, a)
					continue
				}
				newRow = append(newRow, bits)
			} else {
				newRow = append(newRow, a)
			}
		}
		newVals = append(newVals, newRow)
	}
	return newVals
}

type MssqlDriver struct {
	driver        string
	logger        *slog.Logger
	schema, table string
	columns       []string
	options       *InsertOptions
}

func (d *MssqlDriver) BuildInsertQuery(rows [][]any) (query string, queryargs []any, err error) {
	processedCols, processedRow, processedColDefaults := d.filterOutDefaultIdentityColumns(d.columns, rows, d.options.columnDefaults)
	if len(processedRow) == 0 {
		return sqlserverutil.GeSqlServerDefaultValuesInsertSql(d.schema, d.table, len(rows)), []any{}, nil
	}
	var goquRows []exp.Vals
	if d.options.rawInsertMode {
		goquRows = toGoquVals(processedRow)
	} else {
		goquRows = toGoquVals(getMssqlVals(d.logger, processedRow, processedColDefaults))
	}

	insertQuery, args, err := BuildInsertQuery(d.driver, d.schema, d.table, processedCols, goquRows, &d.options.onConflictDoNothing)
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

func (d *MssqlDriver) BuildPreparedInsertQuerySingleRow() (string, error) {
	query, err := BuildPreparedInsertQuery(d.driver, d.schema, d.table, d.columns, 1, d.options.onConflictDoNothing)
	if err != nil {
		return "", err
	}

	if d.options.prefix != nil && *d.options.prefix != "" {
		query = addPrefix(query, *d.options.prefix)
	}
	if d.options.suffix != nil && *d.options.suffix != "" {
		query = addSuffix(query, *d.options.suffix)
	}
	return query, err
}

func (d *MssqlDriver) BuildPreparedInsertArgs(rows [][]any) [][]any {
	_, processedRow, processedColDefaults := d.filterOutDefaultIdentityColumns(d.columns, rows, d.options.columnDefaults)
	if d.options.rawInsertMode {
		return processedRow
	}
	return getMssqlVals(d.logger, processedRow, processedColDefaults)
}

func getMssqlVals(logger *slog.Logger, rows [][]any, columnDefaultProperties []*neosync_benthos.ColumnDefaultProperties) [][]any {
	newVals := [][]any{}
	for _, row := range rows {
		newRow := []any{}
		for idx, a := range row {
			var colDefaults *neosync_benthos.ColumnDefaultProperties
			if idx < len(columnDefaultProperties) {
				colDefaults = columnDefaultProperties[idx]
			}
			if colDefaults != nil && colDefaults.HasDefaultTransformer {
				newRow = append(newRow, goqu.Literal(defaultStr))
			} else if gotypeutil.IsMap(a) {
				bits, err := gotypeutil.MapToJson(a)
				if err != nil {
					logger.Error("unable to marshal map to JSON", "error", err.Error())
					newRow = append(newRow, a)
				} else {
					newRow = append(newRow, bits)
				}
			} else {
				newRow = append(newRow, a)
			}
		}

		newVals = append(newVals, newRow)
	}
	return newVals
}

func (d *MssqlDriver) filterOutDefaultIdentityColumns(
	columnsNames []string,
	dataRows [][]any,
	colDefaultProperties []*neosync_benthos.ColumnDefaultProperties,
) (columns []string, rows [][]any, columnDefaultProperties []*neosync_benthos.ColumnDefaultProperties) {
	newDataRows := sqlserverutil.GoTypeToSqlServerType(dataRows)
	return sqlserverutil.FilterOutSqlServerDefaultIdentityColumns(d.driver, columnsNames, newDataRows, colDefaultProperties)
}

func addPrefix(insertQuery, prefix string) string {
	return prefix + strings.TrimSuffix(insertQuery, ";") + ";"
}

func addSuffix(insertQuery, suffix string) string {
	return strings.TrimSuffix(insertQuery, ";") + ";" + suffix
}

func toGoquVals(rows [][]any) []goqu.Vals {
	gvals := []goqu.Vals{}
	for _, row := range rows {
		gval := goqu.Vals{}
		for _, v := range row {
			gval = append(gval, v)
		}
		gvals = append(gvals, gval)
	}
	return gvals
}

func updateDefaultVals(rows [][]any, columnDefaultProperties []*neosync_benthos.ColumnDefaultProperties) [][]any {
	newVals := [][]any{}
	for _, row := range rows {
		newRow := []any{}
		for i, a := range row {
			var colDefaults *neosync_benthos.ColumnDefaultProperties
			if i < len(columnDefaultProperties) {
				colDefaults = columnDefaultProperties[i]
			}
			if colDefaults != nil && colDefaults.HasDefaultTransformer {
				newRow = append(newRow, goqu.Literal(defaultStr))
			} else {
				newRow = append(newRow, a)
			}
		}
		newVals = append(newVals, newRow)
	}
	return newVals
}
