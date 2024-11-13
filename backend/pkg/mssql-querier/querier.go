package mssql_queries

import (
	"context"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
)

type Querier interface {
	GetCustomSequencesBySchemas(ctx context.Context, db mysql_queries.DBTX, schemas []string) ([]*GetCustomSequencesBySchemasRow, error)
	GetCustomTriggersBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetCustomTriggersBySchemasAndTablesRow, error)
	GetDataTypesBySchemas(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetDataTypesBySchemasRow, error)
	GetDatabaseSchema(ctx context.Context, db mysql_queries.DBTX) ([]*GetDatabaseSchemaRow, error)
	GetDatabaseTableSchemasBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetDatabaseTableSchemasBySchemasAndTablesRow, error)
	GetIndicesBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetIndicesBySchemasAndTablesRow, error)
	GetRolePermissions(ctx context.Context, db mysql_queries.DBTX) ([]*GetRolePermissionsRow, error)
	GetTableConstraintsBySchemas(ctx context.Context, db mysql_queries.DBTX, schemas []string) ([]*GetTableConstraintsBySchemasRow, error)
	GetViewsAndFunctionsBySchemas(ctx context.Context, db mysql_queries.DBTX, schemas []string) ([]*GetViewsAndFunctionsBySchemasRow, error)
}

var _ Querier = (*Queries)(nil)
