---
title: Troubleshooting Neosync Jobs
description: Learn how to troubleshoot common errors in Neosync with this guide
id: troubleshooting
hide_title: false
slug: /guides/troubleshooting
---

## Introduction

Databases are _complex_. Neosync does its best to handle any and all edge cases you throw at it.

This document is a collection of common issues you may face when running Neosync and how to resolve them.

### `Sync: could not successfully complete sync activity: read tcp 0.0.0.0:42454->0.0.0.0:5432: i/o timeout`

This is likely due to an overabundance of connections against the database.

You can resolve this by reducing the number of open connections in the "Connections" tab of Neosync. The default is at 80, the easiest way to debug is to start with something small like 10 and increase it until you hit the error again.

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
