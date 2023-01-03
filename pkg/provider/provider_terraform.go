package provider

import (
	"context"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type TerraformProvider struct {
	// This should be in the /workspaces directory
	WorkingDirectory string
	Version          string
}

func (tp *TerraformProvider) Install() (Output, error) {

	execPath, err := initialize(tp.Version)
	if err != nil {
		return nil, err
	}

	tf, err := tfexec.NewTerraform(tp.WorkingDirectory, execPath)
	if err != nil {
		return nil, err
	}

	if err = tf.Init(context.Background()); err != nil {
		return nil, err
	}

	if err = tf.Apply(context.Background()); err != nil {
		return nil, err
	}

	output, err := tf.Output(context.Background())
	outputs := make(map[string]interface{})
	for key, value := range output {
		outputs[key] = value.Value
	}

	return outputs, nil
}

func (tp *TerraformProvider) Uninstall() error {
	execPath, err := initialize(tp.Version)
	if err != nil {
		return err
	}

	tf, err := tfexec.NewTerraform(tp.WorkingDirectory, execPath)
	if err != nil {
		return err
	}

	if err = tf.Init(context.Background()); err != nil {
		return err
	}

	return tf.Destroy(context.Background())
}

func initialize(tfVersion string) (string, error) {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion(tfVersion)),
	}

	return installer.Install(context.Background())
}
