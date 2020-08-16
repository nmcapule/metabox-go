package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/creasty/defaults"
	"github.com/go-yaml/yaml"
)

type WorkspaceConfig struct {
	RootPath       string   `yaml:"root_path"`
	CachePath      string   `yaml:"cache_path" default:"./cache"`
	VersionsPath   string   `yaml:"versions_path" default:"./backups.txt"`
	UserIdentifier string   `yaml:"user_identifier" default:"anonymous"`
	TagsGenerator  []string `yaml:"tags_generator"`
	Hooks          struct {
		PreBackup   []string `yaml:"pre_backup"`
		PostBackup  []string `yaml:"post_backup"`
		PreRestore  []string `yaml:"pre_restore"`
		PostRestore []string `yaml:"post_restore"`
	} `yaml:"hooks"`
	Options struct {
		Compress string `yaml:"compress" default:"tgz"`
		Hash     string `yaml:"hash" default:"md5"`
	} `yaml:"options"`
}

type TargetDriver string

const (
	TargetDriverLocal = "local"
)

type TargetConfig struct {
	PrefixPath string   `yaml:"prefix_path"`
	Includes   []string `yaml:"includes"`
	Excludes   []string `yaml:"excludes"`
}

type BackupDriver string

const (
	BackupDriverS3    = "s3"
	BackupDriverLocal = "local"
)

type BackupConfig struct {
	Driver string              `yaml:"driver"`
	S3     S3StorageConfig     `yaml:"s3"`
	Local  LocalStorageConfig  `yaml:"local"`
	Remote RemoteStorageConfig `yaml:"remote"`
}

type LocalStorageConfig struct {
	Path string `yaml:"path"`
}

type RemoteStorageConfig struct {
	Path            string `yaml:"path"`
	Host            string `yaml:"host"`
	User            string `yaml:"user"`
	CredentialsPath string `yaml:"ssh_credential_file"`
}

type S3StorageConfig struct {
	PrefixPath      string `yaml:"prefix_path"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region          string `yaml:"region"`
	Bucket          string `yaml:"bucket"`
}

type Config struct {
	Version   string          `yaml:"version"`
	Workspace WorkspaceConfig `yaml:"workspace"`
	Target    TargetConfig    `yaml:"target"`
	Backups   []BackupConfig  `yaml:"backups"`
}

func FromFile(filename string) (*Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Naively expand environment variables in the file before parsing.
	s := []byte(os.ExpandEnv(string(b)))

	var cfg Config
	err = yaml.Unmarshal(s, &cfg)

	if err := defaults.Set(&cfg); err != nil {
		return nil, fmt.Errorf("set defaults: %v", err)
	}

	// Use config file path as rootpath if empty.
	if cfg.Workspace.RootPath == "" {
		path, err := filepath.Abs(filename)
		if err != nil {
			return nil, fmt.Errorf("abs of %q: %v", filename, err)
		}
		cfg.Workspace.RootPath = filepath.Dir(path)
	}

	return &cfg, err
}
