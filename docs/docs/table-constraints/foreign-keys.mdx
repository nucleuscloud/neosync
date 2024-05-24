---
title: Foreign Keys
description: Learn how Neosync handles foreign keys when anonymizing and subsetting data
id: foreign-keys
hide_title: false
slug: /table-constraints/foreign-keys
---

## Introduction

Foreign keys are a great way to enforce relationships between tables and columns and are found in most relational databases. Neosync natively supports foreign key relationships when syncing data from a source to a destination(s) database.

If you have foreign keys in your database, they will appear in the Schema Mapping page with a light orange `Foreign Key` badge:

![foreignkey](/img/fk.png)

If you hover over the badge with your mouse, you can see the reference to the Parent column:

![foreignkeyhover](/img/fkhover.png)

## Transforming Foreign Keys

There are many instances where you might want to transform a column that has a foreign key reference. These are typically columns that have sensitive data and need to be anonymized. Transforming foreign key references is a tricky task because you have to transform the parent column and child column in the exact same way to ensure that you don't break the reference.

![foreignkey](/img/fkref2.png)

In the example above, we have a source column that is storing email addresses. We want to anonymize `bill@gmail.com`. But the email column has a foreign key reference to another table. In the top example, we show that if you transform the email address in the parent column without transforming the email in the child column it breaks the foreign key reference.

In the bottom example, you can see that if you transform the value in the parent column, you must transform the value in the child column in the same way in order to not break the foreign key reference.

Neosync will also update any foreign keys connected to the primary key to reflect the change. However, it's important to note that for this automatic updating to work, the foreign key must be set to `passthrough` or it can be set to a custom javascript transformer but you must be careful in order to ensure that the parent column is set to the same custom javascript transformer.

This setting allows Neosync to correctly identify and update the foreign keys, ensuring that the integrity and relationships within your database are preserved even after primary key changes.
