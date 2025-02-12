# worker

A Helm chart for the Neosync Temporal Worker

**Homepage:** <https://www.neosync.dev>

## Source Code

* <https://github.com/nucleuscloud/neosync>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| autoscaling.behavior | string | `nil` | The behavior of the HPA autoscaler |
| autoscaling.enabled | bool | `false` | Whether or not to install the HPA autoscaler |
| autoscaling.maxReplicas | int | `4` | The maximum number of replicas to scale to |
| autoscaling.minReplicas | int | `1` | The minimum amount of replicas to have running |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | The CPU % utilization to begin a scale up |
| containerPort | int | `8080` | The container port |
| datadog.enabled | bool | `false` | Whether or not to apply the default Datadog annotations/labels to the deployment |
| deploymentAnnotations | object | `{}` | Provide a map of deployment annotations that will be attached to the deployment's annotations |
| ee.license | string | `nil` | Neosync Enterprise-Edition License Key |
| extraEnvVars | list | `[]` | Provide extra environment variables that will be applied to the deployment. |
| fullnameOverride | string | `nil` | Fully overrides the chart name |
| host | string | `"0.0.0.0"` | Sets the host that the backend will listen on. 0.0.0.0 is common for Kubernetes workloads. |
| image.pullPolicy | string | `nil` | Overrides the default K8s pull policy |
| image.repository | string | `"ghcr.io/nucleuscloud/neosync/worker"` | The default image repository |
| image.tag | string | `nil` | Overrides the image tag whose default is {{ printf "v%s" .Chart.AppVersion }} |
| imagePullSecrets | list | `[]` | Define a list of image pull secrets that will be used by the deployment |
| istio.enabled | bool | `false` | Whether or not to apply the default istio annotations/labels to the deployment |
| nameOverride | string | `nil` | Override the name specified on the Chart, which defaults to .Chart.Name |
| neosync.apiKey | string | `nil` | Only required if running the backend in auth-mode |
| neosync.url | string | `"http://neosync-api"` | The url to the Neoysnc API instance |
| neosyncCloud.enabled | bool | `false` | Whether or not this is NeosyncCloud |
| nodeSelector | object | `{}` | Any node selectors that should be applied to the deployment |
| nucleusEnv | string | `nil` | Mostly used by NeosyncCloud. Adds a special tag to the logging to determine what environment is running |
| otel | object | `{"enabled":false,"otlpPort":4317}` | Will eventually allow sending traces. The worker does emit record-based metrics, but does not currently listen to otel.enabled. Must provide the OTEL_SDK_DISABLED=false environment variable separately today. |
| podAnnotations | object | `{}` | Provide a map of pod annotations that will be attached to the deployment's pod template annotations |
| redis.kind | string | `nil` | The kind of redis instance. simpke, cluster, failover |
| redis.master | string | `nil` | Name of redis master when in failover mode |
| redis.tls.clientCerts | list | `[]` | Client TLS Certificate files |
| redis.tls.enableRenegotiation | bool | `false` | Whether to allow the remote server to repeatedly request renegotiation |
| redis.tls.enabled | bool | `false` | Whether or not to enable redis tls |
| redis.tls.rootCertAuthority | string | `nil` | Root certificate authority |
| redis.tls.rootCertAuthorityFile | string | `nil` | Root certificate authority file location |
| redis.tls.skipCertVerify | bool | `false` | Optionally skip cert verification |
| redis.url | string | `nil` | The url to the redis instance that will be used for PK/FK transformation storage cache |
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
| temporal.certificate.certContents | string | `nil` | The full contents of the certificate. Provide this or the certFilePath, not both. |
| temporal.certificate.certFilePath | string | `nil` | The location of the certificate file |
| temporal.certificate.keyContents | string | `nil` | The full contents of the key. Provide this or the keyFilePath, not both. |
| temporal.certificate.keyFilePath | string | `nil` | The location of the certificate key file |
| temporal.namespace | string | `nil` | If not provided, falls back to hardcoded default value |
| temporal.taskQueue | string | `nil` | If not provided, falls back to hardcoded default value |
| temporal.url | string | `"temporal.temporal:7233"` | The default value based on how Temporal manifests are by default configured. Change this based on your temporal configuration |
| terminationGracePeriodSeconds | string | `nil` | The amount of time in seconds to wait for the pod to shut down when a termination event has occurred. |
| tolerations | list | `[]` | Any tolerations that should be applied to the deployment |
| updateStrategy | string | `nil` | The strategy to use when rolling out new replicas |
| volumeMounts | list | `[]` | Volumes that will be mounted to the deployment |
| volumes | list | `[]` | Volumes that will be attached to the deployment |
| tableSync.maxConcurrency | int | `3` | The number of tables to sync concurrently |
