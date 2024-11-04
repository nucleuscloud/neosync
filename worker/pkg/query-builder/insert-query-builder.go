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
	case sqlmanager_shared.PostgresDriver, "postgres":
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

type InsertQueryBuilder interface {
	BuildInsertQuery(rows [][]any) (string, []any, error)

	BuildPreparedInsertQuerySingleRow() (string, error)
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

func WithColumnDataTypes(types []string) InsertOption {
	return func(opts *InsertOptions) {
		opts.columnDataTypes = types
	}
}

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
		goquRows = toGoquVals(rows)
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
				s, err := gotypeutil.ParseSlice(a)
				if err != nil {
					logger.Error("unable to parse slice", "error", err.Error())
					newRow = append(newRow, a)
					continue
				}
				newRow = append(newRow, pq.Array(s))
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
		goquRows = toGoquVals(rows)
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
			fmt.Println("a", a)
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
				fmt.Println("aa", a)
				newRow = append(newRow, a)
			}
		}

		fmt.Println("newRow", newRow)
		newVals = append(newVals, newRow)
	}
	return newVals
}

func (d *MssqlDriver) filterOutDefaultIdentityColumns(
	columnsNames []string,
	dataRows [][]any,
	colDefaultProperties []*neosync_benthos.ColumnDefaultProperties,
) (columns []string, rows [][]any, columnDefaultProperties []*neosync_benthos.ColumnDefaultProperties) {
	defaultIdentityCols := []string{}
	for idx, d := range colDefaultProperties {
		cName := columnsNames[idx]
		if d != nil && d.HasDefaultTransformer && d.NeedsOverride && d.NeedsReset {
			defaultIdentityCols = append(defaultIdentityCols, cName)
		}
	}
	newDataRows := sqlserverutil.GoTypeToSqlServerType(dataRows)
	return sqlserverutil.FilterOutSqlServerDefaultIdentityColumns(d.driver, defaultIdentityCols, columnsNames, newDataRows, colDefaultProperties)
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
