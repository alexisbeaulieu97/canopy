package workspaces

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

func TestNewCanonicalRepoService_PanicsOnNilDependencies(t *testing.T) {
	t.Parallel()

	t.Run("panics on nil gitEngine", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for nil gitEngine")
			}
		}()

		mockStorage := mocks.NewMockWorkspaceStorage()
		NewCanonicalRepoService(nil, mockStorage, "/projects", nil, nil, nil)
	})

	t.Run("panics on nil wsStorage", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for nil wsStorage")
			}
		}()

		mockGit := mocks.NewMockGitOperations()
		NewCanonicalRepoService(mockGit, nil, "/projects", nil, nil, nil)
	})
}

func TestCanonicalRepoService_List(t *testing.T) {
	t.Parallel()

	t.Run("delegates to git engine", func(t *testing.T) {
		t.Parallel()

		mockGit := mocks.NewMockGitOperations()
		mockGit.ListFunc = func(_ context.Context) ([]string, error) {
			return []string{"repo-a", "repo-b"}, nil
		}

		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewCanonicalRepoService(mockGit, mockStorage, "/projects", nil, nil, nil)

		repos, err := svc.List(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(repos) != 2 {
			t.Errorf("expected 2 repos, got %d", len(repos))
		}
	})
}

func TestCanonicalRepoService_Add(t *testing.T) {
	t.Parallel()

	t.Run("extracts name and clones", func(t *testing.T) {
		t.Parallel()

		var clonedURL, clonedName string

		mockGit := mocks.NewMockGitOperations()
		mockGit.CloneFunc = func(_ context.Context, url, name string) error {
			clonedURL = url
			clonedName = name

			return nil
		}

		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewCanonicalRepoService(mockGit, mockStorage, "/projects", nil, nil, nil)

		name, err := svc.Add(context.Background(), "https://github.com/org/my-repo.git")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if name != "my-repo" {
			t.Errorf("expected name 'my-repo', got %q", name)
		}

		if clonedURL != "https://github.com/org/my-repo.git" {
			t.Errorf("expected URL passed to clone, got %q", clonedURL)
		}

		if clonedName != "my-repo" {
			t.Errorf("expected name passed to clone, got %q", clonedName)
		}
	})

	t.Run("returns error for invalid URL", func(t *testing.T) {
		t.Parallel()

		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewCanonicalRepoService(mockGit, mockStorage, "/projects", nil, nil, nil)

		_, err := svc.Add(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty URL")
		}
	})

	t.Run("returns empty name on clone error", func(t *testing.T) {
		t.Parallel()

		mockGit := mocks.NewMockGitOperations()
		mockGit.CloneFunc = func(context.Context, string, string) error {
			return os.ErrPermission
		}

		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewCanonicalRepoService(mockGit, mockStorage, "/projects", nil, nil, nil)

		name, err := svc.Add(context.Background(), "https://github.com/org/my-repo.git")
		if err == nil {
			t.Fatal("expected error from clone")
		}

		if name != "" {
			t.Errorf("expected empty name on error, got %q", name)
		}
	})

	t.Run("rolls back on registry save failure", func(t *testing.T) {
		t.Parallel()

		projectsRoot := t.TempDir()

		registryDir := filepath.Join(t.TempDir(), "registry")
		if err := os.MkdirAll(registryDir, 0o500); err != nil {
			t.Fatalf("failed to create registry dir: %v", err)
		}

		registryPath := filepath.Join(registryDir, "repos.yaml")

		registry, err := config.LoadRepoRegistry(registryPath)
		if err != nil {
			t.Fatalf("failed to load registry: %v", err)
		}

		mockGit := mocks.NewMockGitOperations()
		mockGit.CloneFunc = func(_ context.Context, _, name string) error {
			return os.MkdirAll(filepath.Join(projectsRoot, name), 0o750)
		}

		mockStorage := mocks.NewMockWorkspaceStorage()
		svc := NewCanonicalRepoService(mockGit, mockStorage, projectsRoot, nil, nil, registry)

		_, err = svc.Add(context.Background(), "https://github.com/org/my-repo.git")
		if err == nil {
			t.Fatalf("expected error from registry save failure")
		}

		repoPath := filepath.Join(projectsRoot, "my-repo")
		if _, statErr := os.Stat(repoPath); !os.IsNotExist(statErr) {
			t.Fatalf("expected repo to be removed, got %v", statErr)
		}
	})
}

func TestCanonicalRepoService_Remove(t *testing.T) {
	t.Parallel()

	t.Run("removes repo not in use", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		repoPath := filepath.Join(dir, "test-repo")
		if err := os.MkdirAll(repoPath, 0o750); err != nil {
			t.Fatalf("failed to create repo dir: %v", err)
		}

		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewCanonicalRepoService(mockGit, mockStorage, dir, nil, nil, nil)

		err := svc.Remove(context.Background(), "test-repo", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
			t.Error("expected repo directory to be removed")
		}
	})

	t.Run("returns error when repo in use and not forced", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		repoPath := filepath.Join(dir, "test-repo")
		if err := os.MkdirAll(repoPath, 0o750); err != nil {
			t.Fatalf("failed to create repo dir: %v", err)
		}

		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["ws1"] = domain.Workspace{
			ID:    "ws1",
			Repos: []domain.Repo{{Name: "test-repo", URL: "https://github.com/org/test-repo.git"}},
		}

		svc := NewCanonicalRepoService(mockGit, mockStorage, dir, nil, nil, nil)

		err := svc.Remove(context.Background(), "test-repo", false)
		if err == nil {
			t.Fatal("expected error when repo is in use")
		}
	})

	t.Run("removes repo when in use with force", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		repoPath := filepath.Join(dir, "test-repo")
		if err := os.MkdirAll(repoPath, 0o750); err != nil {
			t.Fatalf("failed to create repo dir: %v", err)
		}

		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["ws1"] = domain.Workspace{
			ID:    "ws1",
			Repos: []domain.Repo{{Name: "test-repo", URL: "https://github.com/org/test-repo.git"}},
		}

		svc := NewCanonicalRepoService(mockGit, mockStorage, dir, nil, nil, nil)

		err := svc.Remove(context.Background(), "test-repo", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
			t.Error("expected repo directory to be removed")
		}
	})

	t.Run("returns error for non-existent repo", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewCanonicalRepoService(mockGit, mockStorage, dir, nil, nil, nil)

		err := svc.Remove(context.Background(), "nonexistent", false)
		if err == nil {
			t.Fatal("expected error for non-existent repo")
		}
	})
}

func TestCanonicalRepoService_Sync(t *testing.T) {
	t.Parallel()

	t.Run("delegates to git engine", func(t *testing.T) {
		t.Parallel()

		var fetchedName string

		mockGit := mocks.NewMockGitOperations()
		mockGit.FetchFunc = func(_ context.Context, name string) error {
			fetchedName = name

			return nil
		}

		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewCanonicalRepoService(mockGit, mockStorage, "/projects", nil, nil, nil)

		err := svc.Sync(context.Background(), "my-repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if fetchedName != "my-repo" {
			t.Errorf("expected fetch for 'my-repo', got %q", fetchedName)
		}
	})
}

func TestCanonicalRepoService_GetWorkspacesUsingRepo(t *testing.T) {
	t.Parallel()

	t.Run("finds workspaces using repo", func(t *testing.T) {
		t.Parallel()

		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["ws1"] = domain.Workspace{
			ID:    "ws1",
			Repos: []domain.Repo{{Name: "shared-repo", URL: "https://github.com/org/shared-repo.git"}},
		}
		mockStorage.Workspaces["ws2"] = domain.Workspace{
			ID:    "ws2",
			Repos: []domain.Repo{{Name: "other-repo", URL: "https://github.com/org/other-repo.git"}},
		}
		mockStorage.Workspaces["ws3"] = domain.Workspace{
			ID:    "ws3",
			Repos: []domain.Repo{{Name: "shared-repo", URL: "https://github.com/org/shared-repo.git"}},
		}

		svc := NewCanonicalRepoService(mockGit, mockStorage, "/projects", nil, nil, nil)

		usedBy, err := svc.GetWorkspacesUsingRepo(context.Background(), "shared-repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(usedBy) != 2 {
			t.Errorf("expected 2 workspaces, got %d", len(usedBy))
		}
	})

	t.Run("returns empty when no workspaces use repo", func(t *testing.T) {
		t.Parallel()

		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewCanonicalRepoService(mockGit, mockStorage, "/projects", nil, nil, nil)

		usedBy, err := svc.GetWorkspacesUsingRepo(context.Background(), "unused-repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(usedBy) != 0 {
			t.Errorf("expected empty list, got %d", len(usedBy))
		}
	})
}

func TestCanonicalRepoService_PreviewRemove(t *testing.T) {
	t.Parallel()

	t.Run("returns preview with usage info", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		repoPath := filepath.Join(dir, "test-repo")
		if err := os.MkdirAll(repoPath, 0o750); err != nil {
			t.Fatalf("failed to create repo dir: %v", err)
		}

		// Create a file to have some disk usage
		testFile := filepath.Join(repoPath, "README.md")
		if err := os.WriteFile(testFile, []byte("# Test"), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		mockGit := mocks.NewMockGitOperations()
		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["ws1"] = domain.Workspace{
			ID:    "ws1",
			Repos: []domain.Repo{{Name: "test-repo", URL: "https://github.com/org/test-repo.git"}},
		}

		diskCalc := DefaultDiskUsageCalculator()
		svc := NewCanonicalRepoService(mockGit, mockStorage, dir, nil, diskCalc, nil)

		preview, err := svc.PreviewRemove(context.Background(), "test-repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if preview.RepoName != "test-repo" {
			t.Errorf("expected repo name 'test-repo', got %q", preview.RepoName)
		}

		if preview.RepoPath != repoPath {
			t.Errorf("expected path %q, got %q", repoPath, preview.RepoPath)
		}

		if len(preview.WorkspacesAffected) != 1 {
			t.Errorf("expected 1 workspace affected, got %d", len(preview.WorkspacesAffected))
		}

		if preview.DiskUsageBytes == 0 {
			t.Error("expected non-zero disk usage")
		}
	})
}
