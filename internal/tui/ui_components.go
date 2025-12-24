package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/tui/components"
)

// UIComponents groups UI component instances.
type UIComponents struct {
	// List is the bubbles list component.
	List list.Model
	// Spinner is the loading spinner component.
	Spinner spinner.Model
	// Keybindings holds the configured keybindings.
	Keybindings config.Keybindings
}

// NewUIComponents creates a new UIComponents with configured components.
func NewUIComponents(keybindings config.Keybindings, staleThreshold int) UIComponents {
	delegate := newWorkspaceDelegate(staleThreshold)
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.Styles.NoItems = components.SubtleTextStyle

	// Build keybinding help based on configured keys
	searchKey := firstKey(keybindings.Search)
	toggleStaleKey := firstKey(keybindings.ToggleStale)
	syncKey := firstKey(keybindings.Sync)
	pushKey := firstKey(keybindings.Push)
	openKey := firstKey(keybindings.OpenEditor)
	selectKey := firstKey(keybindings.Select)

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys(searchKey), key.WithHelp(searchKey, "search")),
			key.NewBinding(key.WithKeys(toggleStaleKey), key.WithHelp(toggleStaleKey, "toggle stale")),
			key.NewBinding(key.WithKeys(syncKey), key.WithHelp(syncKey, "sync selected")),
			key.NewBinding(key.WithKeys(pushKey), key.WithHelp(pushKey, "push selected")),
			key.NewBinding(key.WithKeys(openKey), key.WithHelp(openKey, "open in editor")),
			key.NewBinding(key.WithKeys(selectKey), key.WithHelp(selectKey, "select workspace")),
		}
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(components.ColorPrimary)

	return UIComponents{
		List:        l,
		Spinner:     s,
		Keybindings: keybindings,
	}
}
