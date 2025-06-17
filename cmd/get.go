// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'get' command which is used to
// retrieve secrets from the encrypted store.
package cmd

import (
	"fmt"

	"github.com/kirinyoku/vlxck/internal/store"
	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve a secret from the store",
	Long: `Retrieve a secret from the store and copy it to the clipboard.

In interactive mode, you can select the secret from a list.
In non-interactive mode, you must specify the secret name.

Examples:
  # Interactive mode
  vlxck get -i

  # Non-interactive mode
  vlxck get -n example.com`,

	Run: func(cmd *cobra.Command, args []string) {
		filePath := getStorePath()
		password := utils.PromptForPassword("Enter master password: ")
		s, err := store.LoadStore(filePath, password)
		if err != nil {
			fmt.Println("Error loading store:", err)
			return
		}

		// Check for interactive mode
		interactive, _ := cmd.Flags().GetBool("interactive")
		if interactive {
			getInteractive(s)
			return
		}

		// Non-interactive mode
		getNonInteractive(cmd, s)
	},
}

// getInteractive handles the interactive get flow
func getInteractive(s *store.Store) {
	if len(s.Secrets) == 0 {
		fmt.Println("No secrets found.")
		return
	}

	// Create a list of secret names for selection
	secretNames := make([]string, 0, len(s.Secrets))
	for _, secret := range s.Secrets {
		secretNames = append(secretNames, secret.Name)
	}

	// Prompt user to select a secret
	selectedName, err := utils.PromptForSelect("Select secret to retrieve", secretNames)
	if err != nil {
		fmt.Println("Error selecting secret:", err)
		return
	}

	// Find and copy the selected secret
	for _, secret := range s.Secrets {
		if secret.Name == selectedName {
			copySecretToClipboard(secret)
			return
		}
	}
}

// getNonInteractive handles the non-interactive get flow
func getNonInteractive(cmd *cobra.Command, s *store.Store) {
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		fmt.Println("Error: secret name is required in non-interactive mode")
		return
	}

	for _, secret := range s.Secrets {
		if secret.Name == name {
			copySecretToClipboard(secret)
			return
		}
	}
	fmt.Printf("Secret '%s' not found.\n", name)
}

// copySecretToClipboard copies the secret value to clipboard and provides feedback
func copySecretToClipboard(secret store.Secret) {
	if err := utils.CopyToClipboard(secret.Value); err != nil {
		fmt.Printf("Value: %s (clipboard error: %v)\n", secret.Value, err)
	} else {
		fmt.Printf("Secret '%s' copied to clipboard.\n", secret.Name)
	}
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Define command flags with shorthand and descriptions
	getCmd.Flags().StringP("name", "n", "", "Name of the secret (required in non-interactive mode)")
	getCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode to select from a list")

	// Mark name as required only in non-interactive mode
	getCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		interactive, _ := cmd.Flags().GetBool("interactive")
		name, _ := cmd.Flags().GetString("name")
		if !interactive && name == "" {
			return fmt.Errorf("either --name or --interactive flag is required")
		}
		return nil
	}
}
