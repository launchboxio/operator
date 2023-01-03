package provider

import (
	"errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type RawProvider struct {
	Manifests []string
}

func (rp *RawProvider) Install() (Output, error) {
	return nil, errors.New("Raw Provider not implemented")
}

func (rp *RawProvider) Uninstall() error {
	return nil
}

func (rp *RawProvider) applyRawManifest(u *unstructured.Unstructured, namespaceOverride string) {

}
