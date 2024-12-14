{{/*
Expand the name of the chart.
*/}}
{{- define "neosync-api.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "neosync-api.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "neosync-api.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "neosync-api.labels" -}}
helm.sh/chart: {{ include "neosync-api.chart" . }}
{{ include "neosync-api.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "neosync-api.selectorLabels" -}}
app.kubernetes.io/name: {{ include "neosync-api.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "neosync-api.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "neosync-api.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Generate the stringData section for environment variables
*/}}
{{- define "neosync-api.env-vars" -}}
DB_HOST: {{ .Values.db.host }}
DB_PORT: {{ .Values.db.port | quote }}
DB_NAME: {{ .Values.db.name }}
DB_USER: {{ .Values.db.username }}
DB_PASS: {{ .Values.db.password }}
DB_SSL_DISABLE: {{ .Values.db.disableSsl | quote}}
{{- if .Values.db.options }}
DB_OPTIONS: {{ .Values.db.options | quote}}
{{- end }}
{{- if .Values.host }}
HOST: {{ .Values.host | quote}}
{{- end }}
PORT: {{ .Values.containerPort | quote }}
{{- if .Values.otel.enabled }}
OTEL_EXPORTER_OTLP_PORT: {{ .Values.otel.otlpPort | quote }} # sends to gRPC receiver
{{- end }}
{{- if .Values.nucleusEnv }}
NUCLEUS_ENV: {{ .Values.nucleusEnv }}
{{- end }}
{{- if .Values.shutdownTimeoutSeconds }}
SHUTDOWN_TIMEOUT_SECONDS: {{ .Values.shutdownTimeoutSeconds | quote }}
{{- end }}
{{- if and .Values.auth .Values.auth.enabled }}
AUTH_ENABLED: {{ .Values.auth.enabled | default "false" | quote }}
{{- end }}
{{- if and .Values.auth .Values.auth.baseUrl }}
AUTH_BASEURL: {{ .Values.auth.baseUrl }}
{{- end }}
{{- if and .Values.auth .Values.auth.expectedIss }}
AUTH_EXPECTED_ISS: {{ .Values.auth.expectedIss }}
{{- end }}
{{- if and .Values.auth .Values.auth.audience }}
AUTH_AUDIENCE: {{ .Values.auth.audience }}
{{- end }}
{{- if and .Values.auth .Values.auth.clientMap }}
AUTH_CLIENTID_SECRET: {{ .Values.auth.clientMap | toJson | quote }}
{{- end }}
{{- if and .Values.auth .Values.auth.cliClientId }}
AUTH_CLI_CLIENT_ID: {{ .Values.auth.cliClientId }}
{{- end }}
{{- if and .Values.auth .Values.auth.cliAudience }}
AUTH_CLI_AUDIENCE: {{ .Values.auth.cliAudience }}
{{- end }}
{{- if and .Values.auth .Values.auth.signatureAlgorithm }}
AUTH_SIGNATURE_ALGORITHM: {{ .Values.auth.signatureAlgorithm }}
{{- end }}
{{- if and .Values.temporal .Values.temporal.url }}
TEMPORAL_URL: {{ .Values.temporal.url }}
{{- end }}
{{- if and .Values.temporal .Values.temporal.certificate .Values.temporal.certificate.keyFilePath }}
TEMPORAL_CERT_KEY_PATH: {{ .Values.temporal.certificate.keyFilePath }}
{{- end }}
{{- if and .Values.temporal .Values.temporal.certificate .Values.temporal.certificate.certFilePath }}
TEMPORAL_CERT_PATH: {{ .Values.temporal.certificate.certFilePath }}
{{- end }}
{{- if and .Values.temporal .Values.temporal.certificate .Values.temporal.certificate.keyContents }}
TEMPORAL_CERT_KEY: {{ .Values.temporal.certificate.keyContents }}
{{- end }}
{{- if and .Values.temporal .Values.temporal.certificate .Values.temporal.certificate.certContents }}
TEMPORAL_CERT: {{ .Values.temporal.certificate.certContents }}
{{- end }}
{{- if and .Values.temporal .Values.temporal.defaultNamespace }}
TEMPORAL_DEFAULT_NAMESPACE: {{ .Values.temporal.defaultNamespace }}
{{- end }}
{{- if and .Values.temporal .Values.temporal.defaultSyncJobQueue }}
TEMPORAL_DEFAULT_SYNCJOB_QUEUE: {{ .Values.temporal.defaultSyncJobQueue }}
{{- end }}
{{- if and .Values.auth .Values.auth.api .Values.auth.api.clientId }}
AUTH_API_CLIENT_ID: {{ .Values.auth.api.clientId }}
{{- end }}
{{- if and .Values.auth .Values.auth.api .Values.auth.api.clientSecret }}
AUTH_API_CLIENT_SECRET: {{ .Values.auth.api.clientSecret }}
{{- end }}
{{- if and .Values.auth .Values.auth.api .Values.auth.api.baseUrl }}
AUTH_API_BASEURL: {{ .Values.auth.api.baseUrl }}
{{- end }}
{{- if and .Values.auth .Values.auth.api .Values.auth.api.provider }}
AUTH_API_PROVIDER: {{ .Values.auth.api.provider }}
{{- end }}
NEOSYNC_CLOUD: {{ .Values.neosyncCloud.enabled | default "false" | quote }}
{{- if .Values.neosyncCloud.enabled }}
NEOSYNC_CLOUD_ALLOWED_WORKER_API_KEYS: {{ join "," .Values.neosyncCloud.workerApiKeys }}
{{- end }}
KUBERNETES_ENABLED: {{ .Values.kubernetes.enabled | default "true" | quote }}
KUBERNETES_NAMESPACE: {{ .Values.kubernetes.namespace | default .Release.Namespace }}
{{- if and .Values.kubernetes .Values.kubernetes.workerAppName }}
KUBERNETES_WORKER_APP_NAME: {{ .Values.kubernetes.workerAppName }}
{{- end }}
{{- if and .Values.protometrics .Values.protometrics.enabled }}
METRICS_SERVICE_ENABLED: {{ .Values.protometrics.enabled | default "false" | quote }}
{{- end }}
{{- if and .Values.protometrics .Values.protometrics.url }}
METRICS_URL: {{ .Values.protometrics.url | quote }}
{{- end }}
{{- if and .Values.protometrics .Values.protometrics.apiKey }}
METRICS_API_KEY: {{ .Values.protometrics.apiKey | quote }}
{{- end }}
{{- if and .Values.runLogs .Values.runLogs.enabled }}
RUN_LOGS_ENABLED: {{ .Values.runLogs.enabled | toString | quote }}
RUN_LOGS_TYPE: {{ .Values.runLogs.type | quote }}
{{- if eq .Values.runLogs.type "k8s-pods" }}
RUN_LOGS_PODCONFIG_WORKER_NAMESPACE: {{ default .Release.Namespace .Values.runLogs.podConfig.workerNamespace | quote }}
RUN_LOGS_PODCONFIG_WORKER_APPNAME: {{ .Values.runLogs.podConfig.workerAppName | quote }}
{{- end }}
{{- if eq .Values.runLogs.type "loki" }}
RUN_LOGS_LOKICONFIG_BASEURL: {{ .Values.runLogs.lokiConfig.baseUrl | quote }}
RUN_LOGS_LOKICONFIG_LABELSQUERY: {{ .Values.runLogs.lokiConfig.labelsQuery | quote }}
{{- if .Values.runLogs.lokiConfig.keepLabels }}
RUN_LOGS_LOKICONFIG_KEEPLABELS: {{ .Values.runLogs.lokiConfig.keepLabels | join "," | quote }}
{{- end }}
{{- end }} # ends loki check
{{- end }} # ends runLogs.enabled check
{{- if and .Values.ee .Values.ee.license }}
EE_LICENSE: {{ .Values.ee.license | quote }}
{{- end }}
{{- end -}}

{{/*
Generate the stringData section for environment variables
*/}}
{{- define "neosync-api.migration-env-vars" -}}
DB_HOST: {{ .Values.migrations.db.host }}
DB_PORT: {{ .Values.migrations.db.port | quote }}
DB_NAME: {{ .Values.migrations.db.name }}
DB_USER: {{ .Values.migrations.db.username }}
DB_PASS: {{ .Values.migrations.db.password }}
DB_SSL_DISABLE: {{ .Values.migrations.db.disableSsl | quote}}
{{- if .Values.migrations.db.options }}
DB_MIGRATIONS_OPTIONS: {{ .Values.migrations.db.options | quote}}
{{- end }}
DB_SCHEMA_DIR: {{ .Values.migrations.db.schemaDir }}
DB_MIGRATIONS_TABLE: {{ .Values.migrations.db.migrationsTableName }}
DB_MIGRATIONS_TABLE_QUOTED: {{ .Values.migrations.db.migrationsTableQuoted | quote }}
{{- end -}}
