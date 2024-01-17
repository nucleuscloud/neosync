package dbschemas_postgres

import (
	"context"
	"fmt"
	"strings"

	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
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
	pieces := []string{record.ColumnName, buildDataType(record), buildNullableText(record)}
	if record.ColumnDefault != "" && record.ColumnDefault != "NULL" {
		if strings.HasPrefix(record.ColumnDefault, "nextval") && record.DataType == "integer" {
			pieces = []string{record.ColumnName, "SERIAL"}
		} else {
			pieces = append(pieces, "DEFAULT", record.ColumnDefault)
		}
	}
	return strings.Join(pieces, " ")
}

func buildDataType(record *pg_queries.GetDatabaseTableSchemaRow) string {
	if record.CharacterMaximumLength != nil {
		if strings.EqualFold(record.DataType, "character varying") || strings.EqualFold(record.DataType, "character") || strings.EqualFold(record.DataType, "varchar") || strings.EqualFold(record.DataType, "bpchar") {
			return fmt.Sprintf("%s(%d)", record.DataType, record.CharacterMaximumLength)
		}
	}
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
	constraints []*pg_queries.GetForeignKeyConstraintsRow,
) dbschemas.TableDependency {
	tableConstraints := map[string]*dbschemas.TableConstraints{}
	for _, c := range constraints {
		tableName := dbschemas.BuildTable(c.SchemaName, c.TableName)

		constraint, ok := tableConstraints[tableName]
		if !ok {
			tableConstraints[tableName] = &dbschemas.TableConstraints{
				Constraints: []*dbschemas.ForeignConstraint{
					{Column: c.ColumnName, IsNullable: dbschemas.ConvertIsNullableToBool(c.IsNullable), ForeignKey: &dbschemas.ForeignKey{
						Table:  dbschemas.BuildTable(c.ForeignSchemaName, c.ForeignTableName),
						Column: c.ForeignColumnName,
					}},
				},
			}
		} else {
			constraint.Constraints = append(constraint.Constraints, &dbschemas.ForeignConstraint{
				Column: c.ColumnName, IsNullable: dbschemas.ConvertIsNullableToBool(c.IsNullable), ForeignKey: &dbschemas.ForeignKey{
					Table:  dbschemas.BuildTable(c.ForeignSchemaName, c.ForeignTableName),
					Column: c.ForeignColumnName,
				},
			})
		}
	}
	return tableConstraints
}

func GetUniqueSchemaColMappings(
	schemas []*pg_queries.GetDatabaseSchemaRow,
) map[string]map[string]struct{} {
	groupedSchemas := map[string]map[string]struct{}{} // ex: {public.users: { id: struct{}{}, created_at: struct{}{}}}
	for _, record := range schemas {
		key := dbschemas.BuildTable(record.TableSchema, record.TableName)
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

func GetAllPostgresFkConstraints(
	pgquerier pg_queries.Querier,
	ctx context.Context,
	conn pg_queries.DBTX,
	uniqueSchemas []string,
) ([]*pg_queries.GetForeignKeyConstraintsRow, error) {
	holder := make([][]*pg_queries.GetForeignKeyConstraintsRow, len(uniqueSchemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range uniqueSchemas {
		idx := idx
		schema := uniqueSchemas[idx]
		errgrp.Go(func() error {
			constraints, err := pgquerier.GetForeignKeyConstraints(errctx, conn, schema)
			if err != nil {
				return err
			}
			holder[idx] = constraints
			return nil
		})
	}

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	output := []*pg_queries.GetForeignKeyConstraintsRow{}
	for _, schemas := range holder {
		output = append(output, schemas...)
	}
	return output, nil
}

func GetAllPostgresPkConstraints(
	pgquerier pg_queries.Querier,
	ctx context.Context,
	conn pg_queries.DBTX,
	uniqueSchemas []string,
) ([]*pg_queries.GetPrimaryKeyConstraintsRow, error) {
	holder := make([][]*pg_queries.GetPrimaryKeyConstraintsRow, len(uniqueSchemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range uniqueSchemas {
		idx := idx
		schema := uniqueSchemas[idx]
		errgrp.Go(func() error {
			constraints, err := pgquerier.GetPrimaryKeyConstraints(errctx, conn, schema)
			if err != nil {
				return err
			}
			holder[idx] = constraints
			return nil
		})
	}

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	output := []*pg_queries.GetPrimaryKeyConstraintsRow{}
	for _, schemas := range holder {
		output = append(output, schemas...)
	}
	return output, nil
}

func GetPostgresTablePrimaryKeys(
	primaryKeyConstraints []*pg_queries.GetPrimaryKeyConstraintsRow,
) map[string][]string {
	pkConstraintMap := map[string][]*pg_queries.GetPrimaryKeyConstraintsRow{}
	for _, c := range primaryKeyConstraints {
		_, ok := pkConstraintMap[c.ConstraintName]
		if ok {
			pkConstraintMap[c.ConstraintName] = append(pkConstraintMap[c.ConstraintName], c)
		} else {
			pkConstraintMap[c.ConstraintName] = []*pg_queries.GetPrimaryKeyConstraintsRow{c}
		}
	}
	pkMap := map[string][]string{}
	for _, constraints := range pkConstraintMap {
		for _, c := range constraints {
			key := dbschemas.BuildTable(c.SchemaName, c.TableName)
			_, ok := pkMap[key]
			if ok {
				pkMap[key] = append(pkMap[key], c.ColumnName)
			} else {
				pkMap[key] = []string{c.ColumnName}
			}
		}
	}
	return pkMap
}
