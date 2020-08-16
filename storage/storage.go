package storage

import "io"

// Storage defines a simplified storage interface.
type Storage interface {
	// Exists checks if key name exists on the storage.
	Exists(name string) error
	// Upload writes the source to the storage with key name.
	Upload(name string, source io.Reader) error
	// Download reads the item with key name and writes to destination.
	Download(name string, destination io.Writer) error
}
