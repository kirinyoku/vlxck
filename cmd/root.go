// Package cmd implements the command-line interface for the secure command-line password manager.
// This file contains the implementation of the root command which is used to execute the application.
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kirinyoku/vlxck/internal/cache"
	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/spf13/cobra"
)

const (
	Version      = "0.8.0"         // Version of the application
	cacheTimeout = 5 * time.Minute // Fixed 5-minute cache timeout
)

// getStorePath returns the path to the encrypted store file.
func getStorePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".vlxck", "store.dat")
}

// getPassword retrieves the password from cache or prompts the user
// The password will be cached for 5 minutes after successful verification
// by the caller using cacheVerifiedPassword.
func getPassword(cacheOnSuccess bool) (string, error) {
	// Try to get password from cache
	if !cacheOnSuccess {
		password, err := cache.GetMasterPassword()
		if err != nil {
			// Log but don't fail - we'll prompt for password
			fmt.Fprintf(os.Stderr, "Warning: Failed to read password cache: %v\n", err)
		} else if password != "" {
			return password, nil
		}
	}

	// Not in cache or cache disabled, prompt user
	password := utils.PromptForPassword("Enter master password: ")
	return password, nil
}

// cacheVerifiedPassword caches the password for 5 minutes
func cacheVerifiedPassword(password string) {
	if err := cache.SetMasterPassword(password, cacheTimeout); err != nil {
		// Log but don't fail - the password was still obtained
		fmt.Fprintf(os.Stderr, "Warning: Failed to cache password: %v\n", err)
	}
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
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print the version number")
	rootCmd.PersistentFlags().BoolP("toggle", "t", false, "Help message for toggle")

	// Clear the cache on application exit
	// This ensures we don't leave sensitive data in the cache if the program crashes
	go func() {
		// This will run when the program exits
		// We use a channel to wait for interrupt signals
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cache.ClearMasterPassword()
		os.Exit(0)
	}()
}
