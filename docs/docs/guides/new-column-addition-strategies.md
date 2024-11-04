---
title: New Column Addition Strategies
description: Learn how to configure Neosync to handle new columns detected during a job run
id: new-column-addition-strategies
hide_title: false
slug: /guides/new-column-addition-strategies
# cSpell:words Automap
---

## Introduction

When a Neosync Job is configured for relational databases, all columns for each selected table must have a transformer mapping configured.
This page goes into detail how Neosync handles new columns that may be added to your source database in-between updating a job.

This is a common occurrence for any company that is adding new columns to a database and may not update Neosync straight away.

## Driver Support

| Strategy | Description                                                                                     | PostgreSQL | MySQL | MS SQL Server |
| -------- | ----------------------------------------------------------------------------------------------- | ---------- | ----- | ------------- |
| Halt     | Stops the job run if a new column is detected that is not found in the configured job mappings. | ✅         | ✅    | ✅            |
| AutoMap  | Automatically generates a fake value. See more below.                                           | ✅         | ❌    | ❌            |
| Continue | Ignores new columns; may fail if column doesn't have default. See more below.                   | ✅         | ✅    | ✅            |

## Halt Strategy

This strategy is plain and simple. During the job run, Neosync compares the configured job mappings with the source database.
For the selected tables in the job mappings, a diff is made and if a column is found in the source connection that doesn't exist in the job mappings, the run is halted.

## Continue

This strategy tells Neosync to ignore any difference in job mappings from the source database.

Neosync is able to detect that new columns were added in the source, but it will leave them off of the insert statement.
This may result in failures if any unmapped columns do not have a column default in the destination connection.
However, any additional columns that have a default or are a generated column will not result in a job run failure.

| Example                                                                                                 | Success |
| ------------------------------------------------------------------------------------------------------- | ------- |
| `ALTER TABLE ADD COLUMN foo TEXT NOT NULL DEFAULT "test" `                                              | ✅      |
| `ALTER TABLE ADD COLUMN foo TEXT NULL DEFAULT NULL`                                                     | ✅      |
| `ALTER TABLE ADD COLUMN full_name TEXT GENERATED ALWAYS AS (first_name \|\| ' ' \|\| last_name) STORED` | ✅      |
| `ALTER TABLE ADD COLUMN foo TEXT NOT NULL`                                                              | ❌      |

## Auto Map

Automap is a smart strategy that attempts to do what it can to prevent PII from leaking or from failure modes with additional columns being added.

Not all data types are currently supported and will continue to receive updates over time to improve data type support.

The algorithm works as follows:

1. If the column has a DB Default or is Generated, use the database default
2. If the column is Nullable, set the column to null.
3. Based on the data type, generate a proper value that will fit within that column.
4. If an unsupported data type is detected, halt the run.

### Postgres

Postgres has many data types and not all of them are currently supported in the auto map mode. Support will continue to increase over time.

<!-- cspell:disable  -->

| Data Type        | Support | Generator       |
| ---------------- | ------- | --------------- |
| smallint         | ✅      | GenerateInt64   |
| integer          | ✅      | GenerateInt64   |
| bigint           | ✅      | GenerateInt64   |
| decimal          | ✅      | GenerateFloat64 |
| numeric          | ✅      | GenerateFloat64 |
| real             | ✅      | GenerateFloat64 |
| double precision | ✅      | GenerateFloat64 |
| serial           | ✅      | GenerateDefault |
| smallserial      | ✅      | GenerateDefault |
| bigserial        | ✅      | GenerateDefault |
| money            | ✅      | GenerateFloat64 |
| char             | ✅      | GenerateString  |
| varchar          | ✅      | GenerateString  |
| text             | ✅      | GenerateString  |
| bytea            | ❌      |                 |
| timestamp        | ❌      |                 |
| timestamptz      | ❌      |                 |
| date             | ❌      |                 |
| time             | ❌      |                 |
| timetz           | ❌      |                 |
| interval         | ❌      |                 |
| boolean          | ✅      | GenerateBool    |
| point            | ❌      |                 |
| line             | ❌      |                 |
| lseg             | ❌      |                 |
| box              | ❌      |                 |
| path             | ❌      |                 |
| polygon          | ❌      |                 |
| circle           | ❌      |                 |
| cidr             | ❌      |                 |
| inet             | ❌      |                 |
| macaddr          | ❌      |                 |
| bit              | ❌      |                 |
| tsvector         | ❌      |                 |
| uuid             | ✅      | GenerateUuid    |
| xml              | ❌      |                 |
| json             | ❌      |                 |
| jsonb            | ❌      |                 |
| int4range        | ❌      |                 |
| int8range        | ❌      |                 |
| numrange         | ❌      |                 |
| tsrange          | ❌      |                 |
| tstzrange        | ❌      |                 |
| daterange        | ❌      |                 |
| oid              | ❌      |                 |
| text[]           | ❌      |                 |

<!-- cspell:enable  -->
