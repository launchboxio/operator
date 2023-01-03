package secretstore

import (
	"context"
	"github.com/hashicorp/vault/api"
	"k8s.io/apimachinery/pkg/api/errors"
)

type VaultSecretStore struct {
	Client    api.Client
	MountPath string
}

// Store the requested key value pair in Vault
func (vss *VaultSecretStore) Store(name string, key string, value string) error {
	secret, err := vss.Client.KVv2(vss.MountPath).Get(context.Background(), name)
	if err != nil && errors.IsNotFound(err) {
		_, err = vss.Client.KVv2(vss.MountPath).Put(context.Background(), name, map[string]interface{}{key: value})
		// We need to create the secret
		return err
	} else if err != nil {
		return err
	}

	secret.Data[key] = value
	_, err = vss.Client.KVv2(vss.MountPath).Put(context.Background(), name, secret.Data)
	return err
}
