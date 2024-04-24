package sqlmanager

import (
	"context"
	"fmt"

	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	dbschemas_postgres "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/postgres"
)

type PostgresManager struct {
	querier   pg_queries.Querier
	pool      pg_queries.DBTX
	closePool func()
}

func (p *PostgresManager) GetDatabaseSchema(ctx context.Context) ([]*DatabaseSchemaRow, error) {
	dbschemas, err := p.querier.GetDatabaseSchema(ctx, p.pool)
	if err != nil {
		return nil, fmt.Errorf("unable to get database schema for postgres connection: %w", err)
	}
	result := []*DatabaseSchemaRow{}
	for _, row := range dbschemas {
		var colDefault string
		if row.ColumnDefault != nil {
			val, ok := row.ColumnDefault.(string)
			if ok {
				colDefault = val
			}
		}
		result = append(result, &DatabaseSchemaRow{
			TableSchema:            row.TableSchema,
			TableName:              row.TableName,
			ColumnName:             row.ColumnName,
			DataType:               row.DataType,
			ColumnDefault:          colDefault,
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

func (p *PostgresManager) GetAllPrimaryKeyConstraints(ctx context.Context, schemas []string) ([]*PrimaryKeyConstraintsRow, error) {
	constraints, err := dbschemas_postgres.GetAllPostgresPrimaryKeyConstraints(ctx, p.pool, p.querier, schemas)
	if err != nil {
		return nil, fmt.Errorf("unable to get database primary keys for postgres connection: %w", err)
	}
	result := []*PrimaryKeyConstraintsRow{}
	for _, row := range constraints {
		result = append(result, &PrimaryKeyConstraintsRow{
			SchemaName:     row.SchemaName,
			TableName:      row.TableName,
			ColumnName:     row.ConstraintColumns[0], // todo: hack, this should be fixed to support primary keys
			ConstraintName: row.ConstraintName,
		})
	}
	return result, nil
}
func (p *PostgresManager) ClosePool() {
	if p.pool != nil && p.closePool != nil {
		p.closePool()
	}
}
