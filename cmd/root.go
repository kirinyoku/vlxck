// Package cmd implements the command-line interface for the secure command-line password manager.
// This file contains the implementation of the root command which is used to execute the application.
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const Version = "0.5.1" // Version of the application

// getStorePath returns the path to the encrypted store file.
func getStorePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".vlxck", "store.dat")
}

// rootCmd is the root command for the application.
var rootCmd = &cobra.Command{
	Use:   "vlxck",
	Short: "A secure command-line password manager for storing and managing sensitive data",
	Long: `vlxck is a secure, lightweight password manager that helps you
store and manage your sensitive information with strong encryption.

Getting Started:
  1. Add your first secret: 'vlxck add -n {service name} -v {password}'
  2. Retrieve a secret: 'vlxck get -n {service name}'
  3. List all secrets: 'vlxck list'
  4. Generate a strong password: 'vlxck generate -l 16 -s -n'

Security:
  • All data is encrypted before being written to disk
  • Master password is never stored
  • Uses industry-standard encryption (AES-256-GCM with Argon2id key derivation)

For more information about a specific command, use 'vlxck [command] --help'
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = Version
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
