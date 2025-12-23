package tui

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

func TestLoadWorkspaces_Success(t *testing.T) {
	t.Parallel()

	model, deps := newTUITestModel(t)
	addTUIWorkspace(deps.storage, domain.Workspace{ID: "ws-1"})
	addTUIWorkspace(deps.storage, domain.Workspace{ID: "ws-2"})

	usageByPath := map[string]int64{
		filepath.Join(deps.config.WorkspacesRoot, "ws-1"): 128,
		filepath.Join(deps.config.WorkspacesRoot, "ws-2"): 256,
	}

	deps.disk.CachedUsageFunc = func(root string) (int64, time.Time, error) {
		if usage, ok := usageByPath[root]; ok {
			return usage, deps.disk.DefaultModTime, nil
		}

		return 0, deps.disk.DefaultModTime, nil
	}
	deps.git.ListFunc = func(_ context.Context) ([]string, error) {
		return nil, nil
	}

	msg := model.loadWorkspaces()

	listMsg, ok := msg.(workspaceListMsg)
	if !ok {
		t.Fatalf("expected workspaceListMsg, got %T", msg)
	}

	if len(listMsg.items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(listMsg.items))
	}

	if listMsg.totalUsage != 384 {
		t.Errorf("expected total usage 384, got %d", listMsg.totalUsage)
	}

	for _, item := range listMsg.items {
		if item.OrphanCheckFailed {
			t.Errorf("expected no orphan check failures for %s", item.Workspace.ID)
		}

		if item.Summary.RepoCount != 0 {
			t.Errorf("expected zero repo count for %s, got %d", item.Workspace.ID, item.Summary.RepoCount)
		}
	}
}

func TestLoadWorkspaces_OrphanCheckFailure(t *testing.T) {
	t.Parallel()

	model, deps := newTUITestModel(t)
	addTUIWorkspace(deps.storage, domain.Workspace{ID: "ws-1"})

	deps.git.ListFunc = func(_ context.Context) ([]string, error) {
		return nil, errors.New("boom")
	}

	msg := model.loadWorkspaces()

	listMsg, ok := msg.(workspaceListMsg)
	if !ok {
		t.Fatalf("expected workspaceListMsg, got %T", msg)
	}

	if len(listMsg.items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(listMsg.items))
	}

	if !listMsg.items[0].OrphanCheckFailed {
		t.Error("expected orphan check failure to be recorded")
	}
}

func TestLoadWorkspaces_ListError(t *testing.T) {
	t.Parallel()

	model, deps := newTUITestModel(t)
	deps.storage.ListFunc = func(_ context.Context) ([]domain.Workspace, error) {
		return nil, errors.New("list failed")
	}

	msg := model.loadWorkspaces()
	if _, ok := msg.(loadWorkspacesErrMsg); !ok {
		t.Fatalf("expected loadWorkspacesErrMsg, got %T", msg)
	}
}

func TestLoadWorkspaceStatus_Success(t *testing.T) {
	t.Parallel()

	model, deps := newTUITestModel(t)
	addTUIWorkspace(deps.storage, domain.Workspace{ID: "ws-1"})

	cmd := model.loadWorkspaceStatus("ws-1")
	msg := cmd()

	statusMsg, ok := msg.(workspaceStatusMsg)
	if !ok {
		t.Fatalf("expected workspaceStatusMsg, got %T", msg)
	}

	if statusMsg.id != "ws-1" {
		t.Errorf("expected id ws-1, got %s", statusMsg.id)
	}

	if statusMsg.status == nil {
		t.Fatal("expected status to be non-nil")
	}
}

func TestLoadWorkspaceDetails_NotFound(t *testing.T) {
	t.Parallel()

	model, _ := newTUITestModel(t)

	cmd := model.loadWorkspaceDetails("missing")

	msg := cmd()
	if _, ok := msg.(workspaceDetailsErrMsg); !ok {
		t.Fatalf("expected workspaceDetailsErrMsg, got %T", msg)
	}
}

func TestLoadWorkspaceDetails_Success(t *testing.T) {
	t.Parallel()

	model, deps := newTUITestModel(t)
	workspace := domain.Workspace{ID: "ws-1"}
	addTUIWorkspace(deps.storage, workspace)
	deps.git.ListFunc = func(_ context.Context) ([]string, error) {
		return nil, nil
	}

	model.workspaces.SetItems([]workspaceItem{
		{
			Workspace: workspace,
			Summary:   workspaceSummary{RepoCount: 0},
		},
	}, 0)

	cmd := model.loadWorkspaceDetails("ws-1")
	msg := cmd()

	detailsMsg, ok := msg.(workspaceDetailsMsg)
	if !ok {
		t.Fatalf("expected workspaceDetailsMsg, got %T", msg)
	}

	if detailsMsg.workspace == nil || detailsMsg.workspace.ID != "ws-1" {
		t.Fatalf("expected workspace ws-1, got %+v", detailsMsg.workspace)
	}

	if detailsMsg.status == nil {
		t.Fatal("expected status to be non-nil")
	}
}
