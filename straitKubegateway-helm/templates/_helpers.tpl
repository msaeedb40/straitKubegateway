{{/*
Expand the name of the chart.
*/}}
{{- define "straitKubegateway.name" -}}
{{- default .Chart.Name .Values.nameOverride | lower | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "straitKubegateway.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | lower | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | lower | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | lower | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "straitKubegateway.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | lower | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "straitKubegateway.labels" -}}
helm.sh/chart: {{ include "straitKubegateway.chart" . }}
{{ include "straitKubegateway.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "straitKubegateway.selectorLabels" -}}
app.kubernetes.io/name: {{ include "straitKubegateway.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Daemon Service Account
*/}}
{{- define "straitKubegateway.agentServiceAccount" -}}
{{- printf "%s-agent" (include "straitKubegateway.fullname" .) }}
{{- end }}

{{/*
Controller Service Account
*/}}
{{- define "straitKubegateway.controllerServiceAccount" -}}
{{- printf "%s-controller" (include "straitKubegateway.fullname" .) }}
{{- end }}
