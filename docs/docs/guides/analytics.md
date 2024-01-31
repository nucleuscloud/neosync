---
title: Configuring Analytics
id: analytics
hide_title: false
slug: /guides/analytics
---

## Analytics

We use Posthog to capture usage analytics. This is helpful for us to understand how users are using Neosync and how we can improve the product.

Today, they are only captured in a very minimal sense within Neosync app. We have plans to also start capturing analytics in the CLI.

You can see what information is captured by checking out the [posthog-provider](https://github.com/nucleuscloud/neosync/blob/main/frontend/apps/web/components/providers/posthog-provider.tsx) component that wraps each page's React components.

Analytics are used simply to get a better view into how people use Neosync.

### Disabling Analytics

If allowing Neosync to capture analytics is not desired, simply remove the `POSTHOG_KEY` from the environment, or disable analytics via the `NEOSYNC_ANALYTICS_ENABLED=false` environment variable.
