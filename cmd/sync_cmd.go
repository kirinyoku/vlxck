// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'sync' command which is used to
// synchronize the secret store with Google Drive.
package cmd

import (
	"context"
	"fmt"

	"github.com/kirinyoku/vlxck/internal/sync"
	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
// It synchronizes the secret store with Google Drive.
// The command requires the following flags:
//   - mode (-m): The sync mode, either 'push' or 'pull'
//   - init: Initialize Google Drive sync
var syncCmd = &cobra.Command{
	Use: "sync",
	Short: `Synchronize the secret store with Google Drive

Examples:
  # Initialize Google Drive sync
  vlxck sync --init

  # Push changes to Google Drive
  vlxck sync -m push

  # Pull changes from Google Drive
  vlxck sync -m pull`,
	Run: func(cmd *cobra.Command, args []string) {
		mode, _ := cmd.Flags().GetString("mode")
		initialize, _ := cmd.Flags().GetBool("init")

		storePath := getStorePath()
		masterPassword := utils.PromptForPassword("Enter master password: ")

		ctx := context.Background()

		if initialize {
			cfg, err := sync.InitGoogleDrive(ctx, masterPassword)
			if err != nil {
				fmt.Printf("Error initializing sync: %v\n", err)
				return
			}
			fmt.Printf("Sync initialized successfully. File ID: %s\n", cfg.Sync.FileID)
			return
		}

		manager, err := sync.NewSyncManager(ctx, storePath, masterPassword)
		if err != nil {
			fmt.Printf("Error creating sync manager: %v\n", err)
			return
		}

		if err := manager.Sync(mode); err != nil {
			fmt.Printf("Error syncing: %v\n", err)
		} else {
			fmt.Println("Sync completed successfully")
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Define command flags with shorthand and descriptions
	syncCmd.Flags().StringP("mode", "m", "", "Sync mode: push, pull")
	syncCmd.Flags().Bool("init", false, "Initialize Google Drive sync")
}
