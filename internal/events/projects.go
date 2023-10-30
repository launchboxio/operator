package events

import (
	"context"
	"encoding/json"
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

func (ph *ProjectHandler) syncProjectResource(data []byte) error {
	project := &v1alpha1.Project{}

	resource, err := projectFromPayload(data)
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

func (ph *ProjectHandler) Create(event *ActionCableEvent) error {
	return ph.syncProjectResource(event.Message.Payload)
}

func (ph *ProjectHandler) Update(event *ActionCableEvent) error {
	return ph.syncProjectResource(event.Message.Payload)
}

func (ph *ProjectHandler) Delete(event *ActionCableEvent) error {
	resource, err := projectFromPayload(event.Message.Payload)
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

func (ph *ProjectHandler) Pause(event *ActionCableEvent) error {
	resource, err := projectFromPayload(event.Message.Payload)
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

func (ph *ProjectHandler) Resume(event *ActionCableEvent) error {
	resource, err := projectFromPayload(event.Message.Payload)
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

func projectFromPayload(data []byte) (*v1alpha1.Project, error) {
	input := &ProjectEventPayload{}
	err := json.Unmarshal(data, input)
	if err != nil {
		return nil, err
	}
	var users []v1alpha1.ProjectUser
	for _, user := range input.Users {
		users = append(users, v1alpha1.ProjectUser{
			Email:       user.Email,
			ClusterRole: user.ClusterRole,
		})
	}
	project := &v1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Slug,
			Namespace: "lbx-system",
		},
		Spec: v1alpha1.ProjectSpec{
			Slug: input.Slug,
			Id:   input.Id,
			// TODO: Pull this from the event payload
			KubernetesVersion: "1.25.15",
			Crossplane: v1alpha1.ProjectCrossplaneSpec{
				Providers: []string{},
			},
			Resources: v1alpha1.Resources{
				Cpu:    int32(input.Cpu),
				Memory: int32(input.Memory),
				Disk:   int32(input.Disk),
			},
			Users: users,
		},
	}
	return project, nil
}
