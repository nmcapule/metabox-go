package storage

import (
	"io"

	"github.com/nmcapule/metabox-go/config"
)

type Local struct {
	config *config.LocalStorageConfig
}

func NewLocal(config *config.LocalStorageConfig) (*Local, error) {
	return &Local{
		config: config,
	}, nil
}

func (s *Local) Exists(name string) (bool, error) {
	return false, errUnimplemented
}

func (s *Local) Upload(name string, source io.Reader) error {
	return errUnimplemented
}

func (s *Local) Download(name string, destination io.Writer) error {
	return errUnimplemented
}
