package components

import (
	"strings"
	"testing"
)

func TestConfirmDialog_Show(t *testing.T) {
	dialog := NewConfirmDialog()

	if dialog.Active {
		t.Error("expected new dialog to be inactive")
	}

	dialog.Show(ActionClose, "test-workspace")

	if !dialog.Active {
		t.Error("expected dialog to be active after Show")
	}

	if dialog.Action != ActionClose {
		t.Errorf("expected action %s, got %s", ActionClose, dialog.Action)
	}

	if dialog.TargetLabel != "test-workspace" {
		t.Errorf("expected targetLabel %s, got %s", "test-workspace", dialog.TargetLabel)
	}
}

func TestConfirmDialog_Hide(t *testing.T) {
	dialog := ConfirmDialog{
		Active:      true,
		Action:      ActionPush,
		TargetLabel: "test-workspace",
	}

	dialog.Hide()

	if dialog.Active {
		t.Error("expected dialog to be inactive after Hide")
	}

	if dialog.Action != "" {
		t.Error("expected action to be cleared after Hide")
	}

	if dialog.TargetLabel != "" {
		t.Error("expected targetLabel to be cleared after Hide")
	}
}

func TestConfirmDialog_HandleKey(t *testing.T) {
	tests := []struct {
		name            string
		key             string
		active          bool
		expectConfirmed bool
		expectHandled   bool
		expectActive    bool
	}{
		{
			name:            "inactive dialog ignores keys",
			key:             "y",
			active:          false,
			expectConfirmed: false,
			expectHandled:   false,
			expectActive:    false,
		},
		{
			name:            "y confirms and closes",
			key:             "y",
			active:          true,
			expectConfirmed: true,
			expectHandled:   true,
			expectActive:    false,
		},
		{
			name:            "Y confirms and closes",
			key:             "Y",
			active:          true,
			expectConfirmed: true,
			expectHandled:   true,
			expectActive:    false,
		},
		{
			name:            "n cancels and closes",
			key:             "n",
			active:          true,
			expectConfirmed: false,
			expectHandled:   true,
			expectActive:    false,
		},
		{
			name:            "N cancels and closes",
			key:             "N",
			active:          true,
			expectConfirmed: false,
			expectHandled:   true,
			expectActive:    false,
		},
		{
			name:            "esc cancels and closes",
			key:             "esc",
			active:          true,
			expectConfirmed: false,
			expectHandled:   true,
			expectActive:    false,
		},
		{
			name:            "other keys are handled but do nothing",
			key:             "x",
			active:          true,
			expectConfirmed: false,
			expectHandled:   true,
			expectActive:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := ConfirmDialog{
				Active:      tt.active,
				Action:      ActionClose,
				TargetLabel: "test",
			}

			confirmed, handled := dialog.HandleKey(tt.key)

			if confirmed != tt.expectConfirmed {
				t.Errorf("expected confirmed=%v, got %v", tt.expectConfirmed, confirmed)
			}

			if handled != tt.expectHandled {
				t.Errorf("expected handled=%v, got %v", tt.expectHandled, handled)
			}

			if dialog.Active != tt.expectActive {
				t.Errorf("expected active=%v, got %v", tt.expectActive, dialog.Active)
			}
		})
	}
}

func TestConfirmDialog_Render(t *testing.T) {
	t.Run("inactive dialog renders empty", func(t *testing.T) {
		dialog := NewConfirmDialog()

		result := dialog.Render()
		if result != "" {
			t.Errorf("expected empty render, got %s", result)
		}
	})

	t.Run("active dialog renders prompt", func(t *testing.T) {
		dialog := ConfirmDialog{
			Active:      true,
			Action:      ActionClose,
			TargetLabel: "workspace my-workspace",
		}

		result := dialog.Render()
		if result == "" {
			t.Error("expected non-empty render for active dialog")
		}
		// Should contain the workspace name
		if !strings.Contains(result, "my-workspace") {
			t.Error("expected render to contain workspace name")
		}
	})
}

func TestActionDescription(t *testing.T) {
	tests := []struct {
		action   ConfirmAction
		contains string
	}{
		{ActionClose, "close"},
		{ActionPush, "push"},
		{ActionSync, "sync"},
		{ConfirmAction("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			desc := tt.action.ActionDescription()
			if !strings.Contains(desc, tt.contains) {
				t.Errorf("expected description to contain %q, got %s", tt.contains, desc)
			}
		})
	}
}
