---
title: Schema Initialization Overview
description: Learn how Neosync handles schema initialization
id: schema-initialization-overview
hide_title: false
slug: /schema-init/overview
---
## Introduction
Before launching a data sync, it's essential that your destination database is prepared to handle and mirror the structure of your source. Proper schema initialization lays the foundation by establishing all necessary tables, data types, constraints, indexes, views, sequences, and triggers, ensuring a smooth and consistent data migration process.

## SQL Server Considerations

When using SQL Server as your destination database, please note the following limitations regarding triggers during schema initialization:

- Encrypted triggers are not supported. If a trigger is encrypted, its definition will not be available for schema generation.
- CLR triggers (i.e., triggers implemented using the Common Language Runtime) are not supported. Such triggers will be omitted from the generated schema since they are not written in plain T-SQL.

To ensure a complete and accurate schema initialization, verify that your triggers are implemented using standard T-SQL without encryption or CLR enhancements. Otherwise, the trigger will not be included in the generated schema.
