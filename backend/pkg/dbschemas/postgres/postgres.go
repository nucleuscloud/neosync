package dbschemas_postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"

	"golang.org/x/sync/errgroup"
)

func GetTableCreateStatement(
	ctx context.Context,
	conn pg_queries.DBTX,
	q pg_queries.Querier,
	schema string,
	table string,
) (string, error) {
	errgrp, errctx := errgroup.WithContext(ctx)

	var tableSchemas []*pg_queries.GetDatabaseTableSchemaRow
	errgrp.Go(func() error {
		result, err := q.GetDatabaseTableSchema(errctx, conn, &pg_queries.GetDatabaseTableSchemaParams{
			Schema: schema,
			Table:  table,
		})
		if err != nil {
			return fmt.Errorf("unable to generate database table schema: %w", err)
		}
		tableSchemas = result
		return nil
	})
	var tableConstraints []*pg_queries.GetTableConstraintsRow
	errgrp.Go(func() error {
		result, err := q.GetTableConstraints(errctx, conn, &pg_queries.GetTableConstraintsParams{
			Schema: schema,
			Table:  table,
		})
		if err != nil {
			return fmt.Errorf("unable to generate table constraints: %w", err)
		}
		tableConstraints = result
		return nil
	})
	if err := errgrp.Wait(); err != nil {
		return "", err
	}

	return generateCreateTableStatement(
		schema,
		table,
		tableSchemas,
		tableConstraints,
	), nil
}

// This assumes that the schemas and constraints as for a single table, not an entire db schema
func generateCreateTableStatement(
	schema string,
	table string,
	tableSchemas []*pg_queries.GetDatabaseTableSchemaRow,
	tableConstraints []*pg_queries.GetTableConstraintsRow,
) string {
	columns := make([]string, len(tableSchemas))
	for idx := range tableSchemas {
		record := tableSchemas[idx]
		columns[idx] = buildTableCol(record)
	}

	constraints := make([]string, len(tableConstraints))
	for idx := range tableConstraints {
		constraint := tableConstraints[idx]
		constraints[idx] = fmt.Sprintf("CONSTRAINT %s %s", constraint.ConstraintName, constraint.ConstraintDefinition)
	}
	tableDefs := append(columns, constraints...) //nolint:gocritic
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %q.%q (%s);`, schema, table, strings.Join(tableDefs, ", "))
}

func buildTableCol(record *pg_queries.GetDatabaseTableSchemaRow) string {
	pieces := []string{escapeColumnName(record.ColumnName), buildDataType(record), buildNullableText(record)}
	if record.ColumnDefault != "" {
		if strings.HasPrefix(record.ColumnDefault, "nextval") && record.DataType == "integer" {
			pieces[1] = "SERIAL"
		} else if strings.HasPrefix(record.ColumnDefault, "nextval") && record.DataType == "bigint" {
			pieces[1] = "BIGSERIAL"
		} else if strings.HasPrefix(record.ColumnDefault, "nextval") && record.DataType == "smallint" {
			pieces[1] = "SMALLSERIAL"
		} else if record.ColumnDefault != "NULL" {
			pieces = append(pieces, "DEFAULT", record.ColumnDefault)
		}
	}
	return strings.Join(pieces, " ")
}

// To escape a column name in postgres, they must be wrapped with double quotes ""
func escapeColumnName(columnName string) string {
	return fmt.Sprintf("%q", columnName)
}

func buildDataType(record *pg_queries.GetDatabaseTableSchemaRow) string {
	return record.DataType
}

func buildNullableText(record *pg_queries.GetDatabaseTableSchemaRow) string {
	if record.IsNullable == "NO" {
		return "NOT NULL"
	}
	return "NULL"
}

// Key is schema.table value is list of tables that key depends on
func GetPostgresTableDependencies(
	constraintRows []*pg_queries.GetTableConstraintsBySchemaRow,
) (dbschemas.TableDependency, error) {
	tableConstraints := map[string]*dbschemas.TableConstraints{}
	for _, row := range constraintRows {
		if len(row.ConstraintColumns) != len(row.ForeignColumnNames) {
			return nil, fmt.Errorf("length of columns was not equal to length of foreign key cols: %d %d", len(row.ConstraintColumns), len(row.ForeignColumnNames))
		}
		if len(row.ConstraintColumns) != len(row.Notnullable) {
			return nil, fmt.Errorf("length of columns was not equal to length of not nullable cols: %d %d", len(row.ConstraintColumns), len(row.Notnullable))
		}

		tableName := dbschemas.BuildTable(row.SchemaName, row.TableName)
		for idx, colname := range row.ConstraintColumns {
			fkcol := row.ForeignColumnNames[idx]
			notnullable := row.Notnullable[idx]

			constraints, ok := tableConstraints[tableName]
			constraint := &dbschemas.ForeignConstraint{
				Column:     colname,
				IsNullable: !notnullable, ForeignKey: &dbschemas.ForeignKey{
					Table:  dbschemas.BuildTable(row.ForeignSchemaName, row.ForeignTableName),
					Column: fkcol,
				},
			}
			if ok {
				constraints.Constraints = append(constraints.Constraints, constraint)
			} else {
				tableConstraints[tableName] = &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{constraint},
				}
			}
		}
	}
	return tableConstraints, nil
}

func GetUniqueSchemaColMappings(
	schemas []*pg_queries.GetDatabaseSchemaRow,
) map[string]map[string]*dbschemas.ColumnInfo {
	groupedSchemas := map[string]map[string]*dbschemas.ColumnInfo{} // ex: {public.users: { id: struct{}{}, created_at: struct{}{}}}
	for _, record := range schemas {
		key := dbschemas.BuildTable(record.TableSchema, record.TableName)
		if _, ok := groupedSchemas[key]; ok {
			groupedSchemas[key][record.ColumnName] = toColumnInfo(record)
		} else {
			groupedSchemas[key] = map[string]*dbschemas.ColumnInfo{
				record.ColumnName: toColumnInfo(record),
			}
		}
	}
	return groupedSchemas
}

func toColumnInfo(row *pg_queries.GetDatabaseSchemaRow) *dbschemas.ColumnInfo {
	return &dbschemas.ColumnInfo{
		OrdinalPosition:        int32(row.OrdinalPosition),
		ColumnDefault:          row.ColumnDefault,
		IsNullable:             row.IsNullable,
		DataType:               row.DataType,
		CharacterMaximumLength: ptr(row.CharacterMaximumLength),
		NumericPrecision:       ptr(row.NumericPrecision),
		NumericScale:           ptr(row.NumericScale),
	}
}

func ptr[T any](val T) *T {
	return &val
}

func GetAllPostgresForeignKeyConstraints(
	ctx context.Context,
	conn pg_queries.DBTX,
	pgquerier pg_queries.Querier,
	schemas []string,
) ([]*pg_queries.GetTableConstraintsBySchemaRow, error) {
	if len(schemas) == 0 {
		return []*pg_queries.GetTableConstraintsBySchemaRow{}, nil
	}
	rows, err := pgquerier.GetTableConstraintsBySchema(ctx, conn, schemas)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return []*pg_queries.GetTableConstraintsBySchemaRow{}, nil
	}

	output := []*pg_queries.GetTableConstraintsBySchemaRow{}
	for _, row := range rows {
		if row.ConstraintType != "f" {
			continue
		}
		output = append(output, row)
	}
	return output, nil
}

func BuildTruncateStatement(
	tables []string,
) string {
	return fmt.Sprintf("TRUNCATE TABLE %s;", strings.Join(tables, ", "))
}

func BuildTruncateCascadeStatement(
	schema string,
	table string,
) (string, error) {
	builder := goqu.Dialect("postgres")
	sqltable := goqu.S(schema).Table(table)
	stmt, _, err := builder.From(sqltable).Truncate().Cascade().ToSQL()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s;", stmt), nil
}

func BatchExecStmts(
	ctx context.Context,
	pool pg_queries.DBTX,
	batchSize int,
	statements []string,
) error {
	for i := 0; i < len(statements); i += batchSize {
		end := i + batchSize
		if end > len(statements) {
			end = len(statements)
		}

		batchCmd := strings.Join(statements[i:end], "\n")
		_, err := pool.Exec(ctx, batchCmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func EscapePgColumns(cols []string) []string {
	outcols := make([]string, len(cols))
	for idx := range cols {
		outcols[idx] = EscapePgColumn(cols[idx])
	}
	return outcols
}

func EscapePgColumn(col string) string {
	return fmt.Sprintf("%q", col)
}

// Returns a map by table name and lists all columns that are a part of a unique constraint
func GetAllPostgresUniqueConstraintsByTableCols(
	ctx context.Context,
	conn pg_queries.DBTX,
	pgquerier pg_queries.Querier,
	schemas []string,
) (map[string][]string, error) {
	if len(schemas) == 0 {
		return map[string][]string{}, nil
	}
	rows, err := pgquerier.GetTableConstraintsBySchema(ctx, conn, schemas)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return map[string][]string{}, nil
	}

	output := map[string][]string{}
	for _, row := range rows {
		if row.ConstraintType != "u" {
			continue
		}
		key := dbschemas.BuildTable(row.SchemaName, row.TableName)
		if _, ok := output[key]; ok {
			output[key] = append(output[key], row.ConstraintColumns...)
		} else {
			output[key] = append([]string{}, row.ConstraintColumns...)
		}
	}

	for key, val := range output {
		output[key] = dedupeSlice(val)
	}
	return output, nil
}

func dedupeSlice(input []string) []string {
	set := map[string]any{}
	for _, i := range input {
		set[i] = struct{}{}
	}
	output := make([]string, 0, len(set))
	for key := range set {
		output = append(output, key)
	}
	return output
}

func GetPostgresRolePermissions(
	ctx context.Context,
	conn pg_queries.DBTX,
	pgquerier pg_queries.Querier,
	role string,
) ([]*pg_queries.GetPostgresRolePermissionsRow, error) {
	rows, err := pgquerier.GetPostgresRolePermissions(ctx, conn, role)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return []*pg_queries.GetPostgresRolePermissionsRow{}, nil
	}

	output := []*pg_queries.GetPostgresRolePermissionsRow{}
	for _, row := range rows {
		output = append(output, &pg_queries.GetPostgresRolePermissionsRow{
			TableSchema:   row.TableSchema,
			TableName:     row.TableName,
			PrivilegeType: row.PrivilegeType,
		})
	}
	return output, nil
}
