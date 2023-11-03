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
	c.table_schema NOT IN('sys', 'performance_schema', 'mysql')
	AND t.table_type = 'BASE TABLE';

-- name: GetDatabaseTableSchema :many
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
	c.table_schema = ? AND t.table_name = ?
	AND t.table_type = 'BASE TABLE'
	ORDER BY c.ordinal_position ASC;

-- name: GetTableConstraints :many
SELECT
kcu.constraint_name
,
kcu.table_schema AS db_schema
,
kcu.table_name as table_name
,
kcu.column_name as column_name
,
kcu.referenced_table_schema AS foreign_schema_name
,
kcu.referenced_table_name AS foreign_table_name
,
kcu.referenced_column_name AS foreign_column_name
,
rc.update_rule
,
rc.delete_rule
FROM information_schema.key_column_usage kcu
LEFT JOIN information_schema.referential_constraints rc
	ON
	kcu.constraint_name = rc.constraint_name
WHERE
	kcu.table_schema = ? AND kcu.table_name = ?;

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
WHERE
	kcu.table_schema = ?
ORDER BY
	rc.constraint_name,
	kcu.ordinal_position;
