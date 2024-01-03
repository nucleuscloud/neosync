-- name: GetDatabaseSchema :many
SELECT
	c.table_schema,
	c.table_name,
	c.column_name,
	c.ordinal_position,
	COALESCE(c.column_default, 'NULL') as column_default, -- must coalesce because sqlc doesn't appear to work for system structs to output a *string
	c.is_nullable,
	c.data_type
FROM
	information_schema.columns AS c
	JOIN information_schema.tables AS t ON c.table_schema = t.table_schema
		AND c.table_name = t.table_name
WHERE
	c.table_schema NOT IN('pg_catalog', 'information_schema')
	AND t.table_type = 'BASE TABLE';

-- name: GetDatabaseTableSchema :many
SELECT
	c.table_schema,
	c.table_name,
	c.column_name,
	c.ordinal_position,
	COALESCE(c.column_default, 'NULL') as column_default, -- must coalesce because sqlc doesn't appear to work for system structs to output a *string
	c.is_nullable,
	c.data_type,
    c.character_maximum_length
FROM
	information_schema.columns AS c
	JOIN information_schema.tables AS t ON c.table_schema = t.table_schema
		AND c.table_name = t.table_name
WHERE
	c.table_schema = $1 AND t.table_name = $2
	AND t.table_type = 'BASE TABLE'
	ORDER BY c.ordinal_position ASC;

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
    rc.constraint_name
    ,
    kcu.table_schema AS schema_name
    ,
    kcu.table_name
    ,
    kcu.column_name
    ,
    c.is_nullable
    ,
    kcu2.table_schema AS foreign_schema_name
    ,
    kcu2.table_name AS foreign_table_name
    ,
    kcu2.column_name AS foreign_column_name
FROM
    information_schema.referential_constraints rc
JOIN information_schema.key_column_usage kcu
    ON
    kcu.constraint_name = rc.constraint_name
JOIN information_schema.key_column_usage kcu2
    ON
    kcu2.ordinal_position = kcu.position_in_unique_constraint
    AND kcu2.constraint_name = rc.unique_constraint_name
JOIN information_schema.columns as c
	ON
	c.table_schema = kcu.table_schema 
	AND c.table_name = kcu.table_name 
	AND c.column_name = kcu.column_name
WHERE
    kcu.table_schema = sqlc.arg('tableSchema')
ORDER BY
    rc.constraint_name,
    kcu.ordinal_position;

