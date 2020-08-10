package metabox

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v2"
	"github.com/nmcapule/metabox-go/config"
)

// Metabox implements a configurable backup / restore tool.
type Metabox struct {
	config config.Config
	logger log.Logger
}

// New creates a new Metabox instance.
func New(cfg config.Config) (*Metabox, error) {
	return &Metabox{
		config: cfg,
	}, nil
}

// FromConfigFile creates a new Metabox instance from a config file path.
func FromConfigFile(filename string) (*Metabox, error) {
	cfg, err := config.FromFile(filename)
	if err != nil {
		return nil, err
	}
	return New(*cfg)
}

// StartBackup executes the backup workflow of Metabox.
func (m *Metabox) StartBackup() error {
	// 1. run pre-backup hook
	if err := m.exec("pre-backup", m.config.Workspace.Hooks.PreBackup); err != nil {
		return err
	}

	// 2. walk through included files and hash and targz.
	filepaths, err := m.walk()
	if err != nil {
		return err
	}

	hashsum, err := m.hash(filepaths)
	if err != nil {
		return err
	}

	if err := m.compress(filepaths, fmt.Sprintf("%x", hashsum)); err != nil {
		return err
	}

	// 3. record to versioning file
	// TODO(nmcapule): Implement me.

	// 4. run post-backup hook
	if err := m.exec("post-backup", m.config.Workspace.Hooks.PostBackup); err != nil {
		return err
	}
	return nil
}

// exec executes a shell command.
func (m *Metabox) exec(step string, lines []string) error {
	for _, line := range lines {
		_, err := exec.Command("sh", "-c", line).Output()
		if err != nil {
			return fmt.Errorf("failed exec: %q: %v", line, err)
		}
		log.Printf("%s: %s", step, line)
	}
	return nil
}

func matches(root, matcher, path string) (bool, error) {
	if matcher[len(matcher)-1] == '/' {
		matcher += "**/*"
	}
	pattern := filepath.Join(root, matcher)
	return doublestar.PathMatch(pattern, path)
}

func (m *Metabox) walk() ([]string, error) {
	target, err := filepath.Abs(m.config.Target.Local.PrefixPath)
	if err != nil {
		return nil, fmt.Errorf("retrieving absolute path: %v", err)
	}

	var filepaths []string
	fn := func(path string, info os.FileInfo, err error) error {
		// Skip directories.
		if info.IsDir() {
			return nil
		}

		// Skip symbolic links.
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		// If includes is specified, filter out non-matching paths.
		if len(m.config.Target.Local.Includes) > 0 {
			var include bool
			for _, matcher := range m.config.Target.Local.Includes {
				if ok, err := matches(target, matcher, path); ok {
					include = true
					break
				} else if err != nil {
					return err
				}

			}
			if !include {
				return nil
			}
		}

		// If excludes is specified, filter out matching paths.
		if len(m.config.Target.Local.Excludes) > 0 {
			for _, matcher := range m.config.Target.Local.Excludes {
				ok, err := matches(target, matcher, path)
				if err != nil || ok {
					return err
				}
			}
		}

		// Append to list of filepaths.
		filepaths = append(filepaths, path)

		return nil
	}

	if err := filepath.Walk(target, fn); err != nil {
		return nil, fmt.Errorf("file walk: %v", err)
	}

	return filepaths, nil
}

func (m *Metabox) hash(filepaths []string) ([]byte, error) {
	target, err := filepath.Abs(m.config.Target.Local.PrefixPath)
	if err != nil {
		return nil, fmt.Errorf("retrieving absolute path: %v", err)
	}

	// Declare our hasher accumulator.
	var hasher hash.Hash
	switch m.config.Workspace.Options.Hash {
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

func (m *Metabox) compress(filepaths []string, name string) error {
	target, err := filepath.Abs(m.config.Target.Local.PrefixPath)
	if err != nil {
		return fmt.Errorf("retrieving absolute path: %v", err)
	}

	cachepath := filepath.Join(m.config.Workspace.RootPath, m.config.Workspace.CachePath)

	// Make sure cachepath exists.
	if _, err := os.Stat(cachepath); os.IsNotExist(err) {
		err := os.Mkdir(cachepath, os.FileMode(0777))
		if err != nil {
			return fmt.Errorf("creating %s: %v", cachepath, err)
		}
	}

	// Create target output file.
	outpath := filepath.FromSlash(fmt.Sprintf("%s/%s.tar.gz", cachepath, name))
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
