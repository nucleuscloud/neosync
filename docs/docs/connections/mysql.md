---
title: Mysql
description: Neosync supports most Mysql-compatible databases natively using the Mysql connection.
id: mysql
hide_title: false
slug: /connections/mysql
---

Neosync is an open-source, developer-first product that allows you to create anonymized, secure test data that you can sync across all of your environments for high quality local, stage and CI testing.

Neosync supports most Mysql-compatible databases natively using the Mysql connection.

## MySQL Database Connection Configuration

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
