package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose   bool
	configDir string
	workspace string
)

var rootCmd = &cobra.Command{
	Use:   "dbsmith",
	Short: "A TUI database client for exploring and querying relational databases",
	Long: `dbsmith is a terminal user interface (TUI) application for connecting to,
exploring, and querying relational databases. It supports PostgreSQL, MySQL/MariaDB,
and SQLite databases with features for schema browsing, SQL editing, streaming results,
and exporting data in multiple formats.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
		} else {
			log.SetOutput(nil) // Suppress logs in quiet mode
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "Configuration directory (defaults to ~/.config/dbsmith)")
	rootCmd.PersistentFlags().StringVar(&workspace, "workspace", "", "Workspace file to load")

	// Add subcommands
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(storeSecretCmd)
	rootCmd.AddCommand(listConnectionsCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("dbsmith version 0.1.0-dev")
	},
}
