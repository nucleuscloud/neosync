---
title: MongoDB
description: MongoDB is a source-available, cross-platform, document-oriented database program.
id: mongodb
hide_title: false
slug: /connections/mongodb
# cSpell:words textareas
---

## Introduction

MongoDB is one of the most commonly used NoSQL database in the world and Neosync has emerging support for it.

This connector is still experimental and we are actively adding features to MongoDB to bring it up to the same feature parity as Postgres where relevant.

If you are interested in using MongoDB but don't see a feature that is required for you to use, please reach out to us on Discord.

## Configuring MongoDB

![Mongo New Connection Page](/img/mongonew.png)

This guide will help you to configure your MongoDB database connection properly.

**Connection Name**: Enter a unique name for this connection that you'll easily recognize. This is just a label and does not affect the connection itself.
**URL**: Enter your database connection url that will be used to connect to Mongo. Neosync supports both `mongodb` and `mongodb+srv` protocols.

### TLS Authentication

Mongo allows connection via Client TLS in lieu of or in addition to a username and password.

Your setup will vary based on your specific settings, but generally, you can configure Neosync to connect to Mongo by providing a Client Certificate and Client Key in their respective textareas in the Connection(s) form.

By doing so, Neosync will store them on disk and update the connection url to point to those files so that they can be used to connect.

The `tls=true` query parameter will also be added automatically if both of those fields are specified and the `tls` parameter has not been explicitly set in the connection url.
If using Mongo Atlas, the pem file will contain both the certificate and key. They must be split and put into their respective fields in the Connection form.

You may need to add a few query parameters to your URL in order to properly connect.
For example, using Mongo Atlas, you may need to provide the following parameters: `authMechanism=MONGODB-X509&authSource=$external`.

Read more about the Mongo TLS Options [here](https://www.mongodb.com/docs/manual/reference/connection-string/#tls-options).

## Permissions

### Source Connections

When configuring your MongoDB user, please ensure that you grant it read permissions to relevant Database Collections that you want Neosync to have access to.

### Destination Connections

When configuring your MongoDB user, please ensure that you grant it necessary write permission to the relevant Database and Collections that Neosync will need access to.
