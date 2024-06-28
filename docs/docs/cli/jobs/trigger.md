---
title: Trigger
description: Learn how to trigger a Neosync job with the neosync jobs trigger command.
id: trigger
hide_title: false
slug: /cli/jobs/trigger
---

## Overview

Learn how to trigger a Neosync job with the neosync jobs trigger command.

The `neosync jobs trigger` command is used to trigger an execution of a Neosync job.
This is useful if a Job is configured but is not running on a schedule, or it's desired to trigger a job outside of the normal scheduled flow.

## Usage

```bash
neosync jobs trigger <job-id>
```

### Argument: job-id

A job-id must be provided as the first command-line argument. This is required and will fail otherwise.
This job-id is used to trigger a workflow execution of the relevant Neosync Job.
