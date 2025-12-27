package tui

import (
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/tui/components"
)

func TestHandleDetailKeyWithState_PushShowsConfirm(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)
	state := &DetailViewState{WorkspaceID: "ws-1"}

	next, _, handled := model.handleDetailKeyWithState(state, firstKey(model.ui.Keybindings.Push))
	if !handled {
		t.Fatal("expected push key to be handled in detail view")
	}

	confirm, ok := next.(*ConfirmViewState)
	if !ok {
		t.Fatalf("expected ConfirmViewState, got %T", next)
	}

	if confirm.Action != components.ActionPush {
		t.Fatalf("expected ActionPush, got %s", confirm.Action)
	}

	if len(confirm.TargetIDs) != 1 || confirm.TargetIDs[0] != "ws-1" {
		t.Fatalf("expected target ws-1, got %v", confirm.TargetIDs)
	}

	if confirm.Parent != state {
		t.Fatal("expected confirm state to preserve detail parent")
	}
}

func TestHandleConfirmKeyWithState_FromDetailReturnsDetail(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)
	parent := &DetailViewState{WorkspaceID: "ws-1"}
	state := &ConfirmViewState{
		Action:    components.ActionPush,
		TargetIDs: []string{"ws-1"},
		Parent:    parent,
	}

	next, _, handled := model.handleConfirmKeyWithState(state, firstKey(model.ui.Keybindings.Confirm))
	if !handled {
		t.Fatal("expected confirm key to be handled")
	}

	if next != parent {
		t.Fatalf("expected to return to detail view, got %T", next)
	}
}

func TestHandleConfirmKeyWithState_CloseReturnsList(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)
	parent := &DetailViewState{WorkspaceID: "ws-1"}
	state := &ConfirmViewState{
		Action:    components.ActionClose,
		TargetIDs: []string{"ws-1"},
		Parent:    parent,
	}

	next, _, handled := model.handleConfirmKeyWithState(state, firstKey(model.ui.Keybindings.Confirm))
	if !handled {
		t.Fatal("expected confirm key to be handled")
	}

	if _, ok := next.(*ListViewState); !ok {
		t.Fatalf("expected ListViewState after close, got %T", next)
	}
}
