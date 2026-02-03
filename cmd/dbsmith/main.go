package main

import (
	"fmt"
	"os"

	"github.com/android-lewis/dbsmith/internal/app"
	"github.com/android-lewis/dbsmith/internal/logging"
	"github.com/android-lewis/dbsmith/internal/tui"
	"github.com/spf13/cobra"
)

// Build-time variables injected via ldflags
var (
	Version      = "dev"
	Revision     = "unknown"
	RevisionTime = "unknown"
)

const appName = "dbsmith"

var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "DBSmith - A production-ready database TUI",
	Long: `dbsmith is a comprehensive database client with an interactive terminal UI.
It supports PostgreSQL, MySQL/MariaDB, and SQLite databases.

Features:
- Interactive schema browsing
- SQL query editor with syntax awareness
- Save and load queries
- Export results (CSV, JSON, SQL)
- Secure credential storage`,
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		application, err := app.New(Version)
		if err != nil {
			return err
		}
		defer func() {
			application.Cleanup()
			if err := logging.Close(); err != nil {
				logging.Error().Err(err).Msg("Failed to close logger")
			}
		}()

		if err := tui.Run(application); err != nil {
			logging.Error().Err(err).Msg("TUI runtime error")
			return err
		}

		logging.Info().Msg("DBSmith shutting down")
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
