package mssql_queries

import (
	"context"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
)

type Querier interface {
	GetCustomFunctionsBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetCustomFunctionsBySchemasAndTablesRow, error)
	GetCustomSequencesBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetCustomSequencesBySchemasAndTablesRow, error)
	GetCustomTriggersBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetCustomTriggersBySchemasAndTablesRow, error)
	GetDataTypesBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetDataTypesBySchemasAndTablesRow, error)
	GetDatabaseSchema(ctx context.Context, db mysql_queries.DBTX) ([]*GetDatabaseSchemaRow, error)
	GetDatabaseTableSchemasBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetDatabaseSchemaRow, error)
	GetIndicesBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetIndicesBySchemasAndTablesRow, error)
	GetRolePermissions(ctx context.Context, db mysql_queries.DBTX) ([]*GetRolePermissionsRow, error)
	GetTableConstraintsBySchemas(ctx context.Context, db mysql_queries.DBTX, schemas []string) ([]*GetTableConstraintsBySchemasRow, error)
}

var _ Querier = (*Queries)(nil)
