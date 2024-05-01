-- name: GetDatabaseSchema :many
SELECT
	c.table_schema,
	c.table_name,
	c.column_name,
	c.ordinal_position,
	COALESCE(c.column_default, 'NULL') as column_default, -- must coalesce because sqlc doesn't appear to work for system structs to output a *string
	c.is_nullable,
	c.data_type,
	c.character_maximum_length,
  c.numeric_precision,
  c.numeric_scale,
	c.extra
FROM
	information_schema.columns AS c
	JOIN information_schema.tables AS t ON c.table_schema = t.table_schema
		AND c.table_name = t.table_name
WHERE
	c.table_schema NOT IN('sys', 'performance_schema', 'mysql')
	AND t.table_type = 'BASE TABLE';

-- name: GetForeignKeyConstraints :many
SELECT
rc.constraint_name
,
kcu.table_schema AS schema_name
,
kcu.table_name as table_name
,
kcu.column_name as column_name
,
c.is_nullable as is_nullable
,
kcu.referenced_table_schema AS foreign_schema_name
,
kcu.referenced_table_name AS foreign_table_name
,
kcu.referenced_column_name AS foreign_column_name
FROM
	information_schema.referential_constraints rc
JOIN information_schema.key_column_usage kcu
	ON
	kcu.constraint_name = rc.constraint_name
JOIN information_schema.columns as c
	ON
	c.table_schema = kcu.table_schema
	AND c.table_name = kcu.table_name
	AND c.column_name = kcu.column_name
WHERE
	kcu.table_schema = ?
ORDER BY
	rc.constraint_name,
	kcu.ordinal_position;


-- name: GetPrimaryKeyConstraints :many
SELECT
	table_schema AS schema_name,
	table_name as table_name,
	column_name as column_name,
	constraint_name as constraint_name
FROM
	information_schema.key_column_usage
WHERE
	table_schema = ?
	AND constraint_name = 'PRIMARY'
ORDER BY
	table_name,
	column_name;


-- name: GetUniqueConstraints :many
SELECT
    tc.table_schema AS schema_name,
    tc.table_name AS table_name,
    tc.constraint_name AS constraint_name,
    kcu.column_name AS column_name
FROM
    information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
    AND tc.table_name = kcu.table_name
WHERE
    tc.table_schema = ?
    AND tc.constraint_type = 'UNIQUE'
ORDER BY
    tc.table_name,
    kcu.column_name;


-- name: GetMysqlRolePermissions :many
SELECT
    t.table_schema,
    t.table_name,
	p.grantee,
	p.privilege_type
FROM
    information_schema.tables t
    JOIN information_schema.table_privileges p ON t.table_schema = p.table_schema
    AND t.table_name = p.table_name
WHERE
    t.table_schema NOT IN ('sys', 'performance_schema', 'mysql', 'information_schema')
  AND p.GRANTEE = sqlc.arg('role')
ORDER BY
    t.table_schema, t.table_name, p.GRANTEE;
