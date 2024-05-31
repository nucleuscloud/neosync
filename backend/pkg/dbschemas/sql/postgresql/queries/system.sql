-- name: GetDatabaseSchema :many
SELECT
    n.nspname AS table_schema,
    c.relname AS table_name,
    a.attname AS column_name,
    pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type, -- This formats the type into something that should always be a valid postgres type. It also includes constraints if there are any
 COALESCE(
        pg_catalog.pg_get_expr(d.adbin, d.adrelid),
        ''
    )::text AS column_default,
    CASE
        WHEN a.attnotnull THEN 'NO'
        ELSE 'YES'
    END AS is_nullable,
    CASE
        WHEN pg_catalog.format_type(a.atttypid, a.atttypmod) LIKE 'character varying%' THEN
            a.atttypmod - 4 -- The -4 removes the extra bits that postgres uses for internal use
        ELSE
            -1
    END AS character_maximum_length,
    CASE
        WHEN a.atttypid = pg_catalog.regtype 'numeric'::regtype THEN
            (a.atttypmod - 4) >> 16
        -- Precision is technically only necessary for numeric values, but we are populating these here for simplicity in knowing what the type of integer is.
        -- This operates similar to the precision column in the information_schema.columns table
        WHEN a.atttypid = pg_catalog.regtype 'smallint'::regtype THEN
            16
        WHEN a.atttypid = pg_catalog.regtype 'integer'::regtype THEN
            32
        WHEN a.atttypid = pg_catalog.regtype 'bigint'::regtype THEN
            64
        ELSE
            -1
    END AS numeric_precision,
    CASE
        WHEN a.atttypid = pg_catalog.regtype 'numeric'::regtype THEN
            CASE
                -- If scale is not explicitly set, return -1 (meaning arbitrary scale)
                WHEN (a.atttypmod) = -1 THEN -1
                ELSE (a.atttypmod - 4) & 65535
            END
        -- Scale is technically only necessary for numeric values, but we are populating these here for simplicity in knowing what the type of integer is.
        -- This operates similar to the scake column in the information_schema.columns table
        WHEN a.atttypid = pg_catalog.regtype 'smallint'::regtype THEN
            0
        WHEN a.atttypid = pg_catalog.regtype 'integer'::regtype THEN
            0
        WHEN a.atttypid = pg_catalog.regtype 'bigint'::regtype THEN
            0
        ELSE
            -1
    END AS numeric_scale,
    a.attnum AS ordinal_position,
    a.attgenerated::text as generated_type
FROM
    pg_catalog.pg_attribute a
    INNER JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
    INNER JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    INNER JOIN pg_catalog.pg_type pgt ON pgt.oid = a.atttypid
    LEFT JOIN pg_catalog.pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE
    n.nspname NOT IN('pg_catalog', 'pg_toast', 'information_schema')
    AND a.attnum > 0
    AND NOT a.attisdropped
    AND c.relkind = 'r' -- ensures only tables are present
ORDER BY
    a.attnum;

-- name: GetDatabaseTableSchema :many
SELECT
    n.nspname AS schema_name,
    c.relname AS table_name,
    a.attname AS column_name,
    pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type,  -- This formats the type into something that should always be a valid postgres type. It also includes constraints if there are any
 COALESCE(
        pg_catalog.pg_get_expr(d.adbin, d.adrelid),
        ''
    )::text AS column_default,
    CASE
        WHEN a.attnotnull THEN 'NO'
        ELSE 'YES'
    END AS is_nullable,
    CASE
        WHEN pg_catalog.format_type(a.atttypid, a.atttypmod) LIKE 'character varying%' THEN
            a.atttypmod - 4 -- The -4 removes the extra bits that postgres uses for internal use
        ELSE
            -1
    END AS character_maximum_length,
    CASE
        WHEN a.atttypid = pg_catalog.regtype 'numeric'::regtype THEN
            (a.atttypmod - 4) >> 16
        -- Precision is technically only necessary for numeric values, but we are populating these here for simplicity in knowing what the type of integer is.
        -- This operates similar to the precision column in the information_schema.columns table
        WHEN a.atttypid = pg_catalog.regtype 'smallint'::regtype THEN
            16
        WHEN a.atttypid = pg_catalog.regtype 'integer'::regtype THEN
            32
        WHEN a.atttypid = pg_catalog.regtype 'bigint'::regtype THEN
            64
        ELSE
            -1
    END AS numeric_precision,
    CASE
        WHEN a.atttypid = pg_catalog.regtype 'numeric'::regtype THEN
            CASE
                -- If scale is not explicitly set, return -1 (meaning arbitrary scale)
                WHEN (a.atttypmod) = -1 THEN -1
                ELSE (a.atttypmod - 4) & 65535
            END
        -- Scale is technically only necessary for numeric values, but we are populating these here for simplicity in knowing what the type of integer is.
        -- This operates similar to the scake column in the information_schema.columns table
        WHEN a.atttypid = pg_catalog.regtype 'smallint'::regtype THEN
            0
        WHEN a.atttypid = pg_catalog.regtype 'integer'::regtype THEN
            0
        WHEN a.atttypid = pg_catalog.regtype 'bigint'::regtype THEN
            0
        ELSE
            -1
    END AS numeric_scale,
    a.attnum AS ordinal_position,
    a.attgenerated::text as generated_type
FROM
    pg_catalog.pg_attribute a
    INNER JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
    INNER JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    INNER JOIN pg_catalog.pg_type pgt ON pgt.oid = a.atttypid
    LEFT JOIN pg_catalog.pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE
    c.relname = sqlc.arg('table')
    AND n.nspname = sqlc.arg('schema')
    AND a.attnum > 0
    AND NOT a.attisdropped
    AND c.relkind = 'r' -- ensures only tables are present
ORDER BY
    a.attnum;

-- name: GetDatabaseTableSchemasBySchemasAndTables :many
SELECT
    n.nspname AS schema_name,
    c.relname AS table_name,
    a.attname AS column_name,
    pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type,  -- This formats the type into something that should always be a valid postgres type. It also includes constraints if there are any
 COALESCE(
        pg_catalog.pg_get_expr(d.adbin, d.adrelid),
        ''
    )::text AS column_default,
    CASE
        WHEN a.attnotnull THEN 'NO'
        ELSE 'YES'
    END AS is_nullable,
    CASE
        WHEN pg_catalog.format_type(a.atttypid, a.atttypmod) LIKE 'character varying%' THEN
            a.atttypmod - 4 -- The -4 removes the extra bits that postgres uses for internal use
        ELSE
            -1
    END AS character_maximum_length,
    CASE
        WHEN a.atttypid = pg_catalog.regtype 'numeric'::regtype THEN
            (a.atttypmod - 4) >> 16
        -- Precision is technically only necessary for numeric values, but we are populating these here for simplicity in knowing what the type of integer is.
        -- This operates similar to the precision column in the information_schema.columns table
        WHEN a.atttypid = pg_catalog.regtype 'smallint'::regtype THEN
            16
        WHEN a.atttypid = pg_catalog.regtype 'integer'::regtype THEN
            32
        WHEN a.atttypid = pg_catalog.regtype 'bigint'::regtype THEN
            64
        ELSE
            -1
    END AS numeric_precision,
    CASE
        WHEN a.atttypid = pg_catalog.regtype 'numeric'::regtype THEN
            CASE
                -- If scale is not explicitly set, return -1 (meaning arbitrary scale)
                WHEN (a.atttypmod) = -1 THEN -1
                ELSE (a.atttypmod - 4) & 65535
            END
        -- Scale is technically only necessary for numeric values, but we are populating these here for simplicity in knowing what the type of integer is.
        -- This operates similar to the scake column in the information_schema.columns table
        WHEN a.atttypid = pg_catalog.regtype 'smallint'::regtype THEN
            0
        WHEN a.atttypid = pg_catalog.regtype 'integer'::regtype THEN
            0
        WHEN a.atttypid = pg_catalog.regtype 'bigint'::regtype THEN
            0
        ELSE
            -1
    END AS numeric_scale,
    a.attnum AS ordinal_position,
    a.attgenerated::text as generated_type
FROM
    pg_catalog.pg_attribute a
    INNER JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
    INNER JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    INNER JOIN pg_catalog.pg_type pgt ON pgt.oid = a.atttypid
    LEFT JOIN pg_catalog.pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE
    (n.nspname || '.' || c.relname) = ANY(sqlc.arg('schematables')::TEXT[])
    AND a.attnum > 0
    AND NOT a.attisdropped
    AND c.relkind = 'r' -- ensures only tables are present
ORDER BY
    a.attnum;

-- name: GetTableConstraints :many
SELECT
    con.conname AS constraint_name,
    con.contype::TEXT AS constraint_type,
    con.connamespace::regnamespace::TEXT AS schema_name,
    cls.relname AS table_name,
    array_agg(att.attname)::TEXT[] AS constraint_columns,
    array_agg(att.attnotnull)::BOOL[] AS notnullable,
    CASE
        WHEN con.contype = 'f' THEN fn_cl.relnamespace::regnamespace::TEXT
        ELSE ''
    END AS foreign_schema_name,
    CASE
        WHEN con.contype = 'f' THEN fn_cl.relname
        ELSE ''
    END AS foreign_table_name,
    CASE
        WHEN con.contype = 'f' THEN fk_columns.foreign_column_names
        ELSE NULL::text[]
    END as foreign_column_names,
    pg_get_constraintdef(con.oid)::TEXT AS constraint_definition
FROM
    pg_catalog.pg_constraint con
JOIN
    pg_catalog.pg_attribute att ON
    att.attrelid = con.conrelid
    AND att.attnum = ANY(con.conkey)
JOIN
    pg_catalog.pg_class cls ON
    con.conrelid = cls.oid
JOIN
    pg_catalog.pg_namespace nsp ON
    cls.relnamespace = nsp.oid
LEFT JOIN
    pg_catalog.pg_class fn_cl ON
    fn_cl.oid = con.confrelid
LEFT JOIN LATERAL (
        SELECT
            array_agg(fk_att.attname) AS foreign_column_names
        FROM
            pg_catalog.pg_attribute fk_att
        WHERE
            fk_att.attrelid = con.confrelid
            AND fk_att.attnum = ANY(con.confkey)
    ) AS fk_columns ON
    TRUE
WHERE
    con.connamespace::regnamespace::TEXT = sqlc.arg('schema')
    AND con.conrelid::regclass::TEXT = sqlc.arg('table')
GROUP BY
    con.oid,
    con.connamespace,
    con.conname,
    cls.relname,
    con.contype,
    fn_cl.relnamespace,
    fn_cl.relname,
    fk_columns.foreign_column_names;

-- name: GetTableConstraintsBySchema :many
SELECT
    con.conname AS constraint_name,
    con.contype::TEXT AS constraint_type,
    con.connamespace::regnamespace::TEXT AS schema_name,
    cls.relname AS table_name,
    array_agg(att.attname)::TEXT[] AS constraint_columns,
    array_agg(att.attnotnull)::BOOL[] AS notnullable,
    CASE
        WHEN con.contype = 'f' THEN fn_cl.relnamespace::regnamespace::TEXT
        ELSE ''
    END AS foreign_schema_name,
    CASE
        WHEN con.contype = 'f' THEN fn_cl.relname
        ELSE ''
    END AS foreign_table_name,
    CASE
        WHEN con.contype = 'f' THEN fk_columns.foreign_column_names
        ELSE NULL::text[]
    END as foreign_column_names,
    pg_get_constraintdef(con.oid)::TEXT AS constraint_definition
FROM
    pg_catalog.pg_constraint con
JOIN
    pg_catalog.pg_attribute att ON
    att.attrelid = con.conrelid
    AND att.attnum = ANY(con.conkey)
JOIN
    pg_catalog.pg_class cls ON
    con.conrelid = cls.oid
JOIN
    pg_catalog.pg_namespace nsp ON
    cls.relnamespace = nsp.oid
LEFT JOIN
    pg_catalog.pg_class fn_cl ON
    fn_cl.oid = con.confrelid
LEFT JOIN LATERAL (
        SELECT
            array_agg(fk_att.attname) AS foreign_column_names
        FROM
            pg_catalog.pg_attribute fk_att
        WHERE
            fk_att.attrelid = con.confrelid
            AND fk_att.attnum = ANY(con.confkey)
    ) AS fk_columns ON
    TRUE
WHERE
    con.connamespace::regnamespace::TEXT = ANY(
        sqlc.arg('schema')::TEXT[]
    )
GROUP BY
    con.oid,
    con.connamespace,
    con.conname,
    cls.relname,
    con.contype,
    fn_cl.relnamespace,
    fn_cl.relname,
    fk_columns.foreign_column_names;

-- name: GetPostgresRolePermissions :many
SELECT
    rtg.table_schema as table_schema,
    rtg.table_name as table_name,
    rtg.privilege_type as privilege_type
FROM
    information_schema.role_table_grants as rtg
WHERE
    table_schema NOT IN ('pg_catalog', 'information_schema')
AND grantee =  sqlc.arg('role')
ORDER BY
    table_schema,
    table_name;

-- name: GetIndicesBySchemasAndTables :many
SELECT
    ns.nspname AS schema_name,
    t.relname AS table_name,
    i.relname AS index_name,
    pg_get_indexdef(ix.indexrelid) AS index_definition
FROM
    pg_catalog.pg_class t
    JOIN pg_catalog.pg_index ix ON t.oid = ix.indrelid
    JOIN pg_catalog.pg_class i ON i.oid = ix.indexrelid
    JOIN pg_catalog.pg_namespace ns ON t.relnamespace = ns.oid
LEFT JOIN pg_catalog.pg_constraint con ON con.conindid = ix.indexrelid
WHERE
    con.conindid IS NULL -- Excludes indexes created as part of constraints
    AND (ns.nspname || '.' || t.relname) = ANY(sqlc.arg('schematables')::TEXT[])
GROUP BY
    ns.nspname, t.relname, i.relname, ix.indexrelid
ORDER BY
    schema_name,
    table_name,
    index_name;

-- name: GetDataTypesBySchemaAndTables :many
WITH custom_types AS (
    SELECT
        n.nspname AS schema_name,
        t.typname AS type_name,
        t.oid AS type_oid,
        CASE
            WHEN t.typtype = 'd' THEN 'domain'
            WHEN t.typtype = 'e' THEN 'enum'
            WHEN t.typtype = 'c' THEN 'composite'
        END AS type
    FROM
        pg_catalog.pg_type t
    JOIN
        pg_catalog.pg_namespace n ON n.oid = t.typnamespace
    WHERE
        n.nspname = 'public'
        AND t.typtype IN ('d', 'e', 'c')
),
table_columns AS (
    SELECT
        c.oid AS table_oid,
        a.atttypid AS type_oid
    FROM
        pg_catalog.pg_class c
    JOIN
        pg_catalog.pg_namespace n ON n.oid = c.relnamespace
    JOIN
        pg_catalog.pg_attribute a ON a.attrelid = c.oid
    WHERE
        n.nspname = 'public'
        AND c.relname IN ('example_table') -- replace with your table names
        AND a.attnum > 0
        AND NOT a.attisdropped
),
relevant_custom_types AS (
    SELECT DISTINCT
        ct.schema_name,
        ct.type_name,
        ct.type_oid,
        ct.type
    FROM
        custom_types ct
    JOIN
        table_columns tc ON ct.type_oid = tc.type_oid
),
domain_defs AS (
    SELECT
        rct.schema_name,
        rct.type_name,
        rct.type,
        'CREATE DOMAIN ' || rct.schema_name || '.' || rct.type_name || ' AS ' ||
        pg_catalog.format_type(t.typbasetype, t.typtypmod) ||
        CASE
            WHEN t.typnotnull THEN ' NOT NULL' ELSE ''
        END || ' ' ||
        COALESCE('CONSTRAINT ' || conname || ' ' || pg_catalog.pg_get_constraintdef(c.oid), '') || ';' AS definition
    FROM
        relevant_custom_types rct
    JOIN
        pg_catalog.pg_type t ON rct.type_oid = t.oid
    LEFT JOIN
        pg_catalog.pg_constraint c ON t.oid = c.contypid
    WHERE
        rct.type = 'domain'
),
enum_defs AS (
    SELECT
        rct.schema_name,
        rct.type_name,
        rct.type,
        'CREATE TYPE ' || rct.schema_name || '.' || rct.type_name || ' AS ENUM (' ||
        string_agg('''' || e.enumlabel || '''', ', ') || ');' AS definition
    FROM
        relevant_custom_types rct
    JOIN
        pg_catalog.pg_type t ON rct.type_oid = t.oid
    JOIN
        pg_catalog.pg_enum e ON t.oid = e.enumtypid
    WHERE
        rct.type = 'enum'
    GROUP BY
        rct.schema_name, rct.type_name, rct.type
),
composite_defs AS (
    SELECT
        rct.schema_name,
        rct.type_name,
        rct.type,
        'CREATE TYPE ' || rct.schema_name || '.' || rct.type_name || ' AS (' ||
        string_agg(a.attname || ' ' || pg_catalog.format_type(a.atttypid, a.atttypmod), ', ') || ');' AS definition
    FROM
        relevant_custom_types rct
    JOIN
        pg_catalog.pg_type t ON rct.type_oid = t.oid
    JOIN
        pg_catalog.pg_class c ON c.oid = t.typrelid
    JOIN
        pg_catalog.pg_attribute a ON a.attrelid = c.oid
    WHERE
        rct.type = 'composite'
        AND a.attnum > 0
        AND NOT a.attisdropped
    GROUP BY
        rct.schema_name, rct.type_name, rct.type
)
SELECT
    schema_name,
    type_name,
    "type"::text,
    "definition"::text
FROM
    domain_defs
UNION ALL
SELECT
    schema_name,
    type_name,
    "type"::text,
    "definition"::text
FROM
    enum_defs
UNION ALL
SELECT
    schema_name,
    type_name,
    "type"::text,
    "definition"::text
FROM
    composite_defs;

-- name: GetCustomFunctionsBySchemaAndTables :many
WITH relevant_schemas_tables AS (
    SELECT c.oid, n.nspname AS schema_name, c.relname AS table_name
    FROM pg_catalog.pg_class c
    JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'  -- replace with your schema name
    AND c.relname IN ('new_table')  -- replace with your table names
),
trigger_functions AS (
    SELECT DISTINCT
        n.nspname AS schema_name,
        p.proname AS function_name,
        pg_catalog.pg_get_functiondef(p.oid) AS definition
    FROM pg_catalog.pg_trigger t
    JOIN pg_catalog.pg_proc p ON t.tgfoid = p.oid
    JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
    WHERE t.tgrelid IN (SELECT oid FROM relevant_schemas_tables)
    AND NOT t.tgisinternal
)
SELECT
    schema_name,
    function_name,
    definition
FROM
    trigger_functions
ORDER BY
    schema_name,
    function_name;

-- name: GetCustomTriggersBySchemaAndTables :many
WITH relevant_schemas_tables AS (
    SELECT c.oid, n.nspname AS schema_name, c.relname AS table_name
    FROM pg_catalog.pg_class c
    JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'  -- replace with your schema name
    AND c.relname IN ('new_table')  -- replace with your table names
)
SELECT
    n.nspname AS schema_name,
    c.relname AS table_name,
    t.tgname AS trigger_name,
    pg_catalog.pg_get_triggerdef(t.oid, true) AS definition
FROM pg_catalog.pg_trigger t
JOIN pg_catalog.pg_class c ON c.oid = t.tgrelid
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE c.oid IN (SELECT oid FROM relevant_schemas_tables)
AND NOT t.tgisinternal
ORDER BY
    schema_name,
    table_name,
    trigger_name;
