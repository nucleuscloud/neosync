package sqladapter

import (
	"context"
	"fmt"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	dbschemas_mysql "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/mysql"
)

type MysqlAdapter struct {
	querier   mysql_queries.Querier
	pool      mysql_queries.DBTX
	closePool func()
}

func (m *MysqlAdapter) GetDatabaseSchema(ctx context.Context) ([]*DatabaseSchemaRow, error) {
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

func (m *MysqlAdapter) GetAllForeignKeyConstraints(ctx context.Context, schemas []string) ([]*ForeignKeyConstraintsRow, error) {
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

func (m *MysqlAdapter) GetAllPrimaryKeyConstraints(ctx context.Context, schemas []string) ([]*PrimaryKeyConstraintsRow, error) {
	fkConstraints, err := dbschemas_mysql.GetAllMysqlPkConstraints(m.querier, ctx, m.pool, schemas)
	if err != nil {
		return nil, fmt.Errorf("unable to get database foreign keys for mysql connection: %w", err)
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

func (m *MysqlAdapter) ClosePool() {
	if m.pool != nil && m.closePool != nil {
		m.closePool()
	}
}
