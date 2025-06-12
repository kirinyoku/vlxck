// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'list' command which is used to
// list all secrets in the encrypted store.
package cmd

import (
	"fmt"

	"github.com/kirinyoku/vlxck/internal/store"
	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/spf13/cobra"
)

// listCmd represents the 'list' command that allows users to list all secrets in the store.
// It prompts the user for the master password and displays the list of secrets.
// If the store is not found, it displays a message indicating that the store was not found.
//
// The command requires the following flags:
//   - category (-c): Optional category for filtering secrets
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets",
	Run: func(cmd *cobra.Command, args []string) {
		filePath := getStorePath()
		password := utils.PromptForPassword("Enter master password: ")
		s, err := store.LoadStore(filePath, password)
		if err != nil {
			fmt.Println("Error loading store:", err)
			return
		}
		category, _ := cmd.Flags().GetString("category")
		for _, secret := range s.Secrets {
			if category == "" || secret.Category == category {
				fmt.Printf("Name: %s, Category: %s\n", secret.Name, secret.Category)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Define command flags with shorthand and descriptions
	listCmd.Flags().StringP("category", "c", "", "Filter by category")
}
