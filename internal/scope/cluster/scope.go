package cluster

import (
	"bytes"
	"context"
	"fmt"
	"github.com/launchboxio/operator/api/v1alpha1"
	helmclient "github.com/mittwald/go-helm-client"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"text/template"
)

type Scope struct {
	Cluster *v1alpha1.Cluster
	Client  client.Client
}

const clusterFinalizer = "core.launchboxhq.io/finalizer"

func (s *Scope) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	conf := config.GetConfigOrDie()
	helm, err := helmclient.NewClientFromRestConf(&helmclient.RestConfClientOptions{
		RestConfig: conf,
		Options: &helmclient.Options{
			Debug:     true,
			Namespace: "lbx-system",
		},
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	values, err := generateAgentValues(s.Cluster.Spec)
	if err != nil {
		return ctrl.Result{}, err
	}
	chartSpec := &helmclient.ChartSpec{
		ReleaseName: "agent",
		ChartName:   "oci://ghcr.io/launchboxio/agent/helm/agent",
		Namespace:   "lbx-system",
		Version:     s.Cluster.Spec.Agent.ChartVersion,
		ValuesYaml:  string(values),
	}

	isAgentMarkedToBeDeleted := s.Cluster.GetDeletionTimestamp() != nil
	if isAgentMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(s.Cluster, clusterFinalizer) {
			rel, err := helm.GetRelease(chartSpec.ReleaseName)
			if rel != nil {
				if err := helm.UninstallRelease(chartSpec); err != nil {
					return ctrl.Result{}, err
				}
			}

			controllerutil.RemoveFinalizer(s.Cluster, clusterFinalizer)
			err = s.Client.Update(ctx, s.Cluster)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if _, err = helm.InstallOrUpgradeChart(ctx, &helmclient.ChartSpec{
		ReleaseName: "agent",
		ChartName:   "oci://ghcr.io/launchboxio/agent/helm/agent",
		Namespace:   "lbx-system",
		Version:     s.Cluster.Spec.Agent.ChartVersion,
		ValuesYaml:  string(values),
	}, nil); err != nil {
		return ctrl.Result{}, err
	}

	if !controllerutil.ContainsFinalizer(s.Cluster, clusterFinalizer) {
		controllerutil.AddFinalizer(s.Cluster, clusterFinalizer)
		err = s.Client.Update(ctx, s.Cluster)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Finally, update the status conditions
	fmt.Println("Updating status conditions")
	meta.SetStatusCondition(&s.Cluster.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "Installed",
		Message: fmt.Sprintf("Chart %s has been installed", chartSpec.Version),
	})
	return ctrl.Result{}, s.Client.Status().Update(ctx, s.Cluster)

}

func generateAgentValues(spec v1alpha1.ClusterSpec) ([]byte, error) {
	tmpl, err := template.New("values").Parse(`
image:
  {{- if .Agent.Repository }}
  repository: "{{ .Agent.Repository }}"
  {{- end }}
  {{- if .Agent.PullPolicy }}
  pullPolicy: "{{ .Agent.PullPolicy }}"
  {{- end }}
  {{- if .Agent.Tag }}
  tag: "{{ .Agent.Tag }}"
  {{- end }}
agent:
  tokenUrl: {{ .Launchbox.TokenUrl }}
  apiUrl: {{ .Launchbox.ApiUrl }}
  streamUrl: {{ .Launchbox.StreamUrl }}
  clusterId: {{ .ClusterId }}
  channel: {{ .Launchbox.Channel }}
credentialsSecret:
  name: {{ .CredentialsRef.Name }}
`)
	if err != nil {
		return nil, err
	}
	var values bytes.Buffer
	err = tmpl.Execute(&values, spec)
	return values.Bytes(), err
}
