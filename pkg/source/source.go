package source

type Source interface {
	// Clone the source. Returns the resultant file path, and error
	Clone() error

	// Remove any cloned resources from the operator
	Remove(string) error
}
