package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kirinyoku/vlxck/internal/backup"
	"github.com/kirinyoku/vlxck/internal/utils"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command.
// It restores a backup of the password store to the specified directory.
// If no target directory is provided, restores to the default store location.
var restoreCmd = &cobra.Command{
	Use:   "restore [backup-file] [target-dir]",
	Short: "Restore a backup of the password store",
	Long: `Restore a previously created backup to the specified directory.
If no backup file is provided, you'll be prompted to select one interactively.
If no target directory is provided, restores to the default store location.`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		interactive, err := cmd.Flags().GetBool("interactive")
		if err != nil {
			return fmt.Errorf("failed to get interactive flag: %w", err)
		}

		var backupFile string

		if len(args) == 0 || interactive {
			backupDir := filepath.Join(filepath.Dir(getStorePath()), "backups")
			backups, err := backup.ListBackups(backupDir)
			if err != nil {
				return fmt.Errorf("failed to list backups: %w", err)
			}

			if len(backups) == 0 {
				return fmt.Errorf("no backups found in %s", backupDir)
			}

			var options []string
			for _, b := range backups {
				sizeKB := float64(b.Size) / 1024.0
				option := fmt.Sprintf("%-35s  %8.1f KB  %s",
					b.Name, sizeKB, b.ModTime.Format("2006-01-02 15:04:05"))
				options = append(options, option)
			}

			prompt := promptui.Select{
				Label: "Select backup to restore",
				Items: options,
				Size:  10,
			}

			selected, _, err := prompt.Run()
			if err != nil {
				return fmt.Errorf("backup selection cancelled: %w", err)
			}

			backupFile = backups[selected].Path
		} else {
			backupFile, err = filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("invalid backup file path: %w", err)
			}
		}

		storePath := getStorePath()
		storeDir := filepath.Dir(storePath)
		targetDir := storeDir

		if len(args) > 1 {
			targetDir, err = filepath.Abs(args[1])
			if err != nil {
				return fmt.Errorf("invalid target directory: %w", err)
			}
		}

		fileInfo, err := os.Stat(backupFile)
		if err != nil {
			return fmt.Errorf("backup file not found or not readable: %w", err)
		}

		if fileInfo.Size() == 0 {
			return fmt.Errorf("backup file is empty")
		}

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create target directory: %w", err)
		}

		fmt.Printf("Backup file: %s (%d bytes)\n", backupFile, fileInfo.Size())
		fmt.Printf("Restore to: %s\n", targetDir)
		fmt.Println("WARNING: This will overwrite any existing files in the target directory!")

		confirm, err := utils.PromptForConfirm("Are you sure you want to continue? (y/N): ")
		if err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}

		if !confirm {
			return fmt.Errorf("restore cancelled")
		}

		if err := backup.Restore(backupFile, targetDir); err != nil {
			return fmt.Errorf("restore failed: %w", err)
		}

		fmt.Printf("✓ Backup restored successfully to: %s\n", targetDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	// Define command flags with shorthand and descriptions
	restoreCmd.Flags().BoolP("interactive", "i", false, "Interactive mode to select from available backups")
}
