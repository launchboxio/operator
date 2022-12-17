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
	"github.com/launchboxio/operator/pkg/addons"
	"helm.sh/helm/v3/pkg/repo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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
		return ctrl.Result{}, err
	}

	// TODO: Install the vcluster helm chart
	vclusterOpts := &space.InstallOpts{
		Namespace: s.Namespace,
		Name:      s.Name,
	}

	if err := space.Install(vclusterOpts); err != nil {
		return ctrl.Result{}, err
	}
	fmt.Println("Space successfully started, checking for addons")
	// TODO: Configure the addons installer for vcluster

	vclusterSecret := &v1.Secret{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("vc-%s", s.Name),
		Namespace: s.Namespace,
	}, vclusterSecret)
	if err != nil {
		return ctrl.Result{}, err
	}
	vclusterKubeConfig := string(vclusterSecret.Data["config"])
	client := genericclioptions.NewConfigFlags(false)
	client.KubeConfig = &vclusterKubeConfig

	installer := addons.NewInstaller(client)
	for _, r := range s.Spec.Repos {
		err = installer.InitRepo(&repo.Entry{
			Name:     r.Name,
			URL:      r.Url,
			Username: r.Username,
			Password: r.Password,
		})
		fmt.Printf("[%s/%s] Repo %s successfully initialized\n", s.Namespace, s.Name, r.Name)
	}

	for _, addon := range s.Spec.Addons {
		// Check if the release is already installed? If it is, continue
		release, err := installer.Exists(&addon.HelmRef)
		if err != nil {
			return ctrl.Result{}, err
		}

		if release != nil {
			fmt.Printf("[%s/%s] Addon %s/%s already installed\n", s.Namespace, s.Name, addon.Namespace, addon.Name)
			continue
		}
		_, err = installer.Ensure(&addon.HelmRef)
		if err != nil {
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
func (r *SpaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Space{}).
		Complete(r)
}
