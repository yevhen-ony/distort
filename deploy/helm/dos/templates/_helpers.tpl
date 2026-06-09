{{- define "dos.masterPeers" -}}
{{- $service := .Values.master.serviceName -}}
{{- range $i := until (.Values.master.replicas | int) -}}
{{- if $i }},{{ end -}}
master-{{ $i }}.{{ $service }}
{{- end -}}
{{- end -}}
