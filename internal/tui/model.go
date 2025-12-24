package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

// Model represents the TUI state.
// It coordinates sub-components for views, workspace data, and UI elements.
type Model struct {
	// viewState manages the current view mode (list, detail, confirm).
	viewState ViewState
	// workspaces manages workspace data and caches.
	workspaces *workspaceModel
	// ui groups UI components (list, spinner, keybindings).
	ui UIComponents
	// svc provides workspace operations.
	svc *workspaces.Service
	// symbols provides emoji/ASCII symbol mappings based on config.
	symbols Symbols
	// err holds any error to display.
	err error
	// infoMessage holds an informational message to display.
	infoMessage string
	// printPath enables path-printing mode (exits after selecting).
	printPath bool
	// SelectedPath is set when printPath mode selects a workspace.
	SelectedPath string
	// pushing indicates a push operation is in progress.
	pushing bool
	// pushTarget is the ID of the workspace being pushed.
	pushTarget string
	// selectedWS holds the workspace shown in detail view.
	selectedWS *domain.Workspace
	// wsStatus holds the status of the workspace shown in detail view.
	wsStatus *domain.WorkspaceStatus
	// wsOrphans holds orphaned worktrees for the detail view.
	wsOrphans []domain.OrphanedWorktree
	// lastFilterValue tracks the last filter value for change detection.
	lastFilterValue string
	// selectedIDs tracks workspaces selected for bulk operations.
	selectedIDs map[string]bool
	// selectionMode indicates whether multi-select mode is active.
	selectionMode bool
}

// NewModel creates a new TUI model.
func NewModel(svc *workspaces.Service, printPath bool) Model {
	threshold := svc.StaleThresholdDays()
	kb := svc.Keybindings()
	useEmoji := svc.UseEmoji()

	return Model{
		viewState:   &ListViewState{},
		workspaces:  newWorkspaceModel(threshold),
		ui:          NewUIComponents(kb, threshold),
		svc:         svc,
		symbols:     NewSymbols(useEmoji),
		printPath:   printPath,
		selectedIDs: make(map[string]bool),
	}
}

// Init configures initial commands.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadWorkspaces, m.ui.Spinner.Tick)
}

// matchesKey checks if the pressed key matches any of the configured keybindings.
func matchesKey(key string, bindings []string) bool {
	for _, b := range bindings {
		if key == b {
			return true
		}
	}

	return false
}

// firstKey returns the first key from the bindings, or empty string if none.
func firstKey(bindings []string) string {
	if len(bindings) > 0 {
		return bindings[0]
	}

	return ""
}

// selectedWorkspaceItem returns the currently selected workspace item.
func (m Model) selectedWorkspaceItem() (workspaceItem, bool) {
	if selected, ok := m.ui.List.SelectedItem().(workspaceItem); ok {
		return selected, true
	}

	return workspaceItem{}, false
}

// workspaceItemByID finds a workspace item by its ID.
func (m Model) workspaceItemByID(id string) (workspaceItem, bool) {
	return m.workspaces.FindItemByID(id)
}

// isDetailView returns whether the model is in detail view.
func (m Model) isDetailView() bool {
	_, ok := m.viewState.(*DetailViewState)
	return ok
}

// isConfirming returns whether the model is showing a confirmation dialog.
func (m Model) isConfirming() bool {
	_, ok := m.viewState.(*ConfirmViewState)
	return ok
}

// getConfirmState returns the confirmation state if active, nil otherwise.
func (m Model) getConfirmState() *ConfirmViewState {
	if cs, ok := m.viewState.(*ConfirmViewState); ok {
		return cs
	}

	return nil
}

// getDetailState returns the detail state if active, nil otherwise.
func (m Model) getDetailState() *DetailViewState {
	if ds, ok := m.viewState.(*DetailViewState); ok {
		return ds
	}

	return nil
}

func (m Model) selectionCount() int {
	return len(m.selectedIDs)
}

func (m Model) hasSelection() bool {
	return len(m.selectedIDs) > 0
}

func (m *Model) updateSelectionMode() {
	m.selectionMode = len(m.selectedIDs) > 0
}

func (m *Model) applySelectionToItems() {
	for idx, it := range m.workspaces.allItems {
		it.Selected = m.selectedIDs[it.Workspace.ID]
		m.workspaces.allItems[idx] = it
	}
}

func (m *Model) syncSelectionState() {
	m.applySelectionToItems()
	m.applyFilters()
}

func (m *Model) toggleWorkspaceSelection(id string) {
	if m.selectedIDs[id] {
		delete(m.selectedIDs, id)
	} else {
		m.selectedIDs[id] = true
	}

	m.updateSelectionMode()
	m.syncSelectionState()
}

func (m *Model) selectAllVisible() {
	for _, item := range m.ui.List.Items() {
		wsItem, ok := item.(workspaceItem)
		if !ok {
			continue
		}

		m.selectedIDs[wsItem.Workspace.ID] = true
	}

	m.updateSelectionMode()
	m.syncSelectionState()
}

func (m *Model) clearSelection() {
	if len(m.selectedIDs) == 0 {
		return
	}

	m.selectedIDs = make(map[string]bool)
	m.updateSelectionMode()
	m.syncSelectionState()
}

func (m *Model) pruneSelectionIDs(items []workspaceItem) {
	if len(m.selectedIDs) == 0 {
		return
	}

	valid := make(map[string]bool, len(items))
	for _, it := range items {
		valid[it.Workspace.ID] = true
	}

	for id := range m.selectedIDs {
		if !valid[id] {
			delete(m.selectedIDs, id)
		}
	}

	m.updateSelectionMode()
}

func (m *Model) selectedWorkspaceIDs() []string {
	if len(m.selectedIDs) == 0 {
		return nil
	}

	ids := make([]string, 0, len(m.selectedIDs))
	for _, it := range m.workspaces.Items() {
		if m.selectedIDs[it.Workspace.ID] {
			ids = append(ids, it.Workspace.ID)
		}
	}

	return ids
}

func (m *Model) actionTargetIDs() []string {
	if ids := m.selectedWorkspaceIDs(); len(ids) > 0 {
		return ids
	}

	selected, ok := m.selectedWorkspaceItem()
	if !ok {
		return nil
	}

	return []string{selected.Workspace.ID}
}
