package secretstore

type SecretStore interface {
	Store(key string, value string) error
	Remove(key string) error
}
