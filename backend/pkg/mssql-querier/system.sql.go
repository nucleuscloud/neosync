package mssql_queries

import (
	"context"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
)

const getRolePermissions = `--- name: GetRolePermissions :many
WITH object_list AS (
    SELECT
        s.name COLLATE database_default AS table_schema,
        o.name COLLATE database_default AS table_name
    FROM
        sys.objects o
    JOIN
        sys.schemas s ON o.schema_id = s.schema_id
    WHERE
        o.type IN ('U', 'V') -- Tables and Views
        AND s.name NOT IN ('sys', 'INFORMATION_SCHEMA', 'db_owner', 'db_accessadmin', 'db_securityadmin', 'db_ddladmin', 'db_backupoperator', 'db_datareader', 'db_datawriter', 'db_denydatareader', 'db_denydatawriter')
),
effective_permissions AS (
    SELECT
        ol.table_schema,
        ol.table_name,
        p.permission_name COLLATE database_default AS privilege_type,
        'Effective' AS grant_type
    FROM
        object_list ol
    CROSS APPLY
        sys.fn_my_permissions(QUOTENAME(ol.table_schema) + '.' + QUOTENAME(ol.table_name), 'OBJECT') p
),
explicit_permissions AS (
    SELECT
        s.name COLLATE database_default AS table_schema,
        o.name COLLATE database_default AS table_name,
        dp.permission_name COLLATE database_default AS privilege_type,
        'Explicit' AS grant_type
    FROM
        sys.database_permissions dp
    JOIN
        sys.objects o ON dp.major_id = o.object_id
    JOIN
        sys.schemas s ON o.schema_id = s.schema_id
    WHERE
        dp.grantee_principal_id = DATABASE_PRINCIPAL_ID()
        AND o.type IN ('U', 'V') -- Tables and Views
        AND s.name NOT IN ('sys', 'INFORMATION_SCHEMA', 'db_owner', 'db_accessadmin', 'db_securityadmin', 'db_ddladmin', 'db_backupoperator', 'db_datareader', 'db_datawriter', 'db_denydatareader', 'db_denydatawriter')
)
SELECT * FROM effective_permissions
UNION
SELECT * FROM explicit_permissions
ORDER BY
    table_schema,
    table_name,
    privilege_type;
`

type GetRolePermissionsRow struct {
	TableSchema   string
	TableName     string
	PrivilegeType string
	GrantType     string
}

func (q *Queries) GetRolePermissions(ctx context.Context, db mysql_queries.DBTX) ([]*GetRolePermissionsRow, error) {
	rows, err := db.QueryContext(ctx, getRolePermissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetRolePermissionsRow
	for rows.Next() {
		var i GetRolePermissionsRow
		if err := rows.Scan(&i.TableSchema, &i.TableName, &i.PrivilegeType, &i.GrantType); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
