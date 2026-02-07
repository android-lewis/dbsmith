package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/android-lewis/dbsmith/internal/config"
	"github.com/android-lewis/dbsmith/internal/db"
	"github.com/android-lewis/dbsmith/internal/executor"
	"github.com/android-lewis/dbsmith/internal/explorer"
	"github.com/android-lewis/dbsmith/internal/logging"
	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/secrets"
	wsmgr "github.com/android-lewis/dbsmith/internal/workspace"
	"github.com/rs/zerolog"
)

const (
	DefaultTimeout = 10 * time.Second
	ConfigDirName  = ".config/dbsmith"
	WorkspaceFile  = "workspace.yaml"
)

type App struct {
	Cleanup func()
	Context context.Context

	Config         *config.Config
	Workspace      *wsmgr.Manager
	Connection     *models.Connection
	SecretsManager secrets.Manager
	Driver         db.Driver
	Executor       *executor.QueryExecutor
	Explorer       *explorer.Explorer

	configDir string
}

func New() (*App, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	logCfg := logging.DefaultConfig(configDir)

	switch cfg.Logging.Level {
	case "debug":
		logCfg.Level = zerolog.DebugLevel
	case "warn":
		logCfg.Level = zerolog.WarnLevel
	case "error":
		logCfg.Level = zerolog.ErrorLevel
	default:
		logCfg.Level = zerolog.InfoLevel
	}

	if os.Getenv("DBSMITH_LOG_LEVEL") == "debug" {
		logCfg.Level = zerolog.DebugLevel
	}
	logCfg.MaxSizeMB = cfg.Logging.MaxSizeMB
	logCfg.MaxBackups = cfg.Logging.MaxBackups
	logCfg.MaxAgeDays = cfg.Logging.MaxAgeDays

	if err := logging.Initialize(logCfg); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	logging.Info().
		Str("config_dir", configDir).
		Msg("Starting DBSmith")

	ws, secret, err := loadWorkspace(configDir)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to load workspace")
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	logging.Info().Msg("Application initialized successfully")

	return &App{
		Config:         cfg,
		Workspace:      ws,
		SecretsManager: secret,
		Cleanup:        cancel,
		Context:        ctx,
		configDir:      configDir,
	}, nil
}

func (a *App) SaveWorkspace() error {
	wsPath := filepath.Join(a.configDir, WorkspaceFile)
	return a.Workspace.Save(wsPath)
}

func (a *App) ConnectToDatabase(conn *models.Connection) error {
	if conn == nil {
		logging.Error().Msg("Connection is nil")
		return fmt.Errorf("connection is nil")
	}

	logging.Info().
		Str("connection_name", conn.Name).
		Str("connection_type", string(conn.Type)).
		Msg("Connecting to database")

	driverFactory := db.NewDriverFactory()
	driver, err := driverFactory.Create(conn)
	if err != nil {
		logging.Error().
			Err(err).
			Str("connection_name", conn.Name).
			Msg("Failed to create driver")
		return fmt.Errorf("failed to create driver: %w", err)
	}

	// Use configured timeout
	timeout := a.Config.GetConnectionTimeout()
	ctx, cancel := context.WithTimeout(a.Context, timeout)
	defer cancel()

	if err := driver.Connect(ctx, conn, a.SecretsManager); err != nil {
		logging.Error().
			Err(err).
			Str("connection_name", conn.Name).
			Msg("Failed to connect to database")
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	a.Driver = driver
	a.Connection = conn
	a.Executor = executor.NewQueryExecutor(driver)
	a.Explorer = explorer.NewExplorer(driver)

	logging.Info().
		Str("connection_name", conn.Name).
		Str("connection_type", string(conn.Type)).
		Msg("Successfully connected to database")

	return nil
}

func (a *App) Disconnect() error {
	if a.Driver == nil || !a.Driver.IsConnected() {
		return nil
	}

	connName := "unknown"
	if a.Connection != nil {
		connName = a.Connection.Name
	}

	logging.Info().Str("connection_name", connName).Msg("Disconnecting from database")

	ctx, cancel := context.WithTimeout(a.Context, DefaultTimeout)
	defer cancel()

	if err := a.Driver.Disconnect(ctx); err != nil {
		logging.Error().Err(err).Str("connection_name", connName).Msg("Failed to disconnect")
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	a.Driver = nil
	a.Connection = nil
	a.Executor = nil
	a.Explorer = nil

	logging.Info().Str("connection_name", connName).Msg("Successfully disconnected")

	return nil
}

func loadWorkspace(configDir string) (*wsmgr.Manager, secrets.Manager, error) {
	wsPath := filepath.Join(configDir, WorkspaceFile)

	if _, err := os.Stat(wsPath); os.IsNotExist(err) {
		return createNewWorkspace(configDir, wsPath)
	}

	return loadExistingWorkspace(configDir, wsPath)
}

func createNewWorkspace(configDir, wsPath string) (*wsmgr.Manager, secrets.Manager, error) {
	logging.Info().Str("workspace_path", wsPath).Msg("Creating new workspace")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		logging.Error().Err(err).Str("config_dir", configDir).Msg("Failed to create config directory")
		return nil, nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	wsMgr := wsmgr.New()
	wsMgr.SetName("Default Workspace")

	if err := wsMgr.Save(wsPath); err != nil {
		logging.Error().Err(err).Str("workspace_path", wsPath).Msg("Failed to save workspace")
		return nil, nil, fmt.Errorf("failed to create workspace file: %w", err)
	}

	secretsMgr, err := secrets.NewManager(configDir)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to initialize secrets manager")
		return nil, nil, fmt.Errorf("failed to initialize secrets manager: %w", err)
	}

	logging.Info().Str("workspace_path", wsPath).Msg("New workspace created successfully")

	return wsMgr, secretsMgr, nil
}

func loadExistingWorkspace(configDir, wsPath string) (*wsmgr.Manager, secrets.Manager, error) {
	logging.Info().Str("workspace_path", wsPath).Msg("Loading existing workspace")

	wsMgr, err := wsmgr.Load(wsPath)
	if err != nil {
		logging.Error().Err(err).Str("workspace_path", wsPath).Msg("Failed to load workspace")
		return nil, nil, fmt.Errorf("failed to load workspace: %w", err)
	}

	secretsMgr, err := secrets.NewManager(configDir)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to initialize secrets manager")
		return nil, nil, fmt.Errorf("failed to initialize secrets manager: %w", err)
	}

	logging.Info().
		Str("workspace_path", wsPath).
		Str("workspace_name", wsMgr.GetName()).
		Int("connections", len(wsMgr.ListConnections())).
		Msg("Workspace loaded successfully")

	return wsMgr, secretsMgr, nil
}
