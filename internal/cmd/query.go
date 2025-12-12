package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query <connection> <sql>",
	Short: "Execute a SQL query and display results",
	Long: `Execute a SQL query on a specified connection and display results.
Results can be formatted as text, CSV, JSON, or other formats.`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("query command: %v (not yet implemented)\n", args)
		return nil
	},
}

func init() {
	queryCmd.Flags().String("output-format", "text", "Output format (text, csv, json, jsonl, sql)")
}
