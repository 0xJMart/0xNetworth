{{/*
Expand the name of the chart.
*/}}
{{- define "0xnetworth.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "0xnetworth.fullname" -}}
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
{{- define "0xnetworth.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "0xnetworth.labels" -}}
helm.sh/chart: {{ include "0xnetworth.chart" . }}
{{ include "0xnetworth.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "0xnetworth.selectorLabels" -}}
app.kubernetes.io/name: {{ include "0xnetworth.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Backend fullname
*/}}
{{- define "0xnetworth.backend.fullname" -}}
{{- printf "%s-backend" (include "0xnetworth.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Backend service name
*/}}
{{- define "0xnetworth.backend.serviceName" -}}
{{- printf "%s-backend" (include "0xnetworth.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Frontend fullname
*/}}
{{- define "0xnetworth.frontend.fullname" -}}
{{- printf "%s-frontend" (include "0xnetworth.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Frontend service name
*/}}
{{- define "0xnetworth.frontend.serviceName" -}}
{{- printf "%s-frontend" (include "0xnetworth.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Backend labels
*/}}
{{- define "0xnetworth.backend.labels" -}}
{{ include "0xnetworth.labels" . }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Backend selector labels
*/}}
{{- define "0xnetworth.backend.selectorLabels" -}}
{{ include "0xnetworth.selectorLabels" . }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Frontend labels
*/}}
{{- define "0xnetworth.frontend.labels" -}}
{{ include "0xnetworth.labels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Frontend selector labels
*/}}
{{- define "0xnetworth.frontend.selectorLabels" -}}
{{ include "0xnetworth.selectorLabels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Namespace
*/}}
{{- define "0xnetworth.namespace" -}}
{{- if .Values.namespace.create }}
{{- .Values.namespace.name }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}

