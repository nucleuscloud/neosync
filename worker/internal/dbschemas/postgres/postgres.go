package dbschemas_postgres

import (
	"context"
	"fmt"
	"strings"

	pg_queries "github.com/nucleuscloud/neosync/worker/gen/go/db/postgresql"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
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
			TableSchema: schema,
			TableName:   table,
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
	tableDefs := append(columns, constraints...) //nolint
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (%s);`, schema, table, strings.Join(tableDefs, ", "))
}

func buildTableCol(record *pg_queries.GetDatabaseTableSchemaRow) string {
	pieces := []string{record.ColumnName, record.DataType, buildNullableText(record)}
	if record.ColumnDefault != "" && record.ColumnDefault != "NULL" {
		pieces = append(pieces, "DEFAULT", record.ColumnDefault)
	}
	return strings.Join(pieces, " ")
}
func buildNullableText(record *pg_queries.GetDatabaseTableSchemaRow) string {
	if record.IsNullable == "NO" {
		return "NOT NULL"
	}
	return "NULL"
}

type TableDependency = map[string][]string

// Key is schema.table value is list of tables that key depends on
func GetPostgresTableDependencies(
	constraints []*pg_queries.GetForeignKeyConstraintsRow,
) TableDependency {
	tdmap := map[string][]string{}
	for _, constraint := range constraints {
		tdmap[buildTableKey(constraint.SchemaName, constraint.TableName)] = []string{}
	}

	for _, constraint := range constraints {
		key := buildTableKey(constraint.SchemaName, constraint.TableName)
		tdmap[key] = append(tdmap[key], buildTableKey(constraint.ForeignSchemaName, constraint.ForeignTableName))
	}

	for k, v := range tdmap {
		tdmap[k] = UniqueSlice[string](func(val string) string { return val }, v)
	}
	return tdmap
}

func UniqueSlice[T any](keyFn func(T) string, genSlices ...[]T) []T {
	seen := map[string]struct{}{}
	output := []T{}

	for genIdx := range genSlices {
		for idx := range genSlices[genIdx] {
			val := genSlices[genIdx][idx]
			key := keyFn(val)
			if _, ok := seen[key]; !ok {
				output = append(output, val)
				seen[key] = struct{}{}
			}
		}
	}
	return output
}

func buildTableKey(
	schemaName string,
	tableName string,
) string {
	return fmt.Sprintf("%s.%s", schemaName, tableName)
}

func GetUniqueSchemaColMappings(
	dbschemas []*pg_queries.GetDatabaseSchemaRow,
) map[string]map[string]struct{} {
	groupedSchemas := map[string]map[string]struct{}{} // ex: {public.users: { id: struct{}{}, created_at: struct{}{}}}
	for _, record := range dbschemas {
		key := neosync_benthos.BuildBenthosTable(record.TableSchema, record.TableName)
		if _, ok := groupedSchemas[key]; ok {
			groupedSchemas[key][record.ColumnName] = struct{}{}
		} else {
			groupedSchemas[key] = map[string]struct{}{
				record.ColumnName: {},
			}
		}
	}
	return groupedSchemas
}
