package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Check connection defaults
	if cfg.Connection.Timeout != "30s" {
		t.Errorf("Connection.Timeout = %s, want 30s", cfg.Connection.Timeout)
	}
	if cfg.Connection.MaxOpenConns != 25 {
		t.Errorf("Connection.MaxOpenConns = %d, want 25", cfg.Connection.MaxOpenConns)
	}

	// Check logging defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %s, want info", cfg.Logging.Level)
	}

	// Check editor defaults
	if cfg.Editor.DefaultLimit != 10000 {
		t.Errorf("Editor.DefaultLimit = %d, want 10000", cfg.Editor.DefaultLimit)
	}
	if !cfg.Editor.ConfirmDestructive {
		t.Error("Editor.ConfirmDestructive should be true by default")
	}

	// Check UI defaults
	if !cfg.UI.ShowDataPreview {
		t.Error("UI.ShowDataPreview should be true by default")
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Connection.Timeout = "60s"
	cfg.Editor.DefaultLimit = 5000
	cfg.Logging.Level = "debug"

	if err := cfg.SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load and verify
	loaded, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath() error = %v", err)
	}

	if loaded.Connection.Timeout != "60s" {
		t.Errorf("Connection.Timeout = %s, want 60s", loaded.Connection.Timeout)
	}
	if loaded.Editor.DefaultLimit != 5000 {
		t.Errorf("Editor.DefaultLimit = %d, want 5000", loaded.Editor.DefaultLimit)
	}
	if loaded.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %s, want debug", loaded.Logging.Level)
	}
}

func TestLoadFromPath_NotExists(t *testing.T) {
	cfg, err := LoadFromPath("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("LoadFromPath() should return defaults for non-existent file, got error: %v", err)
	}

	// Should return defaults
	if cfg.Connection.Timeout != "30s" {
		t.Errorf("Expected default timeout, got %s", cfg.Connection.Timeout)
	}
}

func TestLoadFromPath_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFromPath(path)
	if err == nil {
		t.Error("LoadFromPath() should return error for invalid YAML")
	}
}

func TestConfig_GetConnectionTimeout(t *testing.T) {
	tests := []struct {
		timeout string
		want    time.Duration
	}{
		{"30s", 30 * time.Second},
		{"1m", 1 * time.Minute},
		{"invalid", 30 * time.Second}, // Falls back to default
		{"", 30 * time.Second},        // Falls back to default
	}

	for _, tt := range tests {
		cfg := &Config{Connection: ConnectionConfig{Timeout: tt.timeout}}
		got := cfg.GetConnectionTimeout()
		if got != tt.want {
			t.Errorf("GetConnectionTimeout() with %q = %v, want %v", tt.timeout, got, tt.want)
		}
	}
}

func TestConfig_GetConnMaxLifetime(t *testing.T) {
	cfg := &Config{Connection: ConnectionConfig{ConnMaxLifetime: "10m"}}
	got := cfg.GetConnMaxLifetime()
	if got != 10*time.Minute {
		t.Errorf("GetConnMaxLifetime() = %v, want 10m", got)
	}

	// Test fallback
	cfg.Connection.ConnMaxLifetime = "invalid"
	got = cfg.GetConnMaxLifetime()
	if got != 5*time.Minute {
		t.Errorf("GetConnMaxLifetime() fallback = %v, want 5m", got)
	}
}

func TestConfig_PartialOverride(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "partial.yaml")

	// Write partial config - only override some values
	content := `connection:
  timeout: "45s"
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath() error = %v", err)
	}

	// Overridden value
	if cfg.Connection.Timeout != "45s" {
		t.Errorf("Connection.Timeout = %s, want 45s", cfg.Connection.Timeout)
	}

	// Default values preserved
	if cfg.Connection.MaxOpenConns != 25 {
		t.Errorf("Connection.MaxOpenConns = %d, want 25 (default)", cfg.Connection.MaxOpenConns)
	}
	if cfg.Editor.DefaultLimit != 10000 {
		t.Errorf("Editor.DefaultLimit = %d, want 10000 (default)", cfg.Editor.DefaultLimit)
	}
}
