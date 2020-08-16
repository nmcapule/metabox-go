package metabox

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (m *Metabox) extract(name string) error {
	target, err := filepath.Abs(m.derivedTargetPath())
	if err != nil {
		return fmt.Errorf("retrieving absolute path: %v", err)
	}

	// Make sure cachepath exists.
	cache := filepath.FromSlash(filepath.Join(m.derivedCachePath(), m.compressedFilename(name)))
	cachefile, err := os.OpenFile(cache, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening cache file: %v", err)
	}
	defer cachefile.Close()

	gzr, err := gzip.NewReader(cachefile)
	if err != nil {
		return fmt.Errorf("creating gzip reader from %q: %v", cache, err)
	}

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("extract tar: %v", err)
		}

		// Calculate extract path of the new file or directory.
		path := filepath.Join(target, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(path, 0755); err != nil {
				return fmt.Errorf("mkdir %q: %v", path, err)
			}
		case tar.TypeReg:
			out, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("create %q: %v", path, err)
			}
			if _, err := io.Copy(out, tr); err != nil {
				return fmt.Errorf("copy %q: %v", path, err)
			}
			out.Close()
		default:
			return fmt.Errorf("unknown type %q (%q)", hdr.Typeflag, hdr.Name)
		}
	}

	return nil
}
