{{/*
Return the appropriate apiVersion for ingress.
*/}}
{{- define "site-manager.ingress.apiVersion" -}}
  {{- if and (.Capabilities.APIVersions.Has "networking.k8s.io/v1") (semverCompare ">= 1.19-0" .Capabilities.KubeVersion.Version) -}}
      {{- print "networking.k8s.io/v1" -}}
  {{- else if .Capabilities.APIVersions.Has "networking.k8s.io/v1beta1" -}}
    {{- print "networking.k8s.io/v1beta1" -}}
  {{- else -}}
    {{- print "extensions/v1beta1" -}}
  {{- end -}}
{{- end -}}
{{/*
DNS names used to generate SSL certificate with "Subject Alternative Name" field
*/}}
{{- define "site-manager.certDnsNames" -}}
  {{- $dnsNames := list "localhost" "site-manager" (printf "%s.%s" "site-manager" .Release.Namespace)  (printf "%s.%s.svc" "site-manager" .Release.Namespace) -}}
  {{- if .Values.ingress.name -}}
    {{- $dnsNames = append $dnsNames .Values.ingress.name -}}
   {{- end -}}
  {{- $dnsNames = concat $dnsNames .Values.tls.generateCerts.subjectAlternativeName.additionalDnsNames -}}
  {{- $dnsNames | toYaml -}}
{{- end -}}
{{/*
IP addresses used to generate SSL certificate with "Subject Alternative Name" field
*/}}
{{- define "site-manager.certIpAddresses" -}}
  {{- $ipAddresses := list "127.0.0.1" -}}
  {{- $ipAddresses = concat $ipAddresses .Values.tls.generateCerts.subjectAlternativeName.additionalIpAddresses -}}
  {{- $ipAddresses | toYaml -}}
{{- end -}}

{{- define "paas-geo-monitor.port" -}}
  {{- print ( default 8080 .Values.paas_geo_monitor.config.port ) -}}
{{- end -}}
