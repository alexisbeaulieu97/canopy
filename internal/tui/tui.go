// Package tui provides the terminal UI for canopy.
package tui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

// Styles
var (
	statusCleanStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#50FA7B"))

	statusDirtyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5555"))

	statusWarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1FA8C"))

	subtleTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	badgeStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("244"))
)

type workspaceItem struct {
	workspace domain.Workspace
	summary   workspaceSummary
	err       error
	loaded    bool
}

type workspaceSummary struct {
	repoCount     int
	dirtyRepos    int
	unpushedRepos int
	behindRepos   int
}

func (i workspaceItem) Title() string       { return i.workspace.ID }
func (i workspaceItem) Description() string { return "" }
func (i workspaceItem) FilterValue() string { return i.workspace.ID }

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
	confirming         bool
	actionToConfirm    string // "close" | "push"
	confirmingID       string
	allItems           []workspaceItem
	statusCache        map[string]*domain.WorkspaceStatus
	totalDiskUsage     int64
	filterStale        bool
	staleThresholdDays int
}

type workspaceListMsg struct {
	items      []workspaceItem
	totalUsage int64
}

type workspaceStatusMsg struct {
	id     string
	status *domain.WorkspaceStatus
}

type workspaceStatusErrMsg struct {
	id  string
	err error
}

type pushResultMsg struct {
	id  string
	err error
}

type openEditorResultMsg struct {
	err error
}

type workspaceDetailsMsg struct {
	workspace *domain.Workspace
	status    *domain.WorkspaceStatus
}

// NewModel creates a new TUI model.
func NewModel(svc *workspaces.Service, printPath bool) Model {
	threshold := svc.StaleThresholdDays()

	delegate := newWorkspaceDelegate(threshold)
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Workspaces"
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)
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
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

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

func (m Model) loadWorkspaces() tea.Msg {
	workspaces, err := m.svc.ListWorkspaces()
	if err != nil {
		return err
	}

	items := make([]workspaceItem, 0, len(workspaces))

	var totalUsage int64

	for _, w := range workspaces {
		items = append(items, workspaceItem{
			workspace: w,
			summary: workspaceSummary{
				repoCount: len(w.Repos),
			},
		})
		totalUsage += w.DiskUsageBytes
	}

	return workspaceListMsg{
		items:      items,
		totalUsage: totalUsage,
	}
}

func (m Model) loadWorkspaceStatus(id string) tea.Cmd {
	return func() tea.Msg {
		status, err := m.svc.GetStatus(id)
		if err != nil {
			return workspaceStatusErrMsg{id: id, err: err}
		}

		return workspaceStatusMsg{id: id, status: status}
	}
}

// Update handles incoming Tea messages and state transitions.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:gocyclo // message-driven switch covers multiple event types
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if updated, cmd, handled := m.handleKey(msg.String()); handled {
			return updated, cmd
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)

		height := msg.Height - 4
		if height < 8 {
			height = msg.Height
		}

		m.list.SetHeight(height)
	case workspaceListMsg:
		m.totalDiskUsage = msg.totalUsage
		m.allItems = msg.items

		listItems := make([]list.Item, len(msg.items))
		for i := range msg.items {
			listItems[i] = msg.items[i]
		}

		m.list.SetItems(listItems)

		if m.filterStale {
			m.applyFilters()
		}

		var cmds []tea.Cmd
		for _, it := range msg.items {
			cmds = append(cmds, m.loadWorkspaceStatus(it.workspace.ID))
		}

		return m, tea.Batch(cmds...)
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
			m.infoMessage = "Push completed"
			return m, m.loadWorkspaceStatus(msg.id)
		}
	case workspaceDetailsMsg:
		m.selectedWS = msg.workspace
		m.wsStatus = msg.status
		m.loadingDetail = false
	case openEditorResultMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.infoMessage = "Opened in editor"
		}
	case error:
		m.err = msg
		return m, nil
	}

	var cmd tea.Cmd

	if !m.detailView {
		m.list, cmd = m.list.Update(msg)
	}

	var sCmd tea.Cmd

	m.spinner, sCmd = m.spinner.Update(msg)

	return m, tea.Batch(cmd, sCmd)
}

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

func summarizeStatus(status *domain.WorkspaceStatus) workspaceSummary {
	summary := workspaceSummary{
		repoCount: len(status.Repos),
	}

	for _, repo := range status.Repos {
		if repo.IsDirty {
			summary.dirtyRepos++
		}

		if repo.UnpushedCommits > 0 {
			summary.unpushedRepos++
		}

		if repo.BehindRemote > 0 {
			summary.behindRepos++
		}
	}

	return summary
}

type workspaceDelegate struct {
	styles         list.DefaultItemStyles
	staleThreshold int
}

func newWorkspaceDelegate(staleThreshold int) workspaceDelegate {
	styles := list.NewDefaultItemStyles()
	styles.NormalTitle = styles.NormalTitle.Bold(true)
	styles.SelectedTitle = styles.SelectedTitle.Bold(true)

	return workspaceDelegate{
		styles:         styles,
		staleThreshold: staleThreshold,
	}
}

func (d workspaceDelegate) Height() int { return 2 }

func (d workspaceDelegate) Spacing() int { return 0 }

func (d workspaceDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d workspaceDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	wsItem, ok := listItem.(workspaceItem)
	if !ok {
		return
	}

	cursor := " "
	if index == m.Index() {
		cursor = ">"
	}

	statusText, statusStyle := healthForWorkspace(wsItem, d.staleThreshold)

	titleStyle := d.styles.NormalTitle

	descStyle := d.styles.NormalDesc
	if index == m.Index() {
		titleStyle = d.styles.SelectedTitle
		descStyle = d.styles.SelectedDesc
	}

	title := titleStyle.Render(wsItem.workspace.ID)
	badges := renderBadges(wsItem, d.staleThreshold)

	secondary := "loading status..."
	if wsItem.err != nil {
		secondary = fmt.Sprintf("status error: %s", wsItem.err.Error())
	} else if wsItem.loaded {
		lastUpdated := relativeTime(wsItem.workspace.LastModified)
		secondary = fmt.Sprintf(
			"%d repos | %s | Updated %s",
			wsItem.summary.repoCount,
			humanizeBytes(wsItem.workspace.DiskUsageBytes),
			lastUpdated,
		)
	}

	_, _ = fmt.Fprintf(
		w,
		"%s %s %s %s\n",
		cursor,
		statusStyle.Render("● "+statusText),
		title,
		badges,
	)

	_, _ = fmt.Fprintf(w, "  %s\n", descStyle.Render(secondary))
}

func healthForWorkspace(item workspaceItem, staleThreshold int) (string, lipgloss.Style) {
	switch {
	case item.err != nil:
		return "error", statusDirtyStyle
	case !item.loaded:
		return "checking", subtleTextStyle
	case item.summary.dirtyRepos > 0 || item.summary.unpushedRepos > 0:
		return "dirty", statusDirtyStyle
	case item.workspace.IsStale(staleThreshold) || item.summary.behindRepos > 0:
		return "attention", statusWarnStyle
	default:
		return "clean", statusCleanStyle
	}
}

func renderBadges(item workspaceItem, staleThreshold int) string {
	if !item.loaded && item.err == nil {
		return ""
	}

	var badges []string

	dangerBadge := badgeStyle.
		BorderForeground(lipgloss.Color("#FF5555")).
		Foreground(lipgloss.Color("#FF5555"))

	warnBadge := badgeStyle.
		BorderForeground(lipgloss.Color("#F1FA8C")).
		Foreground(lipgloss.Color("#F1FA8C"))

	if item.err != nil {
		badges = append(badges, dangerBadge.Render("STATUS ERROR"))
	}

	if item.summary.dirtyRepos > 0 {
		badges = append(badges, dangerBadge.Render(fmt.Sprintf("%d dirty", item.summary.dirtyRepos)))
	}

	if item.summary.unpushedRepos > 0 {
		badges = append(badges, dangerBadge.Render(fmt.Sprintf("%d unpushed", item.summary.unpushedRepos)))
	}

	if item.summary.behindRepos > 0 {
		badges = append(badges, warnBadge.Render(fmt.Sprintf("%d behind", item.summary.behindRepos)))
	}

	if item.workspace.IsStale(staleThreshold) {
		badges = append(badges, warnBadge.Render("STALE"))
	}

	return strings.Join(badges, " ")
}

func humanizeBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	value := float64(size) / float64(div)

	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	return fmt.Sprintf("%.1f %s", value, units[exp])
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	diff := time.Since(t)
	switch {
	case diff < time.Hour:
		return "just now"
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case diff < 14*24*time.Hour:
		days := int(diff.Hours()) / 24
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("2006-01-02")
	}
}

func (m *Model) applyFilters() {
	var items []list.Item

	for _, it := range m.allItems {
		if m.filterStale && !it.workspace.IsStale(m.staleThresholdDays) {
			continue
		}

		items = append(items, it)
	}

	m.list.SetItems(items)
}

func (m Model) selectedWorkspaceItem() (workspaceItem, bool) {
	if selected, ok := m.list.SelectedItem().(workspaceItem); ok {
		return selected, true
	}

	return workspaceItem{}, false
}

func (m Model) workspaceItemByID(id string) (workspaceItem, bool) {
	for _, it := range m.allItems {
		if it.workspace.ID == id {
			return it, true
		}
	}

	return workspaceItem{}, false
}

func (m Model) loadWorkspaceDetails(id string) tea.Cmd {
	return func() tea.Msg {
		wsItem, ok := m.workspaceItemByID(id)
		if !ok {
			return fmt.Errorf("workspace not found")
		}

		status, err := m.svc.GetStatus(id)
		if err != nil {
			return err
		}

		wsCopy := wsItem.workspace

		return workspaceDetailsMsg{workspace: &wsCopy, status: status}
	}
}

// View renders the UI for the current state.
func (m Model) View() string {
	if m.detailView {
		return m.renderDetailView()
	}

	header := m.renderHeader()
	body := m.list.View()

	if m.pushing {
		body = fmt.Sprintf("%s Pushing %s...\n\n%s", m.spinner.View(), m.pushTarget, body)
	}

	if m.confirming {
		prompt := fmt.Sprintf("Confirm %s workspace %s? (y/n)", m.actionToConfirm, m.confirmingID)

		return fmt.Sprintf("%s\n%s\n\n%s", header, prompt, body)
	}

	return fmt.Sprintf("%s\n%s", header, body)
}

func (m Model) renderHeader() string {
	total := len(m.allItems)
	visible := len(m.list.Items())

	header := fmt.Sprintf("Workspaces: %d", total)
	if visible != total {
		header += fmt.Sprintf(" (showing %d)", visible)
	}

	if m.totalDiskUsage > 0 {
		header += fmt.Sprintf(" | Total disk: %s", humanizeBytes(m.totalDiskUsage))
	}

	var filters []string
	if m.filterStale {
		filters = append(filters, "stale")
	}

	if m.list.FilterState() != list.Unfiltered {
		filters = append(filters, fmt.Sprintf("search:\"%s\"", m.list.FilterValue()))
	}

	if len(filters) > 0 {
		header += " | Filters: " + strings.Join(filters, ", ")
	}

	if m.err != nil {
		header += "\n" + statusDirtyStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.infoMessage != "" {
		header += "\n" + statusCleanStyle.Render(m.infoMessage)
	}

	return header
}

func (m Model) renderDetailView() string {
	if m.loadingDetail {
		return fmt.Sprintf("%s Loading details...", m.spinner.View())
	}

	if m.selectedWS == nil {
		return "No workspace selected. Press 'esc' to return."
	}

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Workspace: %s\n", m.selectedWS.ID))
	builder.WriteString(fmt.Sprintf("Branch: %s\n", m.selectedWS.BranchName))
	builder.WriteString(fmt.Sprintf("Disk: %s\n", humanizeBytes(m.selectedWS.DiskUsageBytes)))
	builder.WriteString(fmt.Sprintf("Last Modified: %s\n\n", relativeTime(m.selectedWS.LastModified)))

	if m.wsStatus == nil {
		builder.WriteString("No status available.\n")
	} else {
		builder.WriteString("Repositories:\n")

		for _, r := range m.wsStatus.Repos {
			flags := []string{}

			if r.IsDirty {
				flags = append(flags, statusDirtyStyle.Render("dirty"))
			}

			if r.UnpushedCommits > 0 {
				flags = append(flags, statusDirtyStyle.Render(fmt.Sprintf("%d unpushed", r.UnpushedCommits)))
			}

			if r.BehindRemote > 0 {
				flags = append(flags, statusWarnStyle.Render(fmt.Sprintf("%d behind", r.BehindRemote)))
			}

			if len(flags) == 0 {
				flags = append(flags, statusCleanStyle.Render("clean"))
			}

			builder.WriteString(fmt.Sprintf("- %-18s [%s] %s\n", r.Name, r.Branch, strings.Join(flags, " ")))
		}
	}

	builder.WriteString("\n(Press 'esc' to go back)")

	return builder.String()
}

func (m Model) closeWorkspace(id string) tea.Cmd {
	return func() tea.Msg {
		err := m.svc.CloseWorkspace(id, false)
		if err != nil {
			return err
		}
		// Reload list
		return m.loadWorkspaces()
	}
}

func (m Model) pushWorkspace(id string) tea.Cmd {
	return func() tea.Msg {
		return pushResultMsg{
			id:  id,
			err: m.svc.PushWorkspace(id),
		}
	}
}

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

func (m Model) handleDetailKey(key string) (Model, tea.Cmd, bool) {
	if key != "esc" && key != "q" {
		return m, nil, false
	}

	m.detailView = false
	m.loadingDetail = false
	m.selectedWS = nil
	m.wsStatus = nil

	return m, nil, true
}

func (m Model) handleConfirmKey(key string) (Model, tea.Cmd, bool) {
	if key == "y" || key == "Y" {
		m.confirming = false

		switch m.actionToConfirm { //nolint:exhaustive // limited action set
		case "close":
			targetID := m.confirmingID
			m.confirmingID = ""

			if targetID != "" {
				return m, m.closeWorkspace(targetID), true
			}
		case "push":
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

func (m Model) handleListKey(key string) (Model, tea.Cmd, bool) {
	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit, true
	case "enter":
		return m.handleEnter()
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

func (m Model) handlePushConfirm() (Model, tea.Cmd, bool) {
	selected, ok := m.selectedWorkspaceItem()
	if !ok {
		return m, nil, true
	}

	m.confirming = true
	m.confirmingID = selected.workspace.ID
	m.actionToConfirm = "push"
	m.infoMessage = ""

	return m, nil, true
}

func (m Model) handleOpenEditor() (Model, tea.Cmd, bool) {
	selected, ok := m.selectedWorkspaceItem()
	if !ok {
		return m, nil, true
	}

	return m, m.openWorkspace(selected.workspace.ID), true
}

func (m Model) handleCloseConfirm() (Model, tea.Cmd, bool) {
	selected, ok := m.selectedWorkspaceItem()
	if !ok {
		return m, nil, true
	}

	m.confirming = true
	m.actionToConfirm = "close"
	m.confirmingID = selected.workspace.ID

	return m, nil, true
}
