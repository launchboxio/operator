package provider

import "errors"

type ShellProvider struct {
}

func (sp *ShellProvider) Install() (Output, error) {
	return nil, errors.New("Shell Provider not implemented")
}

func (sp *ShellProvider) Uninstall() error {
	return nil
}
