---
title: Neosync Cloud Security Overview
id: cloud-security-overview
hide_title: false
slug: /cloud-security-overview
---

At Neosync, we take as many security precautions as we can to ensure that any information saved on our server is safe.

This section will document a few different things that we feel are worth mentioning from a security perspective.
This page is mostly relevant to Neosync Cloud.

## Code

All of the Neosync code is open source and can be found on our [Github](https://github.com/nucleuscloud/neosync).
If you find a security vulnerability, please refer to our [Security.md](https://github.com/nucleuscloud/neosync/blob/main/SECURITY.md) for what to do.
If all else fails, please email `security@neosync.dev` directly.

Otherwise, the code that is found in our repo is the same code that we deploy on our servers.
This is done directly with the helm charts that we publish to the Github Container Registry.

## Production DB Access

Our production postgres instance is not accessible to the internet and is heavily locked down.
Those that access production must go through an approval and review process and must have proper AWS access in order to SSH in and connect.

## SSH Access

For access to our internal cluster we use a Bastion Host. This is an EC2 instance, but we do not directly expose port 22 to the internet.
We use AWS SSM along with IAM Role policies to control who has access to the tunnel.
Any access on this instance is logged.

## Connecting a Production Database to Neosync

We do not recommend connecting a production data directly to Neosync.

This is recommended purely for security purposes, but also due to an increased load that Neosync may put on your database when invoking a sync.
For that reason, we suggest restoring a snapshot of production periodically to another database that is then used by Neosync.
We don't currently support providing snapshots directly, and if this is important to you, please reach out to us.
