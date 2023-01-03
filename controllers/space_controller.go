/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/kube"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/launchboxio/operator/api/v1alpha1"
	"github.com/launchboxio/operator/pkg/space"
)

// SpaceReconciler reconciles a Space object
type SpaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.launchboxhq.io,resources=spaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.launchboxhq.io,resources=spaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.launchboxhq.io,resources=spaces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Space object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *SpaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	s := &corev1alpha1.Space{}
	err := r.Get(ctx, req.NamespacedName, s)
	if err != nil {
		return ctrl.Result{}, nil
	}

	hostConfig := kube.GetConfig("/Users/rwittman/.kube/config", "minikube", "default")

	// TODO: Install the vcluster helm chart
	vclusterOpts := &space.InstallOpts{
		Namespace: s.Namespace,
		Name:      s.Name,
		Client:    hostConfig,
	}

	if err := space.Install(vclusterOpts); err != nil {
		return ctrl.Result{}, err
	}

	// TODO: Configure the addons installer for vcluster
	vclusterSecret := &v1.Secret{}
	if err = r.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("vc-%s", s.Name),
		Namespace: s.Namespace,
	}, vclusterSecret); err != nil {
		return ctrl.Result{}, err
	}

	//
	// This should all be moved to the ServiceCatalog controller
	//

	//vclusterService := &v1.Service{}
	//if err = r.Get(ctx, types.NamespacedName{
	//	Namespace: s.Namespace,
	//	Name:      s.Name,
	//}, vclusterService); err != nil {
	//	return ctrl.Result{}, err
	//}
	//
	//file, err := ioutil.TempFile("/tmp", "vckc")
	//if err != nil {
	//	return ctrl.Result{}, err
	//}
	//
	//defer os.Remove(file.Name())
	//
	//kubeConfigContent := vclusterSecret.Data["config"]
	//if _, present := os.LookupEnv("KUBERNETES_SERVICE_HOST"); !present {
	//	// We aren't running in the cluster. We assume we'll have a single space,
	//	// and just proxy to 443 for the space
	//	kubeConfigContent, err = generateVclusterKubeConfig(vclusterSecret.Data["config"], "https://127.0.0.1:443")
	//}
	//
	//if err := os.WriteFile(file.Name(), kubeConfigContent, 0644); err != nil {
	//	return ctrl.Result{}, err
	//}
	//kubeConfig := file.Name()
	//iClient := genericclioptions.NewConfigFlags(false)
	//iClient.KubeConfig = &kubeConfig
	//
	//installer := addons.NewInstaller(iClient)
	//// TODO: The repos need to be scoped to the space. Since helm installs originate from the operator,
	//// we want to prevent name collisions, as well as prevent other spaces from installing charts from
	//// another space's private repos
	//for _, r := range s.Spec.Repos {
	//	if err = installer.InitRepo(&repo.Entry{
	//		Name:     r.Name,
	//		URL:      r.Url,
	//		Username: r.Username,
	//		Password: r.Password,
	//	}); err != nil {
	//		return ctrl.Result{}, err
	//	}
	//	fmt.Printf("[%s/%s] Repo %s successfully initialized\n", s.Namespace, s.Name, r.Name)
	//}
	//
	//// TODO: Handle removal of addons from the space
	//for _, addon := range s.Spec.Addons {
	//	// Check if the release is already installed? If it is, continue
	//	release, err := installer.Exists(&addon.HelmRef)
	//	if err != nil {
	//		return ctrl.Result{}, err
	//	}
	//
	//	if release != nil {
	//		fmt.Printf("[%s/%s] Addon %s/%s already installed\n", s.Namespace, s.Name, addon.Namespace, addon.Name)
	//		continue
	//	}
	//
	//	if _, err = installer.Ensure(&addon.HelmRef); err != nil {
	//		return ctrl.Result{}, err
	//	}
	//
	//	// TODO: Commit release spec to cluster status
	//	// Requeue to install the next helm chart
	//	return ctrl.Result{Requeue: true}, nil
	//}

	//
	// End of service catalog migration
	//
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SpaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Space{}).
		Complete(r)
}

func generateVclusterKubeConfig(kubeconfig []byte, newHost string) ([]byte, error) {
	result := &KubeConfigValue{}
	if err := yaml.Unmarshal(kubeconfig, result); err != nil {
		return nil, err
	}
	result.Clusters[0].Cluster.Server = newHost
	return yaml.Marshal(result)
}

// KubeConfigValue is a struct used to create a kubectl configuration YAML file.
type KubeConfigValue struct {
	APIVersion     string                   `yaml:"apiVersion"`
	Kind           string                   `yaml:"kind"`
	Clusters       []KubeconfigNamedCluster `yaml:"clusters"`
	Users          []KubeconfigUser         `yaml:"users"`
	Contexts       []KubeconfigNamedContext `yaml:"contexts"`
	CurrentContext string                   `yaml:"current-context"`
	Preferences    struct{}                 `yaml:"preferences"`
}

// KubeconfigUser is a struct used to create a kubectl configuration YAML file
type KubeconfigUser struct {
	Name string                `yaml:"name"`
	User KubeconfigUserKeyPair `yaml:"user"`
}

// KubeconfigUserKeyPair is a struct used to create a kubectl configuration YAML file
type KubeconfigUserKeyPair struct {
	ClientCertificateData string                 `yaml:"client-certificate-data"`
	ClientKeyData         string                 `yaml:"client-key-data"`
	AuthProvider          KubeconfigAuthProvider `yaml:"auth-provider,omitempty"`
}

// KubeconfigAuthProvider is a struct used to create a kubectl authentication provider
type KubeconfigAuthProvider struct {
	Name   string            `yaml:"name"`
	Config map[string]string `yaml:"config"`
}

// KubeconfigNamedCluster is a struct used to create a kubectl configuration YAML file
type KubeconfigNamedCluster struct {
	Name    string            `yaml:"name"`
	Cluster KubeconfigCluster `yaml:"cluster"`
}

// KubeconfigCluster is a struct used to create a kubectl configuration YAML file
type KubeconfigCluster struct {
	Server                   string `yaml:"server"`
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	CertificateAuthority     string `yaml:"certificate-authority"`
}

// KubeconfigNamedContext is a struct used to create a kubectl configuration YAML file
type KubeconfigNamedContext struct {
	Name    string            `yaml:"name"`
	Context KubeconfigContext `yaml:"context"`
}

// KubeconfigContext is a struct used to create a kubectl configuration YAML file
type KubeconfigContext struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace,omitempty"`
	User      string `yaml:"user"`
}
