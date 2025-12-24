package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

func TestToggleWorkspaceSelectionUpdatesItems(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)

	ws1 := domain.Workspace{ID: "ws-1"}
	ws2 := domain.Workspace{ID: "ws-2"}
	items := []workspaceItem{
		{Workspace: ws1},
		{Workspace: ws2},
	}

	model.workspaces.SetItems(items, 0)
	model.ui.List.SetItems([]list.Item{items[0], items[1]})

	model.toggleWorkspaceSelection("ws-1")

	if !model.selectedIDs["ws-1"] {
		t.Fatal("expected ws-1 to be selected")
	}

	item, ok := model.workspaces.FindItemByID("ws-1")
	if !ok || !item.Selected {
		t.Fatal("expected workspace item to be marked selected")
	}

	listItem := model.ui.List.Items()[0].(workspaceItem)
	if !listItem.Selected {
		t.Fatal("expected list item to be marked selected")
	}
}

func TestSelectAllAndClearSelection(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)

	ws1 := domain.Workspace{ID: "ws-1"}
	ws2 := domain.Workspace{ID: "ws-2"}
	items := []workspaceItem{
		{Workspace: ws1},
		{Workspace: ws2},
	}

	model.workspaces.SetItems(items, 0)
	model.ui.List.SetItems([]list.Item{items[0], items[1]})

	model.selectAllVisible()

	if got := model.selectionCount(); got != 2 {
		t.Fatalf("expected 2 selected workspaces, got %d", got)
	}

	model.clearSelection()

	if got := model.selectionCount(); got != 0 {
		t.Fatalf("expected 0 selected workspaces, got %d", got)
	}
}

func TestActionTargetIDsPrefersSelection(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)

	ws1 := domain.Workspace{ID: "ws-1"}
	ws2 := domain.Workspace{ID: "ws-2"}
	items := []workspaceItem{
		{Workspace: ws1},
		{Workspace: ws2},
	}

	model.workspaces.SetItems(items, 0)
	model.ui.List.SetItems([]list.Item{items[0], items[1]})

	model.selectAllVisible()

	ids := model.actionTargetIDs()
	if len(ids) != 2 || ids[0] != "ws-1" || ids[1] != "ws-2" {
		t.Fatalf("expected selected IDs [ws-1 ws-2], got %v", ids)
	}
}

func TestActionTargetIDsFallsBackToSelectedItem(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)

	ws1 := domain.Workspace{ID: "ws-1"}
	ws2 := domain.Workspace{ID: "ws-2"}
	items := []workspaceItem{
		{Workspace: ws1},
		{Workspace: ws2},
	}

	model.workspaces.SetItems(items, 0)
	model.ui.List.SetItems([]list.Item{items[0], items[1]})

	ids := model.actionTargetIDs()
	if len(ids) != 1 || ids[0] != "ws-1" {
		t.Fatalf("expected selected ID [ws-1], got %v", ids)
	}
}
