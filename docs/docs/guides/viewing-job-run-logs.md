---
title: View Job Run Logs
description: Learn how to view job run logs to assist with debugging or to view more details of the job run
id: viewing-job-run-logs
hide_title: false
slug: /guides/viewing-job-run-logs
# cSpell:words LOKICONFIG
---

## Job Run Logs

This section details the variety of ways that job run logs can be accessed depending on your Neosync environment.

![Job Run Logs](/img/runlogs.png)

## Neosync Cloud

Job Run Logs can only be accessed through the UI.
Navigate to [Neosync Cloud](https://app.neosync.dev), then click on the Runs tab in the top nav.

Select the run you wish to see logs for. Select the run.

You'll see a section in the middle of the page for logs.

## Open Source

For the open source variant, you have a few options at your disposal.

### Docker Compose

If you're running the docker compose setup or just trying out Neosync locally, there is currently no option within the UI to view logs.

To see these logs, you'll need to tail the running Neosync `worker` container.
This can be done a variety of ways.

If you ran `make compose/up` from the root, you can run the following command in your terminal:

```console
docker compose logs -f worker
```

This will print out any logs as well as follow the worker. If you just wish to print and not also follow, omit the `-f`

An alternative is to use the `docker` command directly, or navigate to the neosync worker container in Docker UI.

```console
docker logs neosync-worker -f
```

### Kubernetes

If you're running in a Kubernetes environment, there are multiple ways to view worker logs, along with support for showing them natively in Neosync's UI.

#### kubectl

The standard way of viewing the live pod logs:

```console
kubectl logs -n neosync deployment/neosync-worker -f`
```

#### Neosync UI

> **NB:** This requires a valid Neosync Enterprise license for OSS deployments. If you would like to try this out, please contact us.

Neosync can be configured to surface pod logs via the UI by configuring `neosync-api` to surface these.

If you've deployed Neosync via the helm chart, this should come pre-configured with kubernetes pod logs.

This must be enabled both on the Frontend as well as on the backend as the frontend environment variable handles showing the components, while the backend enables the ability to actually surface and scrape the pod logs from the worker.

This are great for basic deployments, but will disappear on worker pod shutdown.

## Persistence with Loki

Neosync has native support for surfacing logs that come from a [Grafana Loki](https://grafana.com/oss/loki/) instance (with a valid Neosync Enterprise license).

Please note that Neosync does not natively handle shipping logs to Loki, however it can natively handle querying a Loki instance to surface logs into Neosync UI.

This can be configured by setting the `RUN_LOGS_TYPE` environment variable to `loki`, along with configuring the `RUN_LOGS_LOKICONFIG_BASEURL`.
To see the full environment variables, view the [api env vars](../deploy/environment-variables.md#backend-api).
