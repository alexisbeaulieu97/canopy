package workspaces

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
	"github.com/alexisbeaulieu97/canopy/internal/storage"
)

func TestCreateWorkspaceRollbackOnCloneFailure(t *testing.T) {
	t.Parallel()

	workspacesRoot := t.TempDir()
	projectsRoot := t.TempDir()
	closedRoot := t.TempDir()

	mockConfig := mocks.NewMockConfigProvider()
	mockConfig.WorkspacesRoot = workspacesRoot
	mockConfig.ProjectsRoot = projectsRoot
	mockConfig.ClosedRoot = closedRoot

	mockGit := mocks.NewMockGitOperations()
	mockGit.EnsureCanonicalFunc = func(_ context.Context, _, _ string) (*git.Repository, error) {
		return nil, nil
	}
	mockGit.CreateWorktreeFunc = func(_ context.Context, _, worktreePath, _ string) error {
		if err := os.MkdirAll(worktreePath, 0o750); err != nil {
			return err
		}

		return errors.New("worktree create failed")
	}
	mockGit.RemoveWorktreeFunc = func(_ context.Context, _, worktreePath string) error {
		return os.RemoveAll(worktreePath)
	}

	wsEngine := storage.New(workspacesRoot, closedRoot)
	svc := NewService(mockConfig, mockGit, wsEngine, nil)

	repos := []domain.Repo{{Name: "repo-a", URL: "https://example.com/repo-a.git"}}

	_, err := svc.CreateWorkspace(context.Background(), "FAIL-CLONE", "", repos)
	if err == nil {
		t.Fatalf("expected error")
	}

	workspacePath := filepath.Join(workspacesRoot, "FAIL-CLONE")
	if _, statErr := os.Stat(workspacePath); !os.IsNotExist(statErr) {
		t.Fatalf("expected workspace directory to be removed, got %v", statErr)
	}
}

func TestAddRepoRollbackOnSaveFailure(t *testing.T) {
	t.Parallel()

	workspacesRoot := t.TempDir()

	mockConfig := mocks.NewMockConfigProvider()
	mockConfig.WorkspacesRoot = workspacesRoot
	mockConfig.ClosedRoot = t.TempDir()

	mockStorage := mocks.NewMockWorkspaceStorage()
	mockStorage.Workspaces["WS-1"] = domain.Workspace{ID: "WS-1", BranchName: "main"}
	mockStorage.SaveFunc = func(_ context.Context, _ domain.Workspace) error {
		return errors.New("save failed")
	}

	mockGit := mocks.NewMockGitOperations()
	mockGit.EnsureCanonicalFunc = func(_ context.Context, _, _ string) (*git.Repository, error) {
		return nil, nil
	}
	mockGit.CreateWorktreeFunc = func(_ context.Context, _, worktreePath, _ string) error {
		return os.MkdirAll(worktreePath, 0o750)
	}
	mockGit.RemoveWorktreeFunc = func(_ context.Context, _, worktreePath string) error {
		return os.RemoveAll(worktreePath)
	}

	svc := NewService(mockConfig, mockGit, mockStorage, nil)

	err := svc.AddRepoToWorkspace(context.Background(), "WS-1", "org/repo-a")
	if err == nil {
		t.Fatalf("expected error")
	}

	ws := mockStorage.Workspaces["WS-1"]
	if len(ws.Repos) != 0 {
		t.Fatalf("expected workspace metadata to remain unchanged")
	}

	worktreePath := filepath.Join(workspacesRoot, "WS-1", "repo-a")
	if _, statErr := os.Stat(worktreePath); !os.IsNotExist(statErr) {
		t.Fatalf("expected worktree directory to be removed, got %v", statErr)
	}
}

func TestRestoreWorkspaceRollbackOnRecreateFailure(t *testing.T) {
	t.Parallel()

	workspacesRoot := t.TempDir()

	mockConfig := mocks.NewMockConfigProvider()
	mockConfig.WorkspacesRoot = workspacesRoot

	closedAt := time.Now().UTC().Add(-1 * time.Hour)
	mockStorage := mocks.NewMockWorkspaceStorage()
	mockStorage.LatestClosedFunc = func(_ context.Context, _ string) (*domain.ClosedWorkspace, error) {
		return &domain.ClosedWorkspace{
			DirName: "WS-RESTORE",
			Path:    filepath.Join(mockConfig.ClosedRoot, "WS-RESTORE"),
			Metadata: domain.Workspace{
				ID:         "WS-RESTORE",
				BranchName: "main",
				Repos:      []domain.Repo{{Name: "repo-a", URL: "https://example.com/repo-a.git"}},
				ClosedAt:   &closedAt,
			},
		}, nil
	}
	deleteCalled := false
	mockStorage.DeleteClosedFunc = func(_ context.Context, _ string, _ time.Time) error {
		deleteCalled = true
		return nil
	}
	mockStorage.CreateFunc = func(_ context.Context, _ domain.Workspace) error {
		return errors.New("create failed")
	}

	mockGit := mocks.NewMockGitOperations()
	mockGit.EnsureCanonicalFunc = func(_ context.Context, _, _ string) (*git.Repository, error) {
		return nil, nil
	}
	mockGit.CreateWorktreeFunc = func(_ context.Context, _, worktreePath, _ string) error {
		if err := os.MkdirAll(worktreePath, 0o750); err != nil {
			return err
		}

		return errors.New("worktree create failed")
	}
	mockGit.RemoveWorktreeFunc = func(_ context.Context, _, worktreePath string) error {
		return os.RemoveAll(worktreePath)
	}

	svc := NewService(mockConfig, mockGit, mockStorage, nil)

	err := svc.RestoreWorkspace(context.Background(), "WS-RESTORE", false)
	if err == nil {
		t.Fatalf("expected error")
	}

	if deleteCalled {
		t.Fatalf("expected closed entry to remain intact")
	}

	workspacePath := filepath.Join(workspacesRoot, "WS-RESTORE")
	if _, statErr := os.Stat(workspacePath); !os.IsNotExist(statErr) {
		t.Fatalf("expected workspace directory to be removed, got %v", statErr)
	}
}
