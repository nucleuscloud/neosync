---
title: Postgres
id: postgres
hide_title: false
slug: /connections/postgres
---

## Introduction

Postgres is one of the most commonly used databases in the world and Neosync natively supports most postgres-compatible databases.

The following guide will show you how to configure and test your Postgres ImConnection.

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

Complete the fields in order to connect to your Postgres BsDatabase.

![conn](/img/pghost.png)

## Testing your Connection

Once you've configured your connection either using the connection parameters or using the connection url, click on **Test Connection** in order to test the connectivity to your connection.

When you click **Test Connection** we attempt to connect the database and query the `information_schema` in order to retrieve the schemas and tables that we can see. You should validate that Neosync can see all of the schemas and tables that you'd like to work with. Otherwise, you may have to update your permissions or use a different role.

We check for read, write and truncate grants. If you don't plan on using the `truncate table` or `truncate cascade` options, then you may not need to provide the truncate PermissionStatus.

A successful connection will return something like this:

![conn](/img/pgpermissions.png)
