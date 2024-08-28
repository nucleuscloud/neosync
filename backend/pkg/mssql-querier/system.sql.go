package mssql_queries

import (
	"context"
	"database/sql"
	"strings"

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
        'SELECT' AS privilege_type,
        HAS_PERMS_BY_NAME(QUOTENAME(ol.table_schema) + '.' + QUOTENAME(ol.table_name), 'OBJECT', 'SELECT') AS perm_state
    FROM object_list ol
    UNION ALL
    SELECT
        ol.table_schema,
        ol.table_name,
        'INSERT' AS privilege_type,
        HAS_PERMS_BY_NAME(QUOTENAME(ol.table_schema) + '.' + QUOTENAME(ol.table_name), 'OBJECT', 'INSERT') AS perm_state
    FROM object_list ol
    UNION ALL
    SELECT
        ol.table_schema,
        ol.table_name,
        'UPDATE' AS privilege_type,
        HAS_PERMS_BY_NAME(QUOTENAME(ol.table_schema) + '.' + QUOTENAME(ol.table_name), 'OBJECT', 'UPDATE') AS perm_state
    FROM object_list ol
    UNION ALL
    SELECT
        ol.table_schema,
        ol.table_name,
        'DELETE' AS privilege_type,
        HAS_PERMS_BY_NAME(QUOTENAME(ol.table_schema) + '.' + QUOTENAME(ol.table_name), 'OBJECT', 'DELETE') AS perm_state
    FROM object_list ol
),
explicit_permissions AS (
    SELECT
        s.name COLLATE database_default AS table_schema,
        o.name COLLATE database_default AS table_name,
        dp.permission_name COLLATE database_default AS privilege_type
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
SELECT table_schema, table_name, privilege_type FROM effective_permissions WHERE perm_state = 1
UNION
SELECT table_schema, table_name, privilege_type FROM explicit_permissions
ORDER BY
    table_name,
    table_schema,
    privilege_type;
`

type GetRolePermissionsRow struct {
	TableSchema   string
	TableName     string
	PrivilegeType string
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
		if err := rows.Scan(&i.TableSchema, &i.TableName, &i.PrivilegeType); err != nil {
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

const getTableConstraintsBySchemas = `
WITH ConstraintColumns AS (
    SELECT
        kc.parent_object_id,
        kc.object_id AS constraint_object_id,
        STRING_AGG(c.name, ', ') WITHIN GROUP (ORDER BY ic.key_ordinal) AS columns,
        STRING_AGG(CASE WHEN c.is_nullable = 1 THEN 'NULL' ELSE 'NOT NULL' END, ', ')
            WITHIN GROUP (ORDER BY ic.key_ordinal) AS nullability
    FROM sys.key_constraints kc
    JOIN sys.index_columns ic ON kc.parent_object_id = ic.object_id AND kc.unique_index_id = ic.index_id
    JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
    GROUP BY kc.parent_object_id, kc.object_id

    UNION ALL

    SELECT
        fkc.parent_object_id,
        fkc.constraint_object_id,
        STRING_AGG(c.name, ', ') WITHIN GROUP (ORDER BY fkc.constraint_column_id) AS columns,
        STRING_AGG(CASE WHEN c.is_nullable = 1 THEN 'NULL' ELSE 'NOT NULL' END, ', ')
            WITHIN GROUP (ORDER BY fkc.constraint_column_id) AS nullability
    FROM sys.foreign_key_columns fkc
    JOIN sys.columns c ON fkc.parent_object_id = c.object_id AND fkc.parent_column_id = c.column_id
    GROUP BY fkc.parent_object_id, fkc.constraint_object_id

    UNION ALL

    SELECT
        cc.parent_object_id,
        cc.object_id,
        STUFF((
            SELECT ', ' + c.name
            FROM sys.columns c
            WHERE c.object_id = cc.parent_object_id
              AND CHARINDEX(QUOTENAME(c.name), cc.definition) > 0
            FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'), 1, 2, '') AS columns,
        STUFF((
            SELECT ', ' + CASE WHEN c.is_nullable = 1 THEN 'NULL' ELSE 'NOT NULL' END
            FROM sys.columns c
            WHERE c.object_id = cc.parent_object_id
              AND CHARINDEX(QUOTENAME(c.name), cc.definition) > 0
            FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'), 1, 2, '') AS nullability
    FROM sys.check_constraints cc
)
SELECT
    s.name AS schema_name,
    t.name AS table_name,
    o.name AS constraint_name,
    CASE
        WHEN o.type = 'PK' THEN 'PRIMARY KEY'
        WHEN o.type = 'UQ' THEN 'UNIQUE'
        WHEN o.type = 'F' THEN 'FOREIGN KEY'
        WHEN o.type = 'C' THEN 'CHECK'
    END AS constraint_type,
    cc.columns AS constraint_columns,
    cc.nullability AS constraint_columns_nullability,
    CASE WHEN o.type = 'F'
        THEN OBJECT_SCHEMA_NAME(fk.referenced_object_id) + '.' + OBJECT_NAME(fk.referenced_object_id)
        ELSE NULL
    END AS referenced_table,
    CASE WHEN o.type = 'F'
        THEN (SELECT STRING_AGG(c.name, ', ') WITHIN GROUP (ORDER BY fc.constraint_column_id)
              FROM sys.foreign_key_columns fc
              JOIN sys.columns c ON fc.referenced_object_id = c.object_id AND fc.referenced_column_id = c.column_id
              WHERE fc.constraint_object_id = o.object_id)
        ELSE NULL
    END AS referenced_columns,
    CASE WHEN o.type = 'F'
        THEN 'ON UPDATE ' + UPPER(fk.update_referential_action_desc) + ', ON DELETE ' + UPPER(fk.delete_referential_action_desc)
        ELSE NULL
    END AS fk_actions,
    CASE WHEN o.type = 'C' THEN cc_def.definition ELSE NULL END AS check_clause
FROM
    sys.objects o
JOIN
    sys.tables t ON o.parent_object_id = t.object_id
JOIN
    sys.schemas s ON t.schema_id = s.schema_id
LEFT JOIN
    ConstraintColumns cc ON o.object_id = cc.constraint_object_id
LEFT JOIN
    sys.foreign_keys fk ON o.object_id = fk.object_id
LEFT JOIN
    sys.check_constraints cc_def ON o.object_id = cc_def.object_id
WHERE
    o.type IN ('PK', 'UQ', 'F', 'C')
    AND s.name IN (SELECT value FROM STRING_SPLIT(@schemas, ','))
ORDER BY
    s.name, t.name, o.name;
`

type GetTableConstraintsBySchemasRow struct {
	SchemaName                   string
	TableName                    string
	ConstraintName               string
	ConstraintType               string
	ConstraintColumns            string
	ConstraintColumnsNullability string
	ReferencedTable              sql.NullString
	ReferencedColumns            sql.NullString
	FKActions                    sql.NullString
	CheckClause                  sql.NullString
}

func (q *Queries) GetTableConstraintsBySchemas(ctx context.Context, db mysql_queries.DBTX, schemas []string) ([]*GetTableConstraintsBySchemasRow, error) {
	// Join schemas into a comma-separated string
	schemaList := strings.Join(schemas, ",")

	// Execute the query with the schema list as a parameter
	rows, err := db.QueryContext(ctx, getTableConstraintsBySchemas, sql.Named("schemas", schemaList))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*GetTableConstraintsBySchemasRow
	for rows.Next() {
		var i GetTableConstraintsBySchemasRow
		if err := rows.Scan(
			&i.SchemaName,
			&i.TableName,
			&i.ConstraintName,
			&i.ConstraintType,
			&i.ConstraintColumns,
			&i.ConstraintColumnsNullability,
			&i.ReferencedTable,
			&i.ReferencedColumns,
			&i.FKActions,
			&i.CheckClause,
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
