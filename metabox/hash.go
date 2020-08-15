package metabox

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
)

func (m *Metabox) hash(filepaths []string) ([]byte, error) {
	target, err := filepath.Abs(m.Config.Target.Local.PrefixPath)
	if err != nil {
		return nil, fmt.Errorf("retrieving absolute path: %v", err)
	}

	// Declare our hasher accumulator.
	var hasher hash.Hash
	switch m.Config.Workspace.Options.Hash {
	case "md5":
		hasher = md5.New()
	case "sha256":
		fallthrough
	default:
		hasher = sha256.New()
	}

	// Do the hash for each file.
	for _, path := range filepaths {
		// Declare relative file path.
		rel, err := filepath.Rel(target, path)
		if err != nil {
			return nil, fmt.Errorf("relpath of %s: %v", path, err)
		}

		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("opening %s: %v", rel, err)
		}

		// Add relative file path to hash accumulator.
		if _, err := hasher.Write([]byte(rel)); err != nil {
			return nil, fmt.Errorf("hashing %s: %v", rel, err)
		}
		// Add file contents to hash accumulator.
		if _, err := io.Copy(hasher, bufio.NewReader(f)); err != nil {
			return nil, fmt.Errorf("hashing %s: %v", rel, err)
		}
	}

	return hasher.Sum(nil), nil
}
