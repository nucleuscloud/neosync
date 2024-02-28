package dbschemas_mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"golang.org/x/sync/errgroup"
)

const (
	DisableForeignKeyChecks = "SET FOREIGN_KEY_CHECKS = 0;"
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
	result.CreateTable = strings.Replace(
		result.CreateTable,
		fmt.Sprintf("CREATE TABLE `%s`", req.Table),
		fmt.Sprintf("CREATE TABLE %s", dbschemas.BuildTableWithEscape(req.Schema, req.Table)),
		1, // do it once
	)
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

func GetMysqlTablePrimaryKeys(
	primaryKeyConstraints []*mysql_queries.GetPrimaryKeyConstraintsRow,
) map[string][]string {
	pkConstraintMap := map[string][]*mysql_queries.GetPrimaryKeyConstraintsRow{}
	for _, c := range primaryKeyConstraints {
		_, ok := pkConstraintMap[c.ConstraintName]
		if ok {
			pkConstraintMap[c.ConstraintName] = append(pkConstraintMap[c.ConstraintName], c)
		} else {
			pkConstraintMap[c.ConstraintName] = []*mysql_queries.GetPrimaryKeyConstraintsRow{c}
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

func GetUniqueSchemaColMappings(
	schemas []*mysql_queries.GetDatabaseSchemaRow,
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

func toColumnInfo(row *mysql_queries.GetDatabaseSchemaRow) *dbschemas.ColumnInfo {
	return &dbschemas.ColumnInfo{
		OrdinalPosition:        row.OrdinalPosition,
		ColumnDefault:          row.ColumnDefault,
		IsNullable:             row.IsNullable,
		DataType:               row.DataType,
		CharacterMaximumLength: fromSqlNullInt32(row.CharacterMaximumLength),
		NumericPrecision:       fromSqlNullInt32(row.NumericPrecision),
		NumericScale:           fromSqlNullInt32(row.NumericScale),
	}
}

func fromSqlNullInt32(nullInt sql.NullInt32) *int32 {
	if nullInt.Valid {
		return &nullInt.Int32
	}
	return nil
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

func GetAllMysqlPkConstraints(
	mysqlquerier mysql_queries.Querier,
	ctx context.Context,
	conn mysql_queries.DBTX,
	schemas []string,
) ([]*mysql_queries.GetPrimaryKeyConstraintsRow, error) {
	holder := make([][]*mysql_queries.GetPrimaryKeyConstraintsRow, len(schemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range schemas {
		idx := idx
		schema := schemas[idx]
		errgrp.Go(func() error {
			constraints, err := mysqlquerier.GetPrimaryKeyConstraints(errctx, conn, schema)
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

	output := []*mysql_queries.GetPrimaryKeyConstraintsRow{}
	for _, schemas := range holder {
		output = append(output, schemas...)
	}
	return output, nil
}

func BuildTruncateStatement(
	schema string,
	table string,
) string {
	return fmt.Sprintf("TRUNCATE TABLE `%s`.`%s`;", schema, table)
}

func BatchExecStmts(
	ctx context.Context,
	pool mysql_queries.DBTX,
	batchSize int,
	statements []string,
	initalStatement *string, // this is run as the first statement in each batch
) error {
	for i := 0; i < len(statements); i += batchSize {
		end := i + batchSize
		if end > len(statements) {
			end = len(statements)
		}

		batchCmd := strings.Join(statements[i:end], " ")
		if initalStatement != nil && *initalStatement != "" {
			batchCmd = fmt.Sprintf("%s %s", *initalStatement, batchCmd)
		}
		_, err := pool.ExecContext(ctx, batchCmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func EscapeMysqlColumns(cols []string) []string {
	outcols := make([]string, len(cols))
	for idx := range cols {
		outcols[idx] = EscapeMysqlColumn(cols[idx])
	}
	return outcols
}

func EscapeMysqlColumn(col string) string {
	return fmt.Sprintf("`%s`", col)
}
