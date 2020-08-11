package config

import (
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
)

type WorkspaceConfig struct {
	RootPath       string   `yaml:"root_path"`
	CachePath      string   `yaml:"cache_path"`
	VersionsPath   string   `yaml:"versions_path"`
	UserIdentifier string   `yaml:"user_identifier"`
	TagsGenerator  []string `yaml:"tags_generator"`
	Hooks          struct {
		PreBackup   []string `yaml:"pre_backup"`
		PostBackup  []string `yaml:"post_backup"`
		PreRestore  []string `yaml:"pre_restore"`
		PostRestore []string `yaml:"post_restore"`
	} `yaml:"hooks"`
	Options struct {
		Compress string `yaml:"compress"`
		Hash     string `yaml:"hash"`
	} `yaml:"options"`
}

type TargetDriver string

const (
	TargetDriverLocal = "local"
)

type TargetConfig struct {
	Driver string `yaml:"driver"`
	Local  struct {
		PrefixPath string   `yaml:"prefix_path"`
		Includes   []string `yaml:"includes"`
		Excludes   []string `yaml:"excludes"`
	} `yaml:"local"`
}

type BackupDriver string

const (
	BackupDriverS3    = "s3"
	BackupDriverLocal = "local"
)

type BackupConfig struct {
	Driver string `yaml:"driver"`
	S3     struct {
		PrefixPath      string `yaml:"prefix_path"`
		AccessKeyID     string `yaml:"access_key_id"`
		SecretAccessKey string `yaml:"secret_access_key"`
		Region          string `yaml:"region"`
		Bucket          string `yaml:"bucket"`
	} `yaml:"s3"`
}

type Config struct {
	Version   string          `yaml:"version"`
	Workspace WorkspaceConfig `yaml:"workspace"`
	Target    TargetConfig    `yaml:"target"`
	Backup    BackupConfig    `yaml:"backup"`
}

func FromFile(filename string) (*Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Expand environment variables in the file before parsing.
	s := []byte(os.ExpandEnv(string(b)))

	var cfg Config
	err = yaml.Unmarshal(s, &cfg)
	return &cfg, err
}
