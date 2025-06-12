// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'add' command which is used to
// add new secrets to the encrypted store.
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/kirinyoku/vlxck/internal/store"
	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/spf13/cobra"
)

// addCmd represents the 'add' command that allows users to add new secrets to the store.
// It handles both the creation of a new store (if none exists) and the addition of
// new secrets to an existing store.
//
// The command requires the following flags:
//   - name (-n): The name/identifier of the secret (required)
//   - value (-v): The secret value to store (required)
//   - category (-c): Optional category for organizing secrets
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new secret to the secure store",
	Run: func(cmd *cobra.Command, args []string) {
		filePath := getStorePath()
		password := utils.PromptForPassword("Enter master password: ")
		s, err := store.LoadStore(filePath, password)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Store not initialized. Setting up new store.")
				newPassword := utils.PromptForPassword("Set master password: ")
				confirmPassword := utils.PromptForPassword("Confirm master password: ")
				if newPassword != confirmPassword {
					fmt.Println("Passwords do not match. Exiting.")
					return
				}
				if err := store.InitializeStore(filePath, newPassword); err != nil {
					fmt.Println("Error initializing store:", err)
					return
				}
				s, err = store.LoadStore(filePath, newPassword)
				if err != nil {
					fmt.Println("Error loading store:", err)
					return
				}
				password = newPassword
			} else {
				fmt.Println("Error loading store:", err)
				return
			}
		}
		name, _ := cmd.Flags().GetString("name")
		value, _ := cmd.Flags().GetString("value")
		category, _ := cmd.Flags().GetString("category")
		newSecret := store.Secret{
			Name:      name,
			Value:     value,
			Category:  category,
			CreatedAt: time.Now(),
		}
		s.Secrets = append(s.Secrets, newSecret)
		if err := store.SaveStore(filePath, password, s); err != nil {
			fmt.Println("Error saving store:", err)
			return
		}
		fmt.Println("Secret added successfully.")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Define command flags with shorthand and descriptions
	addCmd.Flags().StringP("name", "n", "", "Name/identifier of the secret (required)")
	addCmd.Flags().StringP("value", "v", "", "Secret value to store (required)")
	addCmd.Flags().StringP("category", "c", "", "Optional category for organizing secrets")

	// Mark required flags
	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("value")
}
