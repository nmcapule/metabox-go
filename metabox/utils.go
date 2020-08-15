package metabox

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar"
)

func matches(root, matcher, path string) (bool, error) {
	if matcher[len(matcher)-1] == '/' {
		matcher += "**/*"
	}
	pattern := filepath.Join(root, matcher)
	return doublestar.PathMatch(pattern, path)
}

func ensurePathExists(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.Mkdir(path, os.FileMode(0777)); err != nil {
			return fmt.Errorf("creating %q: %v", path, err)
		}
		return nil
	}
	return err
}
