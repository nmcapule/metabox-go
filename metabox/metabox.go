package metabox

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/nmcapule/metabox-go/config"
	"github.com/nmcapule/metabox-go/storage"
	"github.com/nmcapule/metabox-go/tracker"
)

// Metabox implements a configurable backup / restore tool.
type Metabox struct {
	Config *config.Config
	DB     *tracker.SimpleFileDB
	Stores []storage.Storage
	logger log.Logger
}

// New creates a new Metabox instance.
func New(cfg *config.Config) (*Metabox, error) {
	box := &Metabox{Config: cfg}

	// Instantiate tracker db.
	db, err := tracker.NewSimpleFileDB(box.derivedVersionsPath())
	if err != nil {
		return nil, err
	}
	box.DB = db

	// Instantiate storages.
	var stores []storage.Storage
	for i := range cfg.Backups {
		var store storage.Storage
		var err error
		switch cfg.Backups[i].Driver {
		case storage.LocalDriver:
			store, err = storage.NewLocal(&cfg.Backups[i].Local)
		case storage.RemoteDriver:
			store, err = storage.NewRemote(&cfg.Backups[i].Remote)
		case storage.S3Driver:
			store, err = storage.NewS3(&cfg.Backups[i].S3)
		default:
			return nil, fmt.Errorf("unknown storage driver: %q", cfg.Backups[i].Driver)
		}
		if err != nil {
			return nil, err
		}
		stores = append(stores, store)
	}
	box.Stores = stores

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
	// Make sure cachepath exists.
	if err := ensurePathExists(m.derivedCachePath()); err != nil {
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
	var item *tracker.Item
	if m.DB.Exists(sum) {
		if item, err = m.DB.Get(sum); err != nil {
			return nil, err
		}
	} else {
		if err := m.compress(filepaths, sum); err != nil {
			return nil, err
		}

		// 3. upload to backups
		if err := m.uploadToBackups(sum); err != nil {
			return nil, err
		}

		// 4. record to versioning file
		item = &tracker.Item{
			ID:      sum,
			Created: tracker.Time(time.Now()),
			Author:  m.Config.Workspace.UserIdentifier,
			Tags:    m.Config.Workspace.TagsGenerator,
		}

		m.DB.Put(sum, item)
		if err := m.DB.Flush(); err != nil {
			return nil, err
		}
	}

	// 5. run post-backup hook
	if err := m.exec("post-backup", m.Config.Workspace.Hooks.PostBackup); err != nil {
		return nil, err
	}

	return item, nil
}

func (m *Metabox) StartRestore(item *tracker.Item) error {
	// Make sure cachepath and targetpath exists.
	if err := ensurePathExists(m.derivedCachePath()); err != nil {
		return err
	}
	if err := ensurePathExists(m.derivedTargetPath()); err != nil {
		return err
	}

	// 1. run pre-restore hook
	if err := m.exec("pre-restore", m.Config.Workspace.Hooks.PreRestore); err != nil {
		return err
	}

	// 2. download from backups if does not exist in cache
	cache := filepath.FromSlash(filepath.Join(m.derivedCachePath(), m.compressedFilename(item.ID)))
	if _, err := os.Stat(cache); os.IsNotExist(err) {
		if err := m.downloadFromBackups(item.ID); err != nil {
			return err
		}
	}

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
	return filepath.Join(m.Config.Workspace.RootPath, m.Config.Target.PrefixPath)
}

// exec executes a shell command with working directory set to the path of the input yaml file.
func (m *Metabox) exec(step string, lines []string) error {
	for _, line := range lines {
		cmd := exec.Command("sh", "-c", line)
		cmd.Dir = m.Config.Workspace.RootPath

		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed exec: %q: %v", line, err)
		}
		log.Printf("%s (exec) > %s", step, line)
		log.Printf("%s (out): %s", step, string(out))
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
		if len(m.Config.Target.Includes) > 0 {
			var include bool
			for _, matcher := range m.Config.Target.Includes {
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
		if len(m.Config.Target.Excludes) > 0 {
			for _, matcher := range m.Config.Target.Excludes {
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
