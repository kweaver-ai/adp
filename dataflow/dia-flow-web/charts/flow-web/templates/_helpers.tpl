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

{{- define "mergedGlobalValues.image" -}}
{{- $imageRegistry := include "mergedGlobalValues.imageRegistry" . -}}
{{- if $imageRegistry -}}
{{- printf "%s%s:%s" $imageRegistry .Values.image.repository .Values.image.tag -}}
{{- else -}}
{{- printf "%s:%s" (.Values.image.repository | trimPrefix "/") .Values.image.tag -}}
{{- end -}}
{{- end -}}
