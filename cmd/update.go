// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'update' command which is used to
// update existing secrets in the encrypted store.
package cmd

import (
	"fmt"

	"github.com/kirinyoku/vlxck/internal/store"
	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/spf13/cobra"
)

// updateCmd represents the 'update' command that allows users to update existing secrets.
// It updates the value or category of an existing secret in the store.
//
// The command supports two modes of operation:
// 1. Interactive mode (-i): Guides user through prompts
// 2. Non-interactive mode: Uses command-line flags for automation
//
// The command supports the following flags:
//   - name (-n): The name/identifier of the secret (required in non-interactive mode)
//   - value (-v): The new secret value (or use -g to generate)
//   - category (-c): The new category for the secret (use "-" to keep existing)
//   - generate (-g): Generate a new random password for the secret
//   - length (-l): Length of generated password (default: 16)
//   - symbols (-s): Include symbols in generated password
//   - digits (-d): Include digits in generated password
//   - interactive (-i): Use interactive mode (overrides other flags)
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing secret in the secure store",
	Long: `Update an existing secret in the secure store.

In interactive mode, you'll be guided through the update process with prompts.
In non-interactive mode, you must specify the secret name and at least one of:
- A new value (--value/-v)
- The --generate flag to create a new password
- A new category (--category/-c)

Examples:
  # Interactive mode
  vlxck update -i

  # Update just the password (generated)
  vlxck update -n example.com -g

  # Update with specific value and category
  vlxck update -n example.com -v newpassword -c work

  # Generate a 24-char password with symbols and digits
  vlxck update -n example.com -gdsl 24`,

	Run: func(cmd *cobra.Command, args []string) {
		filePath := getStorePath()
		password := utils.PromptForPassword("Enter master password: ")
		s, err := store.LoadStore(filePath, password)
		if err != nil {
			fmt.Println("Error loading store:", err)
			return
		}

		// Check for interactive mode first
		interactive, _ := cmd.Flags().GetBool("interactive")
		if interactive {
			updateInteractive(s, filePath, password)
			return
		}

		// Non-interactive mode
		updateNonInteractive(cmd, s, filePath, password)
	},
}

// updateInteractive handles the interactive update flow
func updateInteractive(s *store.Store, filePath, password string) {
	// Show list of secrets for user to choose from
	if len(s.Secrets) == 0 {
		fmt.Println("No secrets found to update.")
		return
	}

	// Get secret name from user
	secretNames := make([]string, 0, len(s.Secrets))
	for _, secret := range s.Secrets {
		secretNames = append(secretNames, secret.Name)
	}

	selectedName, err := utils.PromptForSelect("Select secret to update", secretNames)
	if err != nil {
		fmt.Println("Error selecting secret:", err)
		return
	}

	// Find the selected secret
	var secretToUpdate *store.Secret
	for i := range s.Secrets {
		if s.Secrets[i].Name == selectedName {
			secretToUpdate = &s.Secrets[i]
			break
		}
	}

	if secretToUpdate == nil {
		fmt.Println("Error: Secret not found")
		return
	}

	// Ask what to update
	updateOptions := []string{"Update value", "Update category", "Update both", "Cancel"}
	action, err := utils.PromptForSelect("What would you like to update?", updateOptions)
	if err != nil || action == "Cancel" {
		return
	}

	// Handle value update if needed
	if action == "Update value" || action == "Update both" {
		updateValue, err := utils.PromptForSelect("Choose value input method",
			[]string{"Enter new value", "Generate password"})
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if updateValue == "Generate password" {
			// Generate password with custom parameters
			length, err := utils.PromptForInt("Enter password length", 16, 1, 100)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			symbols, err := utils.PromptForSelect("Include symbols?", []string{"Yes", "No"})
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			digits, err := utils.PromptForSelect("Include numbers?", []string{"Yes", "No"})
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			value, err := utils.GeneratePassword(length, symbols == "Yes", digits == "Yes")
			if err != nil {
				fmt.Println("Error generating password:", err)
				return
			}

			// Copy to clipboard
			if err := utils.CopyToClipboard(value); err != nil {
				fmt.Println("Warning: Could not copy to clipboard:", err)
			} else {
				fmt.Println("Generated password copied to clipboard.")
			}

			secretToUpdate.Value = value
		} else {
			// Prompt for new value
			value, err := utils.PromptForInput("Enter new value", "", func(input string) error {
				if input == "" {
					return fmt.Errorf("value cannot be empty")
				}
				return nil
			})
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			secretToUpdate.Value = value
		}
	}

	// Handle category update if needed
	if action == "Update category" || action == "Update both" {
		category, err := utils.PromptForInput("Enter new category (leave empty to remove)",
			secretToUpdate.Category, nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		secretToUpdate.Category = category
	}

	// Save changes
	if err := store.SaveStore(filePath, password, s); err != nil {
		fmt.Println("Error saving store:", err)
		return
	}

	// Copy new value to clipboard if it was updated
	if action == "Update value" || action == "Update both" {
		if err := utils.CopyToClipboard(secretToUpdate.Value); err != nil {
			fmt.Println("Warning: Could not copy to clipboard:", err)
		} else {
			fmt.Println("Updated value copied to clipboard.")
		}
	}

	fmt.Println("Secret updated successfully.")
}

// updateNonInteractive handles the non-interactive update flow
func updateNonInteractive(cmd *cobra.Command, s *store.Store, filePath, password string) {
	name, _ := cmd.Flags().GetString("name")
	value, _ := cmd.Flags().GetString("value")
	category, _ := cmd.Flags().GetString("category")
	generate, _ := cmd.Flags().GetBool("generate")
	length, _ := cmd.Flags().GetInt("length")
	symbols, _ := cmd.Flags().GetBool("symbols")
	digits, _ := cmd.Flags().GetBool("digits")
	// Find the secret to update
	secretFound := false
	for i := range s.Secrets {
		if s.Secrets[i].Name == name {
			secretFound = true
			secret := &s.Secrets[i]

			// Update value if provided or if generate is true
			if value != "" {
				secret.Value = value
			} else if generate {
				// Generate new password with specified parameters
				newValue, err := utils.GeneratePassword(length, symbols, digits)
				if err != nil {
					fmt.Println("Error generating password:", err)
					return
				}
				secret.Value = newValue

				// Copy to clipboard
				if err := utils.CopyToClipboard(newValue); err != nil {
					fmt.Println("Warning: Could not copy to clipboard:", err)
				} else {
					fmt.Println("Generated password copied to clipboard.")
				}
			}

			// Update category if provided and not "-"
			if category != "-" {
				secret.Category = category
			}

			// Save changes
			if err := store.SaveStore(filePath, password, s); err != nil {
				fmt.Println("Error saving store:", err)
				return
			}

			fmt.Println("Secret updated successfully.")
			return
		}
	}

	if !secretFound {
		fmt.Printf("Error: Secret with name '%s' not found\n", name)
	}
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Define command flags with their shorthand and default values
	// Note: In interactive mode, these are not required

	// Core flags
	updateCmd.Flags().StringP("name", "n", "", "Name of the secret to update (required in non-interactive mode)")
	updateCmd.Flags().StringP("value", "v", "", "New secret value (or use -g to generate)")
	updateCmd.Flags().StringP("category", "c", "-", "New category (use \"-\" to keep existing)")

	// Password generation flags
	updateCmd.Flags().BoolP("generate", "g", false, "Generate a random password")
	updateCmd.Flags().IntP("length", "l", 16, "Length of the generated password (default: 16)")
	updateCmd.Flags().BoolP("symbols", "s", false, "Include symbols in generated password")
	updateCmd.Flags().BoolP("digits", "d", false, "Include digits in generated password")

	// Mode selection
	updateCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode (overrides other flags)")

	// Mark name as required for non-interactive mode
	updateCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		interactive, _ := cmd.Flags().GetBool("interactive")
		if !interactive {
			return cmd.MarkFlagRequired("name")
		}
		return nil
	}
}
