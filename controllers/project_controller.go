/*
Copyright 2023.

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
	"github.com/launchboxio/operator/internal/scope"
	vclusterv1alpha1 "github.com/loft-sh/cluster-api-provider-vcluster/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/launchboxio/operator/api/v1alpha1"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.launchboxhq.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.launchboxhq.io,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.launchboxhq.io,resources=projects/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=namespaces,verbs=list;get;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Project object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Starting reconcile")
	project := &corev1alpha1.Project{}
	err := r.Get(ctx, req.NamespacedName, project)
	if err != nil {
		// TODO: Operator is not deleting projects / namespaces as expected
		if apierrors.IsNotFound(err) {
			logger.Info("Resource not found, must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed lookup for project resource")
		return ctrl.Result{}, err
	}

	projectLogger := logger.WithValues("project", project.Spec.Slug)

	dynClient, err := r.LoadDynamicClient()
	if err != nil {
		projectLogger.Error(err, "Failed loading dynamic client")
		return ctrl.Result{}, err
	}

	oidcClientId := os.Getenv("OIDC_CLIENT_ID")
	oidcIssuerUrl := os.Getenv("OIDC_ISSUER_URL")

	projectScope := scope.ProjectScope{
		Project:       project,
		Logger:        projectLogger,
		Client:        r.Client,
		DynamicClient: dynClient,
		OidcClientId:  oidcClientId,
		OidcIssuerUrl: oidcIssuerUrl,
	}
	return projectScope.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Project{}).
		Owns(&v1.Namespace{}).
		Owns(&clusterv1.Cluster{}).
		Owns(&vclusterv1alpha1.VCluster{}).
		Complete(r)
}

func (r *ProjectReconciler) LoadDynamicClient() (*dynamic.DynamicClient, error) {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}

		return dynamic.NewForConfig(config)

	} else {
		config, err := clientcmd.BuildConfigFromFlags(
			"", homedir.HomeDir()+"/.kube/config",
		)
		if err != nil {
			return nil, err
		}
		// create the clientset
		return dynamic.NewForConfig(config)
	}
}
