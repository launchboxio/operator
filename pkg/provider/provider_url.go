package provider

import "errors"

type UrlProvider struct {
}

func (up *UrlProvider) Install() (Output, error) {
	return nil, errors.New("URL Provider not implemented")
}

func (up *UrlProvider) Uninstall() error {
	return nil
}
