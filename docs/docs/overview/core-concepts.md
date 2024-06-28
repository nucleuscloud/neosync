---
title: Core Concepts
description: The best way to learn about Neosync is to understand the core concepts of the platform.
id: core-concepts
hide_title: false
slug: /core-concepts
---

## Introduction

The best way to learn about Neosync is to understand the core concepts of the platform.

### Jobs

![job](https://assets.nucleuscloud.com/neosync/docs/jobs-page.png)

Jobs are async workflows that transform data and sync it between source and destination systems. They can run on a set schedule or run ad-hoc and can be paused at any time. Under the covers, Neosync uses [Temporal](https://github.com/temporalio/temporal) as our job scheduling and execution engine and [Benthos](https://github.com/benthosdev/benthos) as our data transformation engine. Temporal handles all of the execution, retries, backoffs and the coordination of tasks within a job. While Benthos handle the data sync'ing and transformation.

Jobs also have <strong>types</strong>. Today, we support two types:

1. Sync Jobs - Replicate and anonymize data between a source and a destination such as two databases.
2. Generation Jobs - Generate synthetic data and sync it to a destination such a database or S3.

You can create multiple jobs with different schedules, schemas and settings. You can also update a jobs connections, schema, transformers and subsets after you've already created the job.

### Runs

![runs](https://assets.nucleuscloud.com/neosync/docs/runs-page.png)

<strong>Runs</strong> are instances of a job that have been executed. Runs can
be paused and restarted at any time. Neosync exposes a lot of useful metadata
for each run which can help you understand if a run completed successfully and
if not, what exactly went wrong. Runs also provide an audit trail of activity
that you can track to see what jobs executed and their status.{' '}

### Connections

![connections](/img/connectionsList.png)

Connections are integrations with upstream and/or downstream systems such as Postgres, S3 and Mysql. Jobs use connections to move data across systems. Connections are created outside of jobs so that you can re-use connections across multiple jobs without re-creating it every time.

We plan to continue to expand the number of connections we offer so that we can cover most major systems. As we mentioned above, we use Benthos for our core data sync'ing and transformation. Benthos ships with a number of pre-built [connections](https://www.benthos.dev/docs/components/inputs/about) that we will work to support.

### Transformers

![transformers](https://assets.nucleuscloud.com/neosync/docs/udt-home.png)

Transformers are data-type specific modules that anonymize or generate data. Transformers are defined in the job workflow and are applied to every piece of data in the column they are assigned. Neosync ships with a number of transformers already built that handle common data types such as email, physical addresses, ssn, strings, integers and more. You can also create your own custom transformers.
