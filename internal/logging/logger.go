package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/android-lewis/dbsmith/internal/constants"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	globalLogger *zerolog.Logger
	logWriter    io.WriteCloser
)

type Config struct {
	LogDir     string
	LogFile    string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Level      zerolog.Level
}

func DefaultConfig(configDir string) Config {
	return Config{
		LogDir:     filepath.Join(configDir, "logs"),
		LogFile:    constants.DefaultLogFileName,
		MaxSizeMB:  constants.DefaultMaxSizeMB,
		MaxBackups: constants.DefaultMaxBackups,
		MaxAgeDays: constants.DefaultMaxAgeDays,
		Level:      zerolog.InfoLevel,
	}
}

func Initialize(cfg Config) error {
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(cfg.LogDir, cfg.LogFile)

	fileWriter := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   true,
	}

	logWriter = fileWriter

	zerolog.TimeFieldFormat = time.RFC3339

	var writers []io.Writer
	writers = append(writers, fileWriter)

	if os.Getenv("DBSMITH_DEBUG") != "" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}
		writers = append(writers, consoleWriter)
	}

	multi := zerolog.MultiLevelWriter(writers...)
	logger := zerolog.New(multi).
		Level(cfg.Level).
		With().
		Timestamp().
		Caller().
		Logger()

	globalLogger = &logger

	logger.Info().
		Str("log_file", logPath).
		Str("level", cfg.Level.String()).
		Msg("Logger initialized")

	return nil
}

func Close() error {
	if logWriter != nil {
		return logWriter.Close()
	}
	return nil
}

func Get() *zerolog.Logger {
	if globalLogger == nil {
		noop := zerolog.Nop()
		return &noop
	}
	return globalLogger
}

func Info() *zerolog.Event {
	return Get().Info()
}

func Debug() *zerolog.Event {
	return Get().Debug()
}

func Warn() *zerolog.Event {
	return Get().Warn()
}

func Error() *zerolog.Event {
	return Get().Error()
}

func Fatal() *zerolog.Event {
	return Get().Fatal()
}

func Panic() *zerolog.Event {
	return Get().Panic()
}

func With() zerolog.Context {
	return Get().With()
}

func SetLevel(level zerolog.Level) {
	if globalLogger != nil {
		updated := globalLogger.Level(level)
		globalLogger = &updated
	}
}
