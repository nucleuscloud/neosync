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

const getDatabaseTableSchemasBySchemasAndTables = `-- name: getDatabaseTableSchemasBySchemasAndTables :many
SELECT
    s.name AS table_schema,
    t.name AS table_name,
    c.name AS column_name,
    c.column_id AS ordinal_position,
	dc.definition as column_default,
    c.is_nullable,
    tp.name AS data_type,
    CASE WHEN tp.name IN ('nchar', 'nvarchar') AND c.max_length != -1 THEN c.max_length / 2
         WHEN tp.name IN ('char', 'varchar') AND c.max_length != -1 THEN c.max_length
         ELSE NULL
    END AS character_maximum_length,
    c.precision AS numeric_precision,
    c.scale AS numeric_scale,
    c.is_identity,
    CASE 
        WHEN c.is_identity = 1 THEN CAST(IDENT_SEED(s.name + '.' + t.name) AS VARCHAR(50))
        ELSE NULL 
    END AS identity_seed,
    CASE 
        WHEN c.is_identity = 1 THEN CAST(IDENT_INCR(s.name + '.' + t.name) AS VARCHAR(50))
        ELSE NULL 
    END AS identity_increment,
    c.is_computed,
    cc.definition as generation_expression,
    cc.is_persisted,
    CASE 
        WHEN c.generated_always_type = 1 THEN 'GENERATED ALWAYS AS ROW START'
        WHEN c.generated_always_type = 2 THEN 'GENERATED ALWAYS AS ROW END'
        WHEN c.generated_always_type = 5 THEN 'GENERATED ALWAYS AS TRANSACTION_ID_START'
        WHEN c.generated_always_type = 6 THEN 'GENERATED ALWAYS AS TRANSACTION_ID_END'
        WHEN c.generated_always_type = 7 THEN 'GENERATED ALWAYS AS SEQUENCE_NUMBER_START'
        WHEN c.generated_always_type = 8 THEN 'GENERATED ALWAYS AS SEQUENCE_NUMBER_END'
        ELSE NULL
    END AS generated_always_type
FROM
    sys.schemas s
    INNER JOIN sys.tables t ON s.schema_id = t.schema_id
    INNER JOIN sys.columns c ON t.object_id = c.object_id
    INNER JOIN sys.types tp ON c.user_type_id = tp.user_type_id
    LEFT JOIN sys.default_constraints dc ON c.default_object_id = dc.object_id
    LEFT JOIN sys.computed_columns cc ON c.object_id = cc.object_id AND c.column_id = cc.column_id
    LEFT JOIN sys.periods p ON t.object_id = p.object_id
WHERE t.type = 'U' AND CONCAT(s.name, '.', t.name) IN (SELECT value FROM STRING_SPLIT(@schematables, ','))
ORDER BY
    s.name, t.name, c.column_id;
`

type GetDatabaseTableSchemasBySchemasAndTablesRow struct {
	TableSchema            string
	TableName              string
	ColumnName             string
	OrdinalPosition        int32
	ColumnDefault          sql.NullString
	IsNullable             bool
	DataType               string
	CharacterMaximumLength sql.NullInt32
	NumericPrecision       sql.NullInt16
	NumericScale           sql.NullInt16
	IsIdentity             bool
	IsComputed             bool
	IsPersisted            bool
	GenerationExpression   sql.NullString
	GeneratedAlwaysType    sql.NullString
	IdentitySeed           sql.NullInt32
	IdentityIncrement      sql.NullInt32
}

func (q *Queries) GetDatabaseTableSchemasBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetDatabaseTableSchemasBySchemasAndTablesRow, error) {
	// Join schemas into a comma-separated string
	schematablesList := strings.Join(schematables, ",")

	rows, err := db.QueryContext(ctx, getDatabaseTableSchemasBySchemasAndTables, sql.Named("schematables", schematablesList))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetDatabaseTableSchemasBySchemasAndTablesRow
	for rows.Next() {
		var i GetDatabaseTableSchemasBySchemasAndTablesRow
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
    table_schema,
    table_name,
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

const getIndicesBySchemasAndTable = `--- name: GetIndicesBySchemaAndTables :many
SELECT 
    SCHEMA_NAME(t.schema_id) AS schema_name,
    t.name AS table_name,
    i.name AS index_name,
    SUBSTRING(
        (
            SELECT CASE 
                -- Clustered index
                WHEN i.type = 1 THEN 'CREATE CLUSTERED INDEX ' + QUOTENAME(i.name) + ' ON ' + QUOTENAME(SCHEMA_NAME(t.schema_id)) + '.' + QUOTENAME(t.name) + ' ('
                -- Nonclustered index
                WHEN i.type = 2 THEN 'CREATE NONCLUSTERED INDEX ' + QUOTENAME(i.name) + ' ON ' + QUOTENAME(SCHEMA_NAME(t.schema_id)) + '.' + QUOTENAME(t.name) + ' ('
                -- XML index
                WHEN i.type = 3 THEN 'CREATE XML INDEX ' + QUOTENAME(i.name) + ' ON ' + QUOTENAME(SCHEMA_NAME(t.schema_id)) + '.' + QUOTENAME(t.name)
                -- Primary XML index
                WHEN i.type = 4 THEN 'CREATE PRIMARY XML INDEX ' + QUOTENAME(i.name) + ' ON ' + QUOTENAME(SCHEMA_NAME(t.schema_id)) + '.' + QUOTENAME(t.name)
                -- Columnstore index
                WHEN i.type = 5 THEN 'CREATE CLUSTERED COLUMNSTORE INDEX ' + QUOTENAME(i.name) + ' ON ' + QUOTENAME(SCHEMA_NAME(t.schema_id)) + '.' + QUOTENAME(t.name)
                -- Nonclustered columnstore index
                WHEN i.type = 6 THEN 'CREATE NONCLUSTERED COLUMNSTORE INDEX ' + QUOTENAME(i.name) + ' ON ' + QUOTENAME(SCHEMA_NAME(t.schema_id)) + '.' + QUOTENAME(t.name) + ' ('
            END +
            -- Key columns
            CASE WHEN i.type IN (1,2) THEN
                STUFF((
                    SELECT ', ' + QUOTENAME(c.name) + 
                        CASE WHEN ic.is_descending_key = 1 
                            THEN ' DESC' 
                            ELSE ' ASC' 
                        END
                    FROM sys.index_columns ic
                    JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
                    WHERE ic.object_id = i.object_id 
                        AND ic.index_id = i.index_id
                        AND ic.is_included_column = 0
                    ORDER BY ic.key_ordinal
                    FOR XML PATH('')
                ), 1, 2, '')
            WHEN i.type = 6 THEN  -- For columnstore
                STUFF((
                    SELECT ', ' + QUOTENAME(c.name)
                    FROM sys.index_columns ic
                    JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
                    WHERE ic.object_id = i.object_id 
                        AND ic.index_id = i.index_id
                    ORDER BY ic.index_column_id
                    FOR XML PATH('')
                ), 1, 2, '')
            ELSE ''
            END + ')' +
            -- Included columns
            CASE WHEN EXISTS (
                SELECT 1
                FROM sys.index_columns ic2
                WHERE ic2.object_id = i.object_id 
                    AND ic2.index_id = i.index_id
                    AND ic2.is_included_column = 1
            ) THEN 
                ' INCLUDE (' + 
                STUFF((
                    SELECT ', ' + QUOTENAME(c.name)
                    FROM sys.index_columns ic
                    JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
                    WHERE ic.object_id = i.object_id 
                        AND ic.index_id = i.index_id
                        AND ic.is_included_column = 1
                    ORDER BY c.name
                    FOR XML PATH('')
                ), 1, 2, '') + ')'
            ELSE ''
            END +
            -- Where clause for filtered indexes
            CASE WHEN i.has_filter = 1 
                THEN ' WHERE ' + i.filter_definition
                ELSE ''
            END +
            -- Index options
            CASE WHEN i.fill_factor <> 0 OR i.is_padded = 1
                THEN ' WITH ('
                    + CASE WHEN i.fill_factor <> 0 
                        THEN 'FILLFACTOR = ' + CAST(i.fill_factor AS varchar(3))
                        ELSE ''
                    END
                    + CASE WHEN i.fill_factor <> 0 AND i.is_padded = 1 THEN ', ' ELSE '' END
                    + CASE WHEN i.is_padded = 1 
                        THEN 'PAD_INDEX = ON'
                        ELSE ''
                    END
                    + ')'
                ELSE ''
            END
        ), 1, 8000) AS index_definition 
FROM sys.indexes i
INNER JOIN sys.tables t ON i.object_id = t.object_id
WHERE i.is_primary_key =0 AND i.type > 0 AND is_unique_constraint = 0
    AND CONCAT(SCHEMA_NAME(t.schema_id), '.', t.name) IN (SELECT value FROM STRING_SPLIT(@schematables, ','))
ORDER BY i.index_id;
`

type GetIndicesBySchemasAndTablesRow struct {
	SchemaName      string
	TableName       string
	IndexName       string
	IndexDefinition string
}

func (q *Queries) GetIndicesBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetIndicesBySchemasAndTablesRow, error) {
	// Join schemas into a comma-separated string
	schematablesList := strings.Join(schematables, ",")

	rows, err := db.QueryContext(ctx, getIndicesBySchemasAndTable, sql.Named("schematables", schematablesList))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetIndicesBySchemasAndTablesRow
	for rows.Next() {
		var i GetIndicesBySchemasAndTablesRow
		if err := rows.Scan(
			&i.SchemaName,
			&i.TableName,
			&i.IndexName,
			&i.IndexDefinition,
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

const getCustomFunctionsBySchemasAndTables = `-- name: GetCustomFunctionsBySchemasAndTables :many
SELECT
   SCHEMA_NAME(o.schema_id) as schema_name,
   o.name AS function_name,
   sm.definition
FROM sys.sql_modules AS sm
LEFT JOIN sys.objects AS o ON sm.object_id = o.object_id
WHERE o.type != 'TR' AND SCHEMA_NAME(o.schema_id) IN (SELECT value FROM STRING_SPLIT(@schemas, ','))
ORDER BY o.type;
`

type GetCustomFunctionsBySchemasRow struct {
	SchemaName   string
	FunctionName string
	Definition   string
}

func (q *Queries) GetCustomFunctionsBySchemas(ctx context.Context, db mysql_queries.DBTX, schemas []string) ([]*GetCustomFunctionsBySchemasRow, error) {
	// Join schemas into a comma-separated string
	schemasList := strings.Join(schemas, ",")
	rows, err := db.QueryContext(ctx, getCustomFunctionsBySchemasAndTables, sql.Named("schemas", schemasList))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetCustomFunctionsBySchemasRow
	for rows.Next() {
		var i GetCustomFunctionsBySchemasRow
		if err := rows.Scan(
			&i.SchemaName,
			&i.FunctionName,
			&i.Definition,
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

const getCustomSequencesBySchemas = `-- name: GetCustomSequencesBySchemasAndTables :many
SELECT 
    SCHEMA_NAME(seq.schema_id) AS schema_name,
    seq.name AS sequence_name,
    -- Build CREATE SEQUENCE statement with proper CASTing
    CONCAT(
        'CREATE SEQUENCE ', QUOTENAME(SCHEMA_NAME(seq.schema_id)), '.', QUOTENAME(seq.name), 
        ' AS ', TYPE_NAME(seq.system_type_id),
        ' START WITH ', CAST(CAST(seq.start_value AS bigint) AS varchar(20)),
        ' INCREMENT BY ', CAST(CAST(seq.increment AS bigint) AS varchar(20)),
        ' MINVALUE ', CAST(CAST(seq.minimum_value AS bigint) AS varchar(20)),
        ' MAXVALUE ', CAST(CAST(seq.maximum_value AS bigint) AS varchar(20)),
        CASE 
            WHEN seq.is_cycling = 1 THEN ' CYCLE' 
            ELSE ' NO CYCLE'
        END,
        CASE 
            WHEN seq.is_cached = 1 THEN ' CACHE ' + CAST(seq.cache_size AS varchar(20))
            ELSE ' NO CACHE'
        END,
        ';'
    ) AS definition
FROM sys.sequences seq
WHERE SCHEMA_NAME(seq.schema_id) IN (SELECT value FROM STRING_SPLIT(@schemas, ','))
ORDER BY seq.schema_id, seq.name;
`

type GetCustomSequencesBySchemasRow struct {
	SchemaName   string
	SequenceName string
	Definition   string
}

func (q *Queries) GetCustomSequencesBySchemas(ctx context.Context, db mysql_queries.DBTX, schemas []string) ([]*GetCustomSequencesBySchemasRow, error) {
	// Join schemas into a comma-separated string
	schemasList := strings.Join(schemas, ",")
	rows, err := db.QueryContext(ctx, getCustomSequencesBySchemas, sql.Named("schemas", schemasList))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetCustomSequencesBySchemasRow
	for rows.Next() {
		var i GetCustomSequencesBySchemasRow
		if err := rows.Scan(
			&i.SchemaName,
			&i.SequenceName,
			&i.Definition,
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

const getCustomTriggersBySchemasAndTables = `-- name: GetCustomTriggersBySchemasAndTables :many
SELECT
   SCHEMA_NAME(o.schema_id) as schema_name,
   oo.name AS table_name,
   o.name AS trigger_name,
   sm.definition
FROM sys.sql_modules AS sm
LEFT JOIN sys.objects AS o ON sm.object_id = o.object_id
LEFT join sys.objects as oo on o.parent_object_id = oo.object_id
WHERE o.type = 'TR' AND CONCAT(SCHEMA_NAME(o.schema_id), '.', oo.name) IN (SELECT value FROM STRING_SPLIT(@schematables, ','))
ORDER BY o.type;
`

type GetCustomTriggersBySchemasAndTablesRow struct {
	SchemaName  string
	TableName   string
	TriggerName string
	Definition  string
}

func (q *Queries) GetCustomTriggersBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetCustomTriggersBySchemasAndTablesRow, error) {
	// Join schemas into a comma-separated string
	schematablesList := strings.Join(schematables, ",")
	rows, err := db.QueryContext(ctx, getCustomTriggersBySchemasAndTables, sql.Named("schematables", schematablesList))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetCustomTriggersBySchemasAndTablesRow
	for rows.Next() {
		var i GetCustomTriggersBySchemasAndTablesRow
		if err := rows.Scan(
			&i.SchemaName,
			&i.TableName,
			&i.TriggerName,
			&i.Definition,
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

const getDataTypesBySchemasAndTables = `-- name: GetDataTypesBySchemasAndTables :many
SELECT 
    SCHEMA_NAME(t.schema_id) AS schema_name,
    t.name AS type_name,
    'domain' as type,
    'CREATE TYPE [' + SCHEMA_NAME(t.schema_id) + '].[' + t.name + '] FROM ' + 
    typ.name + 
    CASE 
        WHEN typ.name IN ('varchar', 'nvarchar', 'char', 'nchar') 
            THEN '(' + CASE WHEN t.max_length = -1 THEN 'MAX' 
                           ELSE CAST(CASE WHEN typ.name LIKE 'n%' 
                                         THEN t.max_length/2 
                                         ELSE t.max_length END AS VARCHAR(10)) 
                      END + ')'
        WHEN typ.name IN ('decimal', 'numeric') 
            THEN '(' + CAST(t.[precision] AS VARCHAR(10)) + ',' + CAST(t.scale AS VARCHAR(10)) + ')'
        ELSE ''
    END + 
    ' ' + CASE WHEN t.is_nullable = 1 THEN 'NULL' ELSE 'NOT NULL' END + ';' AS definition
FROM sys.types t
JOIN sys.types typ ON t.system_type_id = typ.system_type_id
    AND typ.system_type_id = typ.user_type_id
WHERE t.is_user_defined = 1
AND t.is_table_type = 0

UNION ALL

SELECT 
    SCHEMA_NAME(tt.schema_id) AS schema_name,
    tt.name AS type_name,
    'composite' as type,
    'CREATE TYPE [' + SCHEMA_NAME(tt.schema_id) + '].[' + tt.name + '] AS TABLE (' + 
    STUFF((
        SELECT ', ' + c.name + ' ' + 
            CASE 
                WHEN typ.name IN ('varchar', 'nvarchar', 'char', 'nchar') 
                    THEN typ.name + '(' + CASE WHEN c.max_length = -1 THEN 'MAX' 
                                               ELSE CAST(CASE WHEN typ.name LIKE 'n%' 
                                                             THEN c.max_length/2 
                                                             ELSE c.max_length END AS VARCHAR(10)) 
                                          END + ')'
                WHEN typ.name IN ('decimal', 'numeric') 
                    THEN typ.name + '(' + CAST(c.[precision] AS VARCHAR(10)) + ',' + CAST(c.scale AS VARCHAR(10)) + ')'
                ELSE typ.name
            END + 
            CASE WHEN c.is_nullable = 1 THEN ' NULL' ELSE ' NOT NULL' END
        FROM sys.columns c
        JOIN sys.types typ ON c.system_type_id = typ.system_type_id
            AND typ.system_type_id = typ.user_type_id
        WHERE c.object_id = tt.type_table_object_id
        ORDER BY c.column_id
        FOR XML PATH('')
    ), 1, 2, '') + ');' AS definition
FROM sys.table_types tt
WHERE tt.is_user_defined = 1 AND SCHEMA_NAME(tt.schema_id) IN (SELECT value FROM STRING_SPLIT(@schemas, ','));
`

type GetDataTypesBySchemasRow struct {
	SchemaName string
	TypeName   string
	Type       string
	Definition string
}

func (q *Queries) GetDataTypesBySchemas(ctx context.Context, db mysql_queries.DBTX, schemas []string) ([]*GetDataTypesBySchemasRow, error) {
	// Join schemas into a comma-separated string
	schemasList := strings.Join(schemas, ",")
	rows, err := db.QueryContext(ctx, getDataTypesBySchemasAndTables, sql.Named("schemas", schemasList))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetDataTypesBySchemasRow
	for rows.Next() {
		var i GetDataTypesBySchemasRow
		if err := rows.Scan(
			&i.SchemaName,
			&i.TypeName,
			&i.Type,
			&i.Definition,
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
