---
title: Troubleshooting Neosync Jobs
description: Learn how to troubleshoot common errors in Neosync with this guide
id: troubleshooting
hide_title: false
slug: /guides/troubleshooting
---

## Introduction

Databases are _complex_. Neosync does its best to handle any and all edge cases you throw at it. This document is a collection of common issues you may face when running Neosync and how to resolve them.

It's worth reviewing the Neosync [Platform](/platform) overview to understand Neosync's architecture and where possible issues can come up. Generally speaking, issues can come from three different places:

1. **Frontend** - Typically client-side react issues that are fairly rare.
2. **Backend** - API and Database issues that result from some sort of control plane error such as not being able to find jobs or connect to a source/destination. More common than front-end errors but still not the most common of errors.
3. **Worker** - Issues that come up when trying to sync and transform data. This is where the overwhelming majority of issues occur.

Since most issues occur in the worker, let's dive a little deeper there. Let's take a look at the types of worker-related issues:

1. **Orchestration** - Neosync leverages Temporal under the surface as a durable job execution engine. In some situations, Temporal times out if a sync is taking too long because of a large table or something going wrong. In the Job Settings, there are advanced settings to control the number of retries, timeout period and more. Generally, any timeout or errors related to maximum attempts are issues related to the sync taking too long.
2. **Data Syncing and Transformation** - Most of the worker errors fall into this category. This comprises insert/update errors, constraint errors, duplicate key errors and more. These usually come from a combination of data and schema errors.

## Common errors

### `Sync: could not successfully complete sync activity: read tcp 0.0.0.0:42454->0.0.0.0:5432: i/o timeout`

This is likely due to an overabundance of connections against the database.

You can resolve this by reducing the number of open connections in the "Connections" tab of Neosync. The default is at 50, the easiest way to debug is to start with something small like 10 and increase it until you hit the error again.

### `Maximum attempts reached`

This usually occurs when a job is trying to run and failing and due to the retry policy eventually gets canceled because it's already retried too many times. This is usually indicative of a data issue that is happening in the worker that is causing the job to fail, retry, fail, retry, etc.

## Activity start to close timeout

An activity timeout defines the maximum time allowed for a single activity. In Neosync, an activity is responsible for syncing one table. If that table is very big and takes longer than the activity timeout, then you may see this error.

One way to remedy this error is to increase the **Table Sync Timeout** in the job's advanced settings. The default value is 10 minutes but depending on your database, it may need to be longer.

### `could not successfully complete sync activity: pq: insert or update on table \"table_name\" violates foreign key constraint \"table_name_id_foreign_key\"`

This is likely due to records in the source database, that are added _after_ the table on one side of the constraint is synchronized. If you have tables that are constantly in flux on the source database, and your job takes a long time to run this error happens more frequently.

The easiest way to resolve it is to copy or snapshot your source connection to a fresh database with no other connections, and run the job again.

### `ERROR: schema \"schema_name\" does not exist`

Neosync does not support creating schemas on the fly. You must create the schema on the target database before running the job.

### `unable to exec truncate cascade statements: ERROR: relation \"public.table_name\" does not exist`

When you have "Truncate" enabled in the Source Options, Neosync will attempt to truncate the table before inserting new records. If the table does not exist, this error will be thrown.

You can resolve this by disabling the "Truncate" option in the Source Options, and enabling the "Init Table Schema" for your first run, and Neosync will attempt to create the table for you.

### `could not successfully complete sync activity: pq: CONCURRENTLY cannot be used when the materialized view is not populated`

If you have views with triggers that refresh said views on inserts you may see this error when attempting to create records in the table that the view is based on.

The easiest way to resolve this, is to disable and/or remove the triggers on the view, and re-enable them after the job has completed.

### Networking issues with Docker

Some Linux environments have issues with Docker Networking, which may cause problems with Neosync communications.
There have been reports of this on Ubuntu and GCP.

This can possibly be fixed by updating the docker networks to use the correct MTU size.
Be sure to update all of the networks within the compose. Example below:

```yml
networks:
  neosync-network:
    name: neosync-network
    driver_opts:
      com.docker.network.driver.mtu: 1460
```

This topic is further discussed [here](https://www.civo.com/learn/fixing-networking-for-docker) and [here](https://stackoverflow.com/questions/73101754/docker-change-mtu-on-the-fly).

## Debugging Database Queries

Neosync currently supports debugging database queries for Postgres connections.

Database query logging is by default turned off, which is equivalent to setting `DB_LOG_LEVEL=none`.

This log level works similar to Neosync's standard `LOG_LEVEL` where you can set values to emit certain levels of database logging.
To see only queries that result in an error, set the `DB_LOG_LEVEL` to `error`.
If you'd like to see all queries, set the log level to `INFO` or `DEBUG`.

Valid options for `DB_LOG_LEVEL` are (not case sensitive):

- trace
- debug
- info
- warn
- error
- none

**Note** - When turning on database logging, the statements **include arguments** default.

So if planning to run this in any production environment, you may leak PII or other sensitive information.

To disable arguments from being listed in the query, enable the `DB_OMIT_ARGS=true` environment variable.

All database logs will be grouped under the `db` attribute key.
