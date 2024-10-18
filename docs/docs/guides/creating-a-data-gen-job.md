---
title: Creating a Data Generation Job
description: Learn how to create a data generation job in Neosync which allows you to generate synthetic data from scratch
id: creating-a-data-gen-job
hide_title: false
slug: /guides/creating-a-data-gen-job
---

## Introduction

In this guide we will walk through how to create a [data generation job](/core-concepts#jobs). Data generation jobs are used to populate a database or datastore with freshly created synthetic data. Some usecases of data generation jobs are:

1. Creating training data for machine learning usecases such as training a model
2. Augmenting your existing database with more data for performance and scalability testing
3. Generating data for demo environments

## Creating a Data Generation Job

In order to create a data generation job:

1. On the **Jobs** page, click on the **+ New Job** button.

![jobs](https://assets.nucleuscloud.com/neosync/docs/jobs-page.png)

2. Select the **Data Generation** job type.

![job-type](/img/first.png)

3. Then give your job a **Name**. Next, if you want your job to run on a schedule, click on the schedule switch to expose an input where you can provide a cron string. Your job will run on this schedule. Lastly, activate the **Initiate Job Run** switch if you want to immediately trigger a single job run once the job is completed. Click **Next** once you're ready.

![job-define](https://assets.nucleuscloud.com/neosync/docs/new-data-gen-job-define.png)

4. Select your destination(s) connection. You may also configure your destination with the provided configuration options.

![job-connect](https://assets.nucleuscloud.com/neosync/docs/new-data-gen-job-connect.png)

5. Next is the Schema page. Here you can select how you want to transform your tables and columns with [**Transformers**](/core-concepts#transformers). Select your schema and the table you want to transform and then the number of rows you want to generate. There are a number of [transformers](/transformers/system) that Neosync ships with out of the box or you can create your own custom transformer. Once you're done, you can click **Next**.

![job-schema](/img/fourth.png)

7. Congrats! You successfully created a job. From here, you will be taken to the Job Details page where you can pause, resume, run or update the job you created.
