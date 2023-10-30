package events

import (
	"context"
	crossplanepkgv1 "github.com/crossplane/crossplane/apis/pkg/v1"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AddonEventPayload struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	OciRegistry      string `json:"oci_registry"`
	OciVersion       string `json:"oci_version"`
	PullPolicy       string `json:"pull_policy"`
	ActivationPolicy string `json:"activation_policy"`
}

type AddonHandler struct {
	Logger logr.Logger
	Client client.Client
}

func (ah *AddonHandler) syncAddonResource(payload map[string]interface{}) error {
	addon := &crossplanepkgv1.Configuration{}
	resource := addonFromPayload(payload)

	if err := ah.Client.Get(context.TODO(), client.ObjectKey{
		Name: payload["name"].(string),
	}, addon); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return ah.Client.Create(context.TODO(), resource)
	}
	return ah.Client.Update(context.TODO(), resource)
}

func (ah *AddonHandler) Create(event *ActionCableEvent) error {
	data, err := event.Message.GetPayload()
	if err != nil {
		return err
	}
	return ah.syncAddonResource(data)
}

func (ah *AddonHandler) Update(event *ActionCableEvent) error {
	data, err := event.Message.GetPayload()
	if err != nil {
		return err
	}
	return ah.syncAddonResource(data)
}

func (ah *AddonHandler) Delete(event *ActionCableEvent) error {
	data, err := event.Message.GetPayload()
	if err != nil {
		return err
	}
	addon := &crossplanepkgv1.Configuration{}
	if err := ah.Client.Get(context.TODO(), client.ObjectKey{
		Name: data["name"].(string),
	}, addon); err != nil {
		return err
	}

	return ah.Client.Delete(context.TODO(), addon)
}

func addonFromPayload(event map[string]interface{}) *crossplanepkgv1.Configuration {
	pullPolicy := v1.PullAlways
	activationPolicy := crossplanepkgv1.AutomaticActivation
	revisionLimit := int64(5)
	return &crossplanepkgv1.Configuration{
		ObjectMeta: metav1.ObjectMeta{
			Name: event["name"].(string),
		},
		Spec: crossplanepkgv1.ConfigurationSpec{
			PackageSpec: crossplanepkgv1.PackageSpec{
				Package:                  event["oci_registry"].(string) + ":" + event["oci_version"].(string),
				PackagePullPolicy:        &pullPolicy,
				RevisionActivationPolicy: &activationPolicy,
				RevisionHistoryLimit:     &revisionLimit,
			},
		},
	}
}
