package components

import (
	"fmt"
	"strings"
)

// ConfirmAction represents the type of action being confirmed.
type ConfirmAction string

// Confirmation action constants.
const (
	ActionClose ConfirmAction = "close"
	ActionPush  ConfirmAction = "push"
	ActionSync  ConfirmAction = "sync"
)

// ActionDescription returns a human-readable description of the action.
func (a ConfirmAction) ActionDescription() string {
	switch a {
	case ActionClose:
		return "close (delete local files)"
	case ActionPush:
		return "push all changes in"
	case ActionSync:
		return "sync"
	default:
		return string(a)
	}
}

// ActionIcon returns an icon for the action type.
func (a ConfirmAction) ActionIcon() string {
	switch a {
	case ActionClose:
		return IconError // destructive action
	case ActionPush:
		return "↑" // push up
	case ActionSync:
		return "⟳" // sync/refresh
	default:
		return IconWarning
	}
}

// ActionStyle returns the appropriate style for the action type.
func (a ConfirmAction) ActionStyle() func(...string) string {
	switch a {
	case ActionClose:
		return StatusDirtyStyle.Render // destructive = danger
	default:
		return StatusWarnStyle.Render // default = warning
	}
}

// ConfirmDialog represents a confirmation dialog state.
type ConfirmDialog struct {
	Active      bool
	Action      ConfirmAction
	TargetLabel string
}

// Render renders the confirmation dialog prompt with modal styling.
func (d ConfirmDialog) Render() string {
	if !d.Active {
		return ""
	}

	var b strings.Builder

	// Icon based on action type
	icon := d.Action.ActionIcon()
	styleRender := d.Action.ActionStyle()

	// Header line with icon
	header := styleRender(fmt.Sprintf("%s Confirm %s", icon, d.Action.ActionDescription()))
	b.WriteString(header)
	b.WriteString("\n")

	// Target
	b.WriteString(fmt.Sprintf("  Target: %s", d.TargetLabel))
	b.WriteString("\n\n")

	// Action buttons
	confirmBtn := AccentTextStyle.Render("[y] Confirm")
	cancelBtn := SubtleTextStyle.Render("[n/esc] Cancel")
	b.WriteString(fmt.Sprintf("  %s  %s", confirmBtn, cancelBtn))

	return b.String()
}
