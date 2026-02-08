package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/android-lewis/dbsmith/internal/constants"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Connection ConnectionConfig `yaml:"connection"`
	Logging    LoggingConfig    `yaml:"logging"`
	Editor     EditorConfig     `yaml:"editor"`
	UI         UIConfig         `yaml:"ui"`
}

type ConnectionConfig struct {
	Timeout         string `yaml:"timeout"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime string `yaml:"conn_max_idle_time"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
}

type EditorConfig struct {
	DefaultLimit       int  `yaml:"default_limit"`
	ConfirmDestructive bool `yaml:"confirm_destructive"`
	TabSize            int  `yaml:"tab_size"`
}

type UIConfig struct {
	ShowDataPreview     bool `yaml:"show_data_preview"`
	ShowSchemas         bool `yaml:"show_schemas"`
	ShowIndexes         bool `yaml:"show_indexes"`
	MaxPreviewCellWidth int  `yaml:"max_preview_cell_width"`
}

func DefaultConfig() *Config {
	return &Config{
		Connection: ConnectionConfig{
			Timeout:         "30s",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: "5m",
			ConnMaxIdleTime: "5m",
		},
		Logging: LoggingConfig{
			Level:      "info",
			MaxSizeMB:  10,
			MaxBackups: 3,
			MaxAgeDays: 28,
		},
		Editor: EditorConfig{
			DefaultLimit:       10000,
			ConfirmDestructive: true,
			TabSize:            4,
		},
		UI: UIConfig{
			ShowDataPreview:     true,
			ShowSchemas:         true,
			ShowIndexes:         false,
			MaxPreviewCellWidth: 50,
		},
	}
}

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "dbsmith"), nil
}

func GetConfigFilePath(filename string) (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}

func GetConfigPath() (string, error) {
	return GetConfigFilePath(constants.DefaultConfigFileName)
}

func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}
	return LoadFromPath(path)
}

func LoadFromPath(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {

			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	return c.SaveToPath(path)
}

func (c *Config) SaveToPath(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) GetConnectionTimeout() time.Duration {
	d, err := time.ParseDuration(c.Connection.Timeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func (c *Config) GetConnMaxLifetime() time.Duration {
	d, err := time.ParseDuration(c.Connection.ConnMaxLifetime)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}

func (c *Config) GetConnMaxIdleTime() time.Duration {
	d, err := time.ParseDuration(c.Connection.ConnMaxIdleTime)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}
