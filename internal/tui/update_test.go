package tui

import (
	"errors"
	"testing"

	"github.com/charmbracelet/bubbles/list"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

func TestUpdate_WorkspaceListMessage(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)

	items := []workspaceItem{
		{Workspace: domain.Workspace{ID: "ws-1"}},
		{Workspace: domain.Workspace{ID: "ws-2"}},
	}

	updatedModel, _ := model.Update(workspaceListMsg{items: items, totalUsage: 123})
	updated := updatedModel.(Model)

	if got := updated.workspaces.TotalDiskUsage(); got != 123 {
		t.Errorf("expected total usage 123, got %d", got)
	}

	if got := len(updated.ui.List.Items()); got != 2 {
		t.Errorf("expected 2 list items, got %d", got)
	}
}

func TestUpdate_WorkspaceStatusMessage(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)

	ws := domain.Workspace{ID: "ws-1"}
	model.workspaces.SetItems([]workspaceItem{{Workspace: ws}}, 0)
	model.ui.List.SetItems([]list.Item{workspaceItem{Workspace: ws}})

	status := &domain.WorkspaceStatus{ID: "ws-1"}

	updatedModel, _ := model.Update(workspaceStatusMsg{id: "ws-1", status: status})
	updated := updatedModel.(Model)

	item, ok := updated.workspaces.FindItemByID("ws-1")
	if !ok {
		t.Fatal("expected workspace item to be updated")
	}

	if !item.Loaded {
		t.Error("expected item to be marked as loaded")
	}

	listItem := updated.ui.List.Items()[0].(workspaceItem)
	if !listItem.Loaded {
		t.Error("expected list item to be marked as loaded")
	}
}

func TestUpdate_WorkspaceDetailsMessage(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)
	model.viewState = &DetailViewState{Loading: true}

	ws := &domain.Workspace{ID: "ws-1"}
	status := &domain.WorkspaceStatus{ID: "ws-1"}

	updatedModel, _ := model.Update(workspaceDetailsMsg{workspace: ws, status: status})
	updated := updatedModel.(Model)

	if updated.selectedWS == nil || updated.selectedWS.ID != "ws-1" {
		t.Fatalf("expected selected workspace to be set, got %+v", updated.selectedWS)
	}

	if updated.wsStatus == nil {
		t.Fatal("expected workspace status to be set")
	}

	if ds := updated.getDetailState(); ds == nil || ds.Loading {
		t.Error("expected detail state to stop loading")
	}
}

func TestUpdate_PushResultMessage(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)
	model.pushing = true
	model.pushTarget = "ws-1"

	updatedModel, _ := model.Update(pushResultMsg{id: "ws-1"})
	updated := updatedModel.(Model)

	if updated.pushing {
		t.Error("expected pushing to be false")
	}

	if updated.pushTarget != "" {
		t.Errorf("expected pushTarget to be cleared, got %q", updated.pushTarget)
	}

	if updated.infoMessage == "" {
		t.Error("expected info message to be set")
	}
}

func TestUpdate_ErrorMessage(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)
	model.viewState = &DetailViewState{Loading: true}

	err := errors.New("boom")
	updatedModel, _ := model.Update(err)
	updated := updatedModel.(Model)

	if updated.err == nil {
		t.Fatal("expected error to be stored")
	}

	if ds := updated.getDetailState(); ds == nil || ds.Loading {
		t.Error("expected detail state to stop loading")
	}
}
