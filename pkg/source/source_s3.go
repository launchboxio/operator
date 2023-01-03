package source

import (
	"errors"
	"os"
)

type S3Source struct {
	Bucket string
	Prefix string
	Path   string
}

// Clone syncs a bucket, with an optional prefix
func (ss *S3Source) Clone() (string, error) {
	return "", errors.New("S3 Source not implemented")
}

func (ss *S3Source) Remove(path string) error {
	return os.RemoveAll(ss.Path)
}
