package workspaces

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

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

		workspaces, err := svc.ListWorkspaces(context.Background())
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
		mockStorage.ListFunc = func(_ context.Context) ([]domain.Workspace, error) {
			return nil, errors.New("storage unavailable")
		}

		mockConfig := mocks.NewMockConfigProvider()
		mockGit := mocks.NewMockGitOperations()

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		_, err := svc.ListWorkspaces(context.Background())
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

		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

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
		mockGit.RunCommandFunc = func(_ context.Context, _ string, _ ...string) (*ports.CommandResult, error) {
			callCount++

			return &ports.CommandResult{
				Stdout:   "output",
				Stderr:   "",
				ExitCode: 0,
			}, nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		results, err := svc.RunGitInWorkspace(context.Background(), "test-ws", []string{"status"}, GitRunOptions{})
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
		mockGit.RunCommandFunc = func(_ context.Context, _ string, _ ...string) (*ports.CommandResult, error) {
			return &ports.CommandResult{
				Stdout:   "parallel-output",
				Stderr:   "",
				ExitCode: 0,
			}, nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		results, err := svc.RunGitInWorkspace(context.Background(), "test-ws", []string{"fetch", "--all"}, GitRunOptions{
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
		mockGit.RunCommandFunc = func(_ context.Context, _ string, _ ...string) (*ports.CommandResult, error) {
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

		results, err := svc.RunGitInWorkspace(context.Background(), "test-ws", []string{"push"}, GitRunOptions{
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
		mockGit.RunCommandFunc = func(_ context.Context, _ string, _ ...string) (*ports.CommandResult, error) {
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

		results, err := svc.RunGitInWorkspace(context.Background(), "test-ws", []string{"push"}, GitRunOptions{
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

		_, err := svc.RunGitInWorkspace(context.Background(), "non-existent", []string{"status"}, GitRunOptions{})
		if err == nil {
			t.Fatal("expected error for non-existent workspace")
		}
	})

	t.Run("parallel early termination on error", func(t *testing.T) {
		t.Parallel()

		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["test-ws"] = domain.Workspace{
			ID:         "test-ws",
			BranchName: "main",
			Repos: []domain.Repo{
				{Name: "repo-a", URL: "https://example.com/repo-a"},
				{Name: "repo-b", URL: "https://example.com/repo-b"},
				{Name: "repo-c", URL: "https://example.com/repo-c"},
				{Name: "repo-d", URL: "https://example.com/repo-d"},
			},
		}

		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.WorkspacesRoot = "/workspaces"

		var completedCount int32

		mockGit := mocks.NewMockGitOperations()
		mockGit.RunCommandFunc = func(ctx context.Context, _ string, _ ...string) (*ports.CommandResult, error) {
			// Check if context was cancelled (from another goroutine's error)
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			// Use atomic to avoid race condition in concurrent goroutines
			count := atomic.AddInt32(&completedCount, 1)
			// Return error for first call
			if count == 1 {
				return nil, errors.New("git error")
			}

			return &ports.CommandResult{
				Stdout:   "success",
				ExitCode: 0,
			}, nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		_, err := svc.RunGitInWorkspace(context.Background(), "test-ws", []string{"fetch"}, GitRunOptions{
			Parallel:        true,
			ContinueOnError: false,
		})
		if err == nil {
			t.Fatal("expected error in parallel execution")
		}
	})

	t.Run("parallel continue on error completes all", func(t *testing.T) {
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
		mockGit.RunCommandFunc = func(_ context.Context, path string, _ ...string) (*ports.CommandResult, error) {
			// Fail repo-b
			if path == "/workspaces/test-ws/repo-b" {
				return nil, errors.New("repo-b failed")
			}

			return &ports.CommandResult{
				Stdout:   "success",
				ExitCode: 0,
			}, nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		results, err := svc.RunGitInWorkspace(context.Background(), "test-ws", []string{"status"}, GitRunOptions{
			Parallel:        true,
			ContinueOnError: true,
		})
		// Should not return error with ContinueOnError
		if err != nil {
			t.Fatalf("unexpected error with ContinueOnError: %v", err)
		}

		// Should have all 3 results
		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}

		// Check that repo-b failed
		for _, r := range results {
			if r.RepoName == "repo-b" && r.Error == nil {
				t.Error("expected repo-b to have error")
			}
		}
	})

	t.Run("parallel with non-zero exit code triggers early termination", func(t *testing.T) {
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

		mockGit := mocks.NewMockGitOperations()
		mockGit.RunCommandFunc = func(_ context.Context, path string, _ ...string) (*ports.CommandResult, error) {
			// First repo returns non-zero exit code
			if path == "/workspaces/test-ws/repo-a" {
				return &ports.CommandResult{
					Stdout:   "",
					Stderr:   "merge conflict",
					ExitCode: 1,
				}, nil
			}

			return &ports.CommandResult{
				Stdout:   "success",
				ExitCode: 0,
			}, nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		_, err := svc.RunGitInWorkspace(context.Background(), "test-ws", []string{"merge"}, GitRunOptions{
			Parallel:        true,
			ContinueOnError: false,
		})
		if err == nil {
			t.Fatal("expected error on non-zero exit code")
		}
	})
}

// TestServiceWithNewMockInterfaces demonstrates using the new mock interfaces
// for HookExecutor, DiskUsage, and WorkspaceCache.
func TestServiceWithNewMockInterfaces(t *testing.T) {
	t.Parallel()

	t.Run("service with mock hook executor", func(t *testing.T) {
		t.Parallel()

		mockHooks := mocks.NewMockHookExecutor()
		mockConfig := mocks.NewMockConfigProvider()
		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewService(mockConfig, mockGit, mockStorage, nil,
			WithHookExecutor(mockHooks),
		)

		// Verify the mock is used by checking it was injected
		if svc.hookExecutor != mockHooks {
			t.Error("mock hook executor was not injected")
		}
	})

	t.Run("service with mock disk usage", func(t *testing.T) {
		t.Parallel()

		mockDiskUsage := mocks.NewMockDiskUsage()
		mockDiskUsage.DefaultUsage = 1024 * 1024 // 1MB
		mockConfig := mocks.NewMockConfigProvider()
		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewService(mockConfig, mockGit, mockStorage, nil,
			WithDiskUsage(mockDiskUsage),
		)

		// Verify the mock is used by checking it was injected
		if svc.diskUsage != mockDiskUsage {
			t.Error("mock disk usage was not injected")
		}
	})

	t.Run("service with mock cache", func(t *testing.T) {
		t.Parallel()

		mockCache := mocks.NewMockWorkspaceCache()
		mockConfig := mocks.NewMockConfigProvider()
		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewService(mockConfig, mockGit, mockStorage, nil,
			WithCache(mockCache),
		)

		// Verify the mock is used by checking it was injected
		if svc.cache != mockCache {
			t.Error("mock cache was not injected")
		}
	})

	t.Run("service with all mock interfaces", func(t *testing.T) {
		t.Parallel()

		mockHooks := mocks.NewMockHookExecutor()
		mockDiskUsage := mocks.NewMockDiskUsage()
		mockCache := mocks.NewMockWorkspaceCache()
		mockConfig := mocks.NewMockConfigProvider()
		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewService(mockConfig, mockGit, mockStorage, nil,
			WithHookExecutor(mockHooks),
			WithDiskUsage(mockDiskUsage),
			WithCache(mockCache),
		)

		// Verify all mocks are used
		if svc.hookExecutor != mockHooks {
			t.Error("mock hook executor was not injected")
		}

		if svc.diskUsage != mockDiskUsage {
			t.Error("mock disk usage was not injected")
		}

		if svc.cache != mockCache {
			t.Error("mock cache was not injected")
		}
	})

	t.Run("default implementations when no options provided", func(t *testing.T) {
		t.Parallel()

		mockConfig := mocks.NewMockConfigProvider()
		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		// Verify defaults are created
		if svc.hookExecutor == nil {
			t.Error("default hook executor was not created")
		}

		if svc.diskUsage == nil {
			t.Error("default disk usage was not created")
		}

		if svc.cache == nil {
			t.Error("default cache was not created")
		}
	})
}

func TestCloseWorkspaceWithUnpushedCommits(t *testing.T) {
	t.Parallel()

	t.Run("close blocked when repo has unpushed commits", func(t *testing.T) {
		t.Parallel()

		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["ws-unpushed"] = domain.Workspace{
			ID:         "ws-unpushed",
			BranchName: "ws-unpushed",
			Repos: []domain.Repo{
				{Name: "repo-a", URL: "https://example.com/repo-a"},
			},
		}

		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.WorkspacesRoot = t.TempDir()

		mockGit := mocks.NewMockGitOperations()
		mockGit.StatusFunc = func(_ context.Context, _ string) (bool, int, int, string, error) {
			// Return: isDirty=false, unpushed=2, behind=0, branch="ws-unpushed"
			return false, 2, 0, "ws-unpushed", nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		// Attempt to close without force - should fail due to unpushed commits
		_, err := svc.CloseWorkspaceKeepMetadata(context.Background(), "ws-unpushed", false)
		if err == nil {
			t.Fatal("expected close to fail when repo has unpushed commits")
		}
	})

	t.Run("close allowed when repo is clean (no dirty, no unpushed)", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["ws-clean"] = domain.Workspace{
			ID:         "ws-clean",
			BranchName: "ws-clean",
			Repos: []domain.Repo{
				{Name: "repo-a", URL: "https://example.com/repo-a"},
			},
		}

		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.WorkspacesRoot = tmpDir

		mockGit := mocks.NewMockGitOperations()
		mockGit.StatusFunc = func(_ context.Context, _ string) (bool, int, int, string, error) {
			// Return: isDirty=false, unpushed=0, behind=0, branch="ws-clean"
			return false, 0, 0, "ws-clean", nil
		}
		mockGit.RemoveWorktreeFunc = func(_ context.Context, _, _ string) error {
			return nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		// Close should succeed
		_, err := svc.CloseWorkspaceKeepMetadata(context.Background(), "ws-clean", false)
		if err != nil {
			t.Fatalf("expected close to succeed when repo is clean: %v", err)
		}
	})

	t.Run("force bypasses unpushed check", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["ws-force"] = domain.Workspace{
			ID:         "ws-force",
			BranchName: "ws-force",
			Repos: []domain.Repo{
				{Name: "repo-a", URL: "https://example.com/repo-a"},
			},
		}

		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.WorkspacesRoot = tmpDir

		mockGit := mocks.NewMockGitOperations()
		mockGit.StatusFunc = func(_ context.Context, _ string) (bool, int, int, string, error) {
			// Return: isDirty=false, unpushed=5, behind=0, branch="ws-force"
			return false, 5, 0, "ws-force", nil
		}
		mockGit.RemoveWorktreeFunc = func(_ context.Context, _, _ string) error {
			return nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil)

		// Force close should succeed despite unpushed commits
		_, err := svc.CloseWorkspaceKeepMetadata(context.Background(), "ws-force", true)
		if err != nil {
			t.Fatalf("expected force close to succeed despite unpushed commits: %v", err)
		}
	})

	t.Run("preview shows unpushed commit warning", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["ws-preview"] = domain.Workspace{
			ID:         "ws-preview",
			BranchName: "ws-preview",
			Repos: []domain.Repo{
				{Name: "repo-a", URL: "https://example.com/repo-a"},
			},
		}

		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.WorkspacesRoot = tmpDir

		mockGit := mocks.NewMockGitOperations()
		mockGit.StatusFunc = func(_ context.Context, _ string) (bool, int, int, string, error) {
			// Return: isDirty=true, unpushed=3, behind=0, branch="ws-preview"
			return true, 3, 0, "ws-preview", nil
		}

		mockDiskUsage := mocks.NewMockDiskUsage()
		mockDiskUsage.CachedUsageFunc = func(_ string) (int64, time.Time, error) {
			return 1024, time.Time{}, nil
		}

		svc := NewService(mockConfig, mockGit, mockStorage, nil, WithDiskUsage(mockDiskUsage))

		preview, err := svc.PreviewCloseWorkspace("ws-preview", true)
		if err != nil {
			t.Fatalf("PreviewCloseWorkspace failed: %v", err)
		}

		if len(preview.RepoStatuses) != 1 {
			t.Fatalf("expected 1 repo status, got %d", len(preview.RepoStatuses))
		}

		status := preview.RepoStatuses[0]
		if status.Name != "repo-a" {
			t.Errorf("expected repo name repo-a, got %s", status.Name)
		}

		if !status.IsDirty {
			t.Error("expected IsDirty=true")
		}

		if status.UnpushedCount != 3 {
			t.Errorf("expected unpushed count 3, got %d", status.UnpushedCount)
		}
	})
}
