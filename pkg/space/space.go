package space

import (
	"bytes"
	"fmt"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"log"
	"os"
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
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(opts.Client, opts.Namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return err
	}

	client := action.NewUpgrade(actionConfig)
	client.Install = true
	client.Namespace = opts.Namespace

	chart := fmt.Sprintf("%s/%s", opts.Repo, opts.Chart)
	cp, err := client.ChartPathOptions.LocateChart(chart, nil)
	if err != nil {
		return err
	}

	chartReq, err := loader.Load(cp)
	if err != nil {
		return err
	}

	if req := chartReq.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chartReq, req); err != nil {
			return err
		}
	}

	// TODO: Use generate values to get the configurations
	values := map[string]interface{}{}

	hClient := action.NewHistory(actionConfig)
	hClient.Max = 1
	if _, err := hClient.Run(opts.Namespace); err == driver.ErrReleaseNotFound {
		iClient := action.NewInstall(actionConfig)
		iClient.Namespace = opts.Namespace
		//iClient.Wait = true
		iClient.ReleaseName = opts.Name
		_, err = iClient.Run(chartReq, values)
		return err
	}

	_, err = client.Run(opts.Name, chartReq, values)
	return err
}

func generateValues(opts *InstallOpts) (string, error) {
	values := ` 
service:
  {{- if .ServiceType }}
  type: {{ .ServiceType }}
  {{- else }}
  type: ClusterIP
  {{- end }}

  {{- if .CidrRanges }}
  loadBalancerSourceRanges: 
  {{ range .CidrRanges }}
  - {{ . }}
  {{- end }}
  {{- end }}

storage:
  {{- if .DiskSize }}
  size: {{ .DiskSize }}
  {{- else }}
  size: 50Gi
  {{- end }}

{{- with .DnsHostName }}
ingress:
  hostname: {{ . }}
{{- end }}
`
	tmpl, err := template.New("vcluster").Parse(values)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, opts)
	return buf.String(), err
}
