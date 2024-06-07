package sqlmanager_mysql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"golang.org/x/sync/errgroup"
)

type MysqlManager struct {
	querier mysql_queries.Querier
	pool    mysql_queries.DBTX
	close   func()
}

func NewManager(querier mysql_queries.Querier, pool mysql_queries.DBTX, closer func()) *MysqlManager {
	return &MysqlManager{querier: querier, pool: pool, close: closer}
}

func (m *MysqlManager) GetDatabaseSchema(ctx context.Context) ([]*sqlmanager_shared.DatabaseSchemaRow, error) {
	dbSchemas, err := m.querier.GetDatabaseSchema(ctx, m.pool)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return []*sqlmanager_shared.DatabaseSchemaRow{}, nil
	}
	result := []*sqlmanager_shared.DatabaseSchemaRow{}
	for _, row := range dbSchemas {
		var generatedType *string
		if row.Extra.Valid && row.Extra.String == "GENERATED" {
			generatedTypeCopy := row.Extra.String
			generatedType = &generatedTypeCopy
		}
		result = append(result, &sqlmanager_shared.DatabaseSchemaRow{
			TableSchema:   row.TableSchema,
			TableName:     row.TableName,
			ColumnName:    row.ColumnName,
			DataType:      row.DataType,
			ColumnDefault: row.ColumnDefault,
			IsNullable:    row.IsNullable,
			GeneratedType: generatedType,
		})
	}
	return result, nil
}

// returns: {public.users: { id: struct{}{}, created_at: struct{}{}}}
func (m *MysqlManager) GetSchemaColumnMap(ctx context.Context) (map[string]map[string]*sqlmanager_shared.ColumnInfo, error) {
	dbSchemas, err := m.GetDatabaseSchema(ctx)
	if err != nil {
		return nil, err
	}
	result := sqlmanager_shared.GetUniqueSchemaColMappings(dbSchemas)
	return result, nil
}

func (m *MysqlManager) GetTableConstraintsBySchema(ctx context.Context, schemas []string) (*sqlmanager_shared.TableConstraints, error) {
	if len(schemas) == 0 {
		return &sqlmanager_shared.TableConstraints{}, nil
	}

	foreignKeyMap, err := m.GetForeignKeyConstraintsMap(ctx, schemas)
	if err != nil {
		return nil, err
	}

	primaryKeyMap, err := m.GetPrimaryKeyConstraintsMap(ctx, schemas)
	if err != nil {
		return nil, err
	}

	uniqueConstraintsMap, err := m.GetUniqueConstraintsMap(ctx, schemas)
	if err != nil {
		return nil, err
	}

	return &sqlmanager_shared.TableConstraints{
		ForeignKeyConstraints: foreignKeyMap,
		PrimaryKeyConstraints: primaryKeyMap,
		UniqueConstraints:     uniqueConstraintsMap,
	}, nil
}

func (m *MysqlManager) GetForeignKeyConstraints(ctx context.Context, schemas []string) ([]*sqlmanager_shared.ForeignKeyConstraintsRow, error) {
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
	result := []*sqlmanager_shared.ForeignKeyConstraintsRow{}
	for _, row := range output {
		result = append(result, &sqlmanager_shared.ForeignKeyConstraintsRow{
			SchemaName:        row.SchemaName,
			TableName:         row.TableName,
			ColumnName:        row.ColumnName,
			IsNullable:        sqlmanager_shared.ConvertNullableTextToBool(row.IsNullable),
			ConstraintName:    row.ConstraintName,
			ForeignSchemaName: row.ForeignSchemaName,
			ForeignTableName:  row.ForeignTableName,
			ForeignColumnName: row.ForeignColumnName,
		})
	}
	return result, nil
}

// Key is schema.table value is list of tables that key depends on
func (m *MysqlManager) GetForeignKeyConstraintsMap(ctx context.Context, schemas []string) (map[string][]*sqlmanager_shared.ForeignConstraint, error) {
	fkConstraints, err := m.GetForeignKeyConstraints(ctx, schemas)
	if err != nil {
		return nil, err
	}
	groupedFks := map[string][]*sqlmanager_shared.ForeignKeyConstraintsRow{} //  grouped by constraint name
	for _, row := range fkConstraints {
		groupedFks[row.ConstraintName] = append(groupedFks[row.ConstraintName], row)
	}
	constraints := map[string][]*sqlmanager_shared.ForeignConstraint{}
	for _, fks := range groupedFks {
		cols := []string{}
		notNullable := []bool{}
		fkCols := []string{}
		for _, fk := range fks {
			cols = append(cols, fk.ColumnName)
			notNullable = append(notNullable, !fk.IsNullable)
			fkCols = append(fkCols, fk.ForeignColumnName)
		}
		row := fks[0]
		tableName := sqlmanager_shared.BuildTable(row.SchemaName, row.TableName)
		constraints[tableName] = append(constraints[tableName], &sqlmanager_shared.ForeignConstraint{
			Columns:     cols,
			NotNullable: notNullable,
			ForeignKey: &sqlmanager_shared.ForeignKey{
				Table:   sqlmanager_shared.BuildTable(row.ForeignSchemaName, row.ForeignTableName),
				Columns: fkCols,
			},
		})
	}

	return constraints, err
}

func (m *MysqlManager) GetPrimaryKeyConstraints(ctx context.Context, schemas []string) ([]*sqlmanager_shared.PrimaryKey, error) {
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
	result := []*sqlmanager_shared.PrimaryKey{}
	for _, row := range output {
		result = append(result, &sqlmanager_shared.PrimaryKey{
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
		tableName := sqlmanager_shared.BuildTable(row.Schema, row.Table)
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
			key = sqlmanager_shared.BuildTable(c.SchemaName, c.TableName)
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
		key := sqlmanager_shared.BuildTable(permission.TableSchema, permission.TableName)
		schemaTablePrivsMap[key] = append(schemaTablePrivsMap[key], permission.PrivilegeType)
	}
	return schemaTablePrivsMap, err
}

// todo
func (m *MysqlManager) GetTableInitStatements(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableInitStatement, error) {
	return nil, errors.ErrUnsupported
}

// todo
func (m *MysqlManager) GetSchemaTableDataTypes(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) (*sqlmanager_shared.SchemaTableDataTypeResponse, error) {
	return nil, errors.ErrUnsupported
}

// todo
func (m *MysqlManager) GetSchemaTableTriggers(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableTrigger, error) {
	return nil, errors.ErrUnsupported
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

func (m *MysqlManager) BatchExec(ctx context.Context, batchSize int, statements []string, opts *sqlmanager_shared.BatchExecOpts) error {
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

func (m *MysqlManager) GetTableRowCount(
	ctx context.Context,
	schema, table string,
	whereClause *string,
) (int64, error) {
	tableName := sqlmanager_shared.BuildTable(schema, table)
	builder := goqu.Dialect(sqlmanager_shared.MysqlDriver)
	sqltable := goqu.I(tableName)

	query := builder.From(sqltable).Select(goqu.COUNT("*"))
	if whereClause != nil && *whereClause != "" {
		query = query.Where(goqu.L(*whereClause))
	}
	sql, _, err := query.ToSQL()
	if err != nil {
		return 0, err
	}
	var count int64
	err = m.pool.QueryRowContext(ctx, sql).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, err
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
