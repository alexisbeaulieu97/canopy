package components

import (
	"strings"
	"testing"
)

func TestConfirmDialog_Render(t *testing.T) {
	t.Run("inactive dialog renders empty", func(t *testing.T) {
		dialog := ConfirmDialog{}

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
