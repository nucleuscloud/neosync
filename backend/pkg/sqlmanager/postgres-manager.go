package sqlmanager

import (
	"context"
	"fmt"
	"strings"

	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	dbschemas_postgres "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/postgres"
)

type PostgresManager struct {
	querier pg_queries.Querier
	pool    pg_queries.DBTX
	close   func()
}

func (p *PostgresManager) GetDatabaseSchema(ctx context.Context) ([]*DatabaseSchemaRow, error) {
	dbschemas, err := p.querier.GetDatabaseSchema(ctx, p.pool)
	if err != nil {
		return nil, fmt.Errorf("unable to get database schema for postgres connection: %w", err)
	}
	result := []*DatabaseSchemaRow{}
	for _, row := range dbschemas {
		result = append(result, &DatabaseSchemaRow{
			TableSchema:            row.TableSchema,
			TableName:              row.TableName,
			ColumnName:             row.ColumnName,
			DataType:               row.DataType,
			ColumnDefault:          row.ColumnDefault,
			IsNullable:             row.IsNullable,
			CharacterMaximumLength: row.CharacterMaximumLength,
			NumericPrecision:       row.NumericPrecision,
			NumericScale:           row.NumericScale,
			OrdinalPosition:        row.OrdinalPosition,
		})
	}
	return result, nil
}

func (p *PostgresManager) GetAllForeignKeyConstraints(ctx context.Context, schemas []string) ([]*ForeignKeyConstraintsRow, error) {
	constraints, err := dbschemas_postgres.GetAllPostgresForeignKeyConstraints(ctx, p.pool, p.querier, schemas)
	if err != nil {
		return nil, fmt.Errorf("unable to get database foreign keys for postgres connection: %w", err)
	}
	result := []*ForeignKeyConstraintsRow{}
	for _, row := range constraints {
		if len(row.ConstraintColumns) != len(row.ForeignColumnNames) {
			return nil, fmt.Errorf("length of columns was not equal to length of foreign key cols: %d %d", len(row.ConstraintColumns), len(row.ForeignColumnNames))
		}
		if len(row.ConstraintColumns) != len(row.Notnullable) {
			return nil, fmt.Errorf("length of columns was not equal to length of not nullable cols: %d %d", len(row.ConstraintColumns), len(row.Notnullable))
		}

		for idx, colname := range row.ConstraintColumns {
			fkcol := row.ForeignColumnNames[idx]
			notnullable := row.Notnullable[idx]

			result = append(result, &ForeignKeyConstraintsRow{
				SchemaName:        row.SchemaName,
				TableName:         row.TableName,
				ColumnName:        colname,
				IsNullable:        convertNotNullableToNullableText(notnullable),
				ConstraintName:    row.ConstraintName,
				ForeignSchemaName: row.ForeignSchemaName,
				ForeignTableName:  row.ForeignTableName,
				ForeignColumnName: fkcol,
			})
		}
	}
	return result, nil
}

func convertNotNullableToNullableText(notnullable bool) string {
	if notnullable {
		return "NO"
	}
	return "YES"
}

func (p *PostgresManager) GetPrimaryKeyConstraints(ctx context.Context, schemas []string) ([]*PrimaryKey, error) {
	if len(schemas) == 0 {
		return []*PrimaryKey{}, nil
	}
	rows, err := p.querier.GetTableConstraintsBySchema(ctx, p.pool, schemas)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return []*PrimaryKey{}, nil
	}

	constraints := []*pg_queries.GetTableConstraintsBySchemaRow{}
	for _, row := range rows {
		if row.ConstraintType != "p" {
			continue
		}
		constraints = append(constraints, row)
	}
	result := []*PrimaryKey{}
	for _, row := range constraints {
		result = append(result, &PrimaryKey{
			Schema:  row.SchemaName,
			Table:   row.TableName,
			Columns: row.ConstraintColumns,
		})
	}
	return result, nil
}

func (p *PostgresManager) GetPrimaryKeyConstraintsMap(ctx context.Context, schemas []string) (map[string][]string, error) {
	primaryKeys, err := p.GetPrimaryKeyConstraints(ctx, schemas)
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

func (p *PostgresManager) GetCreateTableStatement(ctx context.Context, schema, table string) (string, error) {
	stmt, err := dbschemas_postgres.GetTableCreateStatement(ctx, p.pool, p.querier, schema, table)
	if err != nil {
		return "", err
	}
	return stmt, nil
}

func (p *PostgresManager) BatchExec(ctx context.Context, batchSize int, statements []string, opts *BatchExecOpts) error {
	for i := 0; i < len(statements); i += batchSize {
		end := i + batchSize
		if end > len(statements) {
			end = len(statements)
		}

		batchCmd := strings.Join(statements[i:end], "\n")
		if opts != nil && opts.Prefix != nil && *opts.Prefix != "" {
			batchCmd = fmt.Sprintf("%s %s", *opts.Prefix, batchCmd)
		}
		_, err := p.pool.Exec(ctx, batchCmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresManager) Exec(ctx context.Context, statement string) error {
	_, err := p.pool.Exec(ctx, statement)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresManager) Close() {
	if p.pool != nil && p.close != nil {
		p.close()
	}
}
