package scope

import (
	"github.com/launchboxio/operator/api/v1alpha1"
	"text/template"
)

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
	Users []v1alpha1.ProjectUser
}

var ValuesTemplate = template.Must(template.New("values").Parse(`
globalAnnotations:
  "launchboxhq.io/project-id": "{{ .ProjectId }}"
vcluster:
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
    - --tls-san="api.{{ .ProjectSlug }}.{{ .Ingress.Domain }}"
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
