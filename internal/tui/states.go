package tui

import tea "github.com/charmbracelet/bubbletea"

// ViewState represents a distinct view mode in the TUI.
// Each state handles its own view rendering and key events.
type ViewState interface {
	// View renders the current view state.
	View(m *Model) string
	// HandleKey handles key events for this view state.
	// Returns the new state, command, and whether the key was handled.
	HandleKey(m *Model, key string) (ViewState, tea.Cmd, bool)
}

// Action represents the type of action being confirmed in the TUI.
type Action string

// Action constants for confirmation dialogs.
const (
	ActionClose Action = "close"
	ActionPush  Action = "push"
)

// ListViewState represents the main workspace list view.
type ListViewState struct{}

// DetailViewState represents the workspace detail view.
type DetailViewState struct {
	Loading bool
}

// ConfirmViewState represents a confirmation dialog state.
type ConfirmViewState struct {
	Action   Action
	TargetID string
}

// Ensure states implement ViewState interface.
var (
	_ ViewState = (*ListViewState)(nil)
	_ ViewState = (*DetailViewState)(nil)
	_ ViewState = (*ConfirmViewState)(nil)
)

// View renders the list view.
func (s *ListViewState) View(m *Model) string {
	return m.renderListView()
}

// HandleKey handles key events for the list view.
func (s *ListViewState) HandleKey(m *Model, key string) (ViewState, tea.Cmd, bool) {
	return m.handleListKeyWithState(s, key)
}

// View renders the detail view.
func (s *DetailViewState) View(m *Model) string {
	return m.renderDetailView()
}

// HandleKey handles key events for the detail view.
func (s *DetailViewState) HandleKey(m *Model, key string) (ViewState, tea.Cmd, bool) {
	return m.handleDetailKeyWithState(s, key)
}

// View renders the confirmation dialog over the list view.
func (s *ConfirmViewState) View(m *Model) string {
	return m.renderListViewWithConfirm(s)
}

// HandleKey handles key events for the confirmation dialog.
func (s *ConfirmViewState) HandleKey(m *Model, key string) (ViewState, tea.Cmd, bool) {
	return m.handleConfirmKeyWithState(s, key)
}
