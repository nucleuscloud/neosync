package dbschemas_mysql

import (
	"context"
	"fmt"

	mysql_queries "github.com/nucleuscloud/neosync/worker/gen/go/db/mysql"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
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
	return result.CreateTable, nil
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
	dbschemas []*mysql_queries.GetDatabaseSchemaRow,
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
