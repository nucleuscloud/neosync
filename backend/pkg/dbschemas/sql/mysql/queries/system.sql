-- name: GetDatabaseSchema :many
SELECT
	c.table_schema,
	c.table_name,
	c.column_name,
	c.ordinal_position,
    IFNULL(REPLACE(REPLACE(REPLACE(REPLACE(c.COLUMN_DEFAULT, '_utf8mb4\\\'', '_utf8mb4\''), '_utf8mb3\\\'', '_utf8mb3\''), '\\\'', '\''), '\\\'', '\''), '') AS column_default, -- hack to fix this bug https://bugs.mysql.com/bug.php?
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


-- name: GetTableConstraintsBySchemas :many
SELECT
    tc.table_schema AS schema_name,
    tc.table_name AS table_name,
    JSON_ARRAYAGG(kcu.column_name) AS constraint_columns,
    JSON_ARRAYAGG(CASE WHEN c.is_nullable = 'YES' THEN 0 ELSE 1 END) AS not_nullable,
    tc.constraint_name AS constraint_name,
    tc.constraint_type AS constraint_type,
    COALESCE(kcu.referenced_table_schema, 'NULL') AS referenced_schema_name,
    COALESCE(kcu.referenced_table_name, 'NULL') AS referenced_table_name,
    JSON_ARRAYAGG(kcu.referenced_column_name) AS referenced_column_names,
    rc.update_rule as update_rule,
    rc.delete_rule as delete_rule,
    IFNULL(REPLACE(REPLACE(REPLACE(REPLACE(cc.check_clause, '_utf8mb4\\\'', '_utf8mb4\''), '_utf8mb3\\\'', '_utf8mb3\''), '\\\'', '\''), '\\\'', '\''), '') AS check_clause -- hack to fix this bug https://bugs.mysql.com/
FROM
    information_schema.table_constraints AS tc
LEFT JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
    AND tc.table_name = kcu.table_name
LEFT JOIN information_schema.columns as c
    ON c.table_schema = kcu.table_schema
    AND c.table_name = kcu.table_name
    AND c.column_name = kcu.column_name
LEFT JOIN information_schema.referential_constraints as rc
	ON rc.constraint_schema = tc.table_schema
	AND rc.table_name = tc.table_name
	AND rc.constraint_name = tc.constraint_name
LEFT JOIN information_schema.check_constraints as cc
	ON tc.constraint_schema = cc.constraint_schema
	AND tc.constraint_name = cc.constraint_name
WHERE
    tc.table_schema IN (sqlc.slice('schemas'))
GROUP BY 
    tc.table_schema,
    tc.table_name,
    tc.constraint_name,
    tc.constraint_type,
    kcu.referenced_table_schema,
    kcu.referenced_table_name,
    rc.update_rule,
    rc.delete_rule,
    cc.check_clause;

-- name: GetTableConstraints :many
SELECT
    tc.table_schema AS schema_name,
    tc.table_name AS table_name,
    JSON_ARRAYAGG(kcu.column_name) AS constraint_columns,
    JSON_ARRAYAGG(CASE WHEN c.is_nullable = 'YES' THEN 0 ELSE 1 END) AS not_nullable,
    tc.constraint_name AS constraint_name,
    tc.constraint_type AS constraint_type,
    COALESCE(kcu.referenced_table_schema, 'NULL') AS referenced_schema_name,
    COALESCE(kcu.referenced_table_name, 'NULL') AS referenced_table_name,
    JSON_ARRAYAGG(kcu.referenced_column_name) AS referenced_column_names,
    rc.update_rule as update_rule,
    rc.delete_rule as delete_rule,
    IFNULL(REPLACE(REPLACE(REPLACE(REPLACE(cc.check_clause, '_utf8mb4\\\'', '_utf8mb4\''), '_utf8mb3\\\'', '_utf8mb3\''), '\\\'', '\''), '\\\'', '\''), '') AS check_clause -- hack to fix this bug https://bugs.mysql.com/
FROM
    information_schema.table_constraints AS tc
LEFT JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
    AND tc.table_name = kcu.table_name
LEFT JOIN information_schema.columns as c
    ON c.table_schema = kcu.table_schema
    AND c.table_name = kcu.table_name
    AND c.column_name = kcu.column_name
LEFT JOIN information_schema.referential_constraints as rc
	ON rc.constraint_schema = tc.table_schema
	AND rc.table_name = tc.table_name
	AND rc.constraint_name = tc.constraint_name
LEFT JOIN information_schema.check_constraints as cc
	ON tc.constraint_schema = cc.constraint_schema
	AND tc.constraint_name = cc.constraint_name
WHERE
    tc.table_schema = sqlc.arg('schema')
    AND tc.table_name IN (sqlc.slice('tables'))
GROUP BY 
    tc.table_schema,
    tc.table_name,
    tc.constraint_name,
    tc.constraint_type,
    kcu.referenced_table_schema,
    kcu.referenced_table_name,
    rc.update_rule,
    rc.delete_rule,
    cc.check_clause;


-- name: GetMysqlRolePermissions :many
WITH admin_privileges AS (
	SELECT
		privilege_type
	FROM
		INFORMATION_SCHEMA.USER_PRIVILEGES
	WHERE
		USER_PRIVILEGES.GRANTEE = CONCAT("'",
			SUBSTRING_INDEX(CURRENT_USER(),
				'@',
				1),
			"'@'%'")
),
db_privileges AS (
	SELECT
		TABLE_SCHEMA AS table_schema,
		PRIVILEGE_TYPE AS privilege_type
	FROM
		INFORMATION_SCHEMA.SCHEMA_PRIVILEGES
	WHERE
		SCHEMA_PRIVILEGES.GRANTEE = CONCAT("'",
			SUBSTRING_INDEX(CURRENT_USER(),
				'@',
				1),
			"'@'%'")
)
SELECT
	t.TABLE_SCHEMA AS table_schema, t.TABLE_NAME AS table_name, ap.privilege_type AS privilege_type
FROM
	INFORMATION_SCHEMA.TABLES AS t
	JOIN admin_privileges AS ap
WHERE
	t.TABLE_SCHEMA NOT IN('mysql', 'sys', 'performance_schema', 'information_schema')
UNION
SELECT
	t.TABLE_SCHEMA AS table_schema,
	t.TABLE_NAME AS table_name,
	dp.privilege_type AS privilege_type
FROM
	INFORMATION_SCHEMA.TABLES AS t
	JOIN db_privileges AS dp ON dp.table_schema = t.table_schema
WHERE
	t.TABLE_SCHEMA IN(
		SELECT
			table_schema FROM db_privileges)
UNION
SELECT
	TABLE_PRIVILEGES.TABLE_SCHEMA AS table_schema,
	TABLE_PRIVILEGES.TABLE_NAME AS table_name,
	TABLE_PRIVILEGES.PRIVILEGE_TYPE AS privilege_type
FROM
	INFORMATION_SCHEMA.TABLE_PRIVILEGES
WHERE
	TABLE_PRIVILEGES.GRANTEE = CONCAT("'", SUBSTRING_INDEX(CURRENT_USER(), '@', 1), "'@'%'");


-- sqlc is broken for mysql so can't do CONCAT(EVENT_OBJECT_SCHEMA, '.', EVENT_OBJECT_TABLE) IN (sqlc.slice('schematables'))
-- name: GetCustomTriggersBySchemaAndTables :many
SELECT
    TRIGGER_NAME AS trigger_name,
    TRIGGER_SCHEMA as trigger_schema,
    EVENT_OBJECT_SCHEMA AS schema_name,
    EVENT_OBJECT_TABLE AS table_name,
    ACTION_STATEMENT AS statement,
    EVENT_MANIPULATION AS event_type,
    ACTION_ORIENTATION AS orientation,
    ACTION_TIMING AS timing
FROM
    information_schema.TRIGGERS
WHERE 
    EVENT_OBJECT_SCHEMA = sqlc.arg('schema') AND EVENT_OBJECT_TABLE IN (sqlc.slice('tables'));


-- name: GetDatabaseTableSchemasBySchemasAndTables :many
SELECT
   c.TABLE_SCHEMA AS schema_name,
   c.TABLE_NAME AS table_name,
   c.COLUMN_NAME AS column_name,
   c.COLUMN_TYPE AS data_type,
   IFNULL(REPLACE(REPLACE(REPLACE(REPLACE(c.COLUMN_DEFAULT, '_utf8mb4\\\'', '_utf8mb4\''), '_utf8mb3\\\'', '_utf8mb3\''), '\\\'', '\''), '\\\'', '\''), '') AS column_default, -- hack to fix this bug https://bugs.mysql.com/bug.php?
   CASE WHEN c.IS_NULLABLE = 'YES' THEN 1 ELSE 0 END AS is_nullable,
   IF(c.DATA_TYPE = 'varchar', c.CHARACTER_MAXIMUM_LENGTH, -1) AS character_maximum_length,
   IF(c.DATA_TYPE IN ('decimal', 'numeric'), c.NUMERIC_PRECISION, 
     IF(c.DATA_TYPE = 'smallint', 16, 
        IF(c.DATA_TYPE = 'int', 32, 
           IF(c.DATA_TYPE = 'bigint', 64, -1)))) AS numeric_precision,
   IF(c.DATA_TYPE IN ('decimal', 'numeric'), c.NUMERIC_SCALE, 0) AS numeric_scale,
   c.ORDINAL_POSITION AS ordinal_position,
   c.EXTRA AS identity_generation,
   IFNULL(REPLACE(REPLACE(REPLACE(REPLACE(c.GENERATION_EXPRESSION, '_utf8mb4\\\'', '_utf8mb4\''), '_utf8mb3\\\'', '_utf8mb3\''), '\\\'', '\''), '\\\'', '\''), '') AS generation_exp, -- hack to fix this bug https://bugs.mysql.com/
   t.AUTO_INCREMENT as auto_increment_start_value
FROM
    information_schema.COLUMNS as c
    join information_schema.TABLES as t on t.TABLE_SCHEMA = c.TABLE_SCHEMA and t.TABLE_NAME = c.TABLE_NAME
WHERE
  -- CONCAT(c.TABLE_SCHEMA, '.', c.TABLE_NAME) IN (sqlc.slice('schematables')) broken
	c.TABLE_SCHEMA = sqlc.arg('schema') AND c.TABLE_NAME in (sqlc.slice('tables'))
ORDER BY
    c.ordinal_position;


-- name: GetIndicesBySchemasAndTables :many
SELECT 
    s.TABLE_SCHEMA as schema_name,
    s.TABLE_NAME as table_name,
    s.COLUMN_NAME as column_name,
    s.INDEX_NAME as index_name,
    s.INDEX_TYPE as index_type,
    s.SEQ_IN_INDEX as seq_in_index,
    s.NULLABLE as nullable
FROM 
    INFORMATION_SCHEMA.STATISTICS s
LEFT JOIN 
    INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
    ON s.TABLE_SCHEMA = kcu.CONSTRAINT_SCHEMA
    AND s.TABLE_NAME = kcu.TABLE_NAME
    AND s.COLUMN_NAME = kcu.COLUMN_NAME
WHERE 
    -- CONCAT(s.TABLE_SCHEMA, '.', s.TABLE_NAME) IN (sqlc.slice('schematables')) broken
		s.TABLE_SCHEMA = sqlc.arg('schema') AND s.TABLE_NAME in (sqlc.slice('tables'))
    AND s.INDEX_NAME != 'PRIMARY'
    AND kcu.CONSTRAINT_NAME IS NULL
ORDER BY 
    s.TABLE_NAME,
    s.INDEX_NAME,
    s.SEQ_IN_INDEX;


-- name: GetCustomFunctionsBySchemas :many
SELECT 
    ROUTINE_NAME as function_name, 
    ROUTINE_SCHEMA as schema_name,
    DTD_IDENTIFIER as return_data_type,
    ROUTINE_DEFINITION as definition,
    CASE WHEN IS_DETERMINISTIC = 'YES' THEN 1 ELSE 0 END as is_deterministic
FROM 
    INFORMATION_SCHEMA.ROUTINES 
WHERE 
    ROUTINE_TYPE = 'FUNCTION'
    AND ROUTINE_SCHEMA in (sqlc.slice('schemas'));
