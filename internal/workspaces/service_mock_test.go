package workspaces

import (
	"errors"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
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

func TestRunGitInWorkspace(t *testing.T) {
	t.Parallel()

	t.Run("basic command execution", func(t *testing.T) {
		t.Parallel()

		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["test-ws"] = domain.Workspace{
			ID:         "test-ws",
			BranchName: "main",
			Repos: []domain.Repo{
				{Name: "repo-a", URL: "https://example.com/repo-a"},
				{Name: "repo-b", URL: "https://example.com/repo-b"},
			},
		}

		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.WorkspacesRoot = "/workspaces"

		callCount := 0
		mockGit := mocks.NewMockGitOperations()
		mockGit.RunCommandFunc = func(repoPath string, args ...string) (*ports.CommandResult, error) {
			callCount++

			return &ports.CommandResult{
				Stdout:   "output",
				Stderr:   "",
				ExitCode: 0,
			}, nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		results, err := svc.RunGitInWorkspace("test-ws", []string{"status"}, GitRunOptions{})
		if err != nil {
			t.Fatalf("RunGitInWorkspace failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}

		if callCount != 2 {
			t.Errorf("expected 2 git calls, got %d", callCount)
		}

		for _, r := range results {
			if r.Stdout != "output" {
				t.Errorf("expected stdout 'output', got '%s'", r.Stdout)
			}

			if r.ExitCode != 0 {
				t.Errorf("expected exit code 0, got %d", r.ExitCode)
			}
		}
	})

	t.Run("parallel execution", func(t *testing.T) {
		t.Parallel()

		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["test-ws"] = domain.Workspace{
			ID:         "test-ws",
			BranchName: "main",
			Repos: []domain.Repo{
				{Name: "repo-a", URL: "https://example.com/repo-a"},
				{Name: "repo-b", URL: "https://example.com/repo-b"},
				{Name: "repo-c", URL: "https://example.com/repo-c"},
			},
		}

		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.WorkspacesRoot = "/workspaces"

		mockGit := mocks.NewMockGitOperations()
		mockGit.RunCommandFunc = func(repoPath string, args ...string) (*ports.CommandResult, error) {
			return &ports.CommandResult{
				Stdout:   "parallel-output",
				Stderr:   "",
				ExitCode: 0,
			}, nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		results, err := svc.RunGitInWorkspace("test-ws", []string{"fetch", "--all"}, GitRunOptions{
			Parallel: true,
		})
		if err != nil {
			t.Fatalf("RunGitInWorkspace parallel failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}

		for _, r := range results {
			if r.Stdout != "parallel-output" {
				t.Errorf("expected stdout 'parallel-output', got '%s'", r.Stdout)
			}
		}
	})

	t.Run("error handling - stop on first error", func(t *testing.T) {
		t.Parallel()

		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["test-ws"] = domain.Workspace{
			ID:         "test-ws",
			BranchName: "main",
			Repos: []domain.Repo{
				{Name: "repo-a", URL: "https://example.com/repo-a"},
				{Name: "repo-b", URL: "https://example.com/repo-b"},
			},
		}

		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.WorkspacesRoot = "/workspaces"

		callCount := 0
		mockGit := mocks.NewMockGitOperations()
		mockGit.RunCommandFunc = func(repoPath string, args ...string) (*ports.CommandResult, error) {
			callCount++
			// First call fails
			if callCount == 1 {
				return &ports.CommandResult{
					Stdout:   "",
					Stderr:   "error",
					ExitCode: 1,
				}, nil
			}

			return &ports.CommandResult{
				Stdout:   "success",
				ExitCode: 0,
			}, nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		results, err := svc.RunGitInWorkspace("test-ws", []string{"push"}, GitRunOptions{
			ContinueOnError: false,
		})

		if err == nil {
			t.Fatal("expected error when repo fails")
		}

		// Should have stopped at first failure
		if len(results) != 1 {
			t.Errorf("expected 1 result (stopped at first error), got %d", len(results))
		}
	})

	t.Run("continue on error", func(t *testing.T) {
		t.Parallel()

		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["test-ws"] = domain.Workspace{
			ID:         "test-ws",
			BranchName: "main",
			Repos: []domain.Repo{
				{Name: "repo-a", URL: "https://example.com/repo-a"},
				{Name: "repo-b", URL: "https://example.com/repo-b"},
			},
		}

		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.WorkspacesRoot = "/workspaces"

		callCount := 0
		mockGit := mocks.NewMockGitOperations()
		mockGit.RunCommandFunc = func(repoPath string, args ...string) (*ports.CommandResult, error) {
			callCount++
			// First call fails
			if callCount == 1 {
				return &ports.CommandResult{
					Stdout:   "",
					Stderr:   "error",
					ExitCode: 1,
				}, nil
			}

			return &ports.CommandResult{
				Stdout:   "success",
				ExitCode: 0,
			}, nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		results, err := svc.RunGitInWorkspace("test-ws", []string{"push"}, GitRunOptions{
			ContinueOnError: true,
		})

		// Should not return error with ContinueOnError
		if err != nil {
			t.Fatalf("unexpected error with ContinueOnError: %v", err)
		}

		// Should have all results
		if len(results) != 2 {
			t.Errorf("expected 2 results (continued on error), got %d", len(results))
		}

		// First should have failed
		if results[0].ExitCode != 1 {
			t.Errorf("expected first repo exit code 1, got %d", results[0].ExitCode)
		}

		// Second should have succeeded
		if results[1].ExitCode != 0 {
			t.Errorf("expected second repo exit code 0, got %d", results[1].ExitCode)
		}
	})

	t.Run("workspace not found", func(t *testing.T) {
		t.Parallel()

		mockStorage := mocks.NewMockWorkspaceStorage()
		mockConfig := mocks.NewMockConfigProvider()
		mockGit := mocks.NewMockGitOperations()

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		_, err := svc.RunGitInWorkspace("non-existent", []string{"status"}, GitRunOptions{})
		if err == nil {
			t.Fatal("expected error for non-existent workspace")
		}
	})
}
