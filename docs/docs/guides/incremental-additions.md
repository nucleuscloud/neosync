---
title: Using Neosync to sync incremental data
description: Learn how to use Neosync to sync incremental batches of data instead of a full refresh
id: incremental-data-sync
hide_title: false
slug: /guides/incremental-data-sync
---

## Intro

Data syncs generally have two modes:

1. Full refresh
2. Incremental additions

In this guide, we'll take a look at how to accomplish both in Neosync.

## Destination Options

![new-trans](/img/destoptions.png)

Sync modes are defined in the destination options in the **Connect** page. When you select a destination, you'll see a number of configuration options appear under that destination.

Each destination can have different destination options. For example, one destination database might do a full refresh, while another does incremental. Destination options do differ by destination.

Here is a matrix showing support for each destination option under the **Option** column and which destinations support it.

| Option                      | Description                                                                     | Postgres | MySQL | SQL Server | DynamoDB | MongoDB | S3  |
| --------------------------- | ------------------------------------------------------------------------------- | -------- | ----- | ---------- | -------- | ------- | --- |
| Truncate Before Insert      | Truncates table before inserting data                                           | ✅       | ✅    | ✅         | ❌       | ❌      | ❌  |
| Init Table Schema           | Creates table(s) and their constraints. The database schema must already exist. | ✅       | ✅    | ❌         | ❌       | ❌      | ❌  |
| On Conflict Do Nothing      | If there is a conflict when inserting data do not insert                        | ✅       | ✅    | ❌         | ❌       | ❌      | ❌  |
| Skip Foreign Key Violations | Insert all valid records, bypassing any that violate foreign key constraints.   | ✅       | ✅    | ✅         | ❌       | ❌      | ❌  |
| Truncate CASCADE            | Truncate cascade all tables                                                     | ✅       | ❌    | ❌         | ❌       | ❌      | ❌  |

## Full Refresh

A full refresh means that you're sync'ing all of your data every single time (less any subsets) when you run a sync. This is the default mode that Neosync runs on.

Depending on how much data you have, this may not be feasible. Especially if you're syncing to a smaller database or local environment.

These are the default destination options when you configure a job.

In order to do a full refresh, ensure that your destination has the following config:

- On Conflict Do Nothing - false
- Truncate Before Insert - true
- Truncate Cascade - true

Now whenever you run a sync, the sync will truncate all of the data in the destination and insert a full refresh (less any subsets.)

## Incremental Additions

Incremental additions are great ways to get a smaller, more manageable and potentially more refined data set. There are two ways to accomplish incremental data syncs in Neosync:

1. Using the On Conflict Do Nothing destination option
2. Subsetting

Let's go through both.

### Using the On Conflict Do Nothing destination option

The On Conflict Do Nothing destination option tells Neosync to skip inserting a record if there is a conflict on `INSERT`. The most common reason there is a conflict is because the record already exists given a particular constraint like a `UNIQUE` or `PRIMARY KEY` constraint.

This option can be used to do incremental data additions because on every record conflict it will skip trying to insert the record and move onto the next one. While this might be inefficiently since it has to try to insert every record once, it does result in an incremental data addition since the new records won't conflict with existing records.

#### Subsetting

The other way to do incremental data additions is to use Subsetting. Using subsetting, you can filter your data set before you sync it by passing in SQL filters such as `WHERE created_at > 2024-08-09`. You can create filters for any column and combine multiple WHERE filters. You can even do joins!

This is generally more efficient and more flexible way to do incremental data refreshes. It's also, generally, a more powerful way to subset your data.

# Conclusion

Full refreshes and incremental data syncs are two ways to sync data using Neosync. There are plenty of valid use-cases for both and depending on your requirements, one may make more sense than the other. If you have other requirements or questions, please don't hesitate to submit a feature request or talk to us in Discord.
