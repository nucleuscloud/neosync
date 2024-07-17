create schema information_schema;
CREATE TABLE information_schema.columns (
  table_schema text not null,
  table_name text not null,
  column_name text not null,
  ordinal_position bigint not null,
  column_default text null,
  is_nullable text not null,
  data_type text not null,
  character_maximum_length bigint,
  numeric_precision bigint,
  numeric_scale bigint,
  extra longtext null
);

create table information_schema.tables (
  table_schema text not null,
  table_name text not null
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
