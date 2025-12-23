package workspaces

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

func TestCreateWorkspace_UsesTemplateDefaultBranch(t *testing.T) {
	t.Parallel()

	deps := newMockService(t)

	template := &config.Template{DefaultBranch: "main"}

	_, err := deps.svc.CreateWorkspaceWithOptions(context.Background(), "ws-1", "", nil, CreateOptions{
		Template: template,
	})
	if err != nil {
		t.Fatalf("CreateWorkspaceWithOptions failed: %v", err)
	}

	ws, ok := deps.storage.Workspaces["ws-1"]
	if !ok {
		t.Fatal("expected workspace to be stored")
	}

	if ws.BranchName != "main" {
		t.Errorf("expected branch name main, got %q", ws.BranchName)
	}

	if len(deps.cache.InvalidateCalls) == 0 {
		t.Error("expected cache to be invalidated")
	}
}

func TestCloseWorkspaceKeepMetadata_ArchivesAndDeletes(t *testing.T) {
	t.Parallel()

	deps := newMockService(t)
	addWorkspaceFixture(deps.storage, domain.Workspace{ID: "ws-1"})

	closed, err := deps.svc.CloseWorkspaceKeepMetadataWithOptions(
		context.Background(),
		"ws-1",
		true,
		CloseOptions{SkipHooks: true},
	)
	if err != nil {
		t.Fatalf("CloseWorkspaceKeepMetadataWithOptions failed: %v", err)
	}

	if closed == nil || closed.Metadata.ClosedAt == nil {
		t.Fatal("expected closed workspace metadata to include closed time")
	}

	if _, ok := deps.storage.Workspaces["ws-1"]; ok {
		t.Fatal("expected workspace to be deleted from storage")
	}

	if len(deps.cache.InvalidateCalls) == 0 {
		t.Error("expected cache invalidation after close")
	}
}

func TestSyncWorkspace_AggregatesResults(t *testing.T) {
	t.Parallel()

	deps := newMockService(t)
	addWorkspaceFixture(deps.storage, domain.Workspace{
		ID:      "ws-1",
		DirName: "ws-1",
		Repos: []domain.Repo{
			{Name: "repo-1", URL: "git@example.com:repo-1.git"},
			{Name: "repo-2", URL: "git@example.com:repo-2.git"},
		},
	})

	deps.config.ParallelWorkers = 1
	deps.git.FetchFunc = func(_ context.Context, name string) error {
		if name == "repo-2" {
			return errors.New("fetch failed")
		}

		return nil
	}
	deps.git.StatusFunc = func(_ context.Context, path string) (bool, int, int, string, error) {
		if strings.Contains(path, "repo-1") {
			return false, 0, 2, "main", nil
		}

		return false, 0, 0, "main", nil
	}
	deps.git.PullFunc = func(_ context.Context, _ string) error {
		return nil
	}

	result, err := deps.svc.SyncWorkspace(context.Background(), "ws-1", SyncOptions{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("SyncWorkspace failed: %v", err)
	}

	if result.TotalUpdated != 2 {
		t.Errorf("expected TotalUpdated 2, got %d", result.TotalUpdated)
	}

	if result.TotalErrors != 1 {
		t.Errorf("expected TotalErrors 1, got %d", result.TotalErrors)
	}

	if len(result.Repos) != 2 {
		t.Fatalf("expected 2 repo results, got %d", len(result.Repos))
	}
}

func TestRenameWorkspace_RenamesBranchAndMetadata(t *testing.T) {
	t.Parallel()

	deps := newMockService(t)
	addWorkspaceFixture(deps.storage, domain.Workspace{
		ID:         "old-id",
		DirName:    "old-id",
		BranchName: "old-id",
		Repos:      []domain.Repo{{Name: "repo-1", URL: "git@example.com:repo-1.git"}},
	})

	var renameCalls []string

	deps.git.RenameBranchFunc = func(_ context.Context, _, oldName, newName string) error {
		renameCalls = append(renameCalls, oldName+"->"+newName)
		return nil
	}

	if err := deps.svc.RenameWorkspace(context.Background(), "old-id", "new-id", true, false); err != nil {
		t.Fatalf("RenameWorkspace failed: %v", err)
	}

	ws, ok := deps.storage.Workspaces["new-id"]
	if !ok {
		t.Fatal("expected workspace to be renamed in storage")
	}

	if ws.BranchName != "new-id" {
		t.Errorf("expected branch name new-id, got %q", ws.BranchName)
	}

	if len(renameCalls) != 1 {
		t.Fatalf("expected 1 rename call, got %d", len(renameCalls))
	}
}

func TestRestoreWorkspace_CreatesWorkspaceAndDeletesClosedEntry(t *testing.T) {
	t.Parallel()

	deps := newMockService(t)
	called := false

	deps.storage.LatestClosedFunc = func(_ context.Context, id string) (*domain.ClosedWorkspace, error) {
		if id != "ws-1" {
			return nil, errors.New("unexpected id")
		}

		return &domain.ClosedWorkspace{
			DirName: "ws-1",
			Path:    "closed/ws-1",
			Metadata: domain.Workspace{
				ID:         "ws-1",
				BranchName: "main",
			},
		}, nil
	}
	deps.storage.DeleteClosedFunc = func(_ context.Context, _ string, _ time.Time) error {
		called = true
		return nil
	}

	if err := deps.svc.RestoreWorkspace(context.Background(), "ws-1", false); err != nil {
		t.Fatalf("RestoreWorkspace failed: %v", err)
	}

	if !called {
		t.Fatal("expected closed entry to be deleted")
	}

	if _, ok := deps.storage.Workspaces["ws-1"]; !ok {
		t.Fatal("expected workspace to be restored")
	}
}
