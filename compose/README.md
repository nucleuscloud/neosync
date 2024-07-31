# Compose

These compose files are meant to be layered in to the top-level compose files.

They enable adding in extra databases or also enabling metrics.

Most of these save for compose-metrics.yml (as it hasn't been fully updated yet) cannot be run alone and must be layered.

## compose-db.yml

Stands up two separate postgres databases that can be used for testing Neosync.
This is enabled by default.

## compose-db-mongo.yml

Stands up mongodb databases.

## compose-db-mysql.yml

Stands up mysql databases.

## compose-db-dynamo.yml

Stands up dynamodb databases.

## compose-metrics.yml

Stands up an entire metrics suite.

There are a lot of contains that get added here and this should only be provided if testing metrics or logging.

The sections below detail more about what the containers are used for.
Grafana is used by all of them to surface metrics or logs. This local copy does not come pre-defined with any dashboards.

### Service Metrics

OpenTelemtry, Prometheus are used to retrieve worker metrics.
Otel is used to retrieve the metrics from the worker. These are then exported to Prometheus.

Neosync API can be configured to retrieve these metrics via the metrics service.

### Logs

Loki, Promtail are used to suck up logs from the work and serve them into the dashboard.

Promtail scrapes the docker container logs files and pushes them into Loki.

Loki can be configured as a datasource in Grafana to surface logs.

Loki can also be wired up to Neosync API to surface logs to the app dashboard.
