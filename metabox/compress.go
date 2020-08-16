package metabox

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func (m *Metabox) compress(filepaths []string, name string) error {
	target, err := filepath.Abs(m.derivedTargetPath())
	if err != nil {
		return fmt.Errorf("retrieving absolute path: %v", err)
	}

	// Make sure cachepath exists.
	cachepath := m.derivedCachePath()
	if err := ensurePathExists(cachepath); err != nil {
		return err
	}

	// Create target output file.
	outpath := filepath.FromSlash(filepath.Join(cachepath, m.compressedFilename(name)))
	file, err := os.OpenFile(outpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("creating tmp file: %v", err)
	}
	defer file.Close()

	// Declare our gzip and tar writer.
	gzw, err := gzip.NewWriterLevel(file, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("creating gzip writer: %v", err)
	}
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Compress each file!
	for _, path := range filepaths {
		body, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %v", path, err)
		}

		// Early exit on nil body.
		if body == nil {
			continue
		}

		// Declare relative file path.
		rel, err := filepath.Rel(target, path)
		if err != nil {
			return fmt.Errorf("relpath of %s: %v", path, err)
		}

		log.Printf("compress %s", rel)

		// Do the write to tar.gz!
		hdr := &tar.Header{
			Name: rel,
			Mode: int64(0644),
			Size: int64(len(body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("writing headers: %v", err)
		}
		if _, err := tw.Write(body); err != nil {
			return fmt.Errorf("writing body: %v", err)
		}
	}

	return nil
}
