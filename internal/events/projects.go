package events

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/launchboxio/operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProjectEventPayload struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProjectHandler struct {
	Logger logr.Logger
	Client client.Client
}

func (ph *ProjectHandler) syncProjectResource(payload map[string]interface{}) error {
	project := &v1alpha1.Project{}
	resource := projectFromPayload(payload)

	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name: payload["slug"].(string),
	}, project); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return ph.Client.Create(context.TODO(), resource)
	}
	return ph.Client.Update(context.TODO(), resource)
}

func (ph *ProjectHandler) Create(event *ActionCableEvent) error {
	data, err := event.Message.GetPayload()
	if err != nil {
		return err
	}
	return ph.syncProjectResource(data)
}

func (ph *ProjectHandler) Update(event *ActionCableEvent) error {
	data, err := event.Message.GetPayload()
	if err != nil {
		return err
	}
	return ph.syncProjectResource(data)
}

func (ph *ProjectHandler) Delete(event *ActionCableEvent) error {
	data, err := event.Message.GetPayload()
	if err != nil {
		return err
	}
	project := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name: data["slug"].(string),
	}, project); err != nil {
		return err
	}

	return ph.Client.Delete(context.TODO(), project)
}

func (ph *ProjectHandler) Pause(event *ActionCableEvent) error {
	data, err := event.Message.GetPayload()
	if err != nil {
		return err
	}
	project := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name: data["slug"].(string),
	}, project); err != nil {
		return err
	}
	project.Spec.Paused = true
	return ph.Client.Update(context.TODO(), project)
}

func (ph *ProjectHandler) Resume(event *ActionCableEvent) error {
	data, err := event.Message.GetPayload()
	if err != nil {
		return err
	}

	project := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name: data["slug"].(string),
	}, project); err != nil {
		return err
	}
	project.Spec.Paused = false
	return ph.Client.Update(context.TODO(), project)
}

func projectFromPayload(data map[string]interface{}) *v1alpha1.Project {
	project := &v1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: data["slug"].(string),
		},
		Spec: v1alpha1.ProjectSpec{
			Slug: data["slug"].(string),
			Id:   int(data["id"].(float64)),
			// TODO: Pull this from the event payload
			KubernetesVersion: "1.25.15",
		},
	}

	fmt.Println(project)

	return project
}
