package project

import (
	"github.com/launchboxio/operator/api/v1alpha1"
	"text/template"
)

var ImageMapping = map[string]string{
	"1.28.3":  "rancher/k3s:v1.28.3-k3s2",
	"1.28.2":  "rancher/k3s:v1.28.2-k3s1",
	"1.28.1":  "rancher/k3s:v1.28.1-k3s1",
	"1.28.0":  "rancher/k3s:v1.28.0-rc1-k3s1",
	"1.27.7":  "rancher/k3s:v1.27.7-k3s2",
	"1.27.6":  "rancher/k3s:v1.27.6-k3s1",
	"1.27.5":  "rancher/k3s:v1.27.5-k3s1",
	"1.27.4":  "rancher/k3s:v1.27.4-k3s1",
	"1.27.3":  "rancher/k3s:v1.27.3-k3s1",
	"1.27.2":  "rancher/k3s:v1.27.2-k3s1",
	"1.27.1":  "rancher/k3s:v1.27.1-k3s1",
	"1.26.10": "rancher/k3s:v1.26.10-k3s2",
	"1.26.9":  "rancher/k3s:v1.26.9-k3s1",
	"1.26.8":  "rancher/k3s:v1.26.8-k3s1",
	"1.26.7":  "rancher/k3s:v1.26.8-k3s1",
	"1.26.6":  "rancher/k3s:v1.26.6-k3s1",
	"1.26.5":  "rancher/k3s:v1.26.5-k3s1",
	"1.26.4":  "rancher/k3s:v1.26.4-k3s1",
	"1.26.3":  "rancher/k3s:v1.26.3-k3s1",
	"1.26.2":  "rancher/k3s:v1.26.2-k3s1",
	"1.26.1":  "rancher/k3s:v1.26.1-k3s1",
	"1.26.0":  "rancher/k3s:v1.26.0-k3s2",
}

type ValuesTemplateArgs struct {
	ProjectId   int
	ProjectSlug string
	Cpu         int32
	Memory      int32
	Disk        int32
	Oidc        struct {
		ClientId  string
		IssuerUrl string
	}
	Ingress struct {
		ClassName string
		Domain    string
	}
	Image string
	Users []v1alpha1.ProjectUser
}

var ValuesTemplate = template.Must(template.New("values").Parse(`
globalAnnotations:
  "launchboxhq.io/project-id": "{{ .ProjectId }}"
vcluster:
  image: {{ .Image }}
  {{- with .Oidc }}
  extraArgs:
    - "--kube-apiserver-arg=--oidc-username-claim=preferred_username"
    - "--kube-apiserver-arg=--oidc-issuer-url={{ .IssuerUrl }}"
    - "--kube-apiserver-arg=--oidc-client-id={{ .ClientId }}"
    - "--kube-apiserver-arg=--oidc-username-claim=email"
    - "--kube-apiserver-arg=--oidc-groups-claim=groups"
  {{- end }}
  resources:
    limits:
      cpu: {{ .Cpu }}
      memory: "{{ .Memory }}Mi"

storage:
  persistence: true
  size: "{{ .Disk }}Gi"
sync:
  ingresses:
    enabled: true
  serviceaccounts:
    enabled: true
syncer:
  extraArgs:
    - --tls-san="{{ .ProjectSlug }}.{{ .ProjectSlug }}"
    - --out-kube-config-server=https://{{ .ProjectSlug }}.{{ .ProjectSlug }}
ingress:
  enabled: true
  ingressClassName: "{{ .Ingress.ClassName }}"
  annotations:
    nginx.ingress.kubernetes.io/backend-protocol: HTTPS
    nginx.ingress.kubernetes.io/ssl-passthrough: "true"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
  host: "api.{{ .ProjectSlug }}.{{ .Ingress.Domain }}"

init:
  manifests: |
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: admins
    subjects:
    {{- range $user := .Users }}
    - kind: User
      name: {{ $user.Email }}
      apiGroup: rbac.authorization.k8s.io
    {{- end }}
    roleRef:
      kind: ClusterRole
      name: cluster-admin
      apiGroup: rbac.authorization.k8s.io

`))
