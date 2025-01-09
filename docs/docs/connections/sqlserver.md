---
title: Microsoft SQL Server
description: Microsoft SQL Server is a proprietary relational database management system developed by Microsoft.
id: sqlserver
hide_title: false
slug: /connections/sqlserver
# cSpell:words MSSQL,myuser,mypass,mydbhost,mydatabase,Passw,mssqldb
---

## Introduction

Microsoft SQL Server (MSSQL) is a proprietary relational database management system developed by Microsoft.

This page will document various items of interest regarding MS SQL Server and how to use it to synchronize data with Neosync.

## Configuring MS SQL Server when running Neosync locally with Docker Compose

If you're trying out Neosync locally and are a MS SQL Server user, we make it easy to quickly stand up sql server instances
that will _just work_ with the Neosync setup.

The `compose.yml` has some commented out compose files that may be uncommented to enable or disable various databases.

In the `include` section, you'll find something like this:

```yml
- path: ./compose/compose-db.yml
# - path: ./compose/compose-db-mysql.yml
# - path: ./compose/compose-db-mssql.yml
```

We can change this around to enable the MS Sql Server compose.

```yml
- path: ./compose/compose-db.yml
# - path: ./compose/compose-db-mysql.yml
- path: ./compose/compose-db-mssql.yml
```

Note: If you wish to disable the `./compose/compose-db.yml`, you may, but you may want to also disable the `api-seed` and `temporal-seed` containers as those are set up to automatically seed and populate the test postgres databases.

Afterwards, run `make compose/up` to stand up Neosync with two MS SQL Server containers that are available within the neosync docker network.
They are also both accessible via your laptop on ports `1433` and `1434`.

You can also get away with just one if you wish to only stand up one MSSQL container as SQL Server is capable of database virtualization and is not required to have two physically separate containers for a basic sync.

If you don't want to modify the compose, the compose command may be run directly with the MSSQL snippet layered on.

```console
docker compose -f compose.yml -f compose/compose-db-mssql.yml up -d
```

## Configuring a MSSQL Connection in Neosync

![Configuring a MSSQL Connection](/img/mssql-connection.png)

The form is pretty straightforward and depending on your setup, you may or may not need some of the options.

### Connection Name

Enter a unique name for your MSSQL connection. This is something that you'll see again when configuring a Neosync Job.

### Connection URL

This is the full database url that Neosync will use to connect to your database. This should be in URI format.

Example: `sqlserver://myuser:mypass@mydbhost:1433?database=mydatabase`.

This format supports servers that do not contain a port (if they're behind a web proxy) as well as servers that may or may not require access credentials.

If you've hooked up local Neosync and are wanting to connect to one of the containers, the url would look like this:

Note: the url listed below is using the docker DNS name defined in the compose file.
To access the database from your laptop, you would simply replace `test-prod-db-mssql` with `localhost`.

> **NB:** It is imperative that you provide a `database` in the connection URL. This will define what MSSQL database Neosync connects to.

```bash
sqlserver://sa:YourStrong@Passw0rd@test-prod-db-mssql:1433?database=master
```

Due to MSSQL's design, it is possible to have two separate databases within the same physical database instance and sync between the two. They will simply need to be configured as two separate Neosync connections.

### Environment Variable

To connect using the environment variable, simply paste the environment variable in the **Environment Variable** input.

The value of the environment variable must be in the `Connection URL` format.

This is only available in the OSS version of Neosync. The environment variable must begin with `USER_DEFINED_`.
This is for safety and is to limit the class of environment variables a user of Neosync may configure.

For full support, the environment variable must live on both the `neosync-api` as well as `neosync-worker`.

### Max Open Connection Limit

This is by default set to 50, but it's important to look into this and set this value to the appropriate size given the size of your database as it changes based on the machine type. 50 may be too high, or too low! Ultimately, this will affect how quickly Neosync is able to sync to or from your database as it takes advantage of parallelization where it can.

### Bastion Host Configuration

This section is optional, but may be required for many that use MSSQL databases hosted within cloud infrastructure.
It's generally not a good practice to publicly expose a database and to instead use a bastion host, or jump box. This facilitates Neosync to use SSH tunneling through this jump box to gain access to your database.

If you require a bastion host and need help setting one up or configuring one, check out out [this tutorial from Azure](https://learn.microsoft.com/en-us/azure/private-link/tutorial-private-endpoint-sql-portal) guide.

**Note** For Azure deployments: when trying to connect to your SQL Server instance in a private VPC with a bastion host - make sure to add the database server to the DB user (using an `@`) you're using to connect.

For example, if your database user is `johndoe` and your database host is `test-sql-server.database.windows.net`. Then your connection string should be:

`sqlserver://johndoe@test-sql-server.database.windows.net:<password>@test-sql-server.database.windows.net:1433?database=dev`

Notice how the username has `@<hostname>` appended to it. Once successfully connected, you should see a window like this:

![Connecting](/img/sqlconn.png)

### TLS

Neosync has support for Regular TLS (one-way) as well as mTLS (two-way).

This is configured via the `Client TLS Certificates` section on the database configuration page.

If you simply wish to verify the server certificate, only the `Root certificate` is required.

If wishing to have the client present a certificate, you must specify both the `Client key` as well as the `Client certificate`.
If only one of these is provided, the Neosync will reject the configuration.

The following TLS/SSL modes are available for MSSQL via the `encrypt` query parameter.

> **NB:** Neosync does not automatically add the `encrypt` query parameter if client certificates have been detected. Regardless of this configuration, the `encrypt` parameter is up to the user based on their bespoke configuration and setup.

> **NB:** The current version of the Go driver does _not_ support `encrypt=strict` mode as it has not yet implemented TDS8 yet. [go-mssqldb#87](https://github.com/microsoft/go-mssqldb/issues/87).

```
strict - Data sent between the client and server is encrypted E2E using `TDS8`
disable - Data sent between client and server is not encrypted
false / optional / no / 0 / f - Data sent between client and server si not encrypted beyond the login packet (Default)
true / mandatory / yes / 1/ t - Data sent between client and server is encrypted
```

The `server name` _must_ be provided if using `encrypt=true` otherwise the client will not have enough information to fully verify the host and will fail connection. If this isn't desired, set `trustServerCertificate=true` in your query parameters to allow the client to automatically trust the certificate presented by the server.

### Go Mssql Driver

Neosync uses the official `go-mssqldb` driver provided by Microsoft.

You can find DSN format as well as a full list of supported query parameters by visiting their [Readme](https://github.com/microsoft/go-mssqldb/?tab=readme-ov-file#connection-parameters-and-dsn).

## Permissions

When creating a new connection or checking an existing one, you can click `Test Connection` on the form to check to see if Neosync can connect.

It will first see if the database is connectable, then it will check to see what schemas and tables it has access to within that database.

The response will be something like the following:

![MSSQL Permissions](/img/mssql-permissions.png)

Permission may be different depending on the connection type you are configuring.

Source connections may get away with simple READ permissions on the databases they are reading from.

Destination connections will of course need READ, WRITE permissions so that they can actually insert data.

As more MSSQL features get added to Neosync such as the ability to create schemas, tables as wel as delete or truncate data prior to a run, more permissions for the destination database will be necessary.

## Data Truncation

MSSQL supports table truncation, but only for constraints that do not contain foreign keys.
Due to this, Neosync will run a DELETE FROM on each table in reverse constraint order.
Afterwards, it will reset any identity columns to their original state.

DELETE FROM may take a long time. Depending on your data size, it may be more advantageous for you to drop and re-create the database out of band such that Neosync may only handle the syncing aspect until support is added for drop and recreate in-app.

## Limitations

MSSQL overall has less features that other databases like PostgreSQL and MySQL have easier support for.

That being said, we want to bring MSSQL as close to feature parity as possible with these other databases.
Below is a small section of a few spots that we are actively working on bring MSSQL up to snuff with the other relational DB counterparts.

### Row Conflicts

MSSQL doesn't have a native `ON CONFLICT UPDATE` feature and we are actively working on supporting a similar feature such that no data truncation can occur and only delta rows are added to the dataset with no truncation or deletion being required.
