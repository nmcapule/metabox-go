package metabox

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/nmcapule/metabox-go/config"
	"github.com/nmcapule/metabox-go/tracker"
)

// Metabox implements a configurable backup / restore tool.
type Metabox struct {
	Config *config.Config
	DB     *tracker.SimpleFileDB
	logger log.Logger
}

// New creates a new Metabox instance.
func New(cfg *config.Config) (*Metabox, error) {
	box := &Metabox{Config: cfg}

	db, err := tracker.NewSimpleFileDB(box.derivedVersionsPath())
	if err != nil {
		return nil, err
	}

	box.DB = db

	return box, nil
}

// FromConfigFile creates a new Metabox instance from a config file path.
func FromConfigFile(filename string) (*Metabox, error) {
	cfg, err := config.FromFile(filename)
	if err != nil {
		return nil, err
	}
	return New(cfg)
}

// StartBackup executes the backup workflow of Metabox.
func (m *Metabox) StartBackup() (*tracker.Item, error) {
	// Make sure rootpath exists.
	if err := ensurePathExists(m.Config.Workspace.RootPath); err != nil {
		return nil, err
	}

	// 1. run pre-backup hook
	if err := m.exec("pre-backup", m.Config.Workspace.Hooks.PreBackup); err != nil {
		return nil, err
	}

	// 2. walk through included files and hash and targz.
	filepaths, err := m.walk()
	if err != nil {
		return nil, err
	}

	b, err := m.hash(filepaths)
	if err != nil {
		return nil, err
	}
	sum := fmt.Sprintf("%x", b)

	// interim. check if already exists in versioning before compress and upload.
	// TODO(nmcapule): Implement me.

	if err := m.compress(filepaths, sum); err != nil {
		return nil, err
	}

	// 3. upload to backups
	// TODO(nmcapule): Implement me.

	// 4. record to versioning file
	item := &tracker.Item{
		ID:      sum,
		Created: tracker.Time(time.Now()),
		Author:  m.Config.Workspace.UserIdentifier,
		Tags:    m.Config.Workspace.TagsGenerator,
	}

	m.DB.Put(sum, item)
	if err := m.DB.Flush(); err != nil {
		return nil, err
	}

	// 5. run post-backup hook
	if err := m.exec("post-backup", m.Config.Workspace.Hooks.PostBackup); err != nil {
		return nil, err
	}

	return item, nil
}

func (m *Metabox) StartRestore(item *tracker.Item) error {
	// 1. run pre-restore hook
	if err := m.exec("pre-restore", m.Config.Workspace.Hooks.PreRestore); err != nil {
		return err
	}

	// 2. download from backups if does not exist in cache
	// TODO(nmcapule): Implement me.

	// 3. extract and copy to target path
	if err := m.extract(item.ID); err != nil {
		return err
	}

	// 4. run post-restore hook
	if err := m.exec("post-restore", m.Config.Workspace.Hooks.PostRestore); err != nil {
		return err
	}

	return nil
}

func (m *Metabox) derivedVersionsPath() string {
	return filepath.Join(m.Config.Workspace.RootPath, m.Config.Workspace.VersionsPath)
}

func (m *Metabox) derivedCachePath() string {
	return filepath.Join(m.Config.Workspace.RootPath, m.Config.Workspace.CachePath)
}

func (m *Metabox) derivedTargetPath() string {
	return filepath.Join(m.Config.Workspace.RootPath, m.Config.Target.Local.PrefixPath)
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

func (m *Metabox) walk() ([]string, error) {
	target, err := filepath.Abs(m.derivedTargetPath())
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
		if len(m.Config.Target.Local.Includes) > 0 {
			var include bool
			for _, matcher := range m.Config.Target.Local.Includes {
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
		if len(m.Config.Target.Local.Excludes) > 0 {
			for _, matcher := range m.Config.Target.Local.Excludes {
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
