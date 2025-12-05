package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// Update handles incoming Tea messages and state transitions.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:gocyclo // message-driven switch covers multiple event types
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if updated, cmd, handled := m.handleKey(msg.String()); handled {
			return updated, cmd
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)

		height := msg.Height - 6 // Account for header/footer
		if height < 8 {
			height = msg.Height
		}

		m.list.SetHeight(height)
	case workspaceListMsg:
		m.totalDiskUsage = msg.totalUsage
		m.allItems = msg.items

		m.applyFilters()

		var cmds []tea.Cmd
		for _, it := range msg.items {
			cmds = append(cmds, m.loadWorkspaceStatus(it.workspace.ID))
		}

		return m, tea.Batch(cmds...)
	case loadWorkspacesErrMsg:
		m.err = msg.err
	case workspaceStatusMsg:
		m.statusCache[msg.id] = msg.status
		m.updateWorkspaceSummary(msg.id, msg.status, nil)

		if m.detailView && m.selectedWS != nil && m.selectedWS.ID == msg.id {
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
		m.loadingDetail = false
	case workspaceDetailsErrMsg:
		m.err = msg.err
		m.loadingDetail = false
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
		m.loadingDetail = false

		return m, nil
	}

	var cmd tea.Cmd

	if !m.detailView {
		m.list, cmd = m.list.Update(msg)

		if m.list.FilterValue() != m.lastFilterValue {
			m.lastFilterValue = m.list.FilterValue()
			m.applyFilters()
		}
	}

	var sCmd tea.Cmd

	m.spinner, sCmd = m.spinner.Update(msg)

	return m, tea.Batch(cmd, sCmd)
}

// updateWorkspaceSummary updates the summary for a workspace in both allItems and list.
func (m *Model) updateWorkspaceSummary(id string, status *domain.WorkspaceStatus, err error) {
	for idx, it := range m.allItems {
		if it.workspace.ID != id {
			continue
		}

		if status != nil {
			it.loaded = true
			it.err = nil
			it.summary = summarizeStatus(status)
		}

		if err != nil {
			it.err = err
		}

		m.allItems[idx] = it
	}

	for idx, listItem := range m.list.Items() {
		ws, ok := listItem.(workspaceItem)
		if !ok || ws.workspace.ID != id {
			continue
		}

		if status != nil {
			ws.loaded = true
			ws.err = nil
			ws.summary = summarizeStatus(status)
		}

		if err != nil {
			ws.err = err
		}

		m.list.SetItem(idx, ws)
	}
}

// applyFilters applies current filters to the workspace list.
func (m *Model) applyFilters() {
	var items []list.Item

	search := strings.ToLower(strings.TrimSpace(m.list.FilterValue()))

	for _, it := range m.allItems {
		if m.filterStale && !it.workspace.IsStale(m.staleThresholdDays) {
			continue
		}

		if search != "" && !strings.Contains(strings.ToLower(it.workspace.ID), search) {
			continue
		}

		items = append(items, it)
	}

	m.list.SetItems(items)
}

// handleKey dispatches key events to the appropriate handler.
func (m Model) handleKey(key string) (Model, tea.Cmd, bool) {
	if m.detailView {
		return m.handleDetailKey(key)
	}

	if m.confirming {
		return m.handleConfirmKey(key)
	}

	if m.pushing {
		if key == "ctrl+c" || key == "q" {
			return m, tea.Quit, true
		}

		return m, nil, true
	}

	return m.handleListKey(key)
}

// handleDetailKey handles key events in the detail view.
func (m Model) handleDetailKey(key string) (Model, tea.Cmd, bool) {
	if key != "esc" && key != "q" {
		return m, nil, false
	}

	m.detailView = false
	m.loadingDetail = false
	m.selectedWS = nil
	m.wsStatus = nil
	m.wsOrphans = nil

	return m, nil, true
}

// handleConfirmKey handles key events during confirmation dialogs.
func (m Model) handleConfirmKey(key string) (Model, tea.Cmd, bool) {
	if key == "y" || key == "Y" {
		m.confirming = false

		switch m.actionToConfirm {
		case actionClose:
			targetID := m.confirmingID
			m.confirmingID = ""

			if targetID != "" {
				return m, m.closeWorkspace(targetID), true
			}
		case actionPush:
			targetID := m.confirmingID

			m.confirmingID = ""
			if targetID != "" {
				m.pushing = true
				m.pushTarget = targetID
				m.infoMessage = ""

				return m, m.pushWorkspace(targetID), true
			}
		}

		return m, nil, true
	}

	if key == "n" || key == "N" || key == "esc" {
		m.confirming = false
		m.actionToConfirm = ""
		m.confirmingID = ""

		return m, nil, true
	}

	return m, nil, true
}

// handleListKey handles key events in the main list view.
func (m Model) handleListKey(key string) (Model, tea.Cmd, bool) {
	if m.list.FilterState() == list.Filtering {
		// Allow the list's filter input to consume keys (including our shortcuts).
		return m, nil, false
	}

	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit, true
	case "enter":
		return m.handleEnter()
	case "/":
		m.list.SetFilterState(list.Filtering)
		return m, nil, true
	case "s":
		m.filterStale = !m.filterStale
		m.applyFilters()

		return m, nil, true
	case "p":
		return m.handlePushConfirm()
	case "o":
		return m.handleOpenEditor()
	case "c":
		return m.handleCloseConfirm()
	}

	return m, nil, false
}

// handleEnter handles the enter key to view workspace details.
func (m Model) handleEnter() (Model, tea.Cmd, bool) {
	selected, ok := m.selectedWorkspaceItem()
	if !ok {
		return m, nil, true
	}

	if m.printPath {
		path, err := m.svc.WorkspacePath(selected.workspace.ID)
		if err != nil {
			m.err = err
			return m, nil, true
		}

		m.SelectedPath = path

		return m, tea.Quit, true
	}

	m.detailView = true
	m.loadingDetail = true

	wsCopy := selected.workspace
	if cached, ok := m.statusCache[selected.workspace.ID]; ok {
		return m, func() tea.Msg {
			return workspaceDetailsMsg{workspace: &wsCopy, status: cached}
		}, true
	}

	return m, m.loadWorkspaceDetails(selected.workspace.ID), true
}

// handlePushConfirm initiates push confirmation.
func (m Model) handlePushConfirm() (Model, tea.Cmd, bool) {
	selected, ok := m.selectedWorkspaceItem()
	if !ok {
		return m, nil, true
	}

	m.confirming = true
	m.confirmingID = selected.workspace.ID
	m.actionToConfirm = actionPush
	m.infoMessage = ""

	return m, nil, true
}

// handleOpenEditor opens the selected workspace in an editor.
func (m Model) handleOpenEditor() (Model, tea.Cmd, bool) {
	selected, ok := m.selectedWorkspaceItem()
	if !ok {
		return m, nil, true
	}

	return m, m.openWorkspace(selected.workspace.ID), true
}

// handleCloseConfirm initiates close confirmation.
func (m Model) handleCloseConfirm() (Model, tea.Cmd, bool) {
	selected, ok := m.selectedWorkspaceItem()
	if !ok {
		return m, nil, true
	}

	m.confirming = true
	m.actionToConfirm = actionClose
	m.confirmingID = selected.workspace.ID

	return m, nil, true
}
