package sqlmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"golang.org/x/sync/errgroup"
)

const (
	DisableForeignKeyChecks = "SET FOREIGN_KEY_CHECKS = 0;"
)

type MysqlManager struct {
	querier mysql_queries.Querier
	pool    mysql_queries.DBTX
	close   func()
}

func (m *MysqlManager) GetDatabaseSchema(ctx context.Context) ([]*DatabaseSchemaRow, error) {
	dbSchemas, err := m.querier.GetDatabaseSchema(ctx, m.pool)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return []*DatabaseSchemaRow{}, nil
	}
	result := []*DatabaseSchemaRow{}
	for _, row := range dbSchemas {
		result = append(result, &DatabaseSchemaRow{
			TableSchema:   row.TableSchema,
			TableName:     row.TableName,
			ColumnName:    row.ColumnName,
			DataType:      row.DataType,
			ColumnDefault: row.ColumnDefault,
			IsNullable:    row.IsNullable,
		})
	}
	return result, nil
}

// returns: {public.users: { id: struct{}{}, created_at: struct{}{}}}
func (m *MysqlManager) GetSchemaColumnMap(ctx context.Context) (map[string]map[string]*ColumnInfo, error) {
	dbSchemas, err := m.GetDatabaseSchema(ctx)
	if err != nil {
		return nil, err
	}
	result := getUniqueSchemaColMappings(dbSchemas)
	return result, nil
}

func (m *MysqlManager) GetForeignKeyConstraints(ctx context.Context, schemas []string) ([]*ForeignKeyConstraintsRow, error) {
	holder := make([][]*mysql_queries.GetForeignKeyConstraintsRow, len(schemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range schemas {
		idx := idx
		schema := schemas[idx]
		errgrp.Go(func() error {
			constraints, err := m.querier.GetForeignKeyConstraints(errctx, m.pool, schema)
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
	result := []*ForeignKeyConstraintsRow{}
	for _, row := range output {
		result = append(result, &ForeignKeyConstraintsRow{
			SchemaName:        row.SchemaName,
			TableName:         row.TableName,
			ColumnName:        row.ColumnName,
			IsNullable:        convertNullableTextToBool(row.IsNullable),
			ConstraintName:    row.ConstraintName,
			ForeignSchemaName: row.ForeignSchemaName,
			ForeignTableName:  row.ForeignTableName,
			ForeignColumnName: row.ForeignColumnName,
		})
	}
	return result, nil
}

// Key is schema.table value is list of tables that key depends on
func (m *MysqlManager) GetForeignKeyConstraintsMap(ctx context.Context, schemas []string) (map[string][]*ForeignConstraint, error) {
	fkConstraints, err := m.GetForeignKeyConstraints(ctx, schemas)
	if err != nil {
		return nil, err
	}
	constraints := map[string][]*ForeignConstraint{}
	for _, row := range fkConstraints {
		tableName := BuildTable(row.SchemaName, row.TableName)
		if _, exists := constraints[tableName]; !exists {
			constraints[tableName] = []*ForeignConstraint{}
		}
		constraints[tableName] = append(constraints[tableName], &ForeignConstraint{
			Column:     row.ColumnName,
			IsNullable: row.IsNullable,
			ForeignKey: &ForeignKey{
				Table:  BuildTable(row.ForeignSchemaName, row.ForeignTableName),
				Column: row.ForeignColumnName,
			},
		})

	}
	return constraints, err
}

func (m *MysqlManager) GetPrimaryKeyConstraints(ctx context.Context, schemas []string) ([]*PrimaryKey, error) {
	holder := make([][]*mysql_queries.GetPrimaryKeyConstraintsRow, len(schemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range schemas {
		idx := idx
		schema := schemas[idx]
		errgrp.Go(func() error {
			constraints, err := m.querier.GetPrimaryKeyConstraints(errctx, m.pool, schema)
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
	result := []*PrimaryKey{}
	for _, row := range output {
		result = append(result, &PrimaryKey{
			Schema:  row.SchemaName,
			Table:   row.TableName,
			Columns: []string{row.ColumnName},
		})
	}
	return result, nil
}

func (m *MysqlManager) GetPrimaryKeyConstraintsMap(ctx context.Context, schemas []string) (map[string][]string, error) {
	primaryKeys, err := m.GetPrimaryKeyConstraints(ctx, schemas)
	if err != nil {
		return nil, err
	}
	result := map[string][]string{}
	for _, row := range primaryKeys {
		tableName := fmt.Sprintf("%s.%s", row.Schema, row.Table)
		if _, exists := result[tableName]; !exists {
			result[tableName] = []string{}
		}
		result[tableName] = append(result[tableName], row.Columns...)
	}
	return result, nil
}

func (m *MysqlManager) GetUniqueConstraintsMap(ctx context.Context, schemas []string) (map[string][][]string, error) {
	holder := make([][]*mysql_queries.GetUniqueConstraintsRow, len(schemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range schemas {
		idx := idx
		schema := schemas[idx]
		errgrp.Go(func() error {
			constraints, err := m.querier.GetUniqueConstraints(errctx, m.pool, schema)
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

	rows := []*mysql_queries.GetUniqueConstraintsRow{}
	for _, schemas := range holder {
		rows = append(rows, schemas...)
	}

	uniqueConstraintMap := map[string][]*mysql_queries.GetUniqueConstraintsRow{}
	for _, c := range rows {
		_, ok := uniqueConstraintMap[c.ConstraintName]
		if ok {
			uniqueConstraintMap[c.ConstraintName] = append(uniqueConstraintMap[c.ConstraintName], c)
		} else {
			uniqueConstraintMap[c.ConstraintName] = []*mysql_queries.GetUniqueConstraintsRow{c}
		}
	}
	output := map[string][][]string{}
	for _, constraints := range uniqueConstraintMap {
		uc := []string{}
		var key string
		for _, c := range constraints {
			key = BuildTable(c.SchemaName, c.TableName)
			_, ok := output[key]
			if !ok {
				output[key] = [][]string{}
			}
			uc = append(uc, c.ColumnName)
		}
		output[key] = append(output[key], uc)
	}

	return output, nil
}

func (m *MysqlManager) GetRolePermissionsMap(ctx context.Context, role string) (map[string][]string, error) {
	rows, err := m.querier.GetMysqlRolePermissions(ctx, m.pool, role)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return map[string][]string{}, nil
	}

	schemaTablePrivsMap := map[string][]string{}
	for _, permission := range rows {
		key := fmt.Sprintf("%s.%s", permission.TableSchema, permission.TableName)
		schemaTablePrivsMap[key] = append(schemaTablePrivsMap[key], permission.PrivilegeType)
	}
	return schemaTablePrivsMap, err
}

func (m *MysqlManager) GetCreateTableStatement(ctx context.Context, schema, table string) (string, error) {
	result, err := getShowTableCreate(ctx, m.pool, schema, table)
	if err != nil {
		return "", fmt.Errorf("unable to get table create statement: %w", err)
	}
	result.CreateTable = strings.Replace(
		result.CreateTable,
		fmt.Sprintf("CREATE TABLE `%s`", table),
		fmt.Sprintf("CREATE TABLE `%s`.`%s`", schema, table),
		1, // do it once
	)
	split := strings.Split(result.CreateTable, "CREATE TABLE")
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s;", split[1]), nil
}

type databaseTableShowCreate struct {
	Table       string `db:"Table"`
	CreateTable string `db:"Create Table"`
}

func getShowTableCreate(
	ctx context.Context,
	conn mysql_queries.DBTX,
	schema string,
	table string,
) (*databaseTableShowCreate, error) {
	getShowTableCreateSql := fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`;", schema, table)
	row := conn.QueryRowContext(ctx, getShowTableCreateSql)
	var output databaseTableShowCreate
	err := row.Scan(
		&output.Table,
		&output.CreateTable,
	)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (m *MysqlManager) BatchExec(ctx context.Context, batchSize int, statements []string, opts *BatchExecOpts) error {
	for i := 0; i < len(statements); i += batchSize {
		end := i + batchSize
		if end > len(statements) {
			end = len(statements)
		}

		batchCmd := strings.Join(statements[i:end], " ")
		if opts != nil && opts.Prefix != nil && *opts.Prefix != "" {
			batchCmd = fmt.Sprintf("%s %s", *opts.Prefix, batchCmd)
		}
		_, err := m.pool.ExecContext(ctx, batchCmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MysqlManager) Exec(ctx context.Context, statement string) error {
	_, err := m.pool.ExecContext(ctx, statement)
	if err != nil {
		return err
	}
	return nil
}

func (m *MysqlManager) Close() {
	if m.pool != nil && m.close != nil {
		m.close()
	}
}

func BuildMysqlTruncateStatement(
	schema string,
	table string,
) (string, error) {
	builder := goqu.Dialect("mysql")
	sqltable := goqu.S(schema).Table(table)
	truncateStmt := builder.From(sqltable).Truncate()
	stmt, _, err := truncateStmt.ToSQL()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s;", stmt), nil
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
