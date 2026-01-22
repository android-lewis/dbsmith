package logging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/android-lewis/dbsmith/internal/constants"
	"github.com/rs/zerolog"
)

func TestDefaultConfig(t *testing.T) {
	configDir := "/tmp/test"
	cfg := DefaultConfig(configDir)

	if cfg.LogDir != filepath.Join(configDir, "logs") {
		t.Errorf("expected log dir %s, got %s", filepath.Join(configDir, "logs"), cfg.LogDir)
	}

	if cfg.LogFile != constants.DefaultLogFileName {
		t.Errorf("expected log file %s, got %s", constants.DefaultLogFileName, cfg.LogFile)
	}

	if cfg.MaxSizeMB != constants.DefaultMaxSizeMB {
		t.Errorf("expected max size %d, got %d", constants.DefaultMaxSizeMB, cfg.MaxSizeMB)
	}

	if cfg.MaxBackups != constants.DefaultMaxBackups {
		t.Errorf("expected max backups %d, got %d", constants.DefaultMaxBackups, cfg.MaxBackups)
	}

	if cfg.MaxAgeDays != constants.DefaultMaxAgeDays {
		t.Errorf("expected max age %d, got %d", constants.DefaultMaxAgeDays, cfg.MaxAgeDays)
	}

	if cfg.Level != zerolog.InfoLevel {
		t.Errorf("expected level %s, got %s", zerolog.InfoLevel, cfg.Level)
	}
}

func TestInitialize(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		LogDir:     filepath.Join(tmpDir, "logs"),
		LogFile:    "test.log",
		MaxSizeMB:  5,
		MaxBackups: 2,
		MaxAgeDays: 7,
		Level:      zerolog.DebugLevel,
	}

	err := Initialize(cfg)
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		_ = Close()
	}()

	logPath := filepath.Join(cfg.LogDir, cfg.LogFile)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("log file was not created at %s", logPath)
	}

	if globalLogger == nil {
		t.Error("global logger was not set")
	}
}

func TestGetBeforeInitialize(t *testing.T) {
	globalLogger = nil

	logger := Get()
	if logger == nil {
		t.Error("Get() returned nil")
	}

	logger.Info().Msg("test")
}

func TestLoggerMethods(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		LogDir:     filepath.Join(tmpDir, "logs"),
		LogFile:    "test.log",
		MaxSizeMB:  5,
		MaxBackups: 2,
		MaxAgeDays: 7,
		Level:      zerolog.DebugLevel,
	}

	err := Initialize(cfg)
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		_ = Close()
	}()

	tests := []struct {
		name string
		fn   func() *zerolog.Event
	}{
		{"Info", Info},
		{"Debug", Debug},
		{"Warn", Warn},
		{"Error", Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := tt.fn()
			if event == nil {
				t.Errorf("%s() returned nil", tt.name)
			}
			event.Msg("test message")
		})
	}
}

func TestSetLevel(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		LogDir:     filepath.Join(tmpDir, "logs"),
		LogFile:    "test.log",
		MaxSizeMB:  5,
		MaxBackups: 2,
		MaxAgeDays: 7,
		Level:      zerolog.InfoLevel,
	}

	err := Initialize(cfg)
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		_ = Close()
	}()

	SetLevel(zerolog.DebugLevel)

	if globalLogger.GetLevel() != zerolog.DebugLevel {
		t.Errorf("expected level %s, got %s", zerolog.DebugLevel, globalLogger.GetLevel())
	}
}

func TestWith(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		LogDir:     filepath.Join(tmpDir, "logs"),
		LogFile:    "test.log",
		MaxSizeMB:  5,
		MaxBackups: 2,
		MaxAgeDays: 7,
		Level:      zerolog.InfoLevel,
	}

	err := Initialize(cfg)
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		_ = Close()
	}()

	ctx := With()
	logger := ctx.Str("key", "value").Logger()
	logger.Info().Msg("test")
}
