package storage

import (
	"io"

	"github.com/nmcapule/metabox-go/config"
)

type Remote struct {
	config *config.RemoteStorageConfig
}

func NewRemote(config *config.RemoteStorageConfig) (*Remote, error) {
	return nil, errUnimplemented
}

func (s *Remote) Exists(name string) (bool, error) {
	return false, errUnimplemented
}

func (s *Remote) Upload(name string, source io.Reader) error {
	return errUnimplemented
}

func (s *Remote) Download(name string, destination io.Writer) error {
	return errUnimplemented
}
