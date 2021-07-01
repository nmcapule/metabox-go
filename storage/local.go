package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nmcapule/metabox-go/config"
)

// Local implements a local storage.
type Local struct {
	config *config.LocalStorageConfig
}

// NewLocal creates a Local storage from config.
func NewLocal(config *config.LocalStorageConfig) (*Local, error) {
	return &Local{
		config: config,
	}, nil
}

func (s *Local) Exists(key string) (bool, error) {
	path := filepath.Join(s.config.Path, key)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("stat %q: %v", path, err)
	}
	return true, nil
}

func (s *Local) Upload(key string, source io.Reader) error {
	path := filepath.Join(s.config.Path, key)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("open %q: %v", path, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, source); err != nil {
		return fmt.Errorf("upload to %q: %v", path, err)
	}
	return nil
}

func (s *Local) Download(key string, destination WriterWriterAt) error {
	path := filepath.Join(s.config.Path, key)
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %q: %v", path, err)
	}
	defer file.Close()

	if _, err := io.Copy(destination, file); err != nil {
		return fmt.Errorf("download from %q: %v", path, err)
	}
	return nil
}
