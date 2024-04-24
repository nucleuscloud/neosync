---
title: PostgreSQL
id: postgres
hide_title: false
slug: /connections/postgres
---

## Introduction

Postgres is one of the most commonly used databases in the world and Neosync natively supports most postgres-compatible databases.

The following guide will show you how to configure and test your Postgres Connection.

## Configuring Postgres

In order to connect to your Postgres database, first navigate to **Connections** and then click **+ New Connection**.

![newconn](/img/pgnew.png)

Then select a Postgres compatible database such as Neon, Supabase or just the base Postgres connection.

![conn](/img/conn.png)

You'll now be taken to the Postgres connection form.

First, name your connection in the **Connection Name** field.

Next, decide how you want to connect to your Postgres database. You can configure your connection by providing a connection URL or by entering in the different connection parameters.

To connect using the connection URL, simply paste the connection url in the **Connection URL** input.

![conn](/img/pgstring.png)

To connect using the host and connection parameters, click on the **Host** radio button. You'll see a form appear with the different components of the connection string broken out into individual input fields.

| Field             | Description                                                                                                                                                          |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Host Name         | The address of the server where your PostgreSQL database is hosted. Use 'localhost' for a local database or enter the IP address or domain name for a remote server. |
| Database Port     | PostgreSQL's default port is 5432. If your database listens on a different port, specify it here.                                                                    |
| Database Name     | Enter the name of the database you wish to connect to. The database should already exist on the PostgreSQL server.                                                   |
| Database Username | The username with which you will log in to the database. Ensure it has the necessary permissions for reading or writing data.                                        |
| Database Password | Enter the password associated with the username. Make sure this information is correct for data security.                                                            |
| SSL Mode          | Choose an SSL mode based on your company's security policies. 'Require' enforces SSL connections for enhanced security.                                              |

Complete the fields in order to connect to your Postgres Database.

![conn](/img/pghost.png)

## Permissions

This section details the Postgres role permissions necessary for Neosync to function properly.

This will vary based on the connection, and this section details the minimum permissions required for a source and/or destination connection.
You'll be able to validate your role checks by testing your connection, which can be seen in the below section.

### Source Connections

A source connection is used for sync jobs where you want to synchronize data from one connection to another.
The source connection will be used in a readonly manner, and as such, only requires the `SELECT` permission set on the table or tables that will be synced.

This can be done with the `GRANT` capability in SQL.

Example:

```sql
GRANT SELECT ON public.users TO myrole;
```

Note: Depending on your database configuration, if Neosync is returning permission denied or unable to access any tables, you may need to grant usage on the schema in question.
This can be done for your schema, similar to the example below:

```sql
GRANT USAGE on SCHEMA public TO myrole;
GRANT USAGE ON SEQUENCE my_sequence TO myrole; -- or GRANT USAGE ON ALL SEQUENCE TO myrole;
```

### Destination Connections

A destination connection is used for sync and generate jobs, where data will be inserted into the configured connection.
This requires slighly more permissions, but you can get away with more or less depending on what you are looking to do.

At a bare minimum, this connection requires `CREATE, UPDATE` on all tables that will be written to.
You will also need to grant permissions to any `sequences`, `triggers`, or `functions` that may be invoked during the insertion or update process.

Example:

```sql
GRANT CREATE, UPDATE on public.users TO myrole;

GRANT USAGE ON FUNCTION <name> TO myrole; -- optional if you have functions
GRANT USAGE ON SEQUENCE <name> TO myrole; -- optional if you have sequences
```

If you are planning to allow Neosync to truncate data prior to a job run, then the `TRUNCATE` permissions will need to be added.

```sql
GRANT TRUNCATE on public.users TO myrole;
```

If you are planning to allow Neosync to initialize tables within a schema, you will need to grant more permissions in order to do so.

```sql
GRANT CREATE ON SCHEMA public TO myrole;
```

## Testing your Connection

Once you've configured your connection either using the connection parameters or using the connection url, click on **Test Connection** in order to test the connectivity to your connection.

When you click **Test Connection** we attempt to connect the database and query the `information_schema` in order to retrieve the schemas and tables that the configured role has access to.
You should validate that Neosync can see all of the schemas and tables that you'd like to work with. Otherwise, you may have to update your permissions or use a different role.

Based on the connection type (source or destination) - you may see varying values here. Consult the permissions section above for more information on what you should expect to see based on how you've configured your role.

**Please note that this is not fully encompassing and only checks permissions directly on tables themselves.
This does not currently include functions, sequences, etc.**

If you are running in to issues with permissions, please consult us on Discord.

A successful connection will return something like this:

![conn](/img/pgpermissions.png)
