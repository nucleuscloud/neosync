---
title: sync
description: Learn how to sync data to a local destination with the neosync sync CLI command.
id: sync
hide_title: false
slug: /cli/sync
---

# neosync sync

## Overview

Learn how to sync data to a local destination with the neosync sync CLI command.

The `neosync sync` command is used to sync data from a neosync connection to a local destination.
Supported sources are currently postgres, mysql connections and AWS S3 Sync Job.
Supported are currently postgres and mysql.

## Usage

```bash
neosync sync
```

## Options

The following options can be passed using the `neosync sync` command:

- `--api-key` - Neosync API Key. Takes precedence over `$NEOSYNC_API_KEY`
- `--config` - Path to yaml config. Defaults to `neosync.yaml` in current directory.
- `--connection-id` - Neosync connection id for sync data source. Takes precedence over config.
- `--job-id` - Neosync job id for sync data source. For [AWS S3, GCP Cloud Storage] jobs only. Takes precedence over config.
- `--job-run-id` - Neosync job run id for sync data source. For [AWS S3, GCP Cloud Storage] jobs only. Takes precedence over config.
- `--destination-connection-url` - Local destination connection url to sync data to. Takes precedence over config.
- `--destination-driver` - Destination connection driver (postgres, mysql). Takes precedence over config.
- `--truncate-before-insert` - Truncates the table before inserting data. This will not work with Foreign Keys.
- `--truncate-cascade` - Truncate cascades to all tables. Only supported for postgres.
- `--init-schema` - Creates the table schema and its constraints.
- `--on-conflict-do-nothing` - If there is a conflict when inserting data into SQL database do not insert.
- `--output` - Sets output type (auto, plain, tty). (default `auto`).
- `--debug` - Sets the log level to debug and prints much more information. Works best with `--output plain`.
- `--destination-idle-duration` - Maximum amount of time a connection may be idle (e.g. '5m')
- `--destination-idle-limit` - Maximum number of idle connections
- `--destination-open-duration` - Maximum amount of time a connection may be open (e.g. '30s')
- `--destination-open-limit` - Maximum number of open connections
- `--destination-max-in-flight` - Maximum allowed batched rows to sync. If not provided, uses server default of 64
- `--destination-batch-count` - Batch size of rows that will be sent to the destination.
- `--destination-batch-period` - Duration of tim e that a batch of rows will be sent.

## Yaml Config File

To persist settings, a yaml config may be enabled. It can be provided like so:

```
neosync sync --config ./path/to/config.yaml
```

> **NB:** Flags will take precedence over values provided in the config.

```yaml
source:
  connection-id: d9dc020d-746b-48c1-9319-a165a25ac32e
  connection-opts:
    job-run-id: 43a4aac5-c4a8-4e4f-8554-03b7c5fffc04-2024-06-14T17:43:24Z
destination:
  connection-url: user:pass@tcp(127.0.0.1:3306)/database
  driver: mysql
  truncate-before-insert: false
  truncate-cascade: false
  init-schema: false
  on-conflict:
    do-nothing: false
  connection-opts:
    open-limit: 25 # remove to unset and use system default (CLI falls back to default of 25 if not provided)
    idle-limit: 2 # remove to unset and use system default
    idle-duration: 30s # remove to unset and use system default
    open-duration: 5m # remove to unset and use system default
  max-in-flight: 10
  batch:
    count: 100
    period: 5s
```

## Circular Dependencies

**Support for Circular Dependencies**: The CLI sync feature in Neosync is capable of managing both self-referencing circular dependencies and those involving multiple tables.
In scenarios where the source data is not from a SQL database (like AWS S3) but the destination is a SQL database, Neosync utilizes the foreign key constraints of the destination
SQL database to effectively insert data. This approach ensures data integrity and respects the relational structure of the SQL database.

**Nullable Columns:** For circular dependencies to work, at least one table involved in the dependency must have a column that is nullable.

**Foreign Key Dependencies and Table Constraints:** While a CLI sync does not modify table constraints, it synchronizes data based on foreign key dependencies.

**Data Insertion and Updating Process:** Sync jobs first performs an initial data insertion. Subsequently, it updates the columns involved in the circular dependency.

## Syncing from AWS S3

To synchronize data from a Neosync job with AWS S3 as the destination, you must provide either a job ID or job run ID. Using a job ID will sync data from the most recent job run.
During this process, the table constraints from the `destination-connection-url` database are used to determine the correct order for syncing the data. This ensures that the data is
synchronized in a way that respects the relational structure and integrity of the local database.
