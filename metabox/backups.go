package metabox

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func (m *Metabox) uploadToBackups(sum string) error {
	filename := m.compressedFilename(sum)
	filepath := filepath.Join(m.derivedCachePath(), filename)

	for _, store := range m.Stores {
		file, err := os.Open(filepath)
		if err != nil {
			return fmt.Errorf("open %q: %v", filepath, err)
		}
		defer file.Close()

		if err := store.Upload(filename, file); err != nil {
			return fmt.Errorf("upload: %v", err)
		}
		log.Printf("upload: %s", filepath)
	}

	return nil
}

func (m *Metabox) downloadFromBackups(sum string) error {
	filename := m.compressedFilename(sum)
	filepath := filepath.Join(m.derivedCachePath(), filename)
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("open %q: %v", filepath, err)
	}
	defer file.Close()

	for _, store := range m.Stores {
		if err := store.Download(filename, file); err != nil {
			return fmt.Errorf("download: %v", err)
		}
		log.Printf("download: %s", filepath)
		return nil
	}
	return errNoAvailableStores
}
