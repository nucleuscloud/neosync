---
title: Mysql
description: Neosync supports most Mysql-compatible databases natively using the Mysql connection.
id: mysql
hide_title: false
slug: /connections/mysql
---

Neosync is an open-source, developer-first product that allows you to create anonymized, secure test data that you can sync across all of your environments for high quality local, stage and CI testing.

Neosync supports most Mysql-compatible databases natively using the Mysql connection.

## Things to watch out for

1. Currently Neosync only supports syncing between two physical databases and does not support syncing between two logical databases. This is on the roadmap and will be supported soon.
2. When syncing across two databases, the databases must have the same name.

## MySQL Database Connection Configuration

### Connection URL

A connection url may be provided. DSN format is preferred, limited support for URI format.

### Environment Variable

To connect using the environment variable, simply paste the environment variable in the **Environment Variable** input.

The value of the environment variable must be in the `Connection URL` format.

This is only available in the OSS version of Neosync. The environment variable must begin with `USER_DEFINED_`.
This is for safety and is to limit the class of environment variables a user of Neosync may configure.

For full support, the environment variable must live on both the `neosync-api` as well as `neosync-worker`.

### Discrete Host Parameters

![mysql](https://assets.nucleuscloud.com/neosync/docs/mysql.png)

This guide will help you to configure your MySQL database connection properly.

**Connection Name**: Enter a unique name for this connection that you'll easily recognize. This is just a label and does not affect the connection itself.

**Host Name**: Enter the server address. 'localhost' is used for a local server. For remote servers, use the IP address or domain name.

**Database Port**: MySQL's default port is 3306. Change it if your server uses a different port.

**Database Name**: The name of the database you want to connect to. It should exist on your MySQL server.

**Database Username**: Your database username. It needs to have the appropriate permissions for the operations you intend to perform.

**Database Password**: The password for the database user. Accuracy is essential for security.

**Connection Protocol**: Select the protocol you wish to use to connect to your database. 'tcp' is commonly used for network connections.

Test the connection before saving to ensure all details are correct and the system can connect to the database.

You may also specify a direct DSN via the `URL` tab instead of the split out `Host` view.

## TLS

Neosync has support for Regular TLS (one-way) as well as mTLS (two-way).

This is configured via the `Client TLS Certificates` section on the database configuration page.

If you simply wish to verify the server certificate, only the `Root certificate` is required.

If wishing to have the client present a certificate, you must specify both the `Client key` as well as the `Client certificate`.
If only one of these is provided, the Neosync will reject the configuration.

The following TLS/SSL modes are available for Mysql via the `tls` query parameter.

> **NB:** if using the `URL` configuration, you will need to specify this directly in the query parameters. If using the host configuration, be sure to select the correct option in the dropdown that you intend to use.

```console
true - Enabled TLS/SSL encryption to the server
false - Disables TLS
skip-verify - If you want to use a self-signed or invalid certificate on the server-side. Self-signed may be use if using mTLS.
preferred - Use TLS only when advertised by the server
```

The `server name` _must_ be provided if using `tls=true` otherwise the client will not have enough information to fully verify the host and will fail connection. If this isn't desired, use `tls=skip-verify`.

## Go Mysql Driver

Neosync uses the `go-sql-driver/mysql` for Mysql support.

HTTP urls are not very well supported and you will find much better luck, using the older `DSN` format.

For a full look at query parameters available to you, check the [driver readme](https://github.com/go-sql-driver/mysql?tab=readme-ov-file#parameters)

By default, Neosync will add the following query parameters automatically to your user-provided DSN at runtime:

- `multiStatements=true`

  - This is used to enable sending batched SQL statements to the server at once. If this is turned off, Neosync has to send single statements at a time, which really hurts performance.

- `parseTime=true`
  - This configures the driver to automatically convert date and time values to go's `time.Time` object for better data handling through the system.
