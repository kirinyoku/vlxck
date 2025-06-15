// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'export' command which is used to
// export the store to a specified directory.
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// exportCmd represents the 'export' command that allows users to export the store to a specified directory.
// It prompts the user for the directory to export the store file to and exports the store file to the specified directory.
// If the store file does not exist, it displays a message indicating that the store file does not exist.
// If the directory to export the store file to does not exist, it creates the directory.
// If the store file is successfully exported, it displays a message indicating that the store file was exported successfully.
//
// The command requires the following flags:
//   - dir (-d): The directory to export the store file to (required)
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the store to a specified directory",
	Run: func(cmd *cobra.Command, args []string) {
		storePath := getStorePath()
		dir, _ := cmd.Flags().GetString("dir")

		if _, err := os.Stat(storePath); os.IsNotExist(err) {
			fmt.Println("Error: store file does not exist")
			return
		}

		if err := os.MkdirAll(dir, 0700); err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}

		targetPath := filepath.Join(dir, "store.dat")

		sourceFile, err := os.Open(storePath)
		if err != nil {
			fmt.Println("Error opening store file:", err)
			return
		}
		defer sourceFile.Close()

		targetFile, err := os.Create(targetPath)
		if err != nil {
			fmt.Println("Error creating export file:", err)
			return
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, sourceFile); err != nil {
			fmt.Println("Error copying file:", err)
			return
		}

		if err := os.Chmod(targetPath, 0600); err != nil {
			fmt.Println("Error setting file permissions:", err)
			return
		}

		fmt.Printf("Store exported successfully to %s\n", targetPath)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Directory to export store file to
	exportCmd.Flags().StringP("dir", "d", "", "Directory to export store file to (required)")

	// Mark required flags
	exportCmd.MarkFlagRequired("dir")
}
