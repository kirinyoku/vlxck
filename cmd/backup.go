package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/kirinyoku/vlxck/internal/backup"
	"github.com/spf13/cobra"
)

// backupCmd represents the backup command.
// It creates a timestamped backup of the secret store in the specified directory.
// If no directory is provided, backups will be stored in ~/.vlxck/backups.
var backupCmd = &cobra.Command{
	Use:   "backup [backup-dir]",
	Short: "Create a backup of the secret store",
	Long: `Create a timestamped backup of the secret store in the specified directory.
If no directory is provided, backups will be stored in ~/.vlxck/backups.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		storePath := getStorePath()
		backupDir := filepath.Join(filepath.Dir(storePath), "backups")

		if len(args) > 0 {
			backupDir = args[0]
		}

		backupFile, err := backup.Backup(filepath.Dir(storePath), backupDir)
		if err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}

		checksum, err := backup.GetFileChecksum(backupFile)
		if err != nil {
			return fmt.Errorf("failed to verify backup: %w", err)
		}

		fmt.Printf("✓ Backup created successfully: %s\n", backupFile)
		fmt.Printf("  Checksum: %s\n", checksum)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
