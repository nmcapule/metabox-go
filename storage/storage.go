package storage

import "io"

// Storage defines a simplified storage interface.
type Storage interface {
	// Exists checks if key name exists on the storage.
	Exists(key string) (bool, error)
	// Upload writes the source to the storage with key name.
	Upload(key string, source io.Reader) error
	// Download reads the item with key name and writes to destination.
	Download(key string, destination io.Writer) error
}

// Storage driver names.
const (
	LocalDriver  = "local"
	RemoteDriver = "remote"
	S3Driver     = "s3"
)
