package tui

import (
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/tui/components"
)

// workspaceItem is an alias for the extracted component type.
type workspaceItem = components.WorkspaceItem

// workspaceSummary is an alias for the extracted component type.
type workspaceSummary = components.WorkspaceSummary

// workspaceDelegate is an alias for the extracted component type.
type workspaceDelegate = components.WorkspaceDelegate

// newWorkspaceDelegate creates a new workspace delegate using the components package.
func newWorkspaceDelegate(staleThreshold int) workspaceDelegate {
	return components.NewWorkspaceDelegate(staleThreshold)
}

// summarizeStatus creates a summary from workspace status using the components package.
func summarizeStatus(status *domain.WorkspaceStatus) workspaceSummary {
	return components.SummarizeStatus(status)
}
