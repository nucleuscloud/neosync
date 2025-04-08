# app

A Helm chart for the Neosync App

**Homepage:** <https://www.neosync.dev>

## Source Code

* <https://github.com/nucleuscloud/neosync>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| analytics.enabled | bool | `true` | Enables analytics such as Posthog/Unify (if keys have been provided for them) |
| auth.audience | string | `nil` | The audience that should be present in the JWT token |
| auth.clientId | string | `nil` | The client id that will be used by the app to retrieve user tokens |
| auth.clientSecret | string | `nil` | The client secret that will be used by the app |
| auth.enabled | bool | `false` | Enable/disable authentication |
| auth.issuer | string | `nil` | The issuer url. This is typically the base url of the auth service |
| auth.providerId | string | `nil` | The ID of the provider for your OIDC client. This can be anything |
| auth.providerName | string | `nil` | The display name of the provider |
| auth.scope | string | `nil` | The scopes that should be requested. Standard are "openid email profile offline_access" |
| auth.trustHost | bool | `true` | Whether or not to trust the external host (most likely want this to be true if running behind a load balancer) |
| containerPort | int | `3000` | The container port |
| datadog.enabled | bool | `false` | Whether or not to apply the default Datadog annotations/labels to the deployment |
| deploymentAnnotations | object | `{}` | Provide a map of deployment annotations that will be attached to the deployment's annotations |
| disableGcpCloudStorageConnections | bool | `false` | Feature flag that will disable GCP Cloud Storage Connections from being visible. Note: This only disables the new connections form and is a temporary flag until authentication in the multi-tenant environment is better understood. |
| enableRunLogs | bool | `false` | Feature flag that enables the frontend to show the run logs on the Run [id] page. only enable this if the backend has been configured to surface run logs. Requires EE License |
| extraEnvVars | list | `[]` | Provide extra environment variables that will be applied to the deployment. |
| fullnameOverride | string | `nil` | Fully overrides the chart name |
| host | string | `"0.0.0.0"` | Sets the host that the backend will listen on. 0.0.0.0 is common for Kubernetes workloads. |
| image.pullPolicy | string | `nil` | Overrides the default K8s pull policy |
| image.repository | string | `"ghcr.io/nucleuscloud/neosync/app"` | The default image repository |
| image.tag | string | `nil` | Overrides the image tag whose default is {{ printf "v%s" .Chart.AppVersion }} |
| imagePullSecrets | list | `[]` | Define a list of image pull secrets that will be used by the deployment |
| ingress.enabled | bool | `false` | Enable this if using K8s ingress to expose the backend to the internet |
| istio.enabled | bool | `false` | Whether or not to apply the default istio annotations/labels to the deployment |
| jobHooks.enabled | bool | `false` | Enables Job Hooks on the frontend. Note: This will only work if it has also been enabled via the backend with a valid license |
| nameOverride | string | `nil` | Override the name specified on the Chart, which defaults to .Chart.Name |
| neosyncApi.url | string | `"http://neosync-api"` | The URL to the Neosync API instance |
| neosyncCloud.enabled | bool | `false` | Whether or not this is NeosyncCloud |
| nextAuthSecret | string | `"This is an example"` | next-auth secret that is used to encrypt the session cookie |
| nextAuthUrl | string | `"http://localhost:3000"` | next-auth base url. Should be the public url for the application |
| nextPublic.appBaseUrl | string | `"http://localhost:3000"` | next public app base url. Should be the public url for the application |
| nodeSelector | object | `{}` | Any node selectors that should be applied to the deployment |
| nucleusEnv | string | `nil` | Mostly used by NeosyncCloud. Adds a special tag to the logging to determine what environment is running |
| otel.enabled | bool | `false` | whether or not to enable open telemetry settings |
| otel.otlpPort | int | `4317` | Specifies the port that otel is listening on that the service will export metrics and traces to |
| podAnnotations | object | `{}` | Provide a map of pod annotations that will be attached to the deployment's pod template annotations |
| posthog.key | string | `"phc_qju45RhNvCDwYVdRyUjtWuWsOmLFaQZi3fmztMBaJip"` | Posthog Key |
| protometrics.enabled | bool | `false` |  |
| resources.limits.cpu | string | `"500m"` | Sets the max CPU amount |
| resources.limits.memory | string | `"512Mi"` | Sets the max Memory amount |
| resources.requests.cpu | string | `"100m"` | Sets the CPU amount to be requested |
| resources.requests.memory | string | `"128Mi"` | Sets the Memory amount to be requested |
| serviceAccount.annotations | object | `{}` | Specify annotations here that will be attached to the service account. Useful for specifying role information or other tagging depending on environment. |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `nil` | The name of the service account to use. If namenot set and create is true, a name is generated using fullname template |
| servicePort | int | `80` | The K8s service port |
| shutdownTimeoutSeconds | string | `nil` | Not currently used |
| sidecarContainers | list | `[]` | Provide sidecars that will be appended directly to the deployment next to the user-container |
| terminationGracePeriodSeconds | string | `nil` | The amount of time in seconds to wait for the pod to shut down when a termination event has occurred. |
| tolerations | list | `[]` | Any tolerations that should be applied to the deployment |
| unify.key | string | `nil` | Unify Key |
| updateStrategy | string | `nil` | The strategy to use when rolling out new replicas |
