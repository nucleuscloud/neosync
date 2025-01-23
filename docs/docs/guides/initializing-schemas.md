---
title: Initializing your Schema
description: Learn how to use Neosync to initialize your schema in your destination database to make it easier to hydrate your database with data
id: initializing-your-schema
hide_title: false
slug: /guides/initializing-schemas
---

## Introduction

Neosync has the ability to initialize your database schema prior to running a sync. This is useful if you have a lot of migrations or if you want to ensure that your schema is exactly the same as your source.

## Driver Support

Postgres, MySQL, and MS SQL Server are supported.

> MS SQL Server requires an Enterprise License for OSS users.

## Option 1: Use Neosync's Initialize Schema Feature

The simplest way to get started is to simply enable the feature in your job's configuration.

Once this is enabled, Neosync will generate SQL statements based on the schemas and tables you've selected in your job's configuration to be created prior to syncing the data itself.

Neosync tries to be as smart as possible about what to include in the schema, but it is not always perfect. If you find that you are missing something, you can always manually add it to the schema, or reach out to us to help you out.

> Some actions like schema creation, extension installation, Sequence creation and resets require schema owner access.

## Option 2: Run your Migrations

If you have a set of migrations that you use to set up your database, you can run these migrations on your target database. This is a common way to set up your schema.

This can potentially have some issues if you have a lot of migrations or if your migrations are not idempotent.

If your migrations are any of the following you may run into issues:

- Not Idempotent, meaning that running the migration multiple times (or against a completely blank/fresh database) will cause issues
- Have hard coded database names (if you are choosing to run Neosync against a separate logical database in your single instance)
- Edit database users, global configurations or other settings that are not related to the schema/data

Some of the above issues can be resolved by choosing to run your migrations against a completely fresh database.

Steps to run this method are out of scope for the Neosync docs, but ideally it is as simple as running your migrations against the same database as the Neosync destination.

## Option 3: `pg_dump` your schema (or other driver equivalents)

Use `pg_dump` to dump your schema and then apply it manually to your target database.

This has a few advantages over the above method:

- You can ensure that your schema is exactly the same as your source
- You can remove certain things that cause hangups with Neosync (and inserts in general) like certain triggers, views, and permissions.

To do this, you can run the following command:

```bash
pg_dump -s -h localhost -U postgres -d my_database --schema-only > my_schema.sql
```

Then edit the `my_schema.sql` file to remove anything that is tripping up Neosync, save it in a separate file, and execute it against your target database after the sync is finished.

You also can use this method to fine tune which constraints you want to be honored when doing the sync, in some cases it may be faster to insert with no constraints then apply them after.
