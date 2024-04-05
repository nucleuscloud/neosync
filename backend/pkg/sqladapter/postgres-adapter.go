package sqladapter

import (
	"context"
	"fmt"

	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	dbschemas_postgres "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/postgres"
)

type PostgresAdapter struct {
	querier         pg_queries.Querier
	pool            pg_queries.DBTX
	CloseConnection func()
}

func (p *PostgresAdapter) GetDatabaseSchema(ctx context.Context) ([]*DatabaseSchemaRow, error) {
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

func (p *PostgresAdapter) GetAllForeignKeyConstraints(ctx context.Context, schemas []string) ([]*ForeignKeyConstraintsRow, error) {
	fkConstraints, err := dbschemas_postgres.GetAllPostgresFkConstraints(p.querier, ctx, p.pool, schemas)
	if err != nil {
		return nil, fmt.Errorf("unable to get database foreign keys for postgres connection: %w", err)
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

func (p *PostgresAdapter) GetAllPrimaryKeyConstraints(ctx context.Context, schemas []string) ([]*PrimaryKeyConstraintsRow, error) {
	fkConstraints, err := dbschemas_postgres.GetAllPostgresPkConstraints(p.querier, ctx, p.pool, schemas)
	if err != nil {
		return nil, fmt.Errorf("unable to get database primary keys for postgres connection: %w", err)
	}
	result := []*PrimaryKeyConstraintsRow{}
	for _, row := range fkConstraints {
		result = append(result, &PrimaryKeyConstraintsRow{
			SchemaName:     row.SchemaName,
			TableName:      row.TableName,
			ColumnName:     row.ColumnName,
			ConstraintName: row.ConstraintName,
		})
	}
	return result, nil
}

func (p *PostgresAdapter) Close() error {
	if p.pool != nil && p.CloseConnection != nil {
		p.CloseConnection()
	}
	return nil
}
