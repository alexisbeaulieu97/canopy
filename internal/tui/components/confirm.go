package components

import (
	"fmt"
)

// ConfirmAction represents the type of action being confirmed.
type ConfirmAction string

// Confirmation action constants.
const (
	ActionClose ConfirmAction = "close"
	ActionPush  ConfirmAction = "push"
)

// ActionDescription returns a human-readable description of the action.
func (a ConfirmAction) ActionDescription() string {
	switch a {
	case ActionClose:
		return "close (delete local files)"
	case ActionPush:
		return "push all changes in"
	default:
		return string(a)
	}
}

// ConfirmDialog represents a confirmation dialog state.
type ConfirmDialog struct {
	Active   bool
	Action   ConfirmAction
	TargetID string
}

// NewConfirmDialog creates a new inactive confirmation dialog.
func NewConfirmDialog() ConfirmDialog {
	return ConfirmDialog{
		Active: false,
	}
}

// Show activates the confirmation dialog with the specified action and target.
func (d *ConfirmDialog) Show(action ConfirmAction, targetID string) {
	d.Active = true
	d.Action = action
	d.TargetID = targetID
}

// Hide deactivates the confirmation dialog and clears its state.
func (d *ConfirmDialog) Hide() {
	d.Active = false
	d.Action = ""
	d.TargetID = ""
}

// Render renders the confirmation dialog prompt.
func (d ConfirmDialog) Render() string {
	if !d.Active {
		return ""
	}

	prompt := fmt.Sprintf("⚠️  Confirm %s workspace %s?",
		d.Action.ActionDescription(),
		AccentTextStyle.Render(d.TargetID))

	hint := SubtleTextStyle.Render("Press [y] to confirm, [n] or [esc] to cancel")

	return ConfirmPromptStyle.Render(prompt) + "\n" + hint
}

// HandleKey processes a key press in the confirmation dialog.
// Returns: confirmed (bool), handled (bool)
// - confirmed: true if user pressed y/Y to confirm
// - handled: true if the key was processed by the dialog
//
// Note: Callers should read Action and TargetID before calling HandleKey
// if they need those values, as Hide() clears them.
func (d *ConfirmDialog) HandleKey(key string) (confirmed, handled bool) {
	if !d.Active {
		return false, false
	}

	switch key {
	case "y", "Y":
		d.Hide()

		return true, true
	case "n", "N", "esc":
		d.Hide()

		return false, true
	default:
		// Dialog is active but key not recognized - still consumed
		return false, true
	}
}
