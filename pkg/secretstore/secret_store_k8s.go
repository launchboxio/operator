package secretstore

import (
	"errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Store a secret containing addon configuration in Kubernetes

type KubernetesSecretStore struct {
	Client client.Client
}

func (kss *KubernetesSecretStore) Store(name string, key string, value string) error {
	return errors.New("Kubernetes secret store not implemented")
}

func (kss *KubernetesSecretStore) Remove(key string) error {
	// If the secret exists, and we only have the specified key, remove the secret
	// Otherwise, patch data to remove the single key
	return nil
}
