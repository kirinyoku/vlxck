// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'delete' command which is used to
// delete secrets from the encrypted store.
package cmd

import (
	"fmt"

	"github.com/kirinyoku/vlxck/internal/store"
	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a secret from the store",
	Long: `Delete a secret from the store by its name or through interactive selection.

Examples:
  # Interactive mode
  vlxck delete -i

  # Non-interactive mode
  vlxck delete -n example.com`,
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
			deleteInteractive(s, filePath, password)
			return
		}

		// Non-interactive mode
		deleteNonInteractive(cmd, s, filePath, password)
	},
}

// deleteInteractive handles the interactive delete flow
func deleteInteractive(s *store.Store, filePath, password string) {
	if len(s.Secrets) == 0 {
		fmt.Println("No secrets found to delete.")
		return
	}

	// Create a list of secret names for selection
	secretNames := make([]string, 0, len(s.Secrets))
	for _, secret := range s.Secrets {
		secretNames = append(secretNames, secret.Name)
	}

	// Prompt user to select a secret to delete
	selectedName, err := utils.PromptForSelect("Select secret to delete", secretNames)
	if err != nil {
		fmt.Println("Error selecting secret:", err)
		return
	}

	// Confirm deletion
	confirm, err := utils.PromptForConfirm(fmt.Sprintf("Are you sure you want to delete '%s'?", selectedName))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if !confirm {
		fmt.Println("Deletion cancelled.")
		return
	}

	// Find and delete the selected secret
	for i, secret := range s.Secrets {
		if secret.Name == selectedName {
			s.Secrets = append(s.Secrets[:i], s.Secrets[i+1:]...)
			if err := store.SaveStore(filePath, password, s); err != nil {
				fmt.Println("Error saving store:", err)
				return
			}
			fmt.Printf("Secret '%s' deleted successfully.\n", selectedName)
			return
		}
	}
}

// deleteNonInteractive handles the non-interactive delete flow
func deleteNonInteractive(cmd *cobra.Command, s *store.Store, filePath, password string) {
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		fmt.Println("Error: secret name is required in non-interactive mode")
		return
	}

	for i, secret := range s.Secrets {
		if secret.Name == name {
			s.Secrets = append(s.Secrets[:i], s.Secrets[i+1:]...)
			if err := store.SaveStore(filePath, password, s); err != nil {
				fmt.Println("Error saving store:", err)
				return
			}
			fmt.Printf("Secret '%s' deleted successfully.\n", name)
			return
		}
	}
	fmt.Printf("Secret '%s' not found.\n", name)
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Define command flags with shorthand and descriptions
	deleteCmd.Flags().StringP("name", "n", "", "Name of the secret (required in non-interactive mode)")
	deleteCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode to select from a list")

	// Mark name as required only in non-interactive mode
	deleteCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		interactive, _ := cmd.Flags().GetBool("interactive")
		name, _ := cmd.Flags().GetString("name")
		if !interactive && name == "" {
			return fmt.Errorf("either --name or --interactive flag is required")
		}
		return nil
	}
}
