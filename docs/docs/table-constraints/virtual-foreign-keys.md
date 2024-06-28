---
title: Virtual Foreign Keys
description: Learn how to add virtual foreign keys in Neosync
id: virtual-foreign-keys
hide_title: false
slug: /table-constraints/virtual-foreign-keys
---

## Introduction

Neosync supports virtual foreign keys, which can be added using the Neosync app. Virtual foreign keys allow you to create relationships between tables that do not have explicit foreign key constraints in the database. This is particularly useful when dealing with legacy databases or when you need to create temporary relationships for data syncing purposes.

To add a virtual foreign key in Neosync, follow these steps:

**Navigate to the Schema Mapping Page:**

- Go to the Transformer Mapping page in the Neosync app where you define your data mappings. Click on the `Virtual Foreign Keys` tab.

**Define the Relationship:**

- Specify the source and target columns that you want to link with the virtual foreign key. Ensure that these columns logically relate to each other based on your data structure.

**Save the Virtual Foreign Key:**

- After defining the virtual foreign keys click update to save. The Neosync app will now treat these columns as if they had a real foreign key constraint, maintaining the integrity of your data during syncing operations.

:::info

1. When transforming the source (primary key) of a virtual foreign key the target (foreign key) must be set to passthrough.
2. The column datatypes must align for the source and target of virtual foreign keys
3. The source (primary key) column of a virtual foreign key must be non-nullable and unique.

:::

## Transforming & Subsetting Virtual Foreign Keys

Transforming and subsetting virtual foreign keys behave in the exact same way as foreign keys defined in a database.
Please refer to [Foreign Keys](foreign-keys.md) and [Subsetting](subsetting.md)
