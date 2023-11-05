package cluster

import (
	"context"
	"fmt"
	"github.com/launchboxio/operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Scope struct {
	Cluster *v1alpha1.Cluster
	Client  client.Client
}

func (s *Scope) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO: Create the cluster role and cluster role bindings

	// Create the service account
	serviceAccount := &v1.ServiceAccount{}
	if err := s.Client.Get(ctx, req.NamespacedName, serviceAccount); err != nil {
		if apierrors.IsNotFound(err) {
			sa := &v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      req.Name,
					Namespace: req.Namespace,
				},
			}
			ctrl.SetControllerReference(s.Cluster, sa, s.Client.Scheme())
			err = s.Client.Create(ctx, sa)
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, err
	}

	// Create the deployment
	deployment := &appsv1.Deployment{}
	if err := s.Client.Get(ctx, req.NamespacedName, deployment); err != nil {
		if apierrors.IsNotFound(err) {
			return s.createDeployment()
		}
		return ctrl.Result{}, err
	}

	// Update the status for the cluster
	return ctrl.Result{}, nil
}

func (s *Scope) createDeployment(ctx context.Context) (ctrl.Result, error) {
	agentSpec := s.Cluster.Spec.Agent
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Cluster.Name,
			Namespace: s.Cluster.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "agent",
							Image:           fmt.Sprintf("%s:%s", agentSpec.Repository, agentSpec.Tag),
							ImagePullPolicy: agentSpec.PullPolicy,
						},
					},
				},
			},
		},
	}

	err := s.Client.Create(ctx, deployment)
	return ctrl.Result{}, err
}
