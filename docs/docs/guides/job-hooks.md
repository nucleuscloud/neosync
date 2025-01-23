---
title: Job Hooks
description: Learn how to use Job Hooks to add further customization to your Neosync jobs
id: job-hooks
hide_title: false
slug: /guides/job-hooks
# cSpell:words POSTSYNC,PRESYNC
---

## Introduction

Job Hooks are a way to add further customization to your Neosync jobs.

## Neosync Version Availability

Job Hooks are available for all accounts in Neosync Cloud.

For OSS users, Job Hooks are only available with a valid Enterprise license.

## How to configure Job Hooks

This section will cover how to configure hooks in the Neosync UI.
They can also be configured via the API as well as via the Neosync Terraform provider.

### Getting there

Job Hooks can be configured after a job itself has been created. After creation, navigate to the hooks tab for your job of choice.

![Job Hooks Overview](/img/hooks/job-hooks-overview.png)

### Creating a new hook

From here a new hook may be created. Click on the new hook button and you'll be presented with a new hook form to fill out.

There are a few different configuration options available to you to further fine tune when the hook runs.

Today, only `SQL` hooks are supported, with plans for web hooks in the future.

![New Hook Form](/img/hooks/new-hook-form.png)

Choosing the hook's priority will determine the order in which the hook is executed.

This is useful if you have multiple hooks that need to run in a specific order.

Furthermore, the timing of when the hook runs may also be configured. Today, there are two options available:

- `Pre Sync`: Runs before the first table sync, truncation, and schema initialization.
- `Post Sync`: Runs after the last table syncs (effectively right before the job is marked as complete).

Any SQL connection configured in the job (source or destination) will be available to the hook.

SQL queries are run in an `Exec` manner, meaning that their results are not returned and are ignored. Today, Neosync only checks for errors. This is planned to change in the future where Neosync can handle the returned results to allow users to perform further actions like result verification.

## Execution order strategy

The priority of the hook (0-100). This determines the execution order. Lower values are higher priority (priority=0 is the highest).

Tie Breaking is determined by the following: `(priority, created_at, id)` in ascending order.

This means that if two hooks have the same priority, the hook that was created first will run first.
If the created*at timestamp is the same, which \_could* happen given the right conditions (e.g. if both hooks are created at the same time via API or script), the tie will be broken by the hook's ID.

Hook IDs are uuids, so it is luck of the draw as to which UUID is ordered first (but this will always be consistent once both UUIDs have been created).

If this is of concern, the easiest solution is to simply increase or decrease the priority of the hook in question.

## Hooks in Job Run Details

Each hook timing runs as its own separate activity, which appears in the job run details under the activity timeline and table.

Simply looking for the `RunJobHooksByTiming` activity will show whether or not the hooks ran to success.

In the activity logs, you'll see info messages that look like this:

```console
[INFO] - scheduling "TIMING_POSTSYNC" RunJobHooksByTiming for execution`
[INFO] - completed 0 "TIMING_POSTSYNC" RunJobHooksByTiming
```

The `completed` message will show the number of hooks that were run.

If there were any errors, they will be shown in the activity logs as well.

## Enabling/Disabling Hooks

Hooks can be easily enabled or disabled by clicking the toggle button in the hook creation or edit form.

This is useful if you want to temporarily disable a hook without having to delete it.
