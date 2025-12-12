package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open [workspace-path]",
	Short: "Open a workspace in the TUI",
	Long: `Launch the dbsmith TUI with the specified workspace file.
If no workspace is provided, uses the default workspace or prompts to create one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspacePath string
		if len(args) > 0 {
			workspacePath = args[0]
		}
		fmt.Printf("open command: workspace=%s (not yet implemented)\n", workspacePath)
		return nil
	},
}
