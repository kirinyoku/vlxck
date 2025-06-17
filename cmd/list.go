// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'list' command which is used to
// list all secrets in the encrypted store.
package cmd

import (
	"fmt"

	"github.com/kirinyoku/vlxck/internal/store"
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
		password, err := getPassword(false)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		s, err := store.LoadStore(filePath, password)
		if err == nil {
			// Only cache the password if it was successfully used
			cacheVerifiedPassword(password)
		}
		if err != nil {
			fmt.Println("Error loading store:", err)
			return
		}
		category, _ := cmd.Flags().GetString("category")
		for i, secret := range s.Secrets {
			if category == "" || secret.Category == category {
				fmt.Printf("%d. Name: %s, Category: %s\n", i+1, secret.Name, secret.Category)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Define command flags with shorthand and descriptions
	listCmd.Flags().StringP("category", "c", "", "Filter by category")
}
