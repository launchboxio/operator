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
	"fmt"
	crossplanev1 "github.com/crossplane/crossplane/apis/pkg/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/launchboxio/operator/api/v1alpha1"
)

// AddonReconciler reconciles a Addon object
type AddonReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.launchboxhq.io,resources=addons,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.launchboxhq.io,resources=addons/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.launchboxhq.io,resources=addons/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Addon object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AddonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	addon := &corev1alpha1.Addon{}
	if err := r.Get(ctx, req.NamespacedName, addon); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Resource not found, must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed looking up Addon resource")
		return ctrl.Result{}, err
	}

	addonConfiguration := &crossplanev1.Configuration{}
	if err := r.Get(ctx, req.NamespacedName, addonConfiguration); err != nil {
		if apierrors.IsNotFound(err) {
			// Create the configuration
			c := r.configurationForAddon(addon)
			if err := r.Create(ctx, c); err != nil {
				logger.Error(err, "Failed creating addon configuration")
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		logger.Error(err, "Failed to get Configuration resource")
		return ctrl.Result{}, err
	}

	// TODO: Handle updates to the spec
	hasChanges := false
	packageName := fmt.Sprintf("%s:%s", addon.Spec.OciRegistry, addon.Spec.OciVersion)
	if addonConfiguration.Spec.Package != packageName {
		addonConfiguration.Spec.Package = packageName
		hasChanges = true
	}

	if hasChanges {
		err := r.Update(ctx, addonConfiguration)
		if err != nil {
			logger.Error(err, "Failed updating configuration")
			return ctrl.Result{}, err
		}
		logger.Info("Configuration updated")
	}

	// TODO: Update Addon status
	meta.SetStatusCondition(&addon.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "Installed",
		Message: fmt.Sprintf("Crossplane addon has been installed"),
	})
	err := r.Client.Status().Update(ctx, addon)
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *AddonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Addon{}).
		Complete(r)
}

func (r *AddonReconciler) configurationForAddon(addon *corev1alpha1.Addon) *crossplanev1.Configuration {
	pullPolicy := corev1.PullAlways
	if addon.Spec.PullPolicy == "IfNotPresent" {
		pullPolicy = corev1.PullIfNotPresent
	} else if addon.Spec.PullPolicy == "Never" {
		pullPolicy = corev1.PullNever
	}

	activationPolicy := crossplanev1.AutomaticActivation
	if addon.Spec.ActivationPolicy == "Manual" {
		activationPolicy = crossplanev1.ManualActivation
	}

	revisionHistoryLimit := int64(10)

	c := &crossplanev1.Configuration{
		ObjectMeta: metav1.ObjectMeta{
			Name: addon.Name,
		},
		Spec: crossplanev1.ConfigurationSpec{
			PackageSpec: crossplanev1.PackageSpec{
				Package:                  fmt.Sprintf("%s:%s", addon.Spec.OciRegistry, addon.Spec.OciVersion),
				PackagePullPolicy:        &pullPolicy,
				RevisionActivationPolicy: &activationPolicy,
				RevisionHistoryLimit:     &revisionHistoryLimit,
			},
		},
	}

	ctrl.SetControllerReference(addon, c, r.Scheme)
	return c
}
