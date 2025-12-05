package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

// Model represents the TUI state.
type Model struct {
	list               list.Model
	svc                *workspaces.Service
	err                error
	infoMessage        string
	printPath          bool
	SelectedPath       string
	loadingDetail      bool
	pushing            bool
	pushTarget         string
	spinner            spinner.Model
	detailView         bool
	selectedWS         *domain.Workspace
	wsStatus           *domain.WorkspaceStatus
	wsOrphans          []domain.OrphanedWorktree
	confirming         bool
	actionToConfirm    string // "close" | "push"
	confirmingID       string
	allItems           []workspaceItem
	statusCache        map[string]*domain.WorkspaceStatus
	totalDiskUsage     int64
	filterStale        bool
	staleThresholdDays int
	lastFilterValue    string
}

// NewModel creates a new TUI model.
func NewModel(svc *workspaces.Service, printPath bool) Model {
	threshold := svc.StaleThresholdDays()

	delegate := newWorkspaceDelegate(threshold)
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.Styles.NoItems = subtleTextStyle

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
			key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "toggle stale")),
			key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "push selected")),
			key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open in editor")),
		}
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorPrimary)

	return Model{
		list:               l,
		svc:                svc,
		printPath:          printPath,
		spinner:            s,
		statusCache:        make(map[string]*domain.WorkspaceStatus),
		staleThresholdDays: threshold,
	}
}

// Init configures initial commands.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadWorkspaces, m.spinner.Tick)
}

// selectedWorkspaceItem returns the currently selected workspace item.
func (m Model) selectedWorkspaceItem() (workspaceItem, bool) {
	if selected, ok := m.list.SelectedItem().(workspaceItem); ok {
		return selected, true
	}

	return workspaceItem{}, false
}

// workspaceItemByID finds a workspace item by its ID.
func (m Model) workspaceItemByID(id string) (workspaceItem, bool) {
	for _, it := range m.allItems {
		if it.workspace.ID == id {
			return it, true
		}
	}

	return workspaceItem{}, false
}
