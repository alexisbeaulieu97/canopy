package tui

import (
	"fmt"
	"path/filepath"

	"github.com/alexisbeaulieu97/yard/internal/domain"
	"github.com/alexisbeaulieu97/yard/internal/workspaces"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	statusCleanStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#50FA7B"))

	statusDirtyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5555"))
)

type item struct {
	title, desc string
	id          string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type Model struct {
	list            list.Model
	svc             *workspaces.Service
	workspacesRoot  string
	err             error
	printPath       bool
	SelectedPath    string
	loading         bool
	spinner         spinner.Model
	detailView      bool
	selectedWS      *domain.Workspace
	wsStatus        *domain.WorkspaceStatus
	confirming      bool
	actionToConfirm string // "close"
}

func NewModel(svc *workspaces.Service, workspacesRoot string, printPath bool) Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Workspaces"
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return Model{
		list:           l,
		svc:            svc,
		workspacesRoot: workspacesRoot,
		printPath:      printPath,
		spinner:        s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadWorkspaces, m.spinner.Tick)
}

func (m Model) loadWorkspaces() tea.Msg {
	workspaces, err := m.svc.ListWorkspaces()
	if err != nil {
		return err
	}
	items := make([]list.Item, len(workspaces))
	for i, w := range workspaces {
		items[i] = item{title: w.ID, desc: fmt.Sprintf("%d repos", len(w.Repos)), id: w.ID}
	}
	return items
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.detailView {
			switch msg.String() {
			case "esc", "q":
				m.detailView = false
				m.selectedWS = nil
				m.wsStatus = nil
				return m, nil
			}
		} else {
			// Main list view
			if m.confirming {
				switch msg.String() {
				case "y", "Y":
					m.confirming = false
					if m.actionToConfirm == "close" {
						if i, ok := m.list.SelectedItem().(item); ok {
							return m, m.closeWorkspace(i.id)
						}
					}
				case "n", "N", "esc":
					m.confirming = false
					m.actionToConfirm = ""
				}
				return m, nil
			}

			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				if i, ok := m.list.SelectedItem().(item); ok {
					if m.printPath {
						m.SelectedPath = filepath.Join(m.workspacesRoot, i.id)
						return m, tea.Quit
					}
					// Enter detail view
					m.detailView = true
					m.loading = true
					return m, m.loadWorkspaceDetails(i.id)
				}
			case "s":
				if i, ok := m.list.SelectedItem().(item); ok {
					return m, m.syncWorkspace(i.id)
				}
			case "c":
				if _, ok := m.list.SelectedItem().(item); ok {
					m.confirming = true
					m.actionToConfirm = "close"
				}
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height)
	case []list.Item:
		m.list.SetItems(msg)
	case error:
		m.err = msg
		return m, nil
	case *domain.WorkspaceStatus:
		m.wsStatus = msg
		m.loading = false
	case workspaceDetailsMsg:
		m.selectedWS = msg.workspace
		m.wsStatus = msg.status
		m.loading = false
	}

	var cmd tea.Cmd
	if m.detailView {
		// Handle detail view updates if any
	} else {
		m.list, cmd = m.list.Update(msg)
	}
	var sCmd tea.Cmd
	m.spinner, sCmd = m.spinner.Update(msg)
	return m, tea.Batch(cmd, sCmd)
}

type workspaceDetailsMsg struct {
	workspace *domain.Workspace
	status    *domain.WorkspaceStatus
}

func (m Model) loadWorkspaceDetails(id string) tea.Cmd {
	return func() tea.Msg {
		list, err := m.svc.ListWorkspaces()
		if err != nil {
			return err
		}
		var ws *domain.Workspace
		for _, w := range list {
			if w.ID == id {
				ws = &w
				break
			}
		}
		if ws == nil {
			return fmt.Errorf("workspace not found")
		}

		status, err := m.svc.GetStatus(id)
		if err != nil {
			return err
		}

		return workspaceDetailsMsg{workspace: ws, status: status}
	}
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	if m.detailView {
		if m.loading {
			return fmt.Sprintf("%s Loading details...", m.spinner.View())
		}
		if m.selectedWS != nil && m.wsStatus != nil {
			s := fmt.Sprintf("Workspace: %s\n", m.selectedWS.ID)
			s += fmt.Sprintf("Branch: %s\n\n", m.selectedWS.BranchName)
			s += "Repositories:\n"
			for _, r := range m.wsStatus.Repos {
				statusStyle := statusCleanStyle
				statusText := "Clean"
				if r.IsDirty {
					statusStyle = statusDirtyStyle
					statusText = "Dirty"
				}

				branchInfo := fmt.Sprintf("[%s]", r.Branch)
				if r.UnpushedCommits > 0 {
					branchInfo += fmt.Sprintf(" %d unpushed", r.UnpushedCommits)
				}

				s += fmt.Sprintf("- %-20s %s %s\n", r.Name, branchInfo, statusStyle.Render(statusText))
			}
			s += "\n(Press 'esc' to go back)"
			return s
		}
	}

	if m.confirming {
		return fmt.Sprintf("\n  Are you sure you want to %s this workspace? (y/n)\n\n%s", m.actionToConfirm, m.list.View())
	}

	return m.list.View()
}

func (m Model) loadTicketStatus(id string) tea.Cmd {
	return func() tea.Msg {
		status, err := m.svc.GetStatus(id)
		if err != nil {
			return err
		}
		return status
	}
}

func (m Model) closeWorkspace(id string) tea.Cmd {
	return func() tea.Msg {
		err := m.svc.CloseWorkspace(id, true) // Force close for now as we confirmed
		if err != nil {
			return err
		}
		// Reload list
		return m.loadWorkspaces()
	}
}

func (m Model) syncWorkspace(id string) tea.Cmd {
	return func() tea.Msg {
		err := m.svc.SyncWorkspace(id)
		if err != nil {
			return err
		}
		return nil // Or some success message?
	}
}
