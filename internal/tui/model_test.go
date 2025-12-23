package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

func TestSelectedWorkspaceItem(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)

	ws := domain.Workspace{ID: "ws-1"}
	items := []list.Item{workspaceItem{Workspace: ws}}
	model.ui.List.SetItems(items)

	selected, ok := model.selectedWorkspaceItem()
	if !ok {
		t.Fatal("expected selected workspace item")
	}

	if selected.Workspace.ID != "ws-1" {
		t.Errorf("expected ws-1, got %s", selected.Workspace.ID)
	}
}

func TestWorkspaceItemByID(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)
	model.workspaces.SetItems([]workspaceItem{
		{Workspace: domain.Workspace{ID: "ws-1"}},
		{Workspace: domain.Workspace{ID: "ws-2"}},
	}, 0)

	item, ok := model.workspaceItemByID("ws-2")
	if !ok {
		t.Fatal("expected to find workspace ws-2")
	}

	if item.Workspace.ID != "ws-2" {
		t.Errorf("expected ws-2, got %s", item.Workspace.ID)
	}
}
