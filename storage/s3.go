package storage

import (
	"io"

	"github.com/nmcapule/metabox-go/config"
)

type S3 struct {
	config *config.S3StorageConfig
}

func NewS3(config *config.S3StorageConfig) (*S3, error) {
	return nil, errUnimplemented
}

func (s *S3) Exists(name string) (bool, error) {
	return false, errUnimplemented
}

func (s *S3) Upload(name string, source io.Reader) error {
	return errUnimplemented
}

func (s *S3) Download(name string, destination io.Writer) error {
	return errUnimplemented
}
