package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// loadWorkspaces creates a command to load all workspaces.
func (m Model) loadWorkspaces() tea.Msg {
	workspaces, err := m.svc.ListWorkspaces()
	if err != nil {
		return loadWorkspacesErrMsg{err: err}
	}

	// Detect orphans for all workspaces
	allOrphans, orphanErr := m.svc.DetectOrphans()

	// Count orphans per workspace
	orphanCounts := make(map[string]int)

	if orphanErr == nil {
		for _, orphan := range allOrphans {
			orphanCounts[orphan.WorkspaceID]++
		}
	}

	items := make([]workspaceItem, 0, len(workspaces))

	var totalUsage int64

	for _, w := range workspaces {
		items = append(items, workspaceItem{
			workspace: w,
			summary: workspaceSummary{
				repoCount: len(w.Repos),
			},
			orphanCount: orphanCounts[w.ID],
		})
		totalUsage += w.DiskUsageBytes
	}

	return workspaceListMsg{
		items:      items,
		totalUsage: totalUsage,
	}
}

// loadWorkspaceStatus creates a command to load status for a specific workspace.
func (m Model) loadWorkspaceStatus(id string) tea.Cmd {
	return func() tea.Msg {
		status, err := m.svc.GetStatus(id)
		if err != nil {
			return workspaceStatusErrMsg{id: id, err: err}
		}

		return workspaceStatusMsg{id: id, status: status}
	}
}

// loadWorkspaceDetails creates a command to load detailed info for a workspace.
func (m Model) loadWorkspaceDetails(id string) tea.Cmd {
	return func() tea.Msg {
		wsItem, ok := m.workspaceItemByID(id)
		if !ok {
			return workspaceDetailsErrMsg{id: id, err: fmt.Errorf("workspace not found")}
		}

		status, err := m.svc.GetStatus(id)
		if err != nil {
			return workspaceDetailsErrMsg{id: id, err: err}
		}

		// Get orphans for this workspace
		orphans, _ := m.svc.DetectOrphansForWorkspace(id)

		wsCopy := wsItem.workspace

		return workspaceDetailsMsg{workspace: &wsCopy, status: status, orphans: orphans}
	}
}

// pushWorkspace creates a command to push all changes in a workspace.
func (m Model) pushWorkspace(id string) tea.Cmd {
	return func() tea.Msg {
		return pushResultMsg{
			id:  id,
			err: m.svc.PushWorkspace(id),
		}
	}
}

// closeWorkspace creates a command to close/delete a workspace.
func (m Model) closeWorkspace(id string) tea.Cmd {
	return func() tea.Msg {
		err := m.svc.CloseWorkspace(id, false)
		if err != nil {
			return closeWorkspaceErrMsg{id: id, err: err}
		}
		// Reload list
		return m.loadWorkspaces()
	}
}

// openWorkspace creates a command to open a workspace in an editor.
func (m Model) openWorkspace(id string) tea.Cmd {
	return func() tea.Msg {
		path, err := m.svc.WorkspacePath(id)
		if err != nil {
			return openEditorResultMsg{err: err}
		}

		editor := os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}

		if editor == "" {
			return openEditorResultMsg{err: fmt.Errorf("set $EDITOR or $VISUAL to open workspaces")}
		}

		parts := strings.Fields(editor)
		if len(parts) == 0 {
			return openEditorResultMsg{err: fmt.Errorf("set $EDITOR or $VISUAL to open workspaces")}
		}

		cmd := exec.Command(parts[0], append(parts[1:], path)...) //nolint:gosec // editor command is user-provided
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Dir = path

		if err := cmd.Start(); err != nil {
			return openEditorResultMsg{err: err}
		}

		return openEditorResultMsg{}
	}
}
