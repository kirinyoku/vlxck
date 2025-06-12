// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'change-master' command which is used to
// change the master password for the encrypted store.
package cmd

import (
	"fmt"

	"github.com/kirinyoku/vlxck/internal/store"
	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/spf13/cobra"
)

// changeMasterCmd represents the 'change-master' command that allows users to change the master password for the encrypted store.
// It prompts the user for the current master password and the new master password.
// If the new master password is successfully changed, it displays a message indicating that the master password was changed successfully.
//
// The command requires the following flags:
//   - filePath (-f): The path to the encrypted store file
var changeMasterCmd = &cobra.Command{
	Use:   "change-master",
	Short: "Change the master password",
	Run: func(cmd *cobra.Command, args []string) {
		filePath := getStorePath()
		oldPassword := utils.PromptForPassword("Enter current master password: ")
		s, err := store.LoadStore(filePath, oldPassword)
		if err != nil {
			fmt.Println("Error loading store:", err)
			return
		}
		newPassword := utils.PromptForPassword("Enter new master password: ")
		confirmPassword := utils.PromptForPassword("Confirm new master password: ")
		if newPassword != confirmPassword {
			fmt.Println("Passwords do not match. Exiting.")
			return
		}
		if err := store.SaveStore(filePath, newPassword, s); err != nil {
			fmt.Println("Error saving store:", err)
			return
		}
		fmt.Println("Master password changed successfully.")
	},
}

func init() {
	rootCmd.AddCommand(changeMasterCmd)
}
