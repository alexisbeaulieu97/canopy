package workspaces

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

func TestCanonicalRepoService_List(t *testing.T) {
	t.Parallel()

	t.Run("delegates to git engine", func(t *testing.T) {
		t.Parallel()

		mockGit := mocks.NewMockGitOperations()
		mockGit.ListFunc = func() ([]string, error) {
			return []string{"repo-a", "repo-b"}, nil
		}

		svc := NewCanonicalRepoService(mockGit, nil, "/projects", nil, nil)

		repos, err := svc.List()
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
		mockGit.CloneFunc = func(url, name string) error {
			clonedURL = url
			clonedName = name

			return nil
		}

		svc := NewCanonicalRepoService(mockGit, nil, "/projects", nil, nil)

		name, err := svc.Add("https://github.com/org/my-repo.git")
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

		svc := NewCanonicalRepoService(nil, nil, "/projects", nil, nil)

		_, err := svc.Add("")
		if err == nil {
			t.Fatal("expected error for empty URL")
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

		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewCanonicalRepoService(nil, mockStorage, dir, nil, nil)

		err := svc.Remove("test-repo", false)
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

		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["ws1"] = domain.Workspace{
			ID:    "ws1",
			Repos: []domain.Repo{{Name: "test-repo", URL: "https://github.com/org/test-repo.git"}},
		}

		svc := NewCanonicalRepoService(nil, mockStorage, dir, nil, nil)

		err := svc.Remove("test-repo", false)
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

		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["ws1"] = domain.Workspace{
			ID:    "ws1",
			Repos: []domain.Repo{{Name: "test-repo", URL: "https://github.com/org/test-repo.git"}},
		}

		svc := NewCanonicalRepoService(nil, mockStorage, dir, nil, nil)

		err := svc.Remove("test-repo", true)
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
		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewCanonicalRepoService(nil, mockStorage, dir, nil, nil)

		err := svc.Remove("nonexistent", false)
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
		mockGit.FetchFunc = func(name string) error {
			fetchedName = name

			return nil
		}

		svc := NewCanonicalRepoService(mockGit, nil, "/projects", nil, nil)

		err := svc.Sync("my-repo")
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

		svc := NewCanonicalRepoService(nil, mockStorage, "/projects", nil, nil)

		usedBy, err := svc.GetWorkspacesUsingRepo("shared-repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(usedBy) != 2 {
			t.Errorf("expected 2 workspaces, got %d", len(usedBy))
		}
	})

	t.Run("returns empty when no workspaces use repo", func(t *testing.T) {
		t.Parallel()

		mockStorage := mocks.NewMockWorkspaceStorage()

		svc := NewCanonicalRepoService(nil, mockStorage, "/projects", nil, nil)

		usedBy, err := svc.GetWorkspacesUsingRepo("unused-repo")
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

		mockStorage := mocks.NewMockWorkspaceStorage()
		mockStorage.Workspaces["ws1"] = domain.Workspace{
			ID:    "ws1",
			Repos: []domain.Repo{{Name: "test-repo", URL: "https://github.com/org/test-repo.git"}},
		}

		diskCalc := DefaultDiskUsageCalculator()
		svc := NewCanonicalRepoService(nil, mockStorage, dir, nil, diskCalc)

		preview, err := svc.PreviewRemove("test-repo")
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
