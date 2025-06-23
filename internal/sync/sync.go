// Package sync implements the synchronization functionality
// for the secrets store with Google Drive.
package sync

import (
	"context"
	"fmt"
)

// SyncManager represents the sync manager
type SyncManager struct {
	storePath      string
	masterPassword string
	gdrive         *GoogleDriveSync
	ctx            context.Context
}

// NewSyncManager creates a new sync manager
//
// Parameters:
//   - ctx: The context
//   - storePath: The path to the secrets store
//   - masterPassword: The master password
//
// Returns:
//   - A new sync manager
//   - An error if the sync manager cannot be created
func NewSyncManager(ctx context.Context, storePath, masterPassword string) (*SyncManager, error) {
	gdrive, err := NewGoogleDriveSync(ctx, masterPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Google Drive sync: %v", err)
	}

	return &SyncManager{
		storePath:      storePath,
		masterPassword: masterPassword,
		gdrive:         gdrive,
		ctx:            ctx,
	}, nil
}

// Sync synchronizes the secrets store with Google Drive
//
// Parameters:
//   - mode: The sync mode, either "push" or "pull"
//
// Returns:
//   - An error if the sync fails
func (s *SyncManager) Sync(mode string) error {
	switch mode {
	case "push":
		return s.gdrive.Push(s.storePath)
	case "pull":
		return s.gdrive.Pull(s.storePath)
	default:
		return fmt.Errorf("invalid sync mode: %s", mode)
	}
}
