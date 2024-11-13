---
title: Creating a Sync Job
description: Learn how to create a data sync job to anonymize and transform data from a source database to one or multiple destination databases
id: creating-a-sync-job
hide_title: false
slug: /guides/creating-a-sync-job
---

## Introduction

In this guide we will walk through how to create a [sync job](/core-concepts#jobs). Sync jobs are used to sync data between a source and one or many destinations. Some usecases of sync jobs are:

1. Syncing and anonymizing prod data to lower level environments
2. Syncing data between two lower level environments with no transformations
3. Syncing and anonymizing data to be used for analytical and machine learning use cases such as training a model

## Creating a Sync Job

In order to create a sync job:

1. On the **Jobs** page, click on the **+ New Job** button.

![jobs](https://assets.nucleuscloud.com/neosync/docs/jobs-page.png)

2. Select the **Data Synchronization** job type.

![job-type](/img/third.png)

3. Then give your job a **Name**. Next, if you want your job to run on a schedule, click on the schedule switch to expose an input where you can provide a cron string. Your job will run on this schedule. Lastly, activate the **Initiate Job Run** switch if you want to immediately trigger a single job run once the job is completed. Click **Next** once you're ready.

![job-define](https://assets.nucleuscloud.com/neosync/docs/new-sync-job-definition.png)

4. Select your source and destination(s) connections. You may only select one source but you can select multiple destinations. You may also configure your source and destination with the provided configuration options.

![job-connect](https://assets.nucleuscloud.com/neosync/docs/new-sync-job-connections.png)

5. Next is the Schema page. Here you can select how you want to transform your tables and columns with [**Transformers**](/core-concepts#transformers). There are a number of [transformers](/transformers/system) that Neosync ships with out of the box or you can create your own custom transformer.

![job-schema](/img/second.png)

6. Lastly, you can configure a [subset](/core-features#subsetting). A subset is a way to filter the data that is being synced to the destination(s). A common use-case is to filter the data to reduce the size or dimensionality of the data. You can subset the data using WHERE filters by typing in the filter in the filter box. At the same time, you'll see your `WHERE` filter being constructed and you can click on the **Validate** button to validate that the subset query will successfully execute against the schema. Click **Next** once you're done.

![job-subset](https://assets.nucleuscloud.com/neosync/docs/new-sync-job-subset.png)

7. Congrats! You successfully created a job. From here, you will be taken to the Job Details page where you can pause, resume, run or update the job you created.
