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
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/repo"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/launchboxio/operator/api/v1alpha1"

	"github.com/launchboxio/operator/pkg/addons"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.launchboxhq.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.launchboxhq.io,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.launchboxhq.io,resources=clusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO: Get this from somewhere else
	clientConfig := kube.GetConfig("/Users/rwittman/.kube/config", "launchboxhq", "default")

	cluster := &corev1alpha1.Cluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Install any configured addons
	installer := addons.NewInstaller(clientConfig)

	// Initialize the vcluster Repo as well
	repos := append(cluster.Spec.Repos, corev1alpha1.HelmRepo{
		Name: "loft-sh",
		Url:  "https://charts.loft.sh",
	})
	for _, r := range repos {
		if err := installer.InitRepo(&repo.Entry{
			Name:     r.Name,
			URL:      r.Url,
			Username: r.Username,
			Password: r.Password,
		}); err != nil {
			return ctrl.Result{}, err
		}
		fmt.Printf("[Cluster] Repo %s successfully initialized\n", r.Name)
	}

	for _, addon := range cluster.Spec.Addons {
		// Check if the release is already installed? If it is, continue
		release, err := installer.Exists(&addon.HelmRef)
		if err != nil {
			return ctrl.Result{}, err
		}
		if release != nil {
			fmt.Printf("[Cluster] Addon %s/%s already installed\n", addon.Namespace, addon.Name)
			continue
		}
		if _, err = installer.Ensure(&addon.HelmRef); err != nil {
			// TODO: Commit release error to cluster status
			return ctrl.Result{}, err
		}
		// TODO: Commit release spec to cluster status
		// Requeue to install the next helm chart
		return ctrl.Result{Requeue: true}, nil
	}
	// TODO: Generate a configmap with the OIDC configuration

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Cluster{}).
		Complete(r)
}
