package mssql_queries

import (
	"context"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
)

type Querier interface {
	GetRolePermissions(ctx context.Context, db mysql_queries.DBTX) ([]*GetRolePermissionsRow, error)
}

var _ Querier = (*Queries)(nil)
