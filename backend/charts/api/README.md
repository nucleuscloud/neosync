# api

A Helm chart for the Neosync Backend API

**Homepage:** <https://www.neosync.dev>

## Source Code

* <https://github.com/nucleuscloud/neosync>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| auth.api.baseUrl | string | `nil` | The base url to the auth service's admin url |
| auth.api.clientId | string | `nil` | The service account client id |
| auth.api.clientSecret | string | `nil` | The service account client secret |
| auth.api.provider | string | `nil` | The provider that should be used. Supported types are auth0 and keycloak. Default is keycloak |
| auth.audience | string | `nil` | The audience that is expected to be present in the JWT aud claim |
| auth.baseUrl | string | `nil` | The base url of the auth server |
| auth.cliAudience | string | `nil` | The aud that is expected to be present when the CLI auths to the backend |
| auth.cliClientId | string | `nil` | The client id that the CLI will use to communicate with the backend for authentication (if auth is enabled) |
| auth.clientMap | string | `nil` | A map of clientId->clientSecret of allowed clients |
| auth.enabled | bool | `false` | Enable/Disable authentication |
| autoscaling.enabled | bool | `false` | Whether or not to install the HPA autoscaler |
| autoscaling.maxReplicas | int | `4` | The maximum number of replicas to scale to |
| autoscaling.minReplicas | int | `1` | The minimum amount of replicas to have running |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | The CPU % utilization to begin a scale up |
| containerPort | int | `8080` | The container port |
| datadog.enabled | bool | `false` | Whether or not to apply the default Datadog annotations/labels to the deployment |
| db.disableSsl | bool | `false` | Whether or not to disable SSL when connecting to the database |
| db.host | string | `nil` | The database hostname |
| db.name | string | `nil` | The name of the database to connect to |
| db.options | string | `nil` | Extra database options that will be appended to the query string |
| db.password | string | `nil` | The username's password for authentication |
| db.port | int | `5432` | The database port |
| db.username | string | `nil` | The username that will be used for authentication |
| deploymentAnnotations | object | `{}` | Provide a map of deployment annotations that will be attached to the deployment's annotations |
| ee.license | string | `nil` | Neosync Enterprise-Edition License Key |
| extraEnvVars | list | `[]` | Provide extra environment variables that will be applied to the deployment's user-container. |
| fullnameOverride | string | `nil` | Fully overrides the chart name |
| host | string | `"0.0.0.0"` | Sets the host that the backend will listen on. 0.0.0.0 is common for Kubernetes workloads. |
| image.pullPolicy | string | `nil` | Overrides the default K8s pull policy |
| image.repository | string | `"ghcr.io/nucleuscloud/neosync/api"` | The default image repository |
| image.tag | string | `nil` | Overrides the image tag whose default is {{ printf "v%s" .Chart.AppVersion }} |
| imagePullSecrets | list | `[]` | Define a list of image pull secrets that will be used by the deployment |
| ingress.enabled | bool | `false` | Enable this if using K8s ingress to expose the backend to the internet |
| istio.enabled | bool | `false` | Whether or not to apply the default istio annotations/labels to the deployment |
| kubernetes.enabled | bool | `true` | Whether or not this is kubernetes (should always be true) |
| kubernetes.namespace | string | `nil` | Falls back to the helm release namespace |
| kubernetes.workerAppName | string | `"neosync-worker"` | Corresponds with the app label that is present on the worker pod |
| migrations.db.disableSsl | bool | `false` | Whether or not to disable SSL when connecting to the database |
| migrations.db.host | string | `nil` | The database hostname |
| migrations.db.migrationsTableName | string | `"neosync_api_schema_migrations"` | This is the tablename that will be created in the postgres "public" schema |
| migrations.db.migrationsTableQuoted | bool | `false` | Whether or not the tablename is quoted in the connection string |
| migrations.db.name | string | `nil` | The name of the database |
| migrations.db.options | string | `nil` | Extra database options that will be appended to the query string |
| migrations.db.password | string | `nil` | The username's password for authentication |
| migrations.db.port | int | `5432` | The database port |
| migrations.db.schemaDir | string | `"/migrations"` | The directory where the migrations are located. |
| migrations.db.username | string | `nil` | The username that will be used for authentication |
| migrations.enabled | bool | `true` | Whether or not the migrations init container will be added to the deployment |
| migrations.extraEnvVars | list | `[]` | Provide extra environment variables that will be applied to the migration init container. |
| nameOverride | string | `nil` | Override the name specified on the Chart, which defaults to .Chart.Name |
| neosyncCloud.enabled | bool | `false` | Whether or not this is NeosyncCloud |
| neosyncCloud.workerApiKeys | list | `[]` | Worker API keys that have been allowlisted to for use |
| nodeSelector | object | `{}` | Any node selectors that should be applied to the deployment |
| nucleusEnv | string | `nil` | Mostly used by NeosyncCloud. Adds a special tag to the logging to determine what environment is running |
| otel.enabled | bool | `false` | whether or not to enable open telemetry settings |
| otel.otlpPort | int | `4317` | Specifies the port that otel is listening on that the service will export metrics and traces to |
| podAnnotations | object | `{}` | Provide a map of pod annotations that will be attached to the deployment's pod template annotations |
| protometrics.apiKey | string | `nil` | Optionally provide an API key if the prometheus service requires authentication |
| protometrics.enabled | bool | `false` | Will enable the metrics service proto |
| protometrics.url | string | `nil` | The url, will default to http://localhost:9090 if not provided. Should be a prometheus compliant API |
| resources.limits.cpu | string | `"500m"` | Sets the max CPU amount |
| resources.limits.memory | string | `"512Mi"` | Sets the max Memory amount |
| resources.requests.cpu | string | `"100m"` | Sets the CPU amount to be requested |
| resources.requests.memory | string | `"128Mi"` | Sets the Memory amount to be requested |
| runLogs.enabled | bool | `false` | Enable this if planning to surface logs within Neosync API and UI (requires a valid license). |
| runLogs.lokiConfig.baseUrl | string | `nil` | The base url to the loki instance |
| runLogs.lokiConfig.keepLabels | string | `nil` | List format. |
| runLogs.lokiConfig.labelsQuery | string | `nil` | LogQL labels query (without the {} as those are provided by the system) |
| runLogs.podConfig.workerAppName | string | `"neosync-worker"` | Corresponds to the app label that is present on the worker pod that will be used to surface logs |
| runLogs.podConfig.workerNamespace | string | `nil` | The namespace the worker lives in |
| runLogs.type | string | `"k8s-pods"` | Possible values: k8s-pods, loki |
| serviceAccount.annotations | object | `{}` | Specify annotations here that will be attached to the service account. Useful for specifying role information or other tagging depending on environment. |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `nil` | The name of the service account to use. If namenot set and create is true, a name is generated using fullname template |
| servicePort | int | `80` | The K8s service port |
| shutdownTimeoutSeconds | string | `nil` | Not currently used |
| sidecarContainers | list | `[]` | Provide sidecars that will be appended directly to the deployment next to the user-container |
| temporal.certificate.certContents | string | `nil` | The full contents of the certificate. Provide this or the certFilePath, not both. |
| temporal.certificate.certFilePath | string | `nil` | The location of the certificate file |
| temporal.certificate.keyContents | string | `nil` | The full contents of the key. Provide this or the keyFilePath, not both. |
| temporal.certificate.keyFilePath | string | `nil` | The location of the certificate key file |
| temporal.defaultNamespace | string | `nil` | If not provided, falls back to hardcoded default value |
| temporal.defaultSyncJobQueue | string | `nil` | If not provided, falls back to hardcoded default value |
| temporal.url | string | `"temporal.temporal:7233"` | The default value based on how Temporal manifests are by default configured. Change this based on your temporal configuration |
| terminationGracePeriodSeconds | string | `nil` | The amount of time in seconds to wait for the pod to shut down when a termination event has occurred. |
| tolerations | list | `[]` | Any tolerations that should be applied to the deployment |
| updateStrategy | string | `nil` | The strategy to use when rolling out new replicas |
| volumeMounts | list | `[]` | Volumes that will be mounted to the deployment |
| volumes | list | `[]` | Volumes that will be attached to the deployment |
