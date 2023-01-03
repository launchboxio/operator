package provider

import "errors"

type KustomizeProvider struct {
}

func (kp *KustomizeProvider) Install() (Output, error) {

	return nil, errors.New("Kustomize Provider not implemented")
}

func (kp *KustomizeProvider) Uninstall() error {
	return nil
}
