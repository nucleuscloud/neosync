---
title: Configuring Analytics
description: Learn how to configure analytics within Neosync open source and turn it off or not based on your preferences
id: analytics
hide_title: false
slug: /guides/analytics
---

## Analytics

This section details the analytics tracking Neosync uses to learn more information about its users.

### Posthog

We use Posthog to capture usage analytics. This is helpful for us to understand how users are using Neosync and how we can improve the product.

Today, they are only captured in a very minimal sense within Neosync app. We have plans to also start capturing analytics in the CLI.

You can see what information is captured by checking out the [posthog-provider](https://github.com/nucleuscloud/neosync/blob/main/frontend/apps/web/components/providers/posthog-provider.tsx) component that wraps each page's React components.

Analytics are used simply to get a better view into how people use Neosync.

### Unify

Unify is similar to Posthog in that it is also used to capture user information. We send the same usage information to Unify that we send to Posthog.

You can see what information is captured by checking out the [unify-provider](https://github.com/nucleuscloud/neosync/blob/main/frontend/apps/web/components/providers/unify-provider.tsx) component that wraps each page's React components.

### Disabling Analytics

To fully disable analytics, set the `NEOSYNC_ANALYTICS_ENABLED=false` environment variable on the frontend (and eventually CLI, backend).
All analytics are keyed off of this environment variable to make it easy to disable.

One can also disable analytics by removing the `POSTHOG_KEY` and `KOALA_KEY`.
