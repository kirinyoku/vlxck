// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'import' command which is used to
// import secrets by replacing or merging with an encrypted file.
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kirinyoku/vlxck/internal/store"
	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/spf13/cobra"
)

// importCmd represents the 'import' command that allows users to import secrets by replacing or merging with an encrypted file.
// It prompts the user for the path to the import file and the master password for the import file.
// If the import file is successfully imported, it displays a message indicating that the store was successfully replaced with the import file.
//
// The command requires the following flags:
//   - file (-f): The path to the import file (required)
//   - use-store-password (-p): Whether to use the store's master password for import
//   - merge (-m): Whether to merge secrets from import file into existing store
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import secrets by replacing or merging with an encrypted file",
	Run: func(cmd *cobra.Command, args []string) {
		filePath := getStorePath()
		importPath, _ := cmd.Flags().GetString("file")
		useStorePassword, _ := cmd.Flags().GetBool("use-store-password")
		merge, _ := cmd.Flags().GetBool("merge")
		var importPassword string

		if useStorePassword {
			importPassword = utils.PromptForPassword("Enter master password: ")
		} else {
			importPassword = utils.PromptForPassword("Enter import file password: ")
		}

		importedStore, err := store.LoadStore(importPath, importPassword)
		if err != nil {
			fmt.Println("Error validating import file:", err)
			return
		}

		if merge {
			storePassword := importPassword
			if !useStorePassword {
				storePassword = utils.PromptForPassword("Enter master password for store: ")
			}
			currentStore, err := store.LoadStore(filePath, storePassword)
			if err != nil {
				if os.IsNotExist(err) {
					if err := store.InitializeStore(filePath, storePassword); err != nil {
						fmt.Println("Error initializing store:", err)
						return
					}
					currentStore, err = store.LoadStore(filePath, storePassword)
					if err != nil {
						fmt.Println("Error loading store:", err)
						return
					}
				} else {
					fmt.Println("Error loading store:", err)
					return
				}
			}

			var mergedSecrets []store.Secret
			mergedSecrets = append(mergedSecrets, currentStore.Secrets...)
			importedCount := 0
			skippedCount := 0
			overwrittenCount := 0

			for _, importedSecret := range importedStore.Secrets {
				conflict := false
				for _, currentSecret := range currentStore.Secrets {
					if importedSecret.Name == currentSecret.Name {
						choice := utils.PromptForConflictChoice(currentSecret, importedSecret)
						if choice == "s" {
							skippedCount++
							conflict = true
							break
						} else if choice == "i" {
							for i, secret := range mergedSecrets {
								if secret.Name == importedSecret.Name {
									mergedSecrets = append(mergedSecrets[:i], mergedSecrets[i+1:]...)
									break
								}
							}
							mergedSecrets = append(mergedSecrets, importedSecret)
							importedCount++
							overwrittenCount++
							conflict = true
							break
						} else if choice == "l" {
							skippedCount++
							conflict = true
							break
						}
					}
				}
				if !conflict {
					mergedSecrets = append(mergedSecrets, importedSecret)
					importedCount++
				}
			}

			currentStore.Secrets = mergedSecrets
			if err := store.SaveStore(filePath, storePassword, currentStore); err != nil {
				fmt.Println("Error saving store:", err)
				return
			}

			fmt.Printf("Successfully merged %d secrets (%d overwritten, %d skipped) from %s\n", importedCount, overwrittenCount, skippedCount, importPath)
		} else {
			if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
				fmt.Println("Error creating store directory:", err)
				return
			}

			sourceFile, err := os.Open(importPath)
			if err != nil {
				fmt.Println("Error opening import file:", err)
				return
			}
			defer sourceFile.Close()

			targetFile, err := os.Create(filePath)
			if err != nil {
				fmt.Println("Error creating store file:", err)
				return
			}
			defer targetFile.Close()

			if _, err := io.Copy(targetFile, sourceFile); err != nil {
				fmt.Println("Error copying file:", err)
				return
			}

			if err := os.Chmod(filePath, 0600); err != nil {
				fmt.Println("Error setting file permissions:", err)
				return
			}

			fmt.Printf("Store successfully replaced with %s\n", importPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Define command flags with shorthand and descriptions
	importCmd.Flags().StringP("file", "f", "", "Path to import file (required)")
	importCmd.Flags().BoolP("use-store-password", "p", false, "Use the store's master password for import")
	importCmd.Flags().BoolP("merge", "m", false, "Merge secrets from import file into existing store")

	// Mark required flags
	importCmd.MarkFlagRequired("file")
}
