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

// updateCmd represents the 'update' command that allows users to update existing secrets in the store.
// It prompts the user for the secret name and updates the corresponding secret value.
// If the secret is not found, it displays a message indicating that the secret was not found.
//
// The command requires the following flags:
//   - name (-n): The name/identifier of the secret (required)
//   - value (-v): The secret value to store
//   - category (-c): Optional category for organizing secrets
//   - generate (-g): Generate a random password for the secret
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing secret in the secure store",
	Run: func(cmd *cobra.Command, args []string) {
		filePath := getStorePath()
		password := utils.PromptForPassword("Enter master password: ")
		s, err := store.LoadStore(filePath, password)
		if err != nil {
			fmt.Println("Error loading store:", err)
			return
		}
		name, _ := cmd.Flags().GetString("name")
		value, _ := cmd.Flags().GetString("value")
		category, _ := cmd.Flags().GetString("category")
		generate, _ := cmd.Flags().GetBool("generate")

		if value == "" && category == "-" && !generate {
			fmt.Println("Error: either --value, --category or --generate must be specified")
			return
		}

		if value != "" && generate {
			fmt.Println("Warning: --value takes precedence over --generate")
			generate = false
		}

		for i, secret := range s.Secrets {
			if secret.Name == name {
				if value != "" {
					s.Secrets[i].Value = value
				}
				if category != "" {
					s.Secrets[i].Category = category
				}
				if generate {
					s.Secrets[i].Value, err = utils.GeneratePassword(16, true, true)
					if err != nil {
						fmt.Println("Error generating password:", err)
						return
					}
				}
				if err := store.SaveStore(filePath, password, s); err != nil {
					fmt.Println("Error saving store:", err)
					return
				}
				fmt.Println("Secret updated successfully.")
				return
			}
		}
		fmt.Println("Secret not found.")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Define command flags with shorthand and descriptions
	updateCmd.Flags().StringP("name", "n", "", "The name/identifier of the secret (required)")
	updateCmd.Flags().StringP("value", "v", "", "Secret value to store (required)")
	updateCmd.Flags().StringP("category", "c", "-", "Optional category for organizing secrets")
	updateCmd.Flags().BoolP("generate", "g", false, "Generate a random password for the secret")

	// Mark required flags
	updateCmd.MarkFlagRequired("name")
}
