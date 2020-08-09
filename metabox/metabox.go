package metabox

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v2"
	"github.com/nmcapule/metabox-go/config"
)

type Metabox struct {
	config config.Config
	logger log.Logger
}

func New(cfg config.Config) (*Metabox, error) {
	return &Metabox{
		config: cfg,
	}, nil
}

func FromConfigFile(filename string) (*Metabox, error) {
	cfg, err := config.FromFile(filename)
	if err != nil {
		return nil, err
	}
	return New(*cfg)
}

func (m *Metabox) Start() error {
	// 1. run pre-backup hook
	if err := m.exec("pre-backup", m.config.Workspace.Hooks.PreBackup); err != nil {
		return err
	}
	// 2. walk through included files and hash and targz
	if err := m.walk(); err != nil {
		return err
	}
	// 3. record to versioning file
	// 4. run post-backup hook
	if err := m.exec("post-backup", m.config.Workspace.Hooks.PostBackup); err != nil {
		return err
	}
	return nil
}

func (m *Metabox) exec(step string, lines []string) error {
	for _, line := range lines {
		_, err := exec.Command("bash", "-c", line).Output()
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

func (m *Metabox) walk() error {
	root, err := filepath.Abs(m.config.Workspace.RootPath)
	if err != nil {
		return err
	}

	fn := func(path string, info os.FileInfo, err error) error {
		// Skip directories.
		if info.IsDir() {
			return nil
		}

		// Skip symbolic links.
		if info.Mode() == os.ModeSymlink {
			return nil
		}

		// If includes is specified, filter out non-matching paths.
		if len(m.config.Target.Local.Includes) > 0 {
			var include bool
			for _, matcher := range m.config.Target.Local.Includes {
				if ok, err := matches(root, matcher, path); ok {
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
				ok, err := matches(root, matcher, path)
				if err != nil || ok {
					return err
				}
			}
		}

		log.Printf("include file: %s", path)
		return nil
	}

	return filepath.Walk(root, fn)
}
