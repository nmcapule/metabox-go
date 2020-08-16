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

func (s *S3) Exists(key string) (bool, error) {
	return false, errUnimplemented
}

func (s *S3) Upload(key string, source io.Reader) error {
	return errUnimplemented
}

func (s *S3) Download(key string, destination io.Writer) error {
	return errUnimplemented
}
