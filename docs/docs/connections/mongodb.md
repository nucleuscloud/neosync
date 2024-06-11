---
title: MongoDB
description: MongoDB is a source-available, cross-platform, document-oriented database program.
id: mongodb
hide_title: false
slug: /connections/mongodb
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

## Permissions

### Source Connections

When configuring your MongoDB user, please ensure that you grant it read permissions to relevant Database Collections that you want Neosync to have access to.

### Destination Connections

When configuring your MongoDB user, please ensure that you grant it necessary write permission to the relevant Database and Collections that Neosync will need access to.
