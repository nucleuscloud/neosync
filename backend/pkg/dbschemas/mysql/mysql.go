package dbschemas_mysql

import (
	"context"
	"fmt"
	"strings"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"golang.org/x/sync/errgroup"
)

type GetTableCreateStatementRequest struct {
	Schema string
	Table  string
}

func GetTableCreateStatement(
	ctx context.Context,
	conn mysql_queries.DBTX,
	req *GetTableCreateStatementRequest,
) (string, error) {
	result, err := getShowTableCreate(ctx, conn, req.Schema, req.Table)
	if err != nil {
		return "", fmt.Errorf("unable to get table create statement: %w", err)
	}
	split := strings.Split(result.CreateTable, "CREATE TABLE")
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s;", split[1]), nil
}

type DatabaseTableShowCreate struct {
	Table       string `db:"Table"`
	CreateTable string `db:"Create Table"`
}

func getShowTableCreate(
	ctx context.Context,
	conn mysql_queries.DBTX,
	schema string,
	table string,
) (*DatabaseTableShowCreate, error) {
	getShowTableCreateSql := fmt.Sprintf(`SHOW CREATE TABLE %s.%s;`, schema, table)
	row := conn.QueryRowContext(ctx, getShowTableCreateSql)
	var output DatabaseTableShowCreate
	err := row.Scan(
		&output.Table,
		&output.CreateTable,
	)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

type TableDependency = map[string][]string

// Key is schema.table value is list of tables that key depends on
func GetMysqlTableDependencies(
	constraints []*mysql_queries.GetForeignKeyConstraintsRow,
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
	schemas []*mysql_queries.GetDatabaseSchemaRow,
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

func GetAllMysqlFkConstraints(
	mysqlquerier mysql_queries.Querier,
	ctx context.Context,
	conn mysql_queries.DBTX,
	schemas []string,
) ([]*mysql_queries.GetForeignKeyConstraintsRow, error) {
	holder := make([][]*mysql_queries.GetForeignKeyConstraintsRow, len(schemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range schemas {
		idx := idx
		schema := schemas[idx]
		errgrp.Go(func() error {
			constraints, err := mysqlquerier.GetForeignKeyConstraints(errctx, conn, schema)
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

	output := []*mysql_queries.GetForeignKeyConstraintsRow{}
	for _, schemas := range holder {
		output = append(output, schemas...)
	}
	return output, nil
}
