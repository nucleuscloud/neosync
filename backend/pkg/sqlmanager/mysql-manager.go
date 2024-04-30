package sqlmanager

import (
	"context"
	"fmt"
	"strings"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	dbschemas_mysql "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/mysql"
	"golang.org/x/sync/errgroup"
)

type MySqlManager struct {
	querier mysql_queries.Querier
	pool    mysql_queries.DBTX
	close   func()
}

func (m *MySqlManager) GetDatabaseSchema(ctx context.Context) ([]*DatabaseSchemaRow, error) {
	dbschemas, err := m.querier.GetDatabaseSchema(ctx, m.pool)
	if err != nil {
		return nil, fmt.Errorf("unable to get database schema for postgres connection: %w", err)
	}
	result := []*DatabaseSchemaRow{}
	for _, row := range dbschemas {
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

func (m *MySqlManager) GetAllForeignKeyConstraints(ctx context.Context, schemas []string) ([]*ForeignKeyConstraintsRow, error) {
	fkConstraints, err := dbschemas_mysql.GetAllMysqlFkConstraints(m.querier, ctx, m.pool, schemas)
	if err != nil {
		return nil, fmt.Errorf("unable to get database foreign keys for mysql connection: %w", err)
	}
	result := []*ForeignKeyConstraintsRow{}
	for _, row := range fkConstraints {
		result = append(result, &ForeignKeyConstraintsRow{
			SchemaName:        row.SchemaName,
			TableName:         row.TableName,
			ColumnName:        row.ColumnName,
			IsNullable:        row.IsNullable,
			ConstraintName:    row.ConstraintName,
			ForeignSchemaName: row.ForeignSchemaName,
			ForeignTableName:  row.ForeignTableName,
			ForeignColumnName: row.ForeignColumnName,
		})
	}
	return result, nil
}

func (m *MySqlManager) GetPrimaryKeyConstraints(ctx context.Context, schemas []string) ([]*PrimaryKey, error) {
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

func (m *MySqlManager) GetPrimaryKeyConstraintsMap(ctx context.Context, schemas []string) (map[string][]string, error) {
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

func (m *MySqlManager) GetCreateTableStatement(ctx context.Context, schema, table string) (string, error) {
	stmt, err := dbschemas_mysql.GetTableCreateStatement(ctx, m.pool, &dbschemas_mysql.GetTableCreateStatementRequest{
		Schema: schema,
		Table:  table,
	})
	if err != nil {
		return "", err
	}
	return stmt, nil
}

func (m *MySqlManager) BatchExec(ctx context.Context, batchSize int, statements []string, opts *BatchExecOpts) error {
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

func (m *MySqlManager) Exec(ctx context.Context, statement string) error {
	_, err := m.pool.ExecContext(ctx, statement)
	if err != nil {
		return err
	}
	return nil
}

func (m *MySqlManager) Close() {
	if m.pool != nil && m.close != nil {
		m.close()
	}
}
