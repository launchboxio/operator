package events

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	action_cable "github.com/launchboxio/action-cable"
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

func (ph *ProjectHandler) Create(event *action_cable.ActionCableEvent) {
	if err := ph.syncProjectResource(event.Message); err != nil {
		ph.Logger.Error(err, "projects.created failed")
	}
}

func (ph *ProjectHandler) Update(event *action_cable.ActionCableEvent) {
	if err := ph.syncProjectResource(event.Message); err != nil {
		ph.Logger.Error(err, "projects.updated failed")
	}
}

func (ph *ProjectHandler) Delete(event *action_cable.ActionCableEvent) {
	resource, err := projectFromPayload(event.Message)
	if err != nil {
		ph.Logger.Error(err, "projects.deleted failed")
		return
	}
	project := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name:      resource.ObjectMeta.Name,
		Namespace: resource.ObjectMeta.Namespace,
	}, project); err != nil {
		ph.Logger.Error(err, "projects.deleted failed")
	}

	if err := ph.Client.Delete(context.TODO(), project); err != nil {
		ph.Logger.Error(err, "projects.created failed")
	}
}

func (ph *ProjectHandler) Pause(event *action_cable.ActionCableEvent) {
	resource, err := projectFromPayload(event.Message)
	if err != nil {
		ph.Logger.Error(err, "projects.pause failed")
		return
	}
	project := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name:      resource.ObjectMeta.Name,
		Namespace: resource.ObjectMeta.Namespace,
	}, project); err != nil {
		ph.Logger.Error(err, "projects.pause failed")
		return
	}
	project.Spec.Paused = true
	if err := ph.Client.Update(context.TODO(), project); err != nil {
		ph.Logger.Error(err, "projects.pause failed")
		return
	}
}

func (ph *ProjectHandler) Resume(event *action_cable.ActionCableEvent) {
	resource, err := projectFromPayload(event.Message)
	if err != nil {
		ph.Logger.Error(err, "projects.resume failed")
		return
	}
	project := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name:      resource.ObjectMeta.Name,
		Namespace: resource.ObjectMeta.Namespace,
	}, project); err != nil {
		ph.Logger.Error(err, "projects.resume failed")
		return
	}
	project.Spec.Paused = false
	if err := ph.Client.Update(context.TODO(), project); err != nil {
		ph.Logger.Error(err, "projects.resume failed")
		return
	}
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
