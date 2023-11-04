package events

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/launchboxio/operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProjectEventPayload struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Users []struct {
		Email       string `json:"email"`
		ClusterRole string `json:"clusterRole"`
	} `json:"users"`
	Cpu    int `json:"cpu"`
	Memory int `json:"memory"`
	Disk   int `json:"disk"`
}

type ProjectHandler struct {
	Logger logr.Logger
	Client client.Client
}

func (ph *ProjectHandler) syncProjectResource(event *LaunchboxEvent) error {
	project := &v1alpha1.Project{}

	resource, err := projectFromPayload(event)
	if err != nil {
		return err
	}

	if err := ph.Client.Get(context.TODO(), types.NamespacedName{
		Name:      resource.ObjectMeta.Name,
		Namespace: resource.ObjectMeta.Namespace,
	}, project); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		ph.Logger.Info("Creating new project resource")
		return ph.Client.Create(context.TODO(), resource)
	}
	ph.Logger.Info("Updating existing project resource")
	project.Spec = resource.Spec
	return ph.Client.Update(context.TODO(), project)
}

func (ph *ProjectHandler) Create(event *LaunchboxEvent) error {
	return ph.syncProjectResource(event)
}

func (ph *ProjectHandler) Update(event *LaunchboxEvent) error {
	return ph.syncProjectResource(event)
}

func (ph *ProjectHandler) Delete(event *LaunchboxEvent) error {
	resource, err := projectFromPayload(event)
	if err != nil {
		return err
	}
	project := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name:      resource.ObjectMeta.Name,
		Namespace: resource.ObjectMeta.Namespace,
	}, project); err != nil {
		return err
	}

	return ph.Client.Delete(context.TODO(), project)
}

func (ph *ProjectHandler) Pause(event *LaunchboxEvent) error {
	resource, err := projectFromPayload(event)
	if err != nil {
		return err
	}
	project := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name:      resource.ObjectMeta.Name,
		Namespace: resource.ObjectMeta.Namespace,
	}, project); err != nil {
		return err
	}
	project.Spec.Paused = true

	return ph.Client.Update(context.TODO(), project)
}

func (ph *ProjectHandler) Resume(event *LaunchboxEvent) error {
	resource, err := projectFromPayload(event)
	if err != nil {
		return err
	}
	project := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name:      resource.ObjectMeta.Name,
		Namespace: resource.ObjectMeta.Namespace,
	}, project); err != nil {
		return err
	}
	project.Spec.Paused = false
	return ph.Client.Update(context.TODO(), project)
}

func projectFromPayload(event *LaunchboxEvent) (*v1alpha1.Project, error) {
	var users []v1alpha1.ProjectUser
	if _, ok := event.Data["users"]; ok {
		for _, user := range event.Data["users"].([]struct {
			Email       string
			ClusterRole string
		}) {
			users = append(users, v1alpha1.ProjectUser{
				Email:       user.Email,
				ClusterRole: user.ClusterRole,
			})
		}
	}

	project := &v1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name:      event.Data["slug"].(string),
			Namespace: "lbx-system",
		},
		Spec: v1alpha1.ProjectSpec{
			Slug: event.Data["slug"].(string),
			Id:   int(event.Data["id"].(float64)),
			// TODO: Pull this from the event payload
			KubernetesVersion: "1.25.15",
			Crossplane: v1alpha1.ProjectCrossplaneSpec{
				Providers: []string{},
			},
			Resources: v1alpha1.Resources{
				Cpu:    int32(event.Data["cpu"].(float64)),
				Memory: int32(event.Data["memory"].(float64)),
				Disk:   int32(event.Data["disk"].(float64)),
			},
			Users: users,
		},
	}
	return project, nil
}
