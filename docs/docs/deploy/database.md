---
title: Neosync Database Setup
description: Learn how to configure a Postgres database for use with Neosync
id: database
hide_title: false
slug: /deploy/database
# cSpell:words NOSUPERUSER,NOCREATEDB,NOCREATEROLE,NOLOGIN,NOREPLICATION,NOBYPASSRLS,NOSUPERUSER,NOCREATEDB,NOCREATEROLE,NOREPLICATION
---

## Introduction

This section provides some color to the Postgres environment that is required to run the Neosync database migrations.

Generally, the administration of the database is left as an exercise to the user, but this page details some minimum requirements needed for success.

Neosync Cloud and Neosync Open Source currently run everything against Postgres 15. It may work on earlier versions and will probably work on later versions, but this is a disclaimer that as of June 2024, Neosync is currently setup for Postgres 15.

## Neosync Migrations

Neosync API comes equipped with a set of SQL schema migration files that are used to assert the Neosync database on boot up.
The exact way this is done will vary highly based on your environmental setup as well as risk profile with regards to permissions.

The default setup provided by the open source compose files is to stand up a postgres database using pretty bone default permissions.
This will work fine for development, but is generally not recommended for production deployments.

### Auto Migration Mode

As a part of this, one may run Neosync API in the `DB_AUTO_MIGRATE=true` mode, where, on startup, Neosync API will run the database migration scripts prior to binding to its port. The same credentials used to do this will be what are used to generally connect to the Neosync API during standard crud operations.

### Split Migration Mode

An alternative mode, and the mode that the Helm charts run in by default, are to split the migrations out from the actual API container. This is done today via an init container that will come up first prior to the API coming online. The benefits here are that the init container can run under a different set of database credentials that have elevated permissions to alter the database and schema, where as the API may only have permissions to CRUD the tables.

The `DB_MIGRATIONS_OPTIONS` environment variable may be provided to the migrations container (or the API container if running in auto migrate mode) to allow passing `options` to the postgres url.
This gives one the ability to set the role of the connection. This is extremely useful if you are using more granular RBAC roles in your database and you want the owner of the tables to be different than the standard user that will be inserting records into them.

## Which Migration library is used?

Neosync uses [golang-migrate](https://github.com/golang-migrate/migrate) to handle running database migrations. We wrap it inside of our own CLI so that we can easily run it in containerized environments.

By default, migrate will create a `schema_migrations` table in the `public` schema, which is used to track which migrations have already been invoked.
This table can be changed by providing the `DB_MIGRATIONS_TABLE` environment variable. For example: `DB_MIGRATIONS_TABLE=neosync_api_schema_migrations`.

This is useful if you're already using golang-migrate in your database and don't want conflicts.

## Base Permissions and other requirements

Neosync creates the `schema_migrations` table in the `public` schema and also creates a `neosync_api` custom schema that it dumps all of its table into.
As of the writing of this article, the `neosync_api` schema is hardcoded into the codebase and cannot easily be changed. If this is important to you that it be something else, please reach out to us on Discord.

### Custom Role

> **NB:** This section details an example setup to showcase a more advanced configuration that contains hierarchal roles. It is not necessarily meant for a production usecase.

If you aren't running the database migrations with a SUPERUSER, and instead with a custom role, you'll need to give it some base permissions that coincide with the resources it creates that are detailed in the summary of the parent section.

Let's create a sample role:

```sql
CREATE ROLE neosync_owner WITH
        NOSUPERUSER
        NOCREATEDB
        NOCREATEROLE
        INHERIT
        NOLOGIN
        NOREPLICATION
        NOBYPASSRLS
        CONNECTION LIMIT -1
        VALID UNTIL 'infinity';

```

The following permissions will allow you to run the database migrations successfully.

```sql
GRANT CREATE ON SCHEMA public TO neosync_owner;
GRANT USAGE ON SCHEMA public TO neosync_owner;
GRANT CREATE ON DATABASE postgres TO neosync_owner;
```

We can run the migrations with an admin user that inherits from the `neosync_owner` role.

```sql
CREATE ROLE neosync_admin WITH
    LOGIN
    INHERIT
    PASSWORD 'password'
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOREPLICATION
    NOBYPASSRLS
    CONNECTION LIMIT -1
    VALID UNTIL 'infinity';
GRANT neosync_owner TO neosync_user;
```

In order to retain the ownership of the tables to the `neosync_owner`, you'll need to ensure the `DB_MIGRATIONS_OPTIONS=-c role=neosync_owner` is set.
This will run the migrations with the correct role and ownership permissions.
