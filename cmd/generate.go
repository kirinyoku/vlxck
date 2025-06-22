// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'generate' command which is used to
// generate random passwords.
package cmd

import (
	"fmt"

	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/spf13/cobra"
)

// generateCmd represents the 'generate' command that allows users to generate random passwords.
// It prompts the user for the password length and whether to include symbols and numbers.
// If the password is successfully generated, it copies it to the clipboard.
//
// The command requires the following flags:
//   - length (-l): The desired length of the password
//   - symbols (-s): Whether to include special characters (!@#$%^&*()-_=+)
//   - digits (-d): Whether to include digits (0123456789)
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a random password",
	Run: func(cmd *cobra.Command, args []string) {
		length, _ := cmd.Flags().GetInt("length")
		symbols, _ := cmd.Flags().GetBool("symbols")
		digits, _ := cmd.Flags().GetBool("digits")
		password, err := utils.GeneratePassword(length, symbols, digits)
		if err != nil {
			fmt.Println("Error generating password:", err)
			return
		}
		err = utils.CopyToClipboard(password)
		if err != nil {
			fmt.Printf("Generated password: %s (clipboard error: %v)\n", password, err)
			return
		}
		fmt.Println("Password generated and copied to clipboard.")
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Define command flags with shorthand and descriptions
	generateCmd.Flags().IntP("length", "l", 16, "Length of the password")
	generateCmd.Flags().BoolP("symbols", "s", false, "Include symbols")
	generateCmd.Flags().BoolP("digits", "d", false, "Include digits")
}
