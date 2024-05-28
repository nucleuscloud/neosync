---
title: Initializing your Schema
description: Learn how to use Neosync to initialize your schema in your destination database to make it easier to hydrate your database with data
id: initializing-your-schema
hide_title: false
slug: /guides/initializing-schemas
---

## Background

Neosync has a feature that will "Initialize" your schema. However this feature is not always the best way to get your schema set up. This guide will walk you through ways to set up your schema manually.

## Option 1: Run your Migrations

If you have a set of migrations that you use to set up your database, you can run these migrations on your target database. This is the most common way to set up your schema.

This can potentially have some issues if you have a lot of migrations or if your migrations are not idempotent.

If your migrations are any of the following you may run into issues:

- Not Idempotent, meaning that running the migration multiple times (or against a completely blank/fresh database) will cause issues
- Have hard coded database names (if you are choosing to run Neosync against a separate logical database in your single instance)
- Edit database users, global configurations or other settings that are not related to the schema/data

Some of the above issues can be resolved by choosing to run your migrations against a completely fresh database.

Steps to run this method are out of scope for the Neosync docs, but ideally it is as simple as running your migrations against the same database as the Neosync destination.

## Option 2: `pg_dump` your Schema

Option 2 is to use `pg_dump` to dump your schema and then apply it manually to your target database.

This has a few advantages over the above method:

- You can ensure that your schema is exactly the same as your source
- You can remove certain things that cause hangups with Neosync (and inserts in general) like certain triggers, views, and permissions.

To do this, you can run the following command:

```bash
pg_dump -s -h localhost -U postgres -d my_database --schema-only > my_schema.sql
```

Then edit the `my_schema.sql` file to remove anything that is tripping up Neosync, save it in a separate file, and execute it against your target database after the sync is finished.

You also can use this method to fine tune which constraints you want to be honored when doing the sync, in some cases it may be faster to insert with no constraints then apply them after.
