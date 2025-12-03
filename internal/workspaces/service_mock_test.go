package workspaces

import (
	"errors"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

// TestServiceWithMocks demonstrates using mock implementations for unit testing.
func TestServiceWithMocks(t *testing.T) {
	t.Parallel()

	t.Run("ListWorkspaces uses storage", func(t *testing.T) {
		t.Parallel()

		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["test-ws"] = domain.Workspace{
			ID:         "TEST-123",
			BranchName: "TEST-123",
		}

		mockConfig := mocks.NewMockConfigProvider()
		mockGit := mocks.NewMockGitOperations()

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		workspaces, err := svc.ListWorkspaces()
		if err != nil {
			t.Fatalf("ListWorkspaces failed: %v", err)
		}

		if len(workspaces) != 1 {
			t.Errorf("expected 1 workspace, got %d", len(workspaces))
		}

		if workspaces[0].ID != "TEST-123" {
			t.Errorf("expected ID TEST-123, got %s", workspaces[0].ID)
		}
	})

	t.Run("ListWorkspaces handles storage error", func(t *testing.T) {
		t.Parallel()

		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.ListFunc = func() (map[string]domain.Workspace, error) {
			return nil, errors.New("storage unavailable")
		}

		mockConfig := mocks.NewMockConfigProvider()
		mockGit := mocks.NewMockGitOperations()

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		_, err := svc.ListWorkspaces()
		if err == nil {
			t.Fatal("expected error when storage fails")
		}

		if err.Error() != "storage unavailable" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("ResolveRepos uses config pattern matching", func(t *testing.T) {
		t.Parallel()

		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.GetReposForWorkspaceFunc = func(workspaceID string) []string {
			if workspaceID == "TEST-123" {
				return []string{"org/repo-a", "org/repo-b"}
			}

			return nil
		}

		svc := NewService(mockConfig, nil, nil, nil)

		repos, err := svc.ResolveRepos("TEST-123", nil)
		if err != nil {
			t.Fatalf("ResolveRepos failed: %v", err)
		}

		if len(repos) != 2 {
			t.Errorf("expected 2 repos, got %d", len(repos))
		}
	})
}
