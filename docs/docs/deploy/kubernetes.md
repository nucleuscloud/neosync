---
title: Kubernetes
description: The Kubernetes guide covers how to deploy Neosync specific resources.
id: kubernetes
hide_title: false
slug: /deploy/kubernetes
---

## Kubernetes

The Kubernetes guide covers how to deploy Neosync specific resources.
Deploying a Postgres database and Temporal instances are not covered under this guide.
See the external dependency section below for more information regarding these resources.

## Neosync Helm Chart Layout

Neosync can be deployed to Kubernetes easily with the assistance of Helm charts.

We currently publish four different helm charts for maximum flexibility.
This page will detail the purpose of each one and how it can be used to deploy to Kubernetes.

All of our helm charts are deployed as OCI helm charts and require Helm3 to use.
Our images are published directly to the Github Container registry at: `ghcr.io/nucleuscloud/neosync/helm`.

When a release of Neosync is made, all of these resources are tagged and released at the same version.
If using the Neosync AIO chart at version `v1.0.0`, it will use `v1.0.0` of the api, app, and worker.

### API

The API Helm chart can be used to deploy just the backend API server.
The chart itself can be found [here](https://github.com/nucleuscloud/neosync/tree/main/backend/charts/api).

The local dev edition can be found in the [helmfile](https://github.com/nucleuscloud/neosync/blob/main/backend/dev/helm/api/helmfile.yaml) that is used by our dev Tilt instance.

The full image can be docker pulled via: `docker pull ghcr.io/nucleuscloud/neosync/helm/api:latest`

### App

The APP Helm chart can be used to deploy just the frontend APP.
The chart itself can be found [here](https://github.com/nucleuscloud/neosync/tree/main/frontend/charts/app).

The local dev edition can be found in the [helmfile](https://github.com/nucleuscloud/neosync/blob/main/frontend/apps/web/dev/helm/app/helmfile.yaml) that is used by our dev Tilt instance.

The full image can be docker pulled via: `docker pull ghcr.io/nucleuscloud/neosync/helm/app:latest`

### Worker

The APP Helm chart can be used to deploy just the worker.
The chart itself can be found [here](https://github.com/nucleuscloud/neosync/tree/main/worker/charts/worker).

The local dev edition can be found in the [helmfile](https://github.com/nucleuscloud/neosync/blob/main/worker/dev/helm/helmfile.yaml) that is used by our dev Tilt instance.

The full image can be docker pulled via: `docker pull ghcr.io/nucleuscloud/neosync/helm/worker:latest`

### Neosync Umbrella Chart

The Neosync Umbrella Helm chart can be used to deploy all three resources listed above.
The chart itself can be found [here](https://github.com/nucleuscloud/neosync/blob/main/charts/neosync).

This chart has no templates of its own and merely acts as a single helm entrypoint to deploy all of the Neosync services.
It only contains a `Chart.yaml` that defines the three Neosync dependencies.

The full image can be docker pulled via: `docker pull ghcr.io/nucleuscloud/neosync/helm/neosync:latest`

When running this within the repo, it points to the local copies of the helm chart. The OCI image will point to the OCI images of the published API, APP, and Worker charts.

When defining a values file for this chart, the values will need to be nested underneath the relevant object key for each chart.

```yaml
api:
  # api specific values
app:
  # app specific values
worker:
  # worker specific values
```

These can easily be spread across multiple `values.yaml` files if desired to keep them separate, but they will still need to be nested underneath their respective chart name keys.

## Install Neosync on a Kubernetes Cluster

### Prerequisites

This sequences assumes

- That your system is configured to access a kubernetes cluster
- That your machine has the following installed and is able to access your cluster:
  - kubectl
  - Helm v3

This example uses the umbrella chart for simplicity.

### Basic Installation

Neosync uses the OCI image format for Helm, so there is no need to add a separate registry prior to doing the install.

This example assumes that an external Postgres database has been provisioned as well as a functioning Temporal instance (or cloud).
In the future we will add support for these external dependencies to make it easier to get up and running quickly.

Create a `values.yaml` file.

This is a very minimal install that basically configures the database and temporal, and redis.
There are a number of other options which you'll need to consult the values files for the chart for more information.

Each helm chart has a README that contains the full values spec.

These can be found in the Github repo, and are also available via ArtifactHub.

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/neosync)](https://artifacthub.io/packages/search?repo=neosync)

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/neosync-api)](https://artifacthub.io/packages/search?repo=neosync-api)

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/neosync-app)](https://artifacthub.io/packages/search?repo=neosync-app)

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/neosync-worker)](https://artifacthub.io/packages/search?repo=neosync-worker)

```
api:
  db:
    host: <db-host>
    name: <db-name>
    port: 5432
    username: <user>
    password: <password>
  migrations:
    # The values here should largely be the same as the API. The only different might be the username/password if you intend to separate the roles for the service and migration users.
    db:
      host: <db-host>
      name: <db-name> # should be the same as the api.db.name value
      port: 5432
      username: <user>
      password: <password>

  temporal:
    url: <temporal-url>
    defaultNamespace: <temporal-namespace>
    defaultSyncJobQueue: <temporal-queue>

worker:
  temporal:
    url: <temporal-url>
    namespace: <temporal-namespace>
    taskQueue: <temporal-queue> # if provided a defaultSyncJobQueue for the API, this should be set to the same value

  # required if intending transform primary/foreign keys
  redis:
    url: <redis-url>
```

Install the chart:

When specifying the version, be sure to omit the `v` from the github release. So if the current version if `v0.4.38`, specify `0.4.38` as the helm version.

```
helm install oci://ghcr.io/nucleuscloud/neosync/helm/neosync --version <version> -f values.yaml
```

## External Dependencies for Production Deployments

We do not cover how to deploy a Postgres database for Neosync or the Temporal suite.

### Neosync Postgres DB

Generally, we suggest going with a cloud provider to host the database, but if it's desired to host within Kubernetes, the [Kubernetes Operator](https://postgres-operator.readthedocs.io/en/latest/) is good option.

### Temporal

Temporal can also be deployed to Kubernetes via a Helm chart that can be found in their [Github repo](https://github.com/temporalio/helm-charts). We suggest following their helm chart guide for deploying Temporal into Kubernetes.
Temporal also has [Temporal Cloud](https://cloud.temporal.io/) that can be used if it's not desirable to self-host Temporal in a production environment.

There is also a community maintained [Temporal Operator](https://github.com/alexandrevilain/temporal-operator) that could be of use to assist with Temporal management and upgrades. Use at your own risk.

Here is a link to Temporal's thought on [Helm Chart Deployment](https://docs.temporal.io/self-hosted-guide/setup#helm-charts)

### Redis

Redis serves as an optional external dependency. It's specifically required if primary key transformations are to be utilized. This functionality allows for more advanced data synchronization capabilities by transforming primary keys during the synchronization process, and Redis facilitates this operation by providing the necessary data structure and caching capabilities​ in order to maintain data integrity.

## Tilt for Development Deployments

For development, we use Tilt to set up and deploy Neosync to Kubernetes.
You can find all of the scripts in the various `Tiltfile`'s that can be found throughout our Github repository.
The top level `Tiltfile` is the main driver that can be used to deploy everything (including external dependencies like Temporal).
This is the quickest way to get up and running with a Kubernetes-enabled setup.

## Istio Support

Each chart can set the value `istio.enabled: true`.
This will add the following to each service's `Deployment`.

```yaml
kind: Deployment
spec:
  template:
    metadata:
      annotations:
        proxy.istio.io/config: '{ "holdApplicationUntilProxyStarts": true }'
      labels:
        sidecar.istio.io/inject: 'true'
```

These options will allow injection of the istio sidecar, as well as hold the containers from starting until the Istio sidecar has started.

Istio Gateway's and VirtualServices must be provided separately.

## Datadog Support

Each chart can set the value `datadog.enabled: true`.
This will add the following to each service's `Deployment`.

```yaml
kind: Deployment
metadata:
  labels:
    tags.datadoghq.com/env: { { .Values.nucleusEnv } }
    tags.datadoghq.com/service: { { template "neosync-api.fullname" . } }
    tags.datadoghq.com/version:
      { { .Values.image.tag | default .Chart.AppVersion } }
spec:
  template:
    metadata:
      annotations:
        ad.datadoghq.com/nucleus-api.logs: '[{"source":"nucleus-neosync-api","service":"{{ template "neosync-api.fullname" . }}"}]'
      labels:
        admission.datadoghq.com/enabled: 'true'
        tags.datadoghq.com/env: { { .Values.nucleusEnv } }
        tags.datadoghq.com/service: { { template "neosync-api.fullname" . } }
        tags.datadoghq.com/version:
          { { .Values.image.tag | default .Chart.AppVersion } }
    spec:
      containers:
        - name: user-container
          env:
            - name: DD_ENV
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['tags.datadoghq.com/env']
            - name: DD_SERVICE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['tags.datadoghq.com/service']
            - name: DD_VERSION
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['tags.datadoghq.com/version']
```

## Kubernetes Ingress

Each chart can set the value `ingress.enabled: true`
This will generate a Kubernetes `Ingress` resource that is attached to each chart's service.
See the `ingress.yaml` in each chart for a better understanding of what is available there.
Each ingress has full support for specifying the classname and TLS options.

Below is an example neosync-api Ingress that utilizes nginx, along with cert-manager for TLS decryption, which is required if exposing gRPC/Connect to the outside world.
The TLS setup may be different depending on what your ingress class supports.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cert-manager.io/cluster-issuer: my-cluster-issuer
    nginx.ingress.kubernetes.io/backend-protocol: GRPC
    nginx.org/grpc-services: neosync-api
  name: neosync-api
spec:
  ingressClassName: nginx
  rules:
    - host: neosync-api.example.com
      http:
        paths:
          - backend:
              service:
                name: neosync-api
                port:
                  number: 80
            path: /
            pathType: Prefix
  tls:
    - hosts:
        - neosync-api.example.com
      secretName: neosync-api-certificate
```

The corresponding helm chart values to generate the above would look like this for neosync-api:

```yaml
ingress:
  enabled: true
  className: nginx
  hosts:
    - neosync-api.example.com
  tls:
    - hosts:
        - neosync-api.example.com
      secretName: neosync-api-certificate
  annotations:
    cert-manager.io/cluster-issuer: my-cluster-issuer
    nginx.ingress.kubernetes.io/backend-protocol: GRPC
    nginx.org/grpc-services: neosync-api
```

## Configuring API and Worker charts with Temporal mTLS Certificates

The Temporal mTLS Certificates can be configured as secret values when deploying to a Kubernetes environment.

This section details how that can be set up.

Temporal's guide for generating mTLS Certificates is the recommended way for creating these certs.
That guide can be found [here](https://docs.temporal.io/cloud/certificates#use-tcld-to-generate-certificates).

Once certificates have been created, we can add them to Kubernetes.

The below scripts use kustomize to generate the Certificate secret and store it into Kubernetes.

Both the `temporal.cert` and `temporal.key` must be sitting next to the `kustomization.yaml` file below.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: neosync

secretGenerator:
  - name: temporal-mtls-certs
    files:
      - tls.crt=temporal.cert
      - tls.key=temporal.key
    type: kubernetes.io/tls
    options:
      disableNameSuffixHash: true
```

To apply this: `kustomize build . | kubectl apply -f -`

This will generate a secret and store it in the `neosync` namespace named `temporal-mtls-certs`

To properly configure the `api` and `worker` helm charts, the following values can be specific in the `values.yaml` file.
For simplicity, this section assumes that the same leaf cert will be used by both the API and the Worker.
In a production environment, it may be desirable to utilize separate certificates for maximum security.

```yaml
temporal:
  # ...other configurations omitted...
  certificate:
    keyFilePath: /etc/temporal/certs/tls.key
    certFilePath: /etc/temporal/certs/tls.crt

volumes:
  - name: temporal-certs
    secret:
      secretName: temporal-mtls-certs
volumeMounts:
  - name: temporal-certs
    mountPath: /etc/temporal/certs
```
