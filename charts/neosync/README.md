# neosync

A Helm chart for Neosync that contains the api, app, and worker

**Homepage:** <https://www.neosync.dev>

## Source Code

* <https://github.com/nucleuscloud/neosync>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../../backend/charts/api | api | v0 |
| file://../../frontend/apps/web/charts/app | app | v0 |
| file://../../worker/charts/worker | worker | v0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| api.enabled | bool | `true` | Enable or Disable Neoysnc Api |
| app.enabled | bool | `true` | Enable or Disable Neoysnc App |
| worker.enabled | bool | `true` | Enable or Disable Neosync Worker |
| api.auth.api.baseUrl | string | `nil` | The base url to the auth service's admin url |
| api.auth.api.clientId | string | `nil` | The service account client id |
| api.auth.api.clientSecret | string | `nil` | The service account client secret |
| api.auth.api.provider | string | `nil` | The provider that should be used. Supported types are auth0 and keycloak. Default is keycloak |
| api.auth.audience | string | `nil` | The audience that is expected to be present in the JWT aud claim |
| api.auth.baseUrl | string | `nil` | The base url of the auth server |
| api.auth.cliAudience | string | `nil` | The aud that is expected to be present when the CLI auths to the backend |
| api.auth.cliClientId | string | `nil` | The client id that the CLI will use to communicate with the backend for authentication (if auth is enabled) |
| api.auth.clientMap | string | `nil` | A map of clientId->clientSecret of allowed clients |
| api.auth.enabled | bool | `false` | Enable/Disable authentication |
| api.autoscaling.enabled | bool | `false` | Whether or not to install the HPA autoscaler |
| api.autoscaling.maxReplicas | int | `4` | The maximum number of replicas to scale to |
| api.autoscaling.minReplicas | int | `1` | The minimum amount of replicas to have running |
| api.autoscaling.targetCPUUtilizationPercentage | int | `80` | The CPU % utilization to begin a scale up |
| api.containerPort | int | `8080` | The container port |
| api.datadog.enabled | bool | `false` | Whether or not to apply the default Datadog annotations/labels to the deployment |
| api.db.disableSsl | bool | `false` | Whether or not to disable SSL when connecting to the database |
| api.db.host | string | `nil` | The database hostname |
| api.db.name | string | `nil` | The name of the database to connect to |
| api.db.options | string | `nil` | Extra database options that will be appended to the query string |
| api.db.password | string | `nil` | The username's password for authentication |
| api.db.port | int | `5432` | The database port |
| api.db.username | string | `nil` | The username that will be used for authentication |
| api.deploymentAnnotations | object | `{}` | Provide a map of deployment annotations that will be attached to the deployment's annotations |
| api.ee.license | string | `nil` | Neosync Enterprise-Edition License Key |
| api.extraEnvVars | list | `[]` | Provide extra environment variables that will be applied to the deployment's user-container. |
| api.fullnameOverride | string | `nil` | Fully overrides the chart name |
| api.host | string | `"0.0.0.0"` | Sets the host that the backend will listen on. 0.0.0.0 is common for Kubernetes workloads. |
| api.image.pullPolicy | string | `nil` | Overrides the default K8s pull policy |
| api.image.repository | string | `"ghcr.io/nucleuscloud/neosync/api"` | The default image repository |
| api.image.tag | string | `nil` | Overrides the image tag whose default is {{ printf "v%s" .Chart.AppVersion }} |
| api.imagePullSecrets | list | `[]` | Define a list of image pull secrets that will be used by the deployment |
| api.ingress.enabled | bool | `false` | Enable this if using K8s ingress to expose the backend to the internet |
| api.istio.enabled | bool | `false` | Whether or not to apply the default istio annotations/labels to the deployment |
| api.kubernetes.enabled | bool | `true` | Whether or not this is kubernetes (should always be true) |
| api.kubernetes.namespace | string | `nil` | Falls back to the helm release namespace |
| api.kubernetes.workerAppName | string | `"neosync-worker"` | Corresponds with the app label that is present on the worker pod |
| api.migrations.db.disableSsl | bool | `false` | Whether or not to disable SSL when connecting to the database |
| api.migrations.db.host | string | `nil` | The database hostname |
| api.migrations.db.migrationsTableName | string | `"neosync_api_schema_migrations"` | This is the tablename that will be created in the postgres "public" schema |
| api.migrations.db.migrationsTableQuoted | bool | `false` | Whether or not the tablename is quoted in the connection string |
| api.migrations.db.name | string | `nil` | The name of the database |
| api.migrations.db.options | string | `nil` | Extra database options that will be appended to the query string |
| api.migrations.db.password | string | `nil` | The username's password for authentication |
| api.migrations.db.port | int | `5432` | The database port |
| api.migrations.db.schemaDir | string | `"/migrations"` | The directory where the migrations are located. |
| api.migrations.db.username | string | `nil` | The username that will be used for authentication |
| api.migrations.extraEnvVars | list | `[]` | Provide extra environment variables that will be applied to the migration init container. |
| api.nameOverride | string | `nil` | Override the name specified on the Chart, which defaults to .Chart.Name |
| api.neosyncCloud.enabled | bool | `false` | Whether or not this is NeosyncCloud |
| api.neosyncCloud.workerApiKeys | list | `[]` | Worker API keys that have been allowlisted to for use |
| api.nodeSelector | object | `{}` | Any node selectors that should be applied to the deployment |
| api.nucleusEnv | string | `nil` | Mostly used by NeosyncCloud. Adds a special tag to the logging to determine what environment is running |
| api.otel.enabled | bool | `false` | whether or not to enable open telemetry settings |
| api.otel.otlpPort | int | `4317` | Specifies the port that otel is listening on that the service will export metrics and traces to |
| api.podAnnotations | object | `{}` | Provide a map of pod annotations that will be attached to the deployment's pod template annotations |
| api.protometrics.apiKey | string | `nil` | Optionally provide an API key if the prometheus service requires authentication |
| api.protometrics.enabled | bool | `false` | Will enable the metrics service proto |
| api.protometrics.url | string | `nil` | The url, will default to http://localhost:9090 if not provided. Should be a prometheus compliant API |
| api.resources.limits.cpu | string | `"500m"` | Sets the max CPU amount |
| api.resources.limits.memory | string | `"512Mi"` | Sets the max Memory amount |
| api.resources.requests.cpu | string | `"100m"` | Sets the CPU amount to be requested |
| api.resources.requests.memory | string | `"128Mi"` | Sets the Memory amount to be requested |
| api.runLogs.enabled | bool | `true` | Independently set so that configuration can exist but logs can still be independently disabled. Defaults to true with k8s-pods configuration for backwards compat and because this is a helm chart this is defacto kubernetes |
| api.runLogs.lokiConfig.baseUrl | string | `nil` | The base url to the loki instance |
| api.runLogs.lokiConfig.keepLabels | string | `nil` | List format. |
| api.runLogs.lokiConfig.labelsQuery | string | `nil` | LogQL labels query (without the {} as those are provided by the system) |
| api.runLogs.podConfig.workerAppName | string | `"neosync-worker"` | Corresponds to the app label that is present on the worker pod that will be used to surface logs |
| api.runLogs.podConfig.workerNamespace | string | `nil` | The namespace the worker lives in |
| api.runLogs.type | string | `"k8s-pods"` | Possible values: k8s-pods, loki |
| api.serviceAccount.annotations | object | `{}` | Specify annotations here that will be attached to the service account. Useful for specifying role information or other tagging depending on environment. |
| api.serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| api.serviceAccount.name | string | `nil` | The name of the service account to use. If namenot set and create is true, a name is generated using fullname template |
| api.servicePort | int | `80` | The K8s service port |
| api.shutdownTimeoutSeconds | string | `nil` | Not currently used |
| api.sidecarContainers | list | `[]` | Provide sidecars that will be appended directly to the deployment next to the user-container |
| api.temporal.certificate.certContents | string | `nil` | The full contents of the certificate. Provide this or the certFilePath, not both. |
| api.temporal.certificate.certFilePath | string | `nil` | The location of the certificate file |
| api.temporal.certificate.keyContents | string | `nil` | The full contents of the key. Provide this or the keyFilePath, not both. |
| api.temporal.certificate.keyFilePath | string | `nil` | The location of the certificate key file |
| api.temporal.defaultNamespace | string | `nil` | If not provided, falls back to hardcoded default value |
| api.temporal.defaultSyncJobQueue | string | `nil` | If not provided, falls back to hardcoded default value |
| api.temporal.url | string | `"temporal.temporal:7233"` | The default value based on how Temporal manifests are by default configured. Change this based on your temporal configuration |
| api.terminationGracePeriodSeconds | string | `nil` | The amount of time in seconds to wait for the pod to shut down when a termination event has occurred. |
| api.tolerations | list | `[]` | Any tolerations that should be applied to the deployment |
| api.updateStrategy | string | `nil` | The strategy to use when rolling out new replicas |
| api.volumeMounts | list | `[]` | Volumes that will be mounted to the deployment |
| api.volumes | list | `[]` | Volumes that will be attached to the deployment |
| app.analytics.enabled | bool | `true` | Enables analytics such as Posthog/Koala (if keys have been provided for them) |
| app.auth.audience | string | `nil` | The audience that should be present in the JWT token |
| app.auth.clientId | string | `nil` | The client id that will be used by the app to retrieve user tokens |
| app.auth.clientSecret | string | `nil` | The client secret that will be used by the app |
| app.auth.enabled | bool | `false` | Enable/disable authentication |
| app.auth.issuer | string | `nil` | The issuer url. This is typically the base url of the auth service |
| app.auth.providerId | string | `nil` | The ID of the provider for your OIDC client. This can be anything |
| app.auth.providerName | string | `nil` | The display name of the provider |
| app.auth.scope | string | `nil` | The scopes that should be requested. Standard are "openid email profile offline_access" |
| app.auth.trustHost | bool | `true` | Whether or not to trust the external host (most likely want this to be true if running behind a load balancer) |
| app.containerPort | int | `3000` | The container port |
| app.datadog.enabled | bool | `false` | Whether or not to apply the default Datadog annotations/labels to the deployment |
| app.deploymentAnnotations | object | `{}` | Provide a map of deployment annotations that will be attached to the deployment's annotations |
| app.disableGcpCloudStorageConnections | bool | `false` | Feature flag that will disable GCP Cloud Storage Connections from being visible. Note: This only disables the new connections form and is a temporary flag until authentication in the multi-tenant environment is better understood. |
| app.enableRunLogs | bool | `true` | Feature flag that enables the frontend to show the run logs on the Run [id] page. only enable this if the backend has been configured to surface run logs |
| app.extraEnvVars | list | `[]` | Provide extra environment variables that will be applied to the deployment. |
| app.fullnameOverride | string | `nil` | Fully overrides the chart name |
| app.host | string | `"0.0.0.0"` | Sets the host that the backend will listen on. 0.0.0.0 is common for Kubernetes workloads. |
| app.image.pullPolicy | string | `nil` | Overrides the default K8s pull policy |
| app.image.repository | string | `"ghcr.io/nucleuscloud/neosync/app"` | The default image repository |
| app.image.tag | string | `nil` | Overrides the image tag whose default is {{ printf "v%s" .Chart.AppVersion }} |
| app.imagePullSecrets | list | `[]` | Define a list of image pull secrets that will be used by the deployment |
| app.ingress.enabled | bool | `false` | Enable this if using K8s ingress to expose the backend to the internet |
| app.istio.enabled | bool | `false` | Whether or not to apply the default istio annotations/labels to the deployment |
| app.koala.key | string | `nil` | Koala Key |
| app.nameOverride | string | `nil` | Override the name specified on the Chart, which defaults to .Chart.Name |
| app.neosyncApi.url | string | `"http://neosync-api"` | The URL to the Neosync API instance |
| app.neosyncCloud.enabled | bool | `false` | Whether or not this is NeosyncCloud |
| app.nextAuthSecret | string | `"This is an example"` | next-auth secret that is used to encrypt the session cookie |
| app.nextAuthUrl | string | `"http://localhost:3000"` | next-auth base url. Should be the public url for the application |
| app.nextPublic.appBaseUrl | string | `"http://localhost:3000"` | next public app base url. Should be the public url for the application |
| app.nodeSelector | object | `{}` | Any node selectors that should be applied to the deployment |
| app.nucleusEnv | string | `nil` | Mostly used by NeosyncCloud. Adds a special tag to the logging to determine what environment is running |
| app.otel.enabled | bool | `false` | whether or not to enable open telemetry settings |
| app.otel.otlpPort | int | `4317` | Specifies the port that otel is listening on that the service will export metrics and traces to |
| app.podAnnotations | object | `{}` | Provide a map of pod annotations that will be attached to the deployment's pod template annotations |
| app.posthog.key | string | `"phc_qju45RhNvCDwYVdRyUjtWuWsOmLFaQZi3fmztMBaJip"` | Posthog Key |
| app.protometrics.enabled | bool | `false` |  |
| app.resources.limits.cpu | string | `"500m"` | Sets the max CPU amount |
| app.resources.limits.memory | string | `"512Mi"` | Sets the max Memory amount |
| app.resources.requests.cpu | string | `"100m"` | Sets the CPU amount to be requested |
| app.resources.requests.memory | string | `"128Mi"` | Sets the Memory amount to be requested |
| app.serviceAccount.annotations | object | `{}` | Specify annotations here that will be attached to the service account. Useful for specifying role information or other tagging depending on environment. |
| app.serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| app.serviceAccount.name | string | `nil` | The name of the service account to use. If namenot set and create is true, a name is generated using fullname template |
| app.servicePort | int | `80` | The K8s service port |
| app.shutdownTimeoutSeconds | string | `nil` | Not currently used |
| app.sidecarContainers | list | `[]` | Provide sidecars that will be appended directly to the deployment next to the user-container |
| app.terminationGracePeriodSeconds | string | `nil` | The amount of time in seconds to wait for the pod to shut down when a termination event has occurred. |
| app.tolerations | list | `[]` | Any tolerations that should be applied to the deployment |
| app.updateStrategy | string | `nil` | The strategy to use when rolling out new replicas |
| worker.autoscaling.enabled | bool | `false` | Whether or not to install the HPA autoscaler |
| worker.autoscaling.maxReplicas | int | `4` | The maximum number of replicas to scale to |
| worker.autoscaling.minReplicas | int | `1` | The minimum amount of replicas to have running |
| worker.autoscaling.targetCPUUtilizationPercentage | int | `80` | The CPU % utilization to begin a scale up |
| worker.containerPort | int | `8080` | The container port |
| worker.datadog.enabled | bool | `false` | Whether or not to apply the default Datadog annotations/labels to the deployment |
| worker.deploymentAnnotations | object | `{}` | Provide a map of deployment annotations that will be attached to the deployment's annotations |
| worker.extraEnvVars | list | `[]` | Provide extra environment variables that will be applied to the deployment. |
| worker.fullnameOverride | string | `nil` | Fully overrides the chart name |
| worker.host | string | `"0.0.0.0"` | Sets the host that the backend will listen on. 0.0.0.0 is common for Kubernetes workloads. |
| worker.image.pullPolicy | string | `nil` | Overrides the default K8s pull policy |
| worker.image.repository | string | `"ghcr.io/nucleuscloud/neosync/worker"` | The default image repository |
| worker.image.tag | string | `nil` | Overrides the image tag whose default is {{ printf "v%s" .Chart.AppVersion }} |
| worker.imagePullSecrets | list | `[]` | Define a list of image pull secrets that will be used by the deployment |
| worker.istio.enabled | bool | `false` | Whether or not to apply the default istio annotations/labels to the deployment |
| worker.nameOverride | string | `nil` | Override the name specified on the Chart, which defaults to .Chart.Name |
| worker.neosync.apiKey | string | `nil` | Only required if running the backend in auth-mode |
| worker.neosync.url | string | `"http://neosync-api"` | The url to the Neoysnc API instance |
| worker.nodeSelector | object | `{}` | Any node selectors that should be applied to the deployment |
| worker.nucleusEnv | string | `nil` | Mostly used by NeosyncCloud. Adds a special tag to the logging to determine what environment is running |
| worker.otel | object | `{"enabled":false,"otlpPort":4317}` | Will eventually allow sending traces. The worker does emit record-based metrics, but does not currently listen to otel.enabled. Must provide the OTEL_SDK_DISABLED=false environment variable separately today. |
| worker.podAnnotations | object | `{}` | Provide a map of pod annotations that will be attached to the deployment's pod template annotations |
| worker.redis.kind | string | `nil` | The kind of redis instance. simpke, cluster, failover |
| worker.redis.master | string | `nil` | Name of redis master when in failover mode |
| worker.redis.tls.clientCerts | list | `[]` | Client TLS Certificate files |
| worker.redis.tls.enableRenegotiation | bool | `false` | Whether to allow the remote server to repeatedly request renegotiation |
| worker.redis.tls.enabled | bool | `false` | Whether or not to enable redis tls |
| worker.redis.tls.rootCertAuthority | string | `nil` | Root certificate authority |
| worker.redis.tls.rootCertAuthorityFile | string | `nil` | Root certificate authority file location |
| worker.redis.tls.skipCertVerify | bool | `false` | Optionally skip cert verification |
| worker.redis.url | string | `nil` | The url to the redis instance that will be used for PK/FK transformation storage cache |
| worker.resources.limits.cpu | string | `"500m"` | Sets the max CPU amount |
| worker.resources.limits.memory | string | `"512Mi"` | Sets the max Memory amount |
| worker.resources.requests.cpu | string | `"100m"` | Sets the CPU amount to be requested |
| worker.resources.requests.memory | string | `"128Mi"` | Sets the Memory amount to be requested |
| worker.serviceAccount.annotations | object | `{}` | Specify annotations here that will be attached to the service account. Useful for specifying role information or other tagging depending on environment. |
| worker.serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| worker.serviceAccount.name | string | `nil` | The name of the service account to use. If namenot set and create is true, a name is generated using fullname template |
| worker.servicePort | int | `80` | The K8s service port |
| worker.shutdownTimeoutSeconds | string | `nil` | Not currently used |
| worker.sidecarContainers | list | `[]` | Provide sidecars that will be appended directly to the deployment next to the user-container |
| worker.temporal.certificate.certContents | string | `nil` | The full contents of the certificate. Provide this or the certFilePath, not both. |
| worker.temporal.certificate.certFilePath | string | `nil` | The location of the certificate file |
| worker.temporal.certificate.keyContents | string | `nil` | The full contents of the key. Provide this or the keyFilePath, not both. |
| worker.temporal.certificate.keyFilePath | string | `nil` | The location of the certificate key file |
| worker.temporal.namespace | string | `nil` | If not provided, falls back to hardcoded default value |
| worker.temporal.taskQueue | string | `nil` | If not provided, falls back to hardcoded default value |
| worker.temporal.url | string | `"temporal.temporal:7233"` | The default value based on how Temporal manifests are by default configured. Change this based on your temporal configuration |
| worker.terminationGracePeriodSeconds | string | `nil` | The amount of time in seconds to wait for the pod to shut down when a termination event has occurred. |
| worker.tolerations | list | `[]` | Any tolerations that should be applied to the deployment |
| worker.updateStrategy | string | `nil` | The strategy to use when rolling out new replicas |
| worker.volumeMounts | list | `[]` | Volumes that will be mounted to the deployment |
| worker.volumes | list | `[]` | Volumes that will be attached to the deployment |
