{{/*
Expand the name of the chart.
*/}}
{{- define "vega-metadata.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "vega-metadata.fullname" -}}
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
{{- define "vega-metadata.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "vega-metadata.labels" -}}
helm.sh/chart: {{ include "vega-metadata.chart" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}
{{- define "vega-metadata.labels.backend" -}}
{{ include "vega-metadata.labels" . }}
{{ include "vega-metadata.backend.selectorLabels" . }}
{{- end }}
{{- define "vega-metadata.labels.frontend" -}}
{{ include "vega-metadata.labels" . }}
{{ include "vega-metadata.frontend.selectorLabels" . }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "vega-metadata.backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vega-metadata.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
{{- define "vega-metadata.frontend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vega-metadata.fullname" . }}-frontend
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "vega-metadata.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "vega-metadata.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
