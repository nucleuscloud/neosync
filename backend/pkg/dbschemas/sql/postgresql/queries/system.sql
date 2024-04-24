-- name: GetDatabaseSchema :many
SELECT
    n.nspname AS table_schema,
    c.relname AS table_name,
    a.attname AS column_name,
    pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type, -- This formats the type into something that should always be a valid postgres type. It also includes constraints if there are any
    COALESCE(
        pg_catalog.pg_get_expr(d.adbin, d.adrelid),
        ''
    ) AS column_default,
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
            (a.atttypmod - 4) & 65535
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
    a.attnum AS ordinal_position
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
    ) AS column_default,
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
            (a.atttypmod - 4) & 65535
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
    a.attnum AS ordinal_position
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

-- name: GetTableConstraints :many
SELECT
    con.conname AS constraint_name,
    con.contype::text AS constraint_type,
    con.connamespace::regnamespace::text AS schema_name,
    con.conrelid::regclass::text AS table_name,
    CASE
        WHEN con.contype IN ('f', 'p', 'u') THEN array_agg(att.attname)
        ELSE NULL
    END::text[] AS constraint_columns,
    array_agg(att.attnotnull)::bool[] AS notnullable,
    CASE
        WHEN con.contype = 'f' THEN fn_cl.relnamespace::regnamespace::text
        ELSE ''
    END AS foreign_schema_name,
    CASE
        WHEN con.contype = 'f' THEN con.confrelid::regclass::text
        ELSE ''
    END AS foreign_table_name,
    CASE
        WHEN con.contype = 'f' THEN array_agg(fk_att.attname)::text[]
        ELSE NULL::text[]
    END AS foreign_column_names,
    pg_get_constraintdef(con.oid)::text AS constraint_definition
FROM
    pg_catalog.pg_constraint con
LEFT JOIN
    pg_catalog.pg_attribute fk_att ON fk_att.attrelid = con.confrelid AND fk_att.attnum = ANY(con.confkey)
LEFT JOIN
    pg_catalog.pg_attribute att ON att.attrelid = con.conrelid AND att.attnum = ANY(con.conkey)
LEFT JOIN
    pg_catalog.pg_class fn_cl ON fn_cl.oid = con.confrelid
WHERE
    con.connamespace::regnamespace::text = sqlc.arg('schema') AND con.conrelid::regclass::text = sqlc.arg('table')
GROUP BY
    con.oid, con.conname, con.conrelid, fn_cl.relnamespace, con.confrelid, con.contype;

-- name: GetTableConstraintsBySchema :many
SELECT
    con.conname AS constraint_name,
    con.contype::text AS constraint_type,
    con.connamespace::regnamespace::text AS schema_name,
    con.conrelid::regclass::text AS table_name,
    CASE
        WHEN con.contype IN ('f', 'p', 'u') THEN array_agg(att.attname)
        ELSE NULL
    END::text[] AS constraint_columns,
    array_agg(att.attnotnull)::bool[] AS notnullable,
    CASE
        WHEN con.contype = 'f' THEN fn_cl.relnamespace::regnamespace::text
        ELSE ''
    END AS foreign_schema_name,
    CASE
        WHEN con.contype = 'f' THEN con.confrelid::regclass::text
        ELSE ''
    END AS foreign_table_name,
    CASE
        WHEN con.contype = 'f' THEN array_agg(fk_att.attname)::text[]
        ELSE NULL::text[]
    END AS foreign_column_names,
    pg_get_constraintdef(con.oid)::text AS constraint_definition
FROM
    pg_catalog.pg_constraint con
LEFT JOIN
    pg_catalog.pg_attribute fk_att ON fk_att.attrelid = con.confrelid AND fk_att.attnum = ANY(con.confkey)
LEFT JOIN
    pg_catalog.pg_attribute att ON att.attrelid = con.conrelid AND att.attnum = ANY(con.conkey)
LEFT JOIN
    pg_catalog.pg_class fn_cl ON fn_cl.oid = con.confrelid
WHERE
    con.connamespace::regnamespace::text = ANY(sqlc.arg('schema')::text[])
GROUP BY
    con.oid, con.conname, con.conrelid, fn_cl.relnamespace, con.confrelid, con.contype;

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
