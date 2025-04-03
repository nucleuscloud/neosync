{{/*
Expand the name of the chart.
*/}}
{{- define "neosync-worker.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "neosync-worker.fullname" -}}
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
{{- define "neosync-worker.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "neosync-worker.labels" -}}
helm.sh/chart: {{ include "neosync-worker.chart" . }}
{{ include "neosync-worker.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "neosync-worker.selectorLabels" -}}
app.kubernetes.io/name: {{ include "neosync-worker.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "neosync-worker.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "neosync-worker.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Generate the stringData section for environment variables
*/}}
{{- define "neosync-worker.env-vars" -}}
{{- if .Values.host }}
HOST: {{ .Values.host | quote}}
{{- end }}
{{- if .Values.containerPort }}
PORT: {{ .Values.containerPort | quote }}
{{- end }}
{{- if .Values.otel.enabled }}
OTEL_EXPORTER_OTLP_PORT: {{ .Values.otel.otlpPort | quote }} # sends to gRPC receiver
{{- end }}
{{- if .Values.nucleusEnv }}
NUCLEUS_ENV: {{ .Values.nucleusEnv }}
{{- end }}
{{- if .Values.temporal.url }}
TEMPORAL_URL: {{ .Values.temporal.url }}
{{- end }}
{{- if .Values.temporal.namespace }}
TEMPORAL_NAMESPACE: {{ .Values.temporal.namespace }}
{{- end }}
{{- if .Values.temporal.taskQueue }}
TEMPORAL_TASK_QUEUE: {{ .Values.temporal.taskQueue }}
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
{{- if and .Values.neosync .Values.neosync.url }}
NEOSYNC_URL: {{ .Values.neosync.url }}
{{- end }}
{{- if and .Values.neosync .Values.neosync.apiKey }}
NEOSYNC_API_KEY: {{ .Values.neosync.apiKey }}
{{- end }}
{{- if .Values.redis.url }}
REDIS_URL: {{ .Values.redis.url }}
{{- end }}
{{- if .Values.redis.kind }}
REDIS_KIND: {{ .Values.redis.kind }}
{{- end }}
{{- if .Values.redis.master }}
REDIS_MASTER: {{ .Values.redis.master }}
{{- end }}
REDIS_TLS_ENABLED: {{ .Values.redis.tls.enabled | default "false" | quote }}
REDIS_TLS_SKIP_CERT_VERIFY: {{ .Values.redis.tls.skipCertVerify | default "false" | quote }}
REDIS_TLS_ENABLE_RENEGOTIATION: {{ .Values.redis.tls.enableRenegotiation | default "false" | quote }}
{{- if and .Values.redis .Values.redis.tls .Values.redis.tls.rootCertAuthority }}
REDIS_TLS_ROOT_CERT_AUTHORITY: {{ .Values.redis.tls.rootCertAuthority }}
{{- end }}
{{- if and .Values.redis .Values.redis.tls .Values.redis.tls.rootCertAuthorityFile }}
REDIS_TLS_ROOT_CERT_AUTHORITY_FILE: {{ .Values.redis.tls.rootCertAuthorityFile }}
{{- end }}
NEOSYNC_CLOUD: {{ .Values.neosyncCloud.enabled | default "false" | quote }}
{{- if and .Values.ee .Values.ee.license }}
EE_LICENSE: {{ .Values.ee.license | quote }}
{{- end }}
TABLESYNC_MAX_CONCURRENCY: {{ .Values.tableSync.maxConcurrency | quote }}
{{- end -}}
