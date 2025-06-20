package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/kirinyoku/vlxck/internal/backup"
	"github.com/spf13/cobra"
)

// listBackupsCmd represents the list-backups command.
// It lists all available backups in the specified directory or the default backup location.
var listBackupsCmd = &cobra.Command{
	Use:   "list-backups [backup-dir]",
	Short: "List all available backups",
	Long:  `List all available backups in the specified directory or the default backup location.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupDir := filepath.Join(filepath.Dir(getStorePath()), "backups")

		if len(args) > 0 {
			backupDir = args[0]
		}

		backups, err := backup.ListBackups(backupDir)
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		if len(backups) == 0 {
			fmt.Printf("No backups found in: %s\n", backupDir)
			return nil
		}

		fmt.Printf("Available backups in %s:\n", backupDir)
		for i, b := range backups {
			sizeKB := float64(b.Size) / 1024.0
			modTime := b.ModTime.Format("2006-01-02 15:04:05")
			fmt.Printf("%d. %-35s  %8.1f KB  %s\n",
				i+1,
				b.Name,
				sizeKB,
				modTime)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listBackupsCmd)
}
