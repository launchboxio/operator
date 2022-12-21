package space

import (
	"bytes"
	corev1alpha1 "github.com/launchboxio/operator/api/v1alpha1"
	"github.com/launchboxio/operator/pkg/addons"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"text/template"
)

type InstallOpts struct {
	Namespace   string
	Name        string
	Repo        string
	Chart       string
	Version     string
	Client      genericclioptions.RESTClientGetter
	ServiceType string
	DiskSize    int64
	CidrRanges  []string
	DnsHostName string
}

func Install(opts *InstallOpts) error {
	addon := &corev1alpha1.ClusterAddonSpec{
		HelmRef: corev1alpha1.HelmRef{
			// TODO: Allow configuring to eventually test beta releases
			Repo:      "loft-sh",
			Chart:     "vcluster",
			Namespace: opts.Namespace,
			Name:      opts.Name,
			Version:   opts.Version,
		},
	}

	values, err := generateValues(opts)
	if err != nil {
		return err
	}
	addon.HelmRef.Values = values
	installer := addons.NewInstaller(opts.Client)
	release, err := installer.Exists(&addon.HelmRef)
	if err != nil {
		return err
	}

	if release != nil {
		return nil
	}

	_, err = installer.Ensure(&addon.HelmRef)
	if err != nil {
		return err
	}

	return nil
}

func generateValues(opts *InstallOpts) (string, error) {
	values := ` 
service:
  {{- if opts.ServiceType }}
  type: {{ opts.ServiceType }}
  {{- else }}
  type: ClusterIP
  {{- end }}

  {{- with opts.CidrRanges }}
  loadBalancerSourceRanges: {{ toYaml . | nindent 4 }}
  {{- end }}

storage:
  {{- if opts.DiskSize }}
  size: {{ opts.DiskSize }}
  {{- else }}
  size: 50Gi
  {{- end }}

{{- with opts.DnsHostName }}
ingress:
  hostname: {{ . }}
{{- end }}
`
	tmpl, err := template.New("test").Parse(values)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, opts)
	return buf.String(), err
}
