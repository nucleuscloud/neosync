{{/*
Expand the name of the chart.
*/}}
{{- define "neosync-app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "neosync-app.fullname" -}}
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
{{- define "neosync-app.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "neosync-app.labels" -}}
helm.sh/chart: {{ include "neosync-app.chart" . }}
{{ include "neosync-app.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "neosync-app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "neosync-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "neosync-app.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "neosync-app.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Generate the stringData section for environment variables
*/}}
{{- define "neosync-app.env-vars" -}}
{{- if .Values.host }}
HOSTNAME: {{ .Values.host | quote}}
{{- end }}
PORT: {{ .Values.containerPort | quote }}
{{- if .Values.otel.enabled }}
OTEL_EXPORTER_OTLP_PORT: {{ .Values.otel.otlpPort | quote }} # sends to gRPC receiver
{{- end }}
{{- if .Values.nucleusEnv }}
NUCLEUS_ENV: {{ .Values.nucleusEnv }}
{{- end }}
{{- if .Values.neosyncApi.url }}
NEOSYNC_API_BASE_URL: {{ .Values.neosyncApi.url }}
{{- end }}
NEXTAUTH_SECRET: {{ .Values.nextAuthSecret }}
{{- if .Values.nextAuthUrl }}
NEXTAUTH_URL: {{ .Values.nextAuthUrl }}
{{- end }}
{{- if .Values.auth.clientId }}
AUTH_CLIENT_ID: {{ .Values.auth.clientId }}
{{- end }}
{{- if .Values.auth.clientSecret }}
AUTH_CLIENT_SECRET: {{ .Values.auth.clientSecret }}
{{- end }}
{{- if .Values.auth.issuer }}
AUTH_ISSUER: {{ .Values.auth.issuer }}
{{- end }}
{{- if .Values.auth.expectedIssuer }}
AUTH_EXPECTED_ISSUER: {{ .Values.auth.expectedIssuer }}
{{- end }}
{{- if .Values.auth.authorizeUrl }}
AUTH_AUTHORIZE_URL: {{ .Values.auth.authorizeUrl }}
{{- end }}
{{- if .Values.auth.userinfoUrl }}
AUTH_USERINFO_URL: {{ .Values.auth.userinfoUrl }}
{{- end }}
{{- if .Values.auth.tokenUrl }}
AUTH_TOKEN_URL: {{ .Values.auth.tokenUrl }}
{{- end }}
{{- if .Values.auth.logoutUrl }}
AUTH_LOGOUT_URL: {{ .Values.auth.logoutUrl }}
{{- end}}
{{- if .Values.auth.scope }}
AUTH_SCOPE: {{ .Values.auth.scope }}
{{- end }}
{{- if .Values.auth.audience }}
AUTH_AUDIENCE: {{ .Values.auth.audience }}
{{- end }}
{{- if .Values.auth.providerId }}
AUTH_PROVIDER_ID: {{ .Values.auth.providerId }}
{{- end }}
{{- if .Values.auth.providerName }}
AUTH_PROVIDER_NAME: {{ .Values.auth.providerName }}
{{- end }}
{{- if .Values.nextPublic.appBaseUrl }}
NEXT_PUBLIC_APP_BASE_URL: {{ .Values.nextPublic.appBaseUrl }}
{{- end }}
AUTH_ENABLED: {{ .Values.auth.enabled | default "false" | quote }}
AUTH_TRUST_HOST: {{ .Values.auth.trustHost | default "true" | quote }}
NEOSYNC_ANALYTICS_ENABLED: {{ .Values.analytics.enabled | default "true" | quote }}
{{- if and .Values.posthog .Values.posthog.key }}
POSTHOG_KEY: {{ .Values.posthog.key }}
{{- end }}
{{- if and .Values.posthog .Values.posthog.host }}
POSTHOG_HOST: {{ .Values.posthog.host }}
{{- end }}
{{- if and .Values.unify .Values.unify.key }}
UNIFY_KEY: {{ .Values.unify.key }}
{{- end }}
NEOSYNC_CLOUD: {{ .Values.neosyncCloud.enabled | default "false" | quote }}
ENABLE_RUN_LOGS: {{ .Values.enableRunLogs | default "false" | quote }}
{{- if and .Values.protometrics .Values.protometrics.enabled }}
METRICS_SERVICE_ENABLED: {{ .Values.protometrics.enabled | default "false" | quote }}
{{- end }}
GCP_CS_CONNECTIONS_DISABLED: {{ .Values.disableGcpCloudStorageConnections | default "false" | quote }}
JOBHOOKS_ENABLED: {{ .Values.jobHooks.enabled | default "false" | quote }}
{{- end -}}
