package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Display debug and diagnostic information",
	Long: `Display system information, available drivers, configuration paths,
and other diagnostic details useful for troubleshooting.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("=== dbsmith Debug Information ===")
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("OS: %s\n", runtime.GOOS)
		fmt.Printf("Architecture: %s\n", runtime.GOARCH)
		fmt.Println("\nAvailable Drivers:")
		fmt.Println("  - PostgreSQL (pgx)")
		fmt.Println("  - MySQL/MariaDB")
		fmt.Println("  - SQLite")
		fmt.Println("\nConfiguration:")
		fmt.Printf("  Config Dir: %s\n", configDir)
		fmt.Printf("  Workspace: %s\n", workspace)
		fmt.Println("\n(Full diagnostics not yet implemented)")
		return nil
	},
}
