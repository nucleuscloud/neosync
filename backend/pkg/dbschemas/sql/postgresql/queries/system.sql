-- name: GetDatabaseSchema :many
WITH linked_to_serial AS (
    SELECT
        cl.relname AS sequence_name,
        nsp.nspname AS schema_name,
        cl.oid AS sequence_oid,
        ad.adrelid,
        ad.adnum,
        pg_catalog.pg_get_expr(ad.adbin, ad.adrelid)
    FROM
        pg_catalog.pg_class cl
    JOIN
        pg_catalog.pg_namespace nsp ON cl.relnamespace = nsp.oid
    JOIN
        pg_catalog.pg_depend dep ON dep.objid = cl.oid AND dep.classid = 'pg_catalog.pg_class'::regclass
    JOIN
        pg_catalog.pg_attrdef ad ON dep.refobjid = ad.adrelid AND dep.refobjsubid = ad.adnum
    WHERE
        pg_catalog.pg_get_expr(ad.adbin, ad.adrelid) LIKE 'nextval%' AND cl.relkind = 'S'
),
column_defaults AS (
    SELECT
        n.nspname AS schema_name,
        c.relname AS table_name,
        a.attname AS column_name,
        pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type,
        COALESCE(pg_catalog.pg_get_expr(d.adbin, d.adrelid), '')::text AS column_default,
        CASE WHEN a.attnotnull THEN 'NO' ELSE 'YES' END AS is_nullable,
        CASE
            WHEN pg_catalog.format_type(a.atttypid, a.atttypmod) LIKE 'character varying%' THEN
                a.atttypmod - 4
            WHEN pg_catalog.format_type(a.atttypid, a.atttypmod) LIKE 'character(%' THEN
                a.atttypmod - 4
            ELSE
                -1
        END AS character_maximum_length,
        CASE
            WHEN a.atttypid = pg_catalog.regtype 'numeric'::regtype THEN
                (a.atttypmod - 4) >> 16
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
                    WHEN (a.atttypmod) = -1 THEN -1
                    ELSE (a.atttypmod - 4) & 65535
                END
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
        a.attgenerated::text as generated_type,
        a.attidentity::text as identity_generation,
        c.oid AS table_oid
    FROM
        pg_catalog.pg_attribute a
    INNER JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
    INNER JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    LEFT JOIN pg_catalog.pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
    WHERE
        n.nspname NOT IN('pg_catalog', 'pg_toast', 'information_schema')
        AND a.attnum > 0
        AND NOT a.attisdropped
        AND c.relkind IN ('r', 'p')
        -- exclude child partitions
        AND c.relispartition = FALSE
),
identity_columns AS (
    SELECT
        n.nspname AS schema_name,
        c.relname AS table_name,
        a.attname AS column_name,
        a.attidentity AS identity_generation,
        s.seqincrement AS increment_by,
        s.seqmin AS min_value,
        s.seqmax AS max_value,
        s.seqstart AS start_value,
        s.seqcache AS cache_value,
        s.seqcycle AS cycle_option
    FROM
        pg_catalog.pg_class c
    JOIN
        pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    JOIN
        pg_catalog.pg_attribute a ON a.attrelid = c.oid
    JOIN
        pg_catalog.pg_class seq ON seq.relname = c.relname || '_' || a.attname || '_seq'
        AND seq.relnamespace = c.relnamespace
    JOIN
        pg_catalog.pg_sequence s ON seq.oid = s.seqrelid
    WHERE
        a.attidentity IN ('a', 'd')
)
SELECT
    cd.schema_name, cd.table_name, cd.column_name, cd.data_type, cd.column_default, cd.is_nullable, cd.character_maximum_length, cd.numeric_precision, cd.numeric_scale, cd.ordinal_position, cd.generated_type, cd.identity_generation, cd.table_oid,
    CASE
        WHEN ls.sequence_oid IS NOT NULL THEN 'SERIAL'
        WHEN cd.column_default LIKE 'nextval(%::regclass)' THEN 'USER-DEFINED SEQUENCE'
        WHEN cd.identity_generation != '' THEN 'IDENTITY'
        ELSE ''
    END AS sequence_type,
    ic.increment_by as seq_increment_by,
    ic.min_value as seq_min_value,
    ic.max_value as seq_max_value,
    ic.start_value as seq_start_value,
    ic.cache_value as seq_cache_value,
    ic.cycle_option as seq_cycle_option,
    pg_catalog.col_description(cd.table_oid, cd.ordinal_position) AS column_comment
FROM
    column_defaults cd
LEFT JOIN linked_to_serial ls
    ON cd.table_oid = ls.adrelid
    AND cd.ordinal_position = ls.adnum
LEFT JOIN identity_columns ic
    ON cd.schema_name = ic.schema_name
    AND cd.table_name = ic.table_name
    AND cd.column_name = ic.column_name
ORDER BY
    cd.ordinal_position;

-- name: GetDatabaseTableSchemasBySchemasAndTables :many
WITH linked_to_serial AS (
    SELECT
        cl.relname AS sequence_name,
        nsp.nspname AS schema_name,
        cl.oid AS sequence_oid,
        ad.adrelid,
        ad.adnum,
        pg_catalog.pg_get_expr(ad.adbin, ad.adrelid)
    FROM
        pg_catalog.pg_class cl
    JOIN
        pg_catalog.pg_namespace nsp ON cl.relnamespace = nsp.oid
    JOIN
        pg_catalog.pg_depend dep ON dep.objid = cl.oid AND dep.classid = 'pg_catalog.pg_class'::regclass
    JOIN
        pg_catalog.pg_attrdef ad ON dep.refobjid = ad.adrelid AND dep.refobjsubid = ad.adnum
    WHERE
        pg_catalog.pg_get_expr(ad.adbin, ad.adrelid) LIKE 'nextval%' AND cl.relkind = 'S'
),
column_defaults AS (
    SELECT
        n.nspname AS schema_name,
        c.relname AS table_name,
        a.attname AS column_name,
        pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type,
        COALESCE(pg_catalog.pg_get_expr(d.adbin, d.adrelid), '')::text AS column_default,
        CASE WHEN a.attnotnull THEN 'NO' ELSE 'YES' END AS is_nullable,
        CASE
            WHEN pg_catalog.format_type(a.atttypid, a.atttypmod) LIKE 'character varying%' THEN
                a.atttypmod - 4
            WHEN pg_catalog.format_type(a.atttypid, a.atttypmod) LIKE 'character(%' THEN
                a.atttypmod - 4
            ELSE
                -1
        END AS character_maximum_length,
        CASE
            WHEN a.atttypid = pg_catalog.regtype 'numeric'::regtype THEN
                (a.atttypmod - 4) >> 16
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
                    WHEN (a.atttypmod) = -1 THEN -1
                    ELSE (a.atttypmod - 4) & 65535
                END
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
        a.attgenerated::text as generated_type,
        a.attidentity::text as identity_generation,
        c.oid AS table_oid
    FROM
        pg_catalog.pg_attribute a
    INNER JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
    INNER JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    LEFT JOIN pg_catalog.pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
    WHERE
        (n.nspname || '.' || c.relname) = ANY(sqlc.arg('schematables')::TEXT[])
        AND a.attnum > 0
        AND NOT a.attisdropped
        AND c.relkind IN ('r', 'p')
        -- exclude child partitions
        AND c.relispartition = FALSE
),
identity_columns AS (
    SELECT
        n.nspname AS schema_name,
        c.relname AS table_name,
        a.attname AS column_name,
        a.attidentity AS identity_generation,
        s.seqincrement AS increment_by,
        s.seqmin AS min_value,
        s.seqmax AS max_value,
        s.seqstart AS start_value,
        s.seqcache AS cache_value,
        s.seqcycle AS cycle_option
    FROM
        pg_catalog.pg_class c
    JOIN
        pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    JOIN
        pg_catalog.pg_attribute a ON a.attrelid = c.oid
    JOIN
        pg_catalog.pg_class seq ON seq.relname = c.relname || '_' || a.attname || '_seq'
        AND seq.relnamespace = c.relnamespace
    JOIN
        pg_catalog.pg_sequence s ON seq.oid = s.seqrelid
    WHERE
        a.attidentity IN ('a', 'd')
)
SELECT
    cd.schema_name, cd.table_name, cd.column_name, cd.data_type, cd.column_default, cd.is_nullable, cd.character_maximum_length, cd.numeric_precision, cd.numeric_scale, cd.ordinal_position, cd.generated_type, cd.identity_generation, cd.table_oid,
    CASE
        WHEN ls.sequence_oid IS NOT NULL THEN 'SERIAL'
        WHEN cd.column_default LIKE 'nextval(%::regclass)' THEN 'USER-DEFINED SEQUENCE'
        WHEN cd.identity_generation != '' THEN 'IDENTITY'
        ELSE ''
    END AS sequence_type,
    ic.increment_by as seq_increment_by,
    ic.min_value as seq_min_value,
    ic.max_value as seq_max_value,
    ic.start_value as seq_start_value,
    ic.cache_value as seq_cache_value,
    ic.cycle_option as seq_cycle_option,
    pg_catalog.col_description(cd.table_oid, cd.ordinal_position) AS column_comment
FROM
    column_defaults cd
LEFT JOIN linked_to_serial ls
    ON cd.table_oid = ls.adrelid
    AND cd.ordinal_position = ls.adnum
LEFT JOIN identity_columns ic
    ON cd.schema_name = ic.schema_name
    AND cd.table_name = ic.table_name
    AND cd.column_name = ic.column_name
ORDER BY
    cd.ordinal_position;

-- name: GetAllSchemas :many
SELECT
    nspname AS schema_name
FROM
    pg_catalog.pg_namespace
WHERE
    nspname NOT IN ('information_schema')
    AND nspname NOT LIKE 'pg_%'
ORDER BY
    schema_name;

-- name: GetAllTables :many
SELECT
    n.nspname AS table_schema,
    c.relname AS table_name
FROM
    pg_catalog.pg_class c
JOIN
    pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE
    c.relkind IN ('r', 'p')
    AND c.relispartition = FALSE
    AND n.nspname NOT IN ('information_schema')
    AND n.nspname NOT LIKE 'pg_%'
ORDER BY
    table_schema,
    table_name;

-- name: GetPostgresRolePermissions :many
SELECT
    tp.table_schema::text as table_schema,
    tp.table_name::text as table_name,
    tp.privilege_type::text as privilege_type
FROM
    information_schema.table_privileges AS tp
WHERE
    tp.table_schema NOT IN ('pg_catalog', 'information_schema')
AND (tp.grantee = current_user OR tp.grantee = 'PUBLIC')
ORDER BY
    tp.table_schema,
    tp.table_name;


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
    (ns.nspname || '.' || t.relname) = ANY(sqlc.arg('schematables')::TEXT[])
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
        n.nspname NOT IN('pg_catalog', 'pg_toast', 'information_schema')
        AND t.typtype IN ('d', 'e', 'c')
),
table_columns AS (
    SELECT
        c.oid AS table_oid,
       CASE
            WHEN t.typtype = 'b' THEN t.typelem  -- If it's an array, use the element type
            ELSE a.atttypid                      -- Otherwise use the type directly
        END AS type_oid
    FROM
        pg_catalog.pg_class c
    JOIN
        pg_catalog.pg_namespace n ON n.oid = c.relnamespace
    JOIN
        pg_catalog.pg_attribute a ON a.attrelid = c.oid
    JOIN
        pg_catalog.pg_type t ON t.oid = a.atttypid
    WHERE
        n.nspname = sqlc.arg('schema')
        AND c.relname = ANY(sqlc.arg('tables')::TEXT[])
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
        'CREATE DOMAIN ' || quote_ident(rct.schema_name) || '.' || quote_ident(rct.type_name) || ' AS ' ||
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
        'CREATE TYPE ' || quote_ident(rct.schema_name) || '.' || quote_ident(rct.type_name) || ' AS ENUM (' ||
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
        'CREATE TYPE ' || quote_ident(rct.schema_name) || '.' || quote_ident(rct.type_name) || ' AS (' ||
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
    WHERE n.nspname = sqlc.arg('schema')
    AND c.relname = ANY(sqlc.arg('tables')::TEXT[])
),
trigger_functions AS (
    SELECT DISTINCT
        n.nspname AS schema_name,
        p.proname AS function_name,
        pg_catalog.pg_get_functiondef(p.oid) AS definition,
        pg_catalog.pg_get_function_identity_arguments(p.oid) AS function_signature
    FROM pg_catalog.pg_trigger t
    JOIN pg_catalog.pg_proc p ON t.tgfoid = p.oid
    JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
    WHERE t.tgrelid IN (SELECT oid FROM relevant_schemas_tables)
    AND NOT t.tgisinternal
),
column_default_functions AS (
    SELECT DISTINCT
        n.nspname AS schema_name,
        p.proname AS function_name,
        pg_catalog.pg_get_functiondef(p.oid) AS definition,
        pg_catalog.pg_get_function_identity_arguments(p.oid) AS function_signature
    FROM pg_catalog.pg_attrdef ad
    JOIN pg_catalog.pg_depend d ON ad.oid = d.objid
    JOIN pg_catalog.pg_proc p ON d.refobjid = p.oid
    JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
    WHERE ad.adrelid IN (SELECT oid FROM relevant_schemas_tables)
    AND d.refclassid = 'pg_proc'::regclass
    AND d.classid = 'pg_attrdef'::regclass
    AND p.oid NOT IN(SELECT objid FROM pg_catalog.pg_depend WHERE deptype = 'e') -- excludes extensions
)
SELECT
    schema_name,
    function_name,
    function_signature,
    definition
FROM
    trigger_functions
UNION
SELECT
    schema_name,
    function_name,
    function_signature,
    definition
FROM
    column_default_functions
ORDER BY
    schema_name,
    function_name;


-- name: GetExtensionsBySchemas :many
SELECT
    e.extname AS extension_name,
    e.extversion AS installed_version,
    n.nspname as schema_name
FROM
    pg_catalog.pg_extension e
LEFT JOIN pg_catalog.pg_namespace n ON e.extnamespace = n.oid
WHERE extname != 'plpgsql' AND (n.nspname = ANY(sqlc.arg('schema')::TEXT[]) OR n.nspname = 'public')
ORDER BY
    extname;


-- name: GetCustomTriggersBySchemaAndTables :many
SELECT
    n.nspname AS schema_name,
    c.relname AS table_name,
    t.tgname AS trigger_name,
    pg_catalog.pg_get_triggerdef(t.oid, true) AS definition
FROM pg_catalog.pg_trigger t
JOIN pg_catalog.pg_class c ON c.oid = t.tgrelid
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE  (n.nspname || '.' || c.relname) = ANY(sqlc.arg('schematables')::TEXT[])
AND NOT t.tgisinternal
ORDER BY
    schema_name,
    table_name,
    trigger_name;

-- name: GetCustomSequencesBySchemaAndTables :many
WITH relevant_schemas_tables AS (
    SELECT c.oid AS table_oid, n.nspname AS schema_name, c.relname AS table_name
    FROM pg_catalog.pg_class c
    JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = sqlc.arg('schema')
    AND c.relname = ANY(sqlc.arg('tables')::TEXT[])
),
columns_with_custom_sequences AS (
  SELECT
		at.attrelid AS table_oid,
		sn.nspname AS sequence_schema_name,
		s.relname AS sequence_name,
		st.nspname AS schema_name,
		t.relname AS table_name,
		at.attname AS column_name
	FROM
		pg_catalog.pg_class s
		JOIN pg_catalog.pg_namespace sn ON sn.oid = s.relnamespace
		JOIN pg_catalog.pg_depend d ON d.refobjid = s.oid
		JOIN pg_catalog.pg_attrdef a ON d.objid = a.oid
		JOIN pg_catalog.pg_attribute at ON at.attrelid = a.adrelid
			AND at.attnum = a.adnum
		JOIN pg_catalog.pg_class t ON t.oid = a.adrelid
		JOIN pg_catalog.pg_namespace st ON st.oid = t.relnamespace
	WHERE
		s.relkind = 'S'
		AND d.classid = 'pg_attrdef'::regclass
		AND d.refclassid = 'pg_class'::regclass
)
SELECT
    rst.schema_name,
    rst.table_name,
    cws.column_name,
    cws.sequence_schema_name,
    cws.sequence_name,
   (
        'CREATE SEQUENCE ' || quote_ident(cws.sequence_schema_name) || '.' || quote_ident(cws.sequence_name) ||
        ' START WITH ' || seqs.start_value ||
        ' INCREMENT BY ' || seqs.increment_by ||
        ' MINVALUE ' || seqs.min_value ||
        ' MAXVALUE ' || seqs.max_value ||
        ' CACHE ' || seqs.cache_size ||
        CASE WHEN seqs.cycle THEN ' CYCLE' ELSE ' NO CYCLE' END || ';'
    )::text AS "definition"
FROM
    relevant_schemas_tables rst
JOIN
    columns_with_custom_sequences cws ON rst.table_oid = cws.table_oid
JOIN
    pg_catalog.pg_sequences seqs ON seqs.schemaname = cws.sequence_schema_name AND seqs.sequencename = cws.sequence_name
ORDER BY
    rst.schema_name,
    rst.table_name,
    cws.column_name;

-- name: GetNonForeignKeyTableConstraintsBySchema :many
SELECT
	pn.nspname AS schema_name,
	c.relname AS table_name,
	pgcon.conname AS constraint_name,
	pgcon.contype::TEXT AS constraint_type,
    -- Collect all columns associated with this constraint, if any
	ARRAY_AGG(kcu.column_name ORDER BY kcu.ordinal_position) FILTER (WHERE kcu.column_name IS NOT NULL)::TEXT [] AS constraint_columns,
	pg_get_constraintdef(pgcon.oid)::TEXT AS constraint_definition
FROM
	pg_catalog.pg_constraint pgcon
	JOIN pg_catalog.pg_namespace pn ON pn.oid = pgcon.connamespace /* schema info */
	JOIN pg_catalog.pg_class c ON c.oid = pgcon.conrelid /* table info */
	LEFT JOIN information_schema.key_column_usage AS kcu ON pgcon.conname = kcu.constraint_name /* column info */
		AND pn.nspname = kcu.table_schema
		AND c.relname = kcu.table_name
WHERE
	pn.nspname = ANY(sqlc.arg('schemas')::TEXT[])
	-- Exclude foreign keys, and partition tables
	AND pgcon.contype != 'f' AND c.relispartition = FALSE
GROUP BY
	pgcon.oid,
	pgcon.conname,
	pgcon.contype,
	pn.nspname,
	c.relname;

-- name: GetForeignKeyConstraintsBySchemas :many
SELECT
    -- Name of the foreign key constraint
    constraint_def.conname AS constraint_name,

    -- Schema of the table that contains the foreign key constraint
    referencing_schema.nspname AS referencing_schema,

    -- Name of the table that holds the foreign key constraint
    referencing_tbl.relname AS referencing_table,

    -- Array of column names in the referencing table involved in the constraint,
    -- ordered by the column's ordinal position (attnum) to maintain the defined column order.
    array_agg(referencing_attr.attname ORDER BY referencing_attr.attnum)::TEXT[] AS referencing_columns,

    -- Array of boolean values indicating whether each referencing column is NOT NULL,
    -- ordered to correspond with the column names.
    array_agg(referencing_attr.attnotnull ORDER BY referencing_attr.attnum)::BOOL[] AS not_nullable,

    -- Schema of the referenced table (the table that the foreign key points to)
    referenced_schema.nspname::TEXT AS referenced_schema,

    -- Name of the referenced table
    referenced_tbl.relname::TEXT AS referenced_table,

    -- Array of column names in the referenced table involved in the foreign key constraint
    ref_columns.foreign_column_names::TEXT[] AS referenced_columns
FROM
    pg_catalog.pg_constraint AS constraint_def
    -- Join to retrieve attributes (columns) for the referencing table based on the constraint definition
    JOIN pg_catalog.pg_attribute AS referencing_attr
      ON referencing_attr.attrelid = constraint_def.conrelid
     AND referencing_attr.attnum = ANY(constraint_def.conkey)
    -- Join to retrieve the referencing table details
    JOIN pg_catalog.pg_class AS referencing_tbl
      ON constraint_def.conrelid = referencing_tbl.oid
    -- Join to retrieve the schema details of the referencing table
    JOIN pg_catalog.pg_namespace AS referencing_schema
      ON referencing_tbl.relnamespace = referencing_schema.oid
    -- Left join to retrieve the referenced table details (if the constraint is a foreign key)
    LEFT JOIN pg_catalog.pg_class AS referenced_tbl
      ON referenced_tbl.oid = constraint_def.confrelid
    -- Left join to retrieve the schema details of the referenced table
    LEFT JOIN pg_catalog.pg_namespace AS referenced_schema
      ON referenced_tbl.relnamespace = referenced_schema.oid
    -- Lateral join to aggregate the names of the columns in the referenced table
    LEFT JOIN LATERAL (
        SELECT
            array_agg(referenced_attr.attname) AS foreign_column_names
        FROM
            pg_catalog.pg_attribute AS referenced_attr
        WHERE
            referenced_attr.attrelid = constraint_def.confrelid
            AND referenced_attr.attnum = ANY(constraint_def.confkey)
    ) AS ref_columns ON TRUE
WHERE
    -- Filter to include only constraints from the provided schemas
    referencing_schema.nspname = ANY(sqlc.arg('schemas')::TEXT[])
    -- Limit results to FOREIGN KEY constraints
    AND constraint_def.contype = 'f'
GROUP BY
    constraint_def.oid,
    referencing_schema.nspname,
    constraint_def.conname,
    referencing_tbl.relname,
    constraint_def.contype,
    referenced_schema.nspname,
    referenced_tbl.relname,
    ref_columns.foreign_column_names;


-- name: GetUniqueIndexesBySchema :many
SELECT
  ns.nspname AS table_schema,                      -- Schema name for the table
  tbl.relname AS table_name,                         -- Name of the table the index belongs to
  idx.relname AS index_name,                         -- Name of the index
  array_agg(col.attname ORDER BY key_info.ordinality)::TEXT[] AS index_columns  -- Comma-separated list of index columns
FROM pg_catalog.pg_class AS tbl
  -- Join to get the schema information for the table
  JOIN pg_catalog.pg_namespace AS ns ON tbl.relnamespace = ns.oid
  -- Join to retrieve index metadata for the table
  JOIN pg_catalog.pg_index AS idx_meta ON tbl.oid = idx_meta.indrelid
  -- Join to get the index object details
  JOIN pg_catalog.pg_class AS idx ON idx_meta.indexrelid = idx.oid
  -- Unnest the index key attribute numbers along with their ordinal positions
  JOIN unnest(idx_meta.indkey) WITH ORDINALITY AS key_info(attnum, ordinality) ON true
  -- Join to get the column attributes corresponding to the index keys
  JOIN pg_catalog.pg_attribute AS col ON col.attrelid = tbl.oid AND col.attnum = key_info.attnum
WHERE ns.nspname = ANY(sqlc.arg('schema')::TEXT[])
  AND idx_meta.indisunique = true
GROUP BY ns.nspname, tbl.relname, idx.relname;

-- name: GetPartitionedTablesBySchema :many
SELECT
	n.nspname AS schema_name,
	c.relname AS table_name,
	c.relispartition AS is_partitioned, -- false for partitioned tables, true for child partitions
	pg_get_partkeydef (c.oid) AS partition_key
FROM
	pg_catalog.pg_class c
	JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
WHERE
	c.relkind = 'p' -- 'p' indicates a partitioned table
	AND n.nspname = ANY(sqlc.arg('schema')::TEXT[]);

-- name: GetPartitionHierarchyByTable :many
SELECT
    child_ns.nspname          AS schema_name,
    child_cls.relname         AS table_name,
    parent_ns.nspname         AS parent_schema_name,
    parent_cls.relname        AS parent_table_name,
    COALESCE(
        pg_get_expr(child_cls.relpartbound, child_cls.oid),
        ''
    )::TEXT AS partition_bound
FROM pg_partition_tree(sqlc.arg('table')::TEXT) AS t
JOIN pg_catalog.pg_class child_cls
    ON child_cls.oid = t.relid
JOIN pg_catalog.pg_namespace child_ns
    ON child_ns.oid = child_cls.relnamespace
LEFT JOIN pg_catalog.pg_class parent_cls
    ON parent_cls.oid = t.parentrelid
LEFT JOIN pg_catalog.pg_namespace parent_ns
    ON parent_ns.oid = parent_cls.relnamespace;

