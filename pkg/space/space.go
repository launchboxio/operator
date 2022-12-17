package space

import (
	corev1alpha1 "github.com/launchboxio/operator/api/v1alpha1"
	"github.com/launchboxio/operator/pkg/addons"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type InstallOpts struct {
	Namespace string
	Name      string
	Repo      string
	Chart     string
	Version   string
	Client    genericclioptions.RESTClientGetter
}

func Install(opts *InstallOpts) error {
	addon := &corev1alpha1.ClusterAddonSpec{
		HelmRef: corev1alpha1.HelmRef{
			// TODO: Allow configuring to eventually test beta releases
			Repo:      "loft-sh",
			Chart:     "vcluster",
			Namespace: opts.Namespace,
			Name:      opts.Name,
			Version:   opts.Version,
		},
	}
	installer := addons.NewInstaller(opts.Client)
	release, err := installer.Exists(&addon.HelmRef)
	if err != nil {
		return err
	}

	if release != nil {
		return nil
	}

	_, err = installer.Ensure(&addon.HelmRef)
	if err != nil {
		return err
	}
	return nil
}
