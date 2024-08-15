package mssql_queries

import (
	"context"
	"database/sql"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
)

const getDatabaseSchema = `-- name: GetDatabaseSchema :many
SELECT
    s.name AS table_schema,
    t.name AS table_name,
    c.name AS column_name,
    c.column_id AS ordinal_position,
    ISNULL(dc.definition, '') AS column_default,
    CASE WHEN c.is_nullable = 1 THEN 'YES' ELSE 'NO' END AS is_nullable,
    tp.name AS data_type,
    CASE WHEN tp.name IN ('nchar', 'nvarchar') AND c.max_length != -1 THEN c.max_length / 2
         WHEN tp.name IN ('char', 'varchar') AND c.max_length != -1 THEN c.max_length
         ELSE NULL
    END AS character_maximum_length,
    c.precision AS numeric_precision,
    c.scale AS numeric_scale,
    c.is_identity,
    c.is_computed,
    CASE
        WHEN c.is_computed = 1 THEN cc.definition
        ELSE NULL
    END AS generation_expression
FROM
    sys.schemas s
    INNER JOIN sys.tables t ON s.schema_id = t.schema_id
    INNER JOIN sys.columns c ON t.object_id = c.object_id
    INNER JOIN sys.types tp ON c.user_type_id = tp.user_type_id
    LEFT JOIN sys.default_constraints dc ON c.default_object_id = dc.object_id
    LEFT JOIN sys.computed_columns cc ON c.object_id = cc.object_id AND c.column_id = cc.column_id
WHERE
    s.name NOT IN ('sys', 'INFORMATION_SCHEMA', 'db_owner', 'db_accessadmin', 'db_securityadmin', 'db_ddladmin', 'db_backupoperator', 'db_datareader', 'db_datawriter', 'db_denydatareader', 'db_denydatawriter')
    AND t.type = 'U'
ORDER BY
    s.name, t.name, c.column_id;
`

type GetDatabaseSchemaRow struct {
	TableSchema            string
	TableName              string
	ColumnName             string
	OrdinalPosition        int32
	ColumnDefault          string
	IsNullable             string
	DataType               string
	CharacterMaximumLength sql.NullInt32
	NumericPrecision       sql.NullInt16
	NumericScale           sql.NullInt16
	IsIdentity             bool
	IsComputed             bool
	GenerationExpression   sql.NullString
}

func (q *Queries) GetDatabaseSchema(ctx context.Context, db mysql_queries.DBTX) ([]*GetDatabaseSchemaRow, error) {
	rows, err := db.QueryContext(ctx, getDatabaseSchema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetDatabaseSchemaRow
	for rows.Next() {
		var i GetDatabaseSchemaRow
		if err := rows.Scan(
			&i.TableSchema,
			&i.TableName,
			&i.ColumnName,
			&i.OrdinalPosition,
			&i.ColumnDefault,
			&i.IsNullable,
			&i.DataType,
			&i.CharacterMaximumLength,
			&i.NumericPrecision,
			&i.NumericScale,
			&i.IsIdentity,
			&i.IsComputed,
			&i.GenerationExpression,
		); err != nil {
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
