create schema information_schema;
CREATE TABLE information_schema.columns (
  table_schema text not null,
  table_name text not null,
  column_name text not null,
  column_type text not null,
  column_key text not null,
  ordinal_position bigint not null,
  column_default text null,
  is_nullable text not null,
  data_type text not null,
  character_maximum_length bigint,
  numeric_precision bigint,
  numeric_scale bigint,
  extra longtext null,
  generation_expression longtext null
);

create table information_schema.tables (
  table_schema text not null,
  table_name text not null,
  auto_increment bigint
);

create table information_schema.schemata (
  catalog_name text not null,
  schema_name text not null,
  default_character_set_name text,
  default_collation_name text,
  sql_path text,
  default_encryption text
);

create table information_schema.key_column_usage (
  constraint_name text not null,
  table_schema text not null,
  table_name text not null,
  column_name text not null,
  ordinal_position bigint not null,
  referenced_table_schema text not null,
  referenced_table_name text not null,
  referenced_column_name text not null
);

create table information_schema.referential_constraints (
  constraint_schema text not null,
  constraint_name text not null,
  table_name text not null,
  referenced_table_name text not null,
  update_rule text not null,
  delete_rule text not null
);

create table information_schema.table_constraints (
  table_schema text not null,
  constraint_name text not null,
  constraint_type text not null,
  table_name text not null,
  column_name text not null
);

create table information_schema.table_privileges (
  table_schema text not null,
  table_name text not null,
  grantee text not null,
  privilege_type text not null
);

create table information_schema.statistics (
  table_schema text not null,
  table_name text not null,
  column_name text null,
  expression text null,
  index_name text not null,
  index_type text not null,
  seq_in_index bigint,
  nullable text not null
);

create table information_schema.routines (
  routine_name text not null,
  routine_schema text not null,
  dtd_identifier text not null,
  routine_definition longtext not null,
  is_deterministic text not null,
  created timestamp not null,
  last_altered timestamp not null
);

create table information_schema.triggers (
  trigger_name text not null,
  trigger_schema text not null,
  event_object_schema text not null,
  event_object_table text not null,
  action_statement longtext not null,
  event_manipulation text not null,
  action_orientation text not null,
  action_timing text not null,
  created timestamp not null
);

create table information_schema.check_constraints (
  constraint_schema text not null,
  constraint_name text not null,
  check_clause text not null
);

create table information_schema.user_privileges (
  privilege_type text not null,
  grantee text not null
);

create table information_schema.schema_privileges (
  table_schema text not null,
  privilege_type text not null,
  grantee text not null
);

create table admin_privileges (
  privilege_type text not null,
  grantee text not null
);

create table db_privileges (
  table_schema text not null,
  privilege_type text not null,
  grantee text not null
);
