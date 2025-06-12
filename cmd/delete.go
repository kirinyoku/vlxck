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

// deleteCmd represents the 'delete' command that allows users to delete secrets from the store.
// It prompts the user for the secret name and deletes the corresponding secret from the store.
// If the secret is not found, it displays a message indicating that the secret was not found.
//
// The command requires the following flags:
//   - name (-n): The name/identifier of the secret (required)
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a secret by name",
	Run: func(cmd *cobra.Command, args []string) {
		filePath := getStorePath()
		password := utils.PromptForPassword("Enter master password: ")
		s, err := store.LoadStore(filePath, password)
		if err != nil {
			fmt.Println("Error loading store:", err)
			return
		}
		name, _ := cmd.Flags().GetString("name")
		for i, secret := range s.Secrets {
			if secret.Name == name {
				s.Secrets = append(s.Secrets[:i], s.Secrets[i+1:]...)
				if err := store.SaveStore(filePath, password, s); err != nil {
					fmt.Println("Error saving store:", err)
					return
				}
				fmt.Println("Secret deleted successfully.")
				return
			}
		}
		fmt.Println("Secret not found.")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Define command flags with shorthand and descriptions
	deleteCmd.Flags().StringP("name", "n", "", "Name of the secret (required)")

	// Mark required flags
	deleteCmd.MarkFlagRequired("name")
}
