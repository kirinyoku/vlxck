// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'add' command which is used to
// add new secrets to the encrypted store with options for both interactive and
// non-interactive modes.
package cmd

import (
	"fmt"
	"os"

	"github.com/kirinyoku/vlxck/internal/store"
	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/spf13/cobra"
)

// addCmd represents the 'add' command that allows users to add new secrets to the store.
// It supports two modes of operation:
// 1. Interactive mode (-i): Guides user through prompts
// 2. Non-interactive mode: Uses command-line flags for automation
//
// The command supports the following flags:
//   - name (-n): The name/identifier of the secret (required in non-interactive mode)
//   - value (-v): The secret value to store (or use -g to generate)
//   - category (-c): Optional category for organizing secrets
//   - generate (-g): Generate a random password (overrides -v)
//   - length (-l): Length of generated password (default: 16)
//   - symbols (-s): Include symbols in generated password
//   - digits (-d): Include digits in generated password
//   - interactive (-i): Use interactive mode (overrides other flags)
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new secret to the secure store",
	Long: `Add a new secret to the secure store.

In interactive mode, you'll be guided through the add process with prompts.
In non-interactive mode, you must specify the secret name and at least one of:
- A new value (--value/-v)
- The --generate flag to create a new password
- A new category (--category/-c)

Examples:
  # Interactive mode
  vlxck add -i

  # Add with specific value and category
  vlxck add -n example.com -v newpassword -c work

  # Generate a 24-char password with symbols and digits
  vlxck add -n example.com -gdsl 24`,

	Run: func(cmd *cobra.Command, args []string) {
		// Get store path and check for interactive mode
		filePath := getStorePath()
		interactive, _ := cmd.Flags().GetBool("interactive")

		// Get master password
		password, err := getPassword(false)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Load or initialize the store
		s, err := store.LoadStore(filePath, password)
		if err == nil {
			// Only cache the password if it was successfully used
			cacheVerifiedPassword(password)
		}
		if err != nil {
			if os.IsNotExist(err) {
				if err := store.InitializeStore(filePath, password); err != nil {
					fmt.Println("Error initializing store:", err)
					return
				}
				s, err = store.LoadStore(filePath, password)
				if err != nil {
					fmt.Println("Error loading store:", err)
					return
				}
			} else {
				fmt.Println("Error loading store:", err)
				return
			}
		}

		// Route to appropriate handler
		if interactive {
			addInteractive(s, filePath, password)
		} else {
			addNonInteractive(cmd, s, filePath, password)
		}
	},
}

// addInteractive handles the interactive add flow
func addInteractive(s *store.Store, filePath, password string) {
	// Get secret details from user
	name, err := utils.PromptForSecretName(s.Secrets)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	value, err := utils.PromptForSecretValue()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	category, err := utils.PromptForCategory()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Add the new secret
	s.Secrets = append(s.Secrets, store.Secret{
		Name:     name,
		Value:    value,
		Category: category,
	})

	// Save the updated store
	if err := store.SaveStore(filePath, password, s); err != nil {
		fmt.Println("Error saving store:", err)
		return
	}

	fmt.Printf("Secret '%s' added successfully\n", name)
}

// addNonInteractive handles the non-interactive add flow using command-line flags
func addNonInteractive(cmd *cobra.Command, s *store.Store, filePath, password string) {
	// Parse command-line flags
	name, _ := cmd.Flags().GetString("name")
	value, _ := cmd.Flags().GetString("value")
	generate, _ := cmd.Flags().GetBool("generate")
	length, _ := cmd.Flags().GetInt("length")
	symbols, _ := cmd.Flags().GetBool("symbols")
	digits, _ := cmd.Flags().GetBool("digits")
	category, _ := cmd.Flags().GetString("category")

	// Validate required parameters
	if name == "" {
		fmt.Println("Error: secret name is required")
		return
	}

	if value == "" && !generate {
		fmt.Println("Error: secret value or --generate flag is required")
		return
	}

	// Check for existing secret with the same name
	for _, secret := range s.Secrets {
		if secret.Name == name {
			fmt.Printf("Error: secret with name '%s' already exists\n", name)
			return
		}
	}

	// Generate or use provided value
	var secretValue string
	if generate {
		var err error
		secretValue, err = utils.GeneratePassword(length, symbols, digits)
		if err != nil {
			fmt.Println("Error generating password:", err)
			return
		}
		// Copy generated password to clipboard
		if err := utils.CopyToClipboard(secretValue); err != nil {
			fmt.Println("Warning: Could not copy to clipboard:", err)
		} else {
			fmt.Println("Generated password copied to clipboard.")
		}
	} else {
		secretValue = value
	}

	// Add the new secret
	s.Secrets = append(s.Secrets, store.Secret{
		Name:     name,
		Value:    secretValue,
		Category: category,
	})

	// Save the updated store
	if err := store.SaveStore(filePath, password, s); err != nil {
		fmt.Println("Error saving store:", err)
		return
	}

	fmt.Printf("Secret '%s' added successfully\n", name)
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Define command flags with their shorthand and default values

	// Core flags
	addCmd.Flags().StringP("name", "n", "", "Name of the secret (required in non-interactive mode)")
	addCmd.Flags().StringP("value", "V", "", "Value of the secret (or use -g to generate)")
	addCmd.Flags().StringP("category", "c", "", "Category for organizing secrets")

	// Password generation flags
	addCmd.Flags().BoolP("generate", "g", false, "Generate a random password")
	addCmd.Flags().IntP("length", "l", 16, "Length of the generated password (default: 16)")
	addCmd.Flags().BoolP("symbols", "s", false, "Include symbols in generated password")
	addCmd.Flags().BoolP("digits", "d", false, "Include digits in generated password")

	// Mode selection
	addCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode (overrides other flags)")
}
