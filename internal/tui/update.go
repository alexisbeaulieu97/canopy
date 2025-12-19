package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/tui/components"
)

// Update handles incoming Tea messages and state transitions.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:gocyclo // message-driven switch covers multiple event types
	switch msg := msg.(type) {
	case tea.KeyMsg:
		newState, cmd, handled := m.viewState.HandleKey(&m, msg.String())
		if handled {
			m.viewState = newState
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.ui.List.SetWidth(msg.Width)

		height := msg.Height - 6 // Account for header/footer
		if height < 8 {
			height = msg.Height
		}

		m.ui.List.SetHeight(height)
	case workspaceListMsg:
		m.workspaces.SetItems(msg.items, msg.totalUsage)

		m.applyFilters()

		var cmds []tea.Cmd
		for _, it := range msg.items {
			cmds = append(cmds, m.loadWorkspaceStatus(it.Workspace.ID))
		}

		return m, tea.Batch(cmds...)
	case loadWorkspacesErrMsg:
		m.err = msg.err
	case workspaceStatusMsg:
		m.workspaces.CacheStatus(msg.id, msg.status)
		m.updateWorkspaceSummary(msg.id, msg.status, nil)

		if m.isDetailView() && m.selectedWS != nil && m.selectedWS.ID == msg.id {
			m.wsStatus = msg.status
		}
	case workspaceStatusErrMsg:
		m.updateWorkspaceSummary(msg.id, nil, msg.err)
		m.err = msg.err
	case pushResultMsg:
		m.pushing = false

		m.pushTarget = ""
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.infoMessage = "Push completed successfully"
			return m, m.loadWorkspaceStatus(msg.id)
		}
	case workspaceDetailsMsg:
		m.selectedWS = msg.workspace
		m.wsStatus = msg.status

		m.wsOrphans = msg.orphans
		if ds := m.getDetailState(); ds != nil {
			ds.Loading = false
		}
	case workspaceDetailsErrMsg:
		m.err = msg.err
		if ds := m.getDetailState(); ds != nil {
			ds.Loading = false
		}
	case closeWorkspaceErrMsg:
		m.err = msg.err
	case openEditorResultMsg:
		if msg.err != nil {
			m.err = msg.err
			m.infoMessage = ""
		} else {
			m.infoMessage = "Opened in editor"
		}
	case error:
		m.err = msg
		if ds := m.getDetailState(); ds != nil {
			ds.Loading = false
		}

		return m, nil
	}

	var cmd tea.Cmd

	if !m.isDetailView() {
		m.ui.List, cmd = m.ui.List.Update(msg)

		if m.ui.List.FilterValue() != m.lastFilterValue {
			m.lastFilterValue = m.ui.List.FilterValue()
			m.applyFilters()
		}
	}

	var sCmd tea.Cmd

	m.ui.Spinner, sCmd = m.ui.Spinner.Update(msg)

	return m, tea.Batch(cmd, sCmd)
}

// updateWorkspaceSummary updates the summary for a workspace in both allItems and list.
func (m *Model) updateWorkspaceSummary(id string, status *domain.WorkspaceStatus, err error) {
	m.workspaces.UpdateItemSummary(id, status, err)

	for idx, listItem := range m.ui.List.Items() {
		ws, ok := listItem.(workspaceItem)
		if !ok || ws.Workspace.ID != id {
			continue
		}

		if status != nil {
			ws.Loaded = true
			ws.Err = nil
			ws.Summary = summarizeStatus(status)
		}

		if err != nil {
			ws.Err = err
		}

		m.ui.List.SetItem(idx, ws)
	}
}

// applyFilters applies current filters to the workspace list.
func (m *Model) applyFilters() {
	items := m.workspaces.ApplyFilters(m.ui.List.FilterValue())
	m.ui.List.SetItems(items)
}

// handleListKeyWithState handles key events in the main list view using ViewState pattern.
func (m *Model) handleListKeyWithState(state *ListViewState, key string) (ViewState, tea.Cmd, bool) { //nolint:gocyclo // key-driven switch covers multiple keybindings
	if m.pushing {
		if matchesKey(key, m.ui.Keybindings.Quit) {
			return state, tea.Quit, true
		}

		return state, nil, true
	}

	if m.ui.List.FilterState() == list.Filtering {
		// Allow the list's filter input to consume keys (including our shortcuts).
		return state, nil, false
	}

	switch {
	case matchesKey(key, m.ui.Keybindings.Quit):
		return state, tea.Quit, true
	case matchesKey(key, m.ui.Keybindings.Details):
		return m.handleEnterWithState()
	case matchesKey(key, m.ui.Keybindings.Search):
		m.ui.List.SetFilterState(list.Filtering)
		return state, nil, true
	case matchesKey(key, m.ui.Keybindings.ToggleStale):
		m.workspaces.ToggleStaleFilter()
		m.applyFilters()

		return state, nil, true
	case matchesKey(key, m.ui.Keybindings.Push):
		return m.handlePushConfirmWithState()
	case matchesKey(key, m.ui.Keybindings.OpenEditor):
		return m.handleOpenEditorWithState(state)
	case matchesKey(key, m.ui.Keybindings.Close):
		return m.handleCloseConfirmWithState()
	}

	return state, nil, false
}

// handleDetailKeyWithState handles key events in the detail view using ViewState pattern.
func (m *Model) handleDetailKeyWithState(state *DetailViewState, key string) (ViewState, tea.Cmd, bool) {
	// Only cancel or quit keys exit detail view
	if matchesKey(key, m.ui.Keybindings.Cancel) || matchesKey(key, m.ui.Keybindings.Quit) {
		m.selectedWS = nil
		m.wsStatus = nil
		m.wsOrphans = nil

		return &ListViewState{}, nil, true
	}

	return state, nil, false
}

// handleConfirmKeyWithState handles key events during confirmation dialogs using ViewState pattern.
func (m *Model) handleConfirmKeyWithState(state *ConfirmViewState, key string) (ViewState, tea.Cmd, bool) {
	if matchesKey(key, m.ui.Keybindings.Confirm) {
		switch state.Action {
		case components.ActionClose:
			if state.TargetID != "" {
				return &ListViewState{}, m.closeWorkspace(state.TargetID), true
			}
		case components.ActionPush:
			if state.TargetID != "" {
				m.pushing = true
				m.pushTarget = state.TargetID
				m.infoMessage = ""

				return &ListViewState{}, m.pushWorkspace(state.TargetID), true
			}
		}

		return &ListViewState{}, nil, true
	}

	if matchesKey(key, m.ui.Keybindings.Cancel) {
		return &ListViewState{}, nil, true
	}

	// Swallow all other keys during confirmation to prevent accidental actions.
	// The user must explicitly confirm (y) or cancel (n/esc).
	return state, nil, true
}

// handleEnterWithState handles the enter key to view workspace details.
func (m *Model) handleEnterWithState() (ViewState, tea.Cmd, bool) {
	selected, ok := m.selectedWorkspaceItem()
	if !ok {
		return &ListViewState{}, nil, true
	}

	if m.printPath {
		path, err := m.svc.WorkspacePath(context.Background(), selected.Workspace.ID)
		if err != nil {
			m.err = err
			return &ListViewState{}, nil, true
		}

		m.SelectedPath = path

		return &ListViewState{}, tea.Quit, true
	}

	wsCopy := selected.Workspace
	if cached, ok := m.workspaces.GetCachedStatus(selected.Workspace.ID); ok {
		// Show cached status immediately, but still fetch full details (including orphans)
		// in the background. The UI will update when the full details arrive.
		detailState := &DetailViewState{Loading: false} // Not loading since we have cached data
		cachedMsg := func() tea.Msg {
			return workspaceDetailsMsg{workspace: &wsCopy, status: cached}
		}

		return detailState, tea.Batch(cachedMsg, m.loadWorkspaceDetails(selected.Workspace.ID)), true
	}

	detailState := &DetailViewState{Loading: true}

	return detailState, m.loadWorkspaceDetails(selected.Workspace.ID), true
}

// handlePushConfirmWithState initiates push confirmation using ViewState pattern.
func (m *Model) handlePushConfirmWithState() (ViewState, tea.Cmd, bool) {
	selected, ok := m.selectedWorkspaceItem()
	if !ok {
		return &ListViewState{}, nil, true
	}

	m.infoMessage = ""

	return &ConfirmViewState{
		Action:   components.ActionPush,
		TargetID: selected.Workspace.ID,
	}, nil, true
}

// handleOpenEditorWithState opens the selected workspace in an editor.
func (m *Model) handleOpenEditorWithState(state *ListViewState) (ViewState, tea.Cmd, bool) {
	selected, ok := m.selectedWorkspaceItem()
	if !ok {
		return state, nil, true
	}

	return state, m.openWorkspace(selected.Workspace.ID), true
}

// handleCloseConfirmWithState initiates close confirmation using ViewState pattern.
func (m *Model) handleCloseConfirmWithState() (ViewState, tea.Cmd, bool) {
	selected, ok := m.selectedWorkspaceItem()
	if !ok {
		return &ListViewState{}, nil, true
	}

	return &ConfirmViewState{
		Action:   components.ActionClose,
		TargetID: selected.Workspace.ID,
	}, nil, true
}
