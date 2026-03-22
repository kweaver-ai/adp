{{/*
Expand the name of the chart.
*/}}
{{- define "dataflow.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "dataflow.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
ecron-management fullname
*/}}
{{- define "dataflow.ecronManagement.fullname" -}}
ecron-management
{{- end }}

{{/*
flow-automation fullname
*/}}
{{- define "dataflow.flowAutomation.fullname" -}}
flow-automation
{{- end }}

{{/*
flow-stream-data-pipeline fullname
*/}}
{{- define "dataflow.streamDataPipeline.fullname" -}}
flow-stream-data-pipeline
{{- end }}

{{/*
Common labels
*/}}
{{- define "dataflow.labels" -}}
helm.sh/chart: {{ include "dataflow.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
{{- end }}

{{/*
flow-automation labels
*/}}
{{- define "flow-automation.labels" -}}
helm.sh/chart: {{ include "dataflow.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Helper to merge component and global dependency services
Usage: {{ include "dataflow.utils.mergeDep" (dict "component" .Values.ecronAnalysis.depServices.rds "global" .Values.depServices.rds) }}
*/}}
{{- define "dataflow.utils.mergeDep" -}}
{{- $component := .component | default dict -}}
{{- $global := .global | default dict -}}
{{- merge $component $global | toYaml -}}
{{- end }}


{{/* ========== Universal Global Values Merge Helpers ========== */}}
{{/* All charts use these same helper function names for consistency */}}

{{- define "mergedGlobalValues.imageRegistry" -}}
{{- $globalImage := (.Values.global | default dict).image | default dict -}}
{{- if $globalImage.registry -}}
{{- $globalImage.registry -}}
{{- else -}}
{{- .Values.image.registry -}}
{{- end -}}
{{- end -}}

{{- define "mergedGlobalValues.replicaCount" -}}
{{- $global := .Values.global | default dict -}}
{{- if hasKey $global "replicaCount" -}}
{{- $global.replicaCount -}}
{{- else -}}
{{- .Values.replicaCount -}}
{{- end -}}
{{- end -}}

{{- define "mergedGlobalValues.env" -}}
{{- $globalEnv := (.Values.global | default dict).env | default dict -}}
{{- if $globalEnv -}}
{{- toYaml (mergeOverwrite (deepCopy (.Values.env | default dict)) $globalEnv) -}}
{{- else -}}
{{- toYaml .Values.env -}}
{{- end -}}
{{- end -}}

{{- define "mergedGlobalValues.depServices" -}}
{{- $globalDeps := (.Values.global | default dict).depServices | default dict -}}
{{- if $globalDeps -}}
{{- toYaml (mergeOverwrite (deepCopy (.Values.depServices | default dict)) $globalDeps) -}}
{{- else -}}
{{- toYaml .Values.depServices -}}
{{- end -}}
{{- end -}}

{{- define "mergedGlobalValues.imagePullSecrets" -}}
{{- $localSecrets := .Values.imagePullSecrets | default (list) -}}
{{- $globalSecrets := (.Values.global | default dict).imagePullSecrets | default (list) -}}
{{- if gt (len $localSecrets) 0 -}}
{{- toYaml $localSecrets -}}
{{- else if gt (len $globalSecrets) 0 -}}
{{- toYaml $globalSecrets -}}
{{- end -}}
{{- end -}}

{{- define "mergedGlobalValues.namespace" -}}
{{- $global := .Values.global | default dict -}}
{{- if hasKey $global "namespace" -}}
{{- $global.namespace -}}
{{- else -}}
{{- .Values.namespace -}}
{{- end -}}
{{- end -}}

{{- define "mergedGlobalValues.mode" -}}
{{- $global := .Values.global | default dict -}}
{{- if hasKey $global "mode" -}}
{{- $global.mode -}}
{{- else -}}
{{- .Values.mode | default "Community" -}}
{{- end -}}
{{- end -}}

{{- define "mergedGlobalValues.accessAddress" -}}
{{- $globalAccess := (.Values.global | default dict).accessAddress | default dict -}}
{{- if $globalAccess -}}
{{- toYaml (mergeOverwrite (deepCopy (.Values.accessAddress | default dict)) $globalAccess) -}}
{{- else -}}
{{- toYaml .Values.accessAddress -}}
{{- end -}}
{{- end -}}

{{- define "mergedGlobalValues.ingressClassName" -}}
{{- $global := .Values.global | default dict -}}
{{- if hasKey $global "ingressClassName" -}}
{{- $global.ingressClassName -}}
{{- else -}}
{{- .Values.ingressClassName | default "nginx" -}}
{{- end -}}
{{- end -}}

{{- define "mergedGlobalValues.flowAutomation" -}}
{{- $globalFlow := (.Values.global | default dict).flowAutomation | default dict -}}
{{- if $globalFlow -}}
{{- toYaml (mergeOverwrite (deepCopy (.Values.flowAutomation | default dict)) $globalFlow) -}}
{{- else -}}
{{- toYaml .Values.flowAutomation -}}
{{- end -}}
{{- end -}}

{{- define "mergedGlobalValues.image" -}}
{{- $imageRegistry := include "mergedGlobalValues.imageRegistry" . -}}
{{- if $imageRegistry }}
{{- printf "%s/%s:%s" $imageRegistry .Values.image.repository .Values.image.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.repository .Values.image.tag -}}
{{- end -}}
{{- end -}}
