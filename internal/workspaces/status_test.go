package workspaces

import (
	"context"
	"errors"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

func TestGetStatusSetsRepoError(t *testing.T) {
	t.Parallel()

	mockStorage := mocks.NewMockWorkspaceStorage()
	mockStorage.Workspaces["WS-ERROR"] = domain.Workspace{
		ID:         "WS-ERROR",
		BranchName: "main",
		Repos:      []domain.Repo{{Name: "repo-a"}},
	}

	mockGit := mocks.NewMockGitOperations()
	statusErr := errors.New("status failed")
	mockGit.StatusFunc = func(_ context.Context, _ string) (bool, int, int, string, error) {
		return false, 0, 0, "", statusErr
	}

	mockConfig := mocks.NewMockConfigProvider()
	mockConfig.WorkspacesRoot = t.TempDir()

	svc := NewService(mockConfig, mockGit, mockStorage, nil)

	status, err := svc.GetStatus(context.Background(), "WS-ERROR")
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if len(status.Repos) != 1 {
		t.Fatalf("expected 1 repo status, got %d", len(status.Repos))
	}

	repoStatus := status.Repos[0]
	if repoStatus.Error != domain.StatusError(statusErr.Error()) {
		t.Fatalf("expected error %q, got %q", statusErr.Error(), repoStatus.Error)
	}

	if repoStatus.Branch != "" {
		t.Fatalf("expected empty branch on error, got %q", repoStatus.Branch)
	}
}

func TestGetStatusSetsTimeoutError(t *testing.T) {
	t.Parallel()

	mockStorage := mocks.NewMockWorkspaceStorage()
	mockStorage.Workspaces["WS-TIMEOUT"] = domain.Workspace{
		ID:         "WS-TIMEOUT",
		BranchName: "main",
		Repos:      []domain.Repo{{Name: "repo-a"}},
	}

	mockGit := mocks.NewMockGitOperations()
	mockGit.StatusFunc = func(_ context.Context, _ string) (bool, int, int, string, error) {
		return false, 0, 0, "", context.DeadlineExceeded
	}

	mockConfig := mocks.NewMockConfigProvider()
	mockConfig.WorkspacesRoot = t.TempDir()

	svc := NewService(mockConfig, mockGit, mockStorage, nil)

	status, err := svc.GetStatus(context.Background(), "WS-TIMEOUT")
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	repoStatus := status.Repos[0]
	if repoStatus.Error != domain.StatusErrorTimeout {
		t.Fatalf("expected timeout error, got %q", repoStatus.Error)
	}
}
