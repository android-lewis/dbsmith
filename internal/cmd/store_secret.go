package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var storeSecretCmd = &cobra.Command{
	Use:   "store-secret <key-id>",
	Short: "Store or update a secret in the keyring",
	Long: `Interactively store or update a secret (password) in the system keyring.
The key-id is used to reference the secret when configuring database connections.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyID := args[0]
		fmt.Printf("store-secret command: keyID=%s (not yet implemented)\n", keyID)
		return nil
	},
}
