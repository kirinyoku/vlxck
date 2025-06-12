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

// getCmd represents the 'get' command that allows users to retrieve secrets from the store.
// It prompts the user for the secret name and displays the corresponding secret value.
// If the secret is not found, it displays a message indicating that the secret was not found.
//
// The command requires the following flags:
//   - name (-n): The name/identifier of the secret (required)
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a secret by name",
	Run: func(cmd *cobra.Command, args []string) {
		filePath := getStorePath()
		password := utils.PromptForPassword("Enter master password: ")
		s, err := store.LoadStore(filePath, password)
		if err != nil {
			fmt.Println("Error loading store:", err)
			return
		}
		name, _ := cmd.Flags().GetString("name")
		for _, secret := range s.Secrets {
			if secret.Name == name {
				fmt.Printf("Value: %s\n", secret.Value)
				return
			}
		}
		fmt.Println("Secret not found.")
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Define command flags with shorthand and descriptions
	getCmd.Flags().StringP("name", "n", "", "Name of the secret (required)")

	// Mark required flags
	getCmd.MarkFlagRequired("name")
}
