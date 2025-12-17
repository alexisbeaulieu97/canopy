// Package ports defines interfaces for external dependencies (hexagonal architecture).
package ports

import (
	"context"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// WorkspaceStorage defines the interface for workspace persistence operations.
// All methods accept context.Context as the first parameter for cancellation and timeout support.
// Methods are ID-based to abstract away filesystem implementation details.
type WorkspaceStorage interface {
	// Create creates a new workspace from the provided domain object.
	Create(ctx context.Context, ws domain.Workspace) error

	// Save persists changes to an existing workspace.
	Save(ctx context.Context, ws domain.Workspace) error

	// Close archives a workspace and returns the closed entry.
	Close(ctx context.Context, id string, closedAt time.Time) (*domain.ClosedWorkspace, error)

	// List returns all active workspaces.
	List(ctx context.Context) ([]domain.Workspace, error)

	// ListClosed returns archived workspaces, sorted by newest first.
	ListClosed(ctx context.Context) ([]domain.ClosedWorkspace, error)

	// Load retrieves a workspace by ID.
	Load(ctx context.Context, id string) (*domain.Workspace, error)

	// Delete removes a workspace by ID.
	Delete(ctx context.Context, id string) error

	// Rename changes a workspace's ID.
	Rename(ctx context.Context, oldID, newID string) error

	// LatestClosed returns the most recent closed entry for a workspace.
	LatestClosed(ctx context.Context, id string) (*domain.ClosedWorkspace, error)

	// DeleteClosed removes a closed workspace entry identified by workspace ID and close timestamp.
	DeleteClosed(ctx context.Context, id string, closedAt time.Time) error
}
