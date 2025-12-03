// Package tui provides the terminal UI for canopy.
package tui

import "github.com/alexisbeaulieu97/canopy/internal/domain"

// workspaceListMsg is sent when the list of workspaces is loaded.
type workspaceListMsg struct {
	items      []workspaceItem
	totalUsage int64
}

// workspaceStatusMsg is sent when status for a single workspace is loaded.
type workspaceStatusMsg struct {
	id     string
	status *domain.WorkspaceStatus
}

// workspaceStatusErrMsg is sent when status loading fails.
type workspaceStatusErrMsg struct {
	id  string
	err error
}

// pushResultMsg is sent when a push operation completes.
type pushResultMsg struct {
	id  string
	err error
}

// openEditorResultMsg is sent when opening an editor completes.
type openEditorResultMsg struct {
	err error
}

// workspaceDetailsMsg is sent when detailed info for a workspace is loaded.
type workspaceDetailsMsg struct {
	workspace *domain.Workspace
	status    *domain.WorkspaceStatus
}
