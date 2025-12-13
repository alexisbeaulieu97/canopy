// Package ports defines interfaces for external dependencies (hexagonal architecture).
package ports

import (
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// WorkspaceStorage defines the interface for workspace persistence operations.
type WorkspaceStorage interface {
	// Create creates a new workspace directory and metadata.
	Create(dirName, id, branchName string, repos []domain.Repo) error

	// Save updates the metadata for an existing workspace directory.
	Save(dirName string, ws domain.Workspace) error

	// Close copies workspace metadata into the closed root and returns the closed entry.
	Close(dirName string, ws domain.Workspace, closedAt time.Time) (*domain.ClosedWorkspace, error)

	// List returns all active workspaces.
	List() (map[string]domain.Workspace, error)

	// ListClosed returns closed workspaces stored on disk, sorted by newest first.
	ListClosed() ([]domain.ClosedWorkspace, error)

	// Load reads the metadata for a specific workspace.
	Load(dirName string) (*domain.Workspace, error)

	// LoadByID looks up a workspace by its ID and returns the workspace metadata
	// and directory name. It attempts direct path access first (assuming ID == dirName),
	// then falls back to scanning all workspaces if the direct lookup fails.
	LoadByID(id string) (*domain.Workspace, string, error)

	// Delete removes a workspace.
	Delete(workspaceID string) error

	// Rename renames a workspace directory and updates its metadata.
	// oldDirName is the current directory name, newDirName is the target directory name,
	// and newID is the new workspace ID to be stored in metadata.
	Rename(oldDirName, newDirName, newID string) error

	// LatestClosed returns the newest closed entry for the given workspace ID.
	LatestClosed(workspaceID string) (*domain.ClosedWorkspace, error)

	// DeleteClosed removes a closed workspace entry.
	DeleteClosed(path string) error
}
