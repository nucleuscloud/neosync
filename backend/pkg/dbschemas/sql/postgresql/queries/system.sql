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
    nsp.nspname AS db_schema,
    rel.relname AS table_name,
    con.conname AS constraint_name,
    pg_get_constraintdef(con.oid) AS constraint_definition
FROM
    pg_catalog.pg_constraint con
INNER JOIN pg_catalog.pg_class rel
                       ON
    rel.oid = con.conrelid
INNER JOIN pg_catalog.pg_namespace nsp
                       ON
    nsp.oid = connamespace
WHERE
    nsp.nspname = sqlc.arg('schema') AND rel.relname = sqlc.arg('table');

-- name: GetForeignKeyConstraints :many
SELECT
    rc.constraint_name,
    rc.constraint_schema AS schema_name,
    fk.table_name,
    fk.column_name,
    c.is_nullable,
    pk.table_schema AS foreign_schema_name,
    pk.table_name AS foreign_table_name,
    pk.column_name AS foreign_column_name
FROM
    information_schema.referential_constraints rc
JOIN information_schema.key_column_usage fk ON
    fk.constraint_catalog = rc.constraint_catalog AND
    fk.constraint_schema = rc.constraint_schema AND
    fk.constraint_name = rc.constraint_name
JOIN information_schema.constraint_column_usage pk ON
    pk.constraint_catalog = rc.unique_constraint_catalog AND
    pk.constraint_schema = rc.unique_constraint_schema AND
    pk.constraint_name = rc.unique_constraint_name
JOIN information_schema.columns c ON
    c.table_schema = fk.table_schema AND
    c.table_name = fk.table_name AND
    c.column_name = fk.column_name
WHERE
    rc.constraint_schema = sqlc.arg('tableSchema')
ORDER BY
    rc.constraint_name,
    fk.ordinal_position;

-- name: GetPrimaryKeyConstraints :many
SELECT
    tc.table_schema AS schema_name,
    tc.table_name as table_name,
    tc.constraint_name as constraint_name,
    kcu.column_name as column_name
FROM
    information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
WHERE
    tc.table_schema = sqlc.arg('tableSchema')
    AND tc.constraint_type = 'PRIMARY KEY'
ORDER BY
    tc.table_name,
    kcu.column_name;
