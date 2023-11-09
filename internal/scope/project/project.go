package project

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/launchboxio/operator/api/v1alpha1"
	vclusterv1alpha1 "github.com/loft-sh/cluster-api-provider-vcluster/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type Scope struct {
	Project       *v1alpha1.Project
	Logger        logr.Logger
	Client        client.Client
	DynamicClient *dynamic.DynamicClient
	Cluster       *v1alpha1.Cluster
}

func (scope *Scope) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	identifier := scope.Project.Spec.Slug
	kubernetesVersion := scope.Project.Spec.KubernetesVersion
	defaultMeta := metav1.ObjectMeta{
		Name:      identifier,
		Namespace: identifier,
	}
	//  Ensure our namespace is created
	namespace := &v1.Namespace{}
	if err := scope.Client.Get(ctx, types.NamespacedName{Name: identifier}, namespace); err != nil {
		if apierrors.IsNotFound(err) {
			scope.Logger.Info("Creating namespace")
			ns := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: identifier,
				},
			}
			ctrl.SetControllerReference(scope.Project, ns, scope.Client.Scheme())
			// Create the namespace
			if err = scope.Client.Create(ctx, ns); err != nil {
				scope.Logger.Error(err, "Failed creating namespace")
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		scope.Logger.Error(err, "Failed lookup for namespace")
		return ctrl.Result{}, err
	}

	var values bytes.Buffer
	if err := ValuesTemplate.Execute(&values, getValuesArgs(scope)); err != nil {
		scope.Logger.Error(err, "Failed generating vcluster values")
		return ctrl.Result{}, err
	}

	// Create the infrastructure and vcluster resource
	infrastructure := &vclusterv1alpha1.VCluster{}
	if err := scope.Client.Get(ctx, types.NamespacedName{
		Name:      identifier,
		Namespace: identifier,
	}, infrastructure); err != nil {
		if apierrors.IsNotFound(err) {
			scope.Logger.Info("Creating new vcluster resource")
			// Create the namespace
			if err = scope.Client.Create(ctx, &vclusterv1alpha1.VCluster{
				ObjectMeta: defaultMeta,
				Spec: vclusterv1alpha1.VClusterSpec{
					ControlPlaneEndpoint: clusterv1.APIEndpoint{
						Host: "",
						Port: 0,
					},
					HelmRelease: &vclusterv1alpha1.VirtualClusterHelmRelease{
						Chart: vclusterv1alpha1.VirtualClusterHelmChart{},
						// TODO: We have to reconcile these as well, in the event they change
						Values: values.String(),
					},
					KubernetesVersion: &kubernetesVersion,
				},
			}); err != nil {
				scope.Logger.Error(err, "Failed creating vcluster resource")
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		scope.Logger.Error(err, "Failed lookup for vcluster resource")
		return ctrl.Result{}, err
	}

	if values.String() != infrastructure.Spec.HelmRelease.Values {
		scope.Logger.Info("Updating helm values")
		infrastructure.Spec.HelmRelease.Values = values.String()
		err := scope.Client.Update(ctx, infrastructure)
		return ctrl.Result{Requeue: true}, err
	}

	cluster := &clusterv1.Cluster{}
	if err := scope.Client.Get(ctx, types.NamespacedName{
		Name:      identifier,
		Namespace: identifier,
	}, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			scope.Logger.Info("Creating new cluster resource")
			if err = scope.Client.Create(ctx, &clusterv1.Cluster{
				ObjectMeta: defaultMeta,
				Spec: clusterv1.ClusterSpec{
					ControlPlaneRef: &v1.ObjectReference{
						Name:       identifier,
						Kind:       "VCluster",
						APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha1",
					},
					InfrastructureRef: &v1.ObjectReference{
						Name:       identifier,
						Kind:       "VCluster",
						APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha1",
					},
				},
			}); err != nil {
				scope.Logger.Error(err, "Failed creating new cluster resource")
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		scope.Logger.Error(err, "Failed lookup for cluster resource")
		return ctrl.Result{}, err
	}

	// TODO: Wait for the vcluster instance to be ready
	secret := &v1.Secret{}
	if err := scope.Client.Get(ctx, types.NamespacedName{
		Name:      "vc-" + identifier,
		Namespace: identifier,
	}, secret); err != nil {
		if apierrors.IsNotFound(err) {
			scope.Logger.Info("Waiting for vcluster secret to be available")
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}
		scope.Logger.Error(err, "Failed quering vcluster secret")
		return ctrl.Result{}, err
	}

	// Update the CaCertificate of our project
	if scope.Project.Status.CaCertificate != string(secret.Data["certificate-authority"]) {
		scope.Logger.Info("Storing CA certificate for project")
		scope.Project.Status.CaCertificate = string(secret.Data["certificate-authority"])
		scope.Project.Status.Status = "provisioned"
		if err := scope.Client.Status().Update(context.TODO(), scope.Project); err != nil {
			scope.Logger.Error(err, "Failed updating project status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Install any necessary crossplane providers
	// TODO: Support dynamic provisioning. For now, we just install Kubernetes and Helm
	if err := scope.installProviders(ctx); err != nil {
		scope.Logger.Error(err, "Failed creating provider resources")
		return ctrl.Result{}, err
	}

	// TODO: Install any subscribed addons

	statefulSet := &appsv1.StatefulSet{}
	if err := scope.Client.Get(ctx, types.NamespacedName{
		Name:      identifier,
		Namespace: identifier,
	}, statefulSet); err != nil {
		scope.Logger.Error(err, "Failed querying statefulset")
		return ctrl.Result{}, err
	}

	var desiredReplicas int32
	if scope.Project.Spec.Paused == true {
		desiredReplicas = 0
	} else {
		desiredReplicas = 1
	}

	if *statefulSet.Spec.Replicas != desiredReplicas {
		scope.Logger.Info(fmt.Sprintf("Updating statefulset to %d replicas", desiredReplicas))
		statefulSet.Spec.Replicas = &desiredReplicas
		if err := scope.Client.Update(ctx, statefulSet); err != nil {
			scope.Logger.Error(err, "Failed updating desired replicas")
			return ctrl.Result{}, err
		}

	}

	// If paused, we also need to terminate all the running pods
	if scope.Project.Spec.Paused == true {
		pod := &v1.Pod{}
		if err := scope.Client.DeleteAllOf(ctx, pod, []client.DeleteAllOfOption{
			client.InNamespace(identifier),
			client.MatchingLabels{"vcluster.loft.sh/managed-by": identifier},
			client.GracePeriodSeconds(5),
		}...); err != nil {
			scope.Logger.Error(err, "Failed to delete running pods")
			return ctrl.Result{}, err
		}
	}
	// TODO: We shouldn't manually requeue. Instead, we should fix the generation
	// observation to start execution on changes
	return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
}

func getValuesArgs(scope *Scope) ValuesTemplateArgs {
	project := scope.Project
	image := ImageMapping["1.28.3"]
	if scope.Project.Spec.KubernetesVersion != "" {
		if val, ok := ImageMapping[scope.Project.Spec.KubernetesVersion]; ok {
			image = val
		}
	}
	args := ValuesTemplateArgs{
		ProjectId:   project.Spec.Id,
		ProjectSlug: project.Spec.Slug,
		Cpu:         project.Spec.Resources.Cpu,
		Memory:      project.Spec.Resources.Memory,
		Disk:        project.Spec.Resources.Disk,
		Users:       project.Spec.Users,
		Oidc: struct {
			ClientId  string
			IssuerUrl string
		}{
			ClientId:  scope.Cluster.Spec.Oidc.ClientId,
			IssuerUrl: scope.Cluster.Spec.Oidc.IssuerUrl},
		Ingress: struct {
			ClassName string
			Domain    string
		}{
			ClassName: scope.Cluster.Spec.Ingress.ClassName,
			Domain:    scope.Cluster.Spec.Ingress.Domain,
		},
		Image: image,
	}
	return args
}

func (scope *Scope) installProviders(ctx context.Context) error {

	for _, provider := range []schema.GroupVersionResource{
		{Group: "helm.crossplane.io", Version: "v1beta1", Resource: "providerconfigs"},
		{Group: "kubernetes.crossplane.io", Version: "v1alpha1", Resource: "providerconfigs"},
	} {
		scope.Logger.Info("Creating provider " + provider.Group + "/" + provider.Version)
		providerConfig := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": provider.Group + "/" + provider.Version,
				"kind":       "ProviderConfig",
				"metadata": map[string]interface{}{
					"name": scope.Project.Spec.Slug,
				},
				"spec": map[string]interface{}{
					"credentials": map[string]interface{}{
						"source": "Secret",
						"secretRef": map[string]interface{}{
							"namespace": scope.Project.Spec.Slug,
							"name":      "vc-" + scope.Project.Spec.Slug,
							"key":       "config",
						},
					},
				},
			},
		}
		_, err := scope.DynamicClient.Resource(provider).Get(ctx, scope.Project.Spec.Slug, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				_, err := scope.DynamicClient.Resource(provider).Create(ctx, providerConfig, metav1.CreateOptions{})
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}
