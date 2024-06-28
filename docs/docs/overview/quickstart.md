---
title: Quick start
description: Learn how to seed a PostgreSQL database with synthetic data for better data security and privacy
id: quickstart
hide_title: false
slug: /quickstart
---

## Getting started

In this quick start, we're going to walk through how you can seed a PostgreSQL database with synthetic data for better data security and privacy in your staging and local environments. This guide is meant to be the shortest path to setting up Neosync and running your first Job.

For this quick start, we're going to use [Neon](https://neon.tech) as our PostgreSQL provider. Neosync supports any PostgreSQL compatible database. Neosync also supports MySQL compatible databases. However, this quick start will focus on Postgres but the two can be configured very similarly.

Let's get started.

## Setting up Neon

First, create a [Neon](https://neon.tech) account by clicking on **Sign up** and signing in with an email/password or OAuth provider of your choice.

Once you're signed in, create a new project. If you don't have a Neon account then give your project and database a name and select a region like below:

![neon-create-project](/img/neon-create.png)

Then, click on "Create Project". You'll be prompted with a pop-up containing the connection string to your newly created database. We'll come back to this later, so you can close this for now.

Next, you'll need to define your database schema. For this demo, we'll just create one table.

On the left-hand menu, click on **SQL editor**. In the SQL editor code input box, click into the box and press **Command + A** on your keyboard to highlight all of the text, then click **Delete** to delete everything.

Here is the SQL script I ran to create the **Users** table in the public schema. Create your table by copying and pasting the script below into the SQL code editor box in Neon and click **Run**.

```sql
CREATE TABLE public.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    age INTEGER NOT NULL
);
```

You can do a quick sanity check by going to **Tables** and expanding the **Users** menu to check that the table creation was successful.

![neon-created-tables](/img/neon-show-table.gif)

Nice! Okay, last step for Neon. Let's get our connection string so we can connect to Neon from Neosync.
We can find our connection string by going to **Dashboard**.
The **Connection Details** section, under the **Connection String** sub-header, will be the connection string. Click on the copy icon to copy it to your computer's clipboard.

## Setting up Neosync

Let's get started with Neosync. First, sign up for a [Neosync account](https://app.neosync.dev).

Sign in with an email/password or Google OAuth. Once you're logged in, you'll be directed to the **Jobs** page.

Next, you'll want to create a connection to your Neon database and then create a Job to generate data. Let's get started.

### Creating a Connection

Navigate to **Connections** -> **New Connection** then click on **Neon**.

![neosync-connect-form](/img/neon-integration.png)

Let's name your connection **neon-db**. Then click on the **Connection URL** tab and paste in the Neon database connection string you copied earlier from the Neon dashboard into the **Connection URL** field.

Once you've pasted in the string, you can click on **Test Connection** to test that you're connected. You should see this if it passes:

![neosync-test](/img/neon-test.png)

Click on **Submit** to move onto the last part.

### Creating a Job

In order to generate data, we need to create a **Job** in Neosync. Let's click on **Jobs** and then click on **+ New Job**. We're now presented with two options:

![neosync-test](/img/data-gen.png)

- Data Synchronization - Synchronize and anonymize data between a source and destination.
- Data Generation - Generate synthetic data from scratch for a chosen destination.

Since we're seeding a table from scratch, you can select the **Data Generation** Job and click **Next**.

Name your job **neon-sync** and then set **Initiate Job Run** to **Yes**. We can leave the Schedule and Advanced Settings alone for now.

![neosync-test](/img/define.png)

Click **Next** to move onto the **Connect** page. Select the Neon connection you previously connected from the Destination dropdown.

![neosync-test](/img/neon-connect.png)

There are some other options here that can be useful in the future but we'll skip these for now and click **Next**.

Now for the fun part.

First, decide how many rows you want to create. For this run, we'll do 1000 rows.

Next, in the **Table Selection** selection, click on the **public.users** checkbox then click on the **right arrow** to move the table from the Source -> Destination.

![neosync-test](/img/rows-table.gif)

Once you've completed this, you'll see the **Transformer Mapping** table populate with the columns of the table.

Lastly, we need to determine what kind of synthetic data we want to create and map that to our schema. Neosync has **Transformers** which are ways of creating synthetic data. Click on the **Select a Transformer** button in the **Transformer Mapping** table row to select a Transformer.

Here is what to set up for the users table.

| Column     | Transformer           | Options          |
| ---------- | --------------------- | ---------------- |
| ID         | Generate UUID         | Default          |
| first_name | Generate First Name   | Default          |
| last_name  | Generate Last Name    | Default          |
| email      | Generate Email        | Default          |
| age        | Generate Random Int64 | Min: 18, Max: 40 |

For the age column, use the `Generate Random Int64` Transformer to randomly generate ages between 18 and 40. You can configure that by clicking on the pencil icon next to the transformer and setting your min and max.

![neosync-test](/img/update-age.gif)

Now that we've configured everything, you can click on **Submit** and create the Job! We'll get routed to the Job page and the Job will start to run.
After a few seconds, the **Status** will update to **Complete** in the **Recent Job Runs** table.
You may need to manually click the reload button next to the **Recent Job Runs** heading.

![neosync-test](/img/success-job.png)

Success!

Now we can head back over to Neon and check on our data. First, let's check the count and make sure we generated 1000 rows.

Click on **SQL Editor** on the left-hand menu and in the SQL editor code box, delete any existing queries and then copy and paste the following SQL query and click **Run**.

```sql
SELECT COUNT(*) FROM users;
```

![neosync-test](/img/users-count.png)

Nice, we wanted to generate 1000 rows and the count tells us we did.

Next, let's check the data. Delete the previous query from the SQL editor code box and copy and paste the following query:

```sql
SELECT * FROM users;
```

![neosync-test](/img/data-users.png)

Checking the results, we have the columns that we defined above (id, first_name, last_name, email, age).

Looking pretty good! We have seeded our Neon database with 1000 rows of completely synthetic data!

## Conclusion

In this quick start, we walked through how to seed a PostgreSQL database with 1000 rows of synthetic data using Neosync. This is just a small test and you can expand this to generate tens and hundreds of thousands or even millions of rows of data across any relational database. Neosync handles the referential integrity.

Lastly, if you want to anonymize existing data then it's a similar workflow. The only difference is that you select the [Data Sync Job](https://www.neosync.dev/blog/neosync-neon-sync-job) and select a destination database.
