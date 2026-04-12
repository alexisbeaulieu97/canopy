package storage

import (
	"github.com/alexisbeaulieu97/canopy/internal/ports"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
)

// Compile-time check that Engine implements ports.WorkspaceStorage.
var _ ports.WorkspaceStorage = (*Engine)(nil)

// Engine manages workspaces.
type Engine struct {
	WorkspacesRoot string
	ClosedRoot     string
	dirNameForID   func(string) (string, error)
}

// New creates a new Workspace Engine.
func New(workspacesRoot, closedRoot string) *Engine {
	return NewWithNaming(workspacesRoot, closedRoot, nil)
}

// NewWithNaming creates a new Workspace Engine with a custom naming resolver.
func NewWithNaming(workspacesRoot, closedRoot string, dirNameForID func(string) (string, error)) *Engine {
	if dirNameForID == nil {
		dirNameForID = validation.NormalizeWorkspaceDirName
	}

	return &Engine{
		WorkspacesRoot: workspacesRoot,
		ClosedRoot:     closedRoot,
		dirNameForID:   dirNameForID,
	}
}
