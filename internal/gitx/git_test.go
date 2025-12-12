package gitx

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// Helper function to create a test repository with commits
func createTestRepo(t *testing.T, path string, bare bool) *git.Repository {
	t.Helper()

	repo, err := git.PlainInit(path, bare)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	if !bare {
		// Create an initial commit
		wt, err := repo.Worktree()
		if err != nil {
			t.Fatalf("failed to get worktree: %v", err)
		}

		// Create a file
		filePath := filepath.Join(path, "README.md")
		if err := os.WriteFile(filePath, []byte("# Test\n"), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		// Add and commit
		if _, err := wt.Add("README.md"); err != nil {
			t.Fatalf("failed to add file: %v", err)
		}

		_, err = wt.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test",
				Email: "test@test.com",
			},
		})
		if err != nil {
			t.Fatalf("failed to commit: %v", err)
		}
	}

	return repo
}

// Helper to create a bare clone from a non-bare repo
func cloneToBare(t *testing.T, sourceRepo *git.Repository, destPath string) *git.Repository {
	t.Helper()

	// Get source worktree root path
	wt, err := sourceRepo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	repo, err := git.PlainClone(destPath, true, &git.CloneOptions{
		URL: wt.Filesystem.Root(),
	})
	if err != nil {
		t.Fatalf("failed to clone to bare: %v", err)
	}

	return repo
}

func TestGitEngine_New(t *testing.T) {
	t.Parallel()

	engine := New("/tmp/projects")
	if engine.ProjectsRoot != "/tmp/projects" {
		t.Errorf("expected ProjectsRoot to be /tmp/projects, got %s", engine.ProjectsRoot)
	}
}

func TestGitEngine_Clone(t *testing.T) {
	t.Parallel()

	t.Run("clones repository as bare", func(t *testing.T) {
		t.Parallel()

		// Create source repo
		sourceDir := t.TempDir()
		sourcePath := filepath.Join(sourceDir, "source")
		createTestRepo(t, sourcePath, false)

		// Create projects root
		projectsRoot := t.TempDir()
		engine := New(projectsRoot)

		// Clone
		err := engine.Clone(context.Background(), sourcePath, "test-repo")
		if err != nil {
			t.Fatalf("Clone failed: %v", err)
		}

		// Verify it's a bare repo
		clonedPath := filepath.Join(projectsRoot, "test-repo")

		repo, err := git.PlainOpen(clonedPath)
		if err != nil {
			t.Fatalf("failed to open cloned repo: %v", err)
		}

		cfg, err := repo.Config()
		if err != nil {
			t.Fatalf("failed to get config: %v", err)
		}

		if !cfg.Core.IsBare {
			t.Error("expected cloned repo to be bare")
		}
	})

	t.Run("returns error if repo already exists", func(t *testing.T) {
		t.Parallel()

		// Create source repo
		sourceDir := t.TempDir()
		sourcePath := filepath.Join(sourceDir, "source")
		createTestRepo(t, sourcePath, false)

		// Create projects root with existing repo dir
		projectsRoot := t.TempDir()

		existingPath := filepath.Join(projectsRoot, "test-repo")
		if err := os.MkdirAll(existingPath, 0o755); err != nil {
			t.Fatalf("failed to create existing dir: %v", err)
		}

		engine := New(projectsRoot)

		// Clone should fail
		err := engine.Clone(context.Background(), sourcePath, "test-repo")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var cerr *cerrors.CanopyError
		if !errors.As(err, &cerr) || cerr.Code != cerrors.ErrRepoAlreadyExists {
			t.Errorf("expected ErrRepoAlreadyExists, got %v", err)
		}
	})
}

func TestGitEngine_CreateWorktree(t *testing.T) {
	t.Parallel()

	t.Run("creates worktree with new branch", func(t *testing.T) {
		t.Parallel()

		// Create source repo
		sourceDir := t.TempDir()
		sourcePath := filepath.Join(sourceDir, "source")
		sourceRepo := createTestRepo(t, sourcePath, false)

		// Create bare clone in projects root
		projectsRoot := t.TempDir()
		canonicalPath := filepath.Join(projectsRoot, "test-repo")
		cloneToBare(t, sourceRepo, canonicalPath)

		// Create worktree destination
		worktreeDir := t.TempDir()
		worktreePath := filepath.Join(worktreeDir, "workspace")

		engine := New(projectsRoot)

		// Create worktree
		err := engine.CreateWorktree("test-repo", worktreePath, "feature-branch")
		if err != nil {
			t.Fatalf("CreateWorktree failed: %v", err)
		}

		// Verify worktree exists and is on the correct branch
		repo, err := git.PlainOpen(worktreePath)
		if err != nil {
			t.Fatalf("failed to open worktree: %v", err)
		}

		head, err := repo.Head()
		if err != nil {
			t.Fatalf("failed to get HEAD: %v", err)
		}

		if head.Name().Short() != "feature-branch" {
			t.Errorf("expected branch feature-branch, got %s", head.Name().Short())
		}

		// Verify files exist
		readmePath := filepath.Join(worktreePath, "README.md")
		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			t.Error("expected README.md to exist in worktree")
		}
	})
}

func TestGitEngine_Status(t *testing.T) {
	t.Parallel()

	t.Run("reports clean status", func(t *testing.T) {
		t.Parallel()

		// Create repo
		repoPath := t.TempDir()
		createTestRepo(t, repoPath, false)

		engine := New("")

		isDirty, unpushed, behind, branch, err := engine.Status(repoPath)
		if err != nil {
			t.Fatalf("Status failed: %v", err)
		}

		if isDirty {
			t.Error("expected clean repo, got dirty")
		}

		if unpushed != 0 {
			t.Errorf("expected 0 unpushed, got %d", unpushed)
		}

		if behind != 0 {
			t.Errorf("expected 0 behind, got %d", behind)
		}

		if branch != "master" && branch != "main" {
			t.Errorf("expected master or main branch, got %s", branch)
		}
	})

	t.Run("reports dirty status", func(t *testing.T) {
		t.Parallel()

		// Create repo
		repoPath := t.TempDir()
		createTestRepo(t, repoPath, false)

		// Modify a file
		filePath := filepath.Join(repoPath, "README.md")
		if err := os.WriteFile(filePath, []byte("# Modified\n"), 0o644); err != nil {
			t.Fatalf("failed to modify file: %v", err)
		}

		engine := New("")

		isDirty, _, _, _, err := engine.Status(repoPath)
		if err != nil {
			t.Fatalf("Status failed: %v", err)
		}

		if !isDirty {
			t.Error("expected dirty repo, got clean")
		}
	})
}

func TestGitEngine_Checkout(t *testing.T) {
	t.Parallel()

	t.Run("creates new branch", func(t *testing.T) {
		t.Parallel()

		// Create repo
		repoPath := t.TempDir()
		createTestRepo(t, repoPath, false)

		engine := New("")

		err := engine.Checkout(repoPath, "new-branch", true)
		if err != nil {
			t.Fatalf("Checkout failed: %v", err)
		}

		// Verify branch
		repo, err := git.PlainOpen(repoPath)
		if err != nil {
			t.Fatalf("failed to open repo: %v", err)
		}

		head, err := repo.Head()
		if err != nil {
			t.Fatalf("failed to get HEAD: %v", err)
		}

		if head.Name().Short() != "new-branch" {
			t.Errorf("expected new-branch, got %s", head.Name().Short())
		}
	})

	t.Run("switches to existing branch", func(t *testing.T) {
		t.Parallel()

		// Create repo
		repoPath := t.TempDir()
		repo := createTestRepo(t, repoPath, false)

		// Create a second branch
		wt, _ := repo.Worktree()

		err := wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName("other-branch"),
			Create: true,
		})
		if err != nil {
			t.Fatalf("failed to create other branch: %v", err)
		}

		// Switch back to master
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName("master"),
		})
		if err != nil {
			t.Fatalf("failed to switch to master: %v", err)
		}

		engine := New("")

		// Use engine to checkout other-branch
		err = engine.Checkout(repoPath, "other-branch", false)
		if err != nil {
			t.Fatalf("Checkout failed: %v", err)
		}

		head, _ := repo.Head()
		if head.Name().Short() != "other-branch" {
			t.Errorf("expected other-branch, got %s", head.Name().Short())
		}
	})
}

func TestGitEngine_Fetch(t *testing.T) {
	t.Parallel()

	t.Run("fetches from remote", func(t *testing.T) {
		t.Parallel()

		// Create source repo
		sourceDir := t.TempDir()
		sourcePath := filepath.Join(sourceDir, "source")
		sourceRepo := createTestRepo(t, sourcePath, false)

		// Create bare clone in projects root
		projectsRoot := t.TempDir()
		canonicalPath := filepath.Join(projectsRoot, "test-repo")
		cloneToBare(t, sourceRepo, canonicalPath)

		// Add a commit to source
		wt, _ := sourceRepo.Worktree()
		filePath := filepath.Join(sourcePath, "NEW.md")
		_ = os.WriteFile(filePath, []byte("# New\n"), 0o644)
		_, _ = wt.Add("NEW.md")
		_, _ = wt.Commit("New commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test",
				Email: "test@test.com",
			},
		})

		engine := New(projectsRoot)

		// Fetch
		err := engine.Fetch(context.Background(), "test-repo")
		if err != nil {
			t.Fatalf("Fetch failed: %v", err)
		}
	})

	t.Run("returns error for non-existent repo", func(t *testing.T) {
		t.Parallel()

		projectsRoot := t.TempDir()
		engine := New(projectsRoot)

		err := engine.Fetch(context.Background(), "non-existent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var cerr *cerrors.CanopyError
		if !errors.As(err, &cerr) || cerr.Code != cerrors.ErrRepoNotFound {
			t.Errorf("expected ErrRepoNotFound, got %v", err)
		}
	})
}

func TestGitEngine_List(t *testing.T) {
	t.Parallel()

	t.Run("lists repositories", func(t *testing.T) {
		t.Parallel()

		projectsRoot := t.TempDir()

		// Create some bare repo dirs with HEAD files (simulating bare git repos)
		repoA := filepath.Join(projectsRoot, "repo-a")
		repoB := filepath.Join(projectsRoot, "repo-b")
		_ = os.MkdirAll(repoA, 0o755)
		_ = os.MkdirAll(repoB, 0o755)
		_ = os.WriteFile(filepath.Join(repoA, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644)
		_ = os.WriteFile(filepath.Join(repoB, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644)

		engine := New(projectsRoot)

		repos, err := engine.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(repos) != 2 {
			t.Errorf("expected 2 repos, got %d", len(repos))
		}
	})

	t.Run("ignores non-git directories", func(t *testing.T) {
		t.Parallel()

		projectsRoot := t.TempDir()

		// Create a mix of git repos and regular dirs
		repoA := filepath.Join(projectsRoot, "repo-a")
		notRepo := filepath.Join(projectsRoot, "not-a-repo")
		_ = os.MkdirAll(repoA, 0o755)
		_ = os.MkdirAll(notRepo, 0o755)
		_ = os.WriteFile(filepath.Join(repoA, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644)
		// notRepo has no HEAD file

		engine := New(projectsRoot)

		repos, err := engine.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(repos) != 1 {
			t.Errorf("expected 1 repo, got %d", len(repos))
		}

		if len(repos) > 0 && repos[0] != "repo-a" {
			t.Errorf("expected repo-a, got %s", repos[0])
		}
	})

	t.Run("returns empty for non-existent dir", func(t *testing.T) {
		t.Parallel()

		engine := New("/non/existent/path")

		repos, err := engine.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(repos) != 0 {
			t.Errorf("expected empty slice, got %v", repos)
		}
	})
}

func TestGitEngine_RunCommand(t *testing.T) {
	t.Parallel()

	// Skip if git is not installed
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	t.Run("executes git command", func(t *testing.T) {
		t.Parallel()

		repoPath := t.TempDir()
		createTestRepo(t, repoPath, false)

		engine := New("")

		result, err := engine.RunCommand(context.Background(), repoPath, "status", "--short")
		if err != nil {
			t.Fatalf("RunCommand failed: %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
	})

	t.Run("returns error for empty args", func(t *testing.T) {
		t.Parallel()

		engine := New("")

		_, err := engine.RunCommand(context.Background(), "/tmp")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var cerr *cerrors.CanopyError
		if !errors.As(err, &cerr) || cerr.Code != cerrors.ErrInvalidArgument {
			t.Errorf("expected ErrInvalidArgument, got %v", err)
		}
	})
}

func TestGitEngine_countAheadBehind(t *testing.T) {
	t.Parallel()

	t.Run("returns 0,0 for same hash", func(t *testing.T) {
		t.Parallel()

		repoPath := t.TempDir()
		repo := createTestRepo(t, repoPath, false)

		head, _ := repo.Head()

		engine := New("")

		ahead, behind, err := engine.countAheadBehind(repo, head.Hash(), head.Hash())
		if err != nil {
			t.Fatalf("countAheadBehind failed: %v", err)
		}

		if ahead != 0 || behind != 0 {
			t.Errorf("expected 0,0 for same hash, got %d,%d", ahead, behind)
		}
	})

	t.Run("counts commits correctly", func(t *testing.T) {
		t.Parallel()

		repoPath := t.TempDir()
		repo := createTestRepo(t, repoPath, false)

		// Get initial commit hash
		head, _ := repo.Head()
		initialHash := head.Hash()

		// Add another commit
		wt, _ := repo.Worktree()
		filePath := filepath.Join(repoPath, "NEW.md")
		_ = os.WriteFile(filePath, []byte("# New\n"), 0o644)
		_, _ = wt.Add("NEW.md")
		_, _ = wt.Commit("New commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test",
				Email: "test@test.com",
			},
		})

		// Get new head
		newHead, _ := repo.Head()

		engine := New("")

		// New head should be 1 ahead of initial
		ahead, behind, err := engine.countAheadBehind(repo, newHead.Hash(), initialHash)
		if err != nil {
			t.Fatalf("countAheadBehind failed: %v", err)
		}

		if ahead != 1 {
			t.Errorf("expected 1 ahead, got %d", ahead)
		}

		if behind != 0 {
			t.Errorf("expected 0 behind, got %d", behind)
		}

		// Initial should be 0 ahead, 1 behind new head
		ahead, behind, err = engine.countAheadBehind(repo, initialHash, newHead.Hash())
		if err != nil {
			t.Fatalf("countAheadBehind failed: %v", err)
		}

		if ahead != 0 {
			t.Errorf("expected 0 ahead, got %d", ahead)
		}

		if behind != 1 {
			t.Errorf("expected 1 behind, got %d", behind)
		}
	})
}

func TestGitEngine_EnsureCanonical(t *testing.T) {
	t.Parallel()

	t.Run("clones if not exists", func(t *testing.T) {
		t.Parallel()

		// Create source repo
		sourceDir := t.TempDir()
		sourcePath := filepath.Join(sourceDir, "source")
		createTestRepo(t, sourcePath, false)

		// Create projects root
		projectsRoot := t.TempDir()
		engine := New(projectsRoot)

		repo, err := engine.EnsureCanonical(context.Background(), sourcePath, "test-repo")
		if err != nil {
			t.Fatalf("EnsureCanonical failed: %v", err)
		}

		if repo == nil {
			t.Fatal("expected repo, got nil")
		}
	})

	t.Run("opens if exists", func(t *testing.T) {
		t.Parallel()

		// Create source repo
		sourceDir := t.TempDir()
		sourcePath := filepath.Join(sourceDir, "source")
		sourceRepo := createTestRepo(t, sourcePath, false)

		// Create projects root with existing clone
		projectsRoot := t.TempDir()
		canonicalPath := filepath.Join(projectsRoot, "test-repo")
		cloneToBare(t, sourceRepo, canonicalPath)

		engine := New(projectsRoot)

		repo, err := engine.EnsureCanonical(context.Background(), sourcePath, "test-repo")
		if err != nil {
			t.Fatalf("EnsureCanonical failed: %v", err)
		}

		if repo == nil {
			t.Fatal("expected repo, got nil")
		}
	})
}
