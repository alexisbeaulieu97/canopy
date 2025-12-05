package workspaces

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/gitx"
	"github.com/alexisbeaulieu97/canopy/internal/workspace"
)

type testServiceDeps struct {
	svc            *Service
	wsEngine       *workspace.Engine
	projectsRoot   string
	workspacesRoot string
	closedRoot     string
}

func newTestService(t *testing.T) testServiceDeps {
	t.Helper()

	base := t.TempDir()
	projectsRoot := filepath.Join(base, "projects")
	workspacesRoot := filepath.Join(base, "workspaces")
	closedRoot := filepath.Join(base, "closed")

	mustMkdir(t, projectsRoot)
	mustMkdir(t, workspacesRoot)

	cfg := &config.Config{
		ProjectsRoot:   projectsRoot,
		WorkspacesRoot: workspacesRoot,
		ClosedRoot:     closedRoot,
	}

	gitEngine := gitx.New(projectsRoot)
	wsEngine := workspace.New(workspacesRoot, closedRoot)

	return testServiceDeps{
		svc:            NewService(cfg, gitEngine, wsEngine, nil),
		wsEngine:       wsEngine,
		projectsRoot:   projectsRoot,
		workspacesRoot: workspacesRoot,
		closedRoot:     closedRoot,
	}
}

func TestResolveRepos(t *testing.T) {
	t.Parallel()

	registry := config.RepoRegistry{
		Repos: map[string]config.RegistryEntry{
			"myorg/repo-a": {Alias: "myorg/repo-a", URL: "https://github.com/myorg/repo-a.git"},
			"alias/repo":   {Alias: "alias/repo", URL: "https://github.com/org/repo.git"},
		},
	}

	cfg := &config.Config{
		Registry: &registry,
		Defaults: config.Defaults{
			WorkspacePatterns: []config.WorkspacePattern{
				{Pattern: "^TEST-", Repos: []string{"myorg/repo-a"}},
			},
		},
	}

	// We don't need real engines for ResolveRepos
	svc := NewService(cfg, nil, nil, nil)

	// Test case 1: Pattern match
	repos, err := svc.ResolveRepos("TEST-123", nil)
	if err != nil {
		t.Fatalf("ResolveRepos failed: %v", err)
	}

	if len(repos) != 1 || repos[0].Name != "myorg/repo-a" {
		t.Errorf("expected [myorg/repo-a], got %v", repos)
	}

	// Test case 2: Explicit repos
	repos, err = svc.ResolveRepos("OTHER-123", []string{"myorg/repo-b", "https://github.com/org/repo-c.git"})
	if err != nil {
		t.Fatalf("ResolveRepos failed: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}

	if repos[0].Name != "repo-b" {
		t.Errorf("expected repo-b, got %s", repos[0].Name)
	}

	if repos[1].Name != "repo-c" {
		t.Errorf("expected repo-c, got %s", repos[1].Name)
	}

	// URL should use alias when registry contains that URL.
	repos, err = svc.ResolveRepos("OTHER-123", []string{"https://github.com/org/repo.git"})
	if err != nil {
		t.Fatalf("ResolveRepos failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}

	if repos[0].Name != "alias/repo" {
		t.Errorf("expected alias/repo, got %s", repos[0].Name)
	}
}

func TestCreateWorkspace(t *testing.T) {
	// Setup temp dirs
	tmpDir, err := os.MkdirTemp("", "canopy-service-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	projectsRoot := filepath.Join(tmpDir, "projects")
	workspacesRoot := filepath.Join(tmpDir, "workspaces")
	closedRoot := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(projectsRoot, 0o750); err != nil {
		t.Fatalf("failed to create projects root: %v", err)
	}

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	cfg := &config.Config{
		ProjectsRoot:    projectsRoot,
		WorkspacesRoot:  workspacesRoot,
		ClosedRoot:      closedRoot,
		WorkspaceNaming: "{{.ID}}",
	}

	gitEngine := gitx.New(projectsRoot)
	wsEngine := workspace.New(workspacesRoot, closedRoot)
	svc := NewService(cfg, gitEngine, wsEngine, nil)

	// We can't easily test full CreateWorkspace because it calls git commands.
	// But we can test the directory creation part if we mock git or use bare repos.
	// For now, let's test a "bare" workspace creation (no repos) if allowed.
	// CreateWorkspace requires repos? No, it iterates over them.

	// Test creating a workspace with NO repos
	dirName, err := svc.CreateWorkspace("TEST-EMPTY", "", []domain.Repo{})
	if err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	expectedPath := filepath.Join(workspacesRoot, dirName)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("workspace directory not created at %s", expectedPath)
	}

	// Check metadata
	ws, err := wsEngine.Load(dirName)
	if err != nil {
		t.Fatalf("failed to load workspace: %v", err)
	}

	if ws.ID != "TEST-EMPTY" {
		t.Errorf("expected ID TEST-EMPTY, got %s", ws.ID)
	}
}

func TestRepoNameFromURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{name: "standard https", url: "https://github.com/org/repo.git", want: "repo"},
		{name: "scp style", url: "git@github.com:org/repo.git", want: "repo"},
		{name: "trailing slash", url: "https://github.com/org/repo/", want: "repo"},
		{name: "multiple trailing slashes", url: "https://github.com/org/repo///", want: "repo"},
		{name: "empty input", url: "", want: ""},
		{name: "slash only", url: "///", want: ""},
		{name: "file scheme", url: "file:///tmp/repo.git", want: "repo"},
		{name: "ssh scheme", url: "ssh://git@example.com/org/repo.git", want: "repo"},
		{name: "https with user info", url: "https://user:token@github.com/org/repo.git", want: "repo"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := repoNameFromURL(tt.url); got != tt.want {
				t.Fatalf("repoNameFromURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestCloseWorkspaceStoresMetadata(t *testing.T) {
	deps := newTestService(t)

	if _, err := deps.svc.CreateWorkspace("TEST-ARCHIVE", "", []domain.Repo{}); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	archived, err := deps.svc.CloseWorkspaceKeepMetadata("TEST-ARCHIVE", true)
	if err != nil {
		t.Fatalf("CloseWorkspaceKeepMetadata failed: %v", err)
	}

	if archived == nil {
		t.Fatalf("expected closed entry details")
	}

	if _, err := os.Stat(filepath.Join(deps.workspacesRoot, "TEST-ARCHIVE")); !os.IsNotExist(err) {
		t.Fatalf("expected workspace directory to be removed")
	}

	closedEntries, err := deps.wsEngine.ListClosed()
	if err != nil {
		t.Fatalf("ListClosed failed: %v", err)
	}

	if len(closedEntries) != 1 {
		t.Fatalf("expected 1 closed entry, got %d", len(closedEntries))
	}

	if closedEntries[0].Metadata.ClosedAt == nil {
		t.Fatalf("expected closed metadata to include timestamp")
	}
}

func TestCloseWorkspaceNonexistent(t *testing.T) {
	deps := newTestService(t)

	if _, err := deps.svc.CloseWorkspaceKeepMetadata("MISSING", false); err == nil {
		t.Fatalf("expected error when closing nonexistent workspace")
	}
}

func TestRestoreWorkspaceConflict(t *testing.T) {
	deps := newTestService(t)

	if _, err := deps.svc.CreateWorkspace("TEST-CONFLICT", "", []domain.Repo{}); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	_, err := deps.wsEngine.Close("TEST-CONFLICT", domain.Workspace{ID: "TEST-CONFLICT"}, time.Now())
	if err != nil {
		t.Fatalf("failed to seed closed entry: %v", err)
	}

	if err := deps.svc.RestoreWorkspace("TEST-CONFLICT", false); err == nil {
		t.Fatalf("expected restore conflict error")
	}
}

func TestCloseRestoreCycle(t *testing.T) {
	deps := newTestService(t)

	sourceRepo := filepath.Join(deps.projectsRoot, "source")
	createRepoWithCommit(t, sourceRepo)

	canonicalPath := filepath.Join(deps.projectsRoot, "sample")
	runGit(t, "", "clone", "--bare", sourceRepo, canonicalPath)

	repoURL := "file://" + sourceRepo

	if _, err := deps.svc.CreateWorkspace("PROJ-1", "", []domain.Repo{{Name: "sample", URL: repoURL}}); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	worktreePath := filepath.Join(deps.workspacesRoot, "PROJ-1", "sample")

	if _, err := os.Stat(worktreePath); err != nil {
		t.Fatalf("expected worktree at %s: %v", worktreePath, err)
	}

	archived, err := deps.svc.CloseWorkspaceKeepMetadata("PROJ-1", false)
	if err != nil {
		t.Fatalf("CloseWorkspaceKeepMetadata failed: %v", err)
	}

	if archived.Metadata.ClosedAt == nil {
		t.Fatalf("expected closed timestamp to be set")
	}

	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Fatalf("expected worktree to be removed when keeping metadata")
	}

	if err := deps.svc.RestoreWorkspace("PROJ-1", false); err != nil {
		t.Fatalf("RestoreWorkspace failed: %v", err)
	}

	if _, err := os.Stat(worktreePath); err != nil {
		t.Fatalf("expected restored worktree at %s: %v", worktreePath, err)
	}

	if _, err := os.Stat(archived.Path); !os.IsNotExist(err) {
		t.Fatalf("expected closed entry path to be removed after restore")
	}

	branch := runGitOutput(t, worktreePath, "rev-parse", "--abbrev-ref", "HEAD")
	if branch != "PROJ-1" {
		t.Fatalf("expected branch PROJ-1 after restore, got %s", branch)
	}
}

func TestCloseWorkspaceDirtyFailsWithoutForce(t *testing.T) {
	deps := newTestService(t)

	sourceRepo := filepath.Join(deps.projectsRoot, "source-dirty")
	createRepoWithCommit(t, sourceRepo)

	canonicalPath := filepath.Join(deps.projectsRoot, "sample-dirty")
	runGit(t, "", "clone", "--bare", sourceRepo, canonicalPath)

	repoURL := "file://" + sourceRepo

	if _, err := deps.svc.CreateWorkspace("PROJ-2", "", []domain.Repo{{Name: "sample-dirty", URL: repoURL}}); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	worktreePath := filepath.Join(deps.workspacesRoot, "PROJ-2", "sample-dirty")
	if err := os.WriteFile(filepath.Join(worktreePath, "WIP.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatalf("failed to write dirty file: %v", err)
	}

	if _, err := deps.svc.CloseWorkspaceKeepMetadata("PROJ-2", false); err == nil {
		t.Fatalf("expected close keep-metadata to fail on dirty workspace")
	}
}

func TestRestoreWorkspaceForceDoesNotDeleteWithoutClosedEntry(t *testing.T) {
	deps := newTestService(t)

	if _, err := deps.svc.CreateWorkspace("PROJ-NO-ARCHIVE", "", []domain.Repo{}); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	if err := deps.svc.RestoreWorkspace("PROJ-NO-ARCHIVE", true); err == nil {
		t.Fatalf("expected restore to fail without closed entry present")
	}

	if _, err := os.Stat(filepath.Join(deps.workspacesRoot, "PROJ-NO-ARCHIVE")); err != nil {
		t.Fatalf("workspace should remain when restore fails: %v", err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o750); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

func createRepoWithCommit(t *testing.T, path string) {
	t.Helper()

	mustMkdir(t, path)
	runGit(t, path, "init")
	runGit(t, path, "config", "user.email", "test@example.com")
	runGit(t, path, "config", "user.name", "Test User")
	runGit(t, path, "config", "credential.helper", "")

	filePath := filepath.Join(path, "README.md")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	runGit(t, path, "add", ".")
	runGit(t, path, "commit", "-m", "init")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...) //nolint:gosec // test helper
	cmd.Dir = dir

	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %s (%v)", args, strings.TrimSpace(string(output)), err)
	}
}

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...) //nolint:gosec // test helper
	cmd.Dir = dir

	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %s (%v)", args, strings.TrimSpace(string(output)), err)
	}

	return strings.TrimSpace(string(output))
}

func TestDetectOrphans_MissingCanonicalRepo(t *testing.T) {
	deps := newTestService(t)

	// Create a workspace manually with a repo reference but no canonical repo
	ws := domain.Workspace{
		ID:         "ORPHAN-TEST-1",
		BranchName: "feature-branch",
		Repos: []domain.Repo{
			{Name: "missing-repo", URL: "https://github.com/org/missing-repo.git"},
		},
	}

	// Save workspace metadata (without actually creating the canonical repo)
	if err := deps.wsEngine.Create("ORPHAN-TEST-1", ws.ID, ws.BranchName, ws.Repos); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	orphans, err := deps.svc.DetectOrphans()
	if err != nil {
		t.Fatalf("DetectOrphans failed: %v", err)
	}

	if len(orphans) != 1 {
		t.Fatalf("expected 1 orphan, got %d", len(orphans))
	}

	if orphans[0].WorkspaceID != "ORPHAN-TEST-1" {
		t.Errorf("expected workspace ID ORPHAN-TEST-1, got %s", orphans[0].WorkspaceID)
	}

	if orphans[0].RepoName != "missing-repo" {
		t.Errorf("expected repo name missing-repo, got %s", orphans[0].RepoName)
	}

	if orphans[0].Reason != domain.OrphanReasonCanonicalMissing {
		t.Errorf("expected reason canonical_missing, got %s", orphans[0].Reason)
	}
}

func TestDetectOrphans_MissingWorktreeDirectory(t *testing.T) {
	deps := newTestService(t)

	// Create a bare canonical repo
	sourceRepo := filepath.Join(deps.projectsRoot, "source-orphan")
	createRepoWithCommit(t, sourceRepo)

	canonicalPath := filepath.Join(deps.projectsRoot, "existing-repo")
	runGit(t, "", "clone", "--bare", sourceRepo, canonicalPath)

	// Create workspace metadata referencing the canonical repo
	// but don't create the actual worktree directory
	ws := domain.Workspace{
		ID:         "ORPHAN-TEST-2",
		BranchName: "main",
		Repos: []domain.Repo{
			{Name: "existing-repo", URL: "file://" + sourceRepo},
		},
	}

	if err := deps.wsEngine.Create("ORPHAN-TEST-2", ws.ID, ws.BranchName, ws.Repos); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	// Don't create the worktree directory - it should be detected as orphaned

	orphans, err := deps.svc.DetectOrphans()
	if err != nil {
		t.Fatalf("DetectOrphans failed: %v", err)
	}

	if len(orphans) != 1 {
		t.Fatalf("expected 1 orphan, got %d", len(orphans))
	}

	if orphans[0].Reason != domain.OrphanReasonDirectoryMissing {
		t.Errorf("expected reason directory_missing, got %s", orphans[0].Reason)
	}
}

func TestDetectOrphans_InvalidGitDir(t *testing.T) {
	deps := newTestService(t)

	// Create a bare canonical repo
	sourceRepo := filepath.Join(deps.projectsRoot, "source-invalid-git")
	createRepoWithCommit(t, sourceRepo)

	canonicalPath := filepath.Join(deps.projectsRoot, "invalid-git-repo")
	runGit(t, "", "clone", "--bare", sourceRepo, canonicalPath)

	// Create workspace metadata
	ws := domain.Workspace{
		ID:         "ORPHAN-TEST-3",
		BranchName: "main",
		Repos: []domain.Repo{
			{Name: "invalid-git-repo", URL: "file://" + sourceRepo},
		},
	}

	if err := deps.wsEngine.Create("ORPHAN-TEST-3", ws.ID, ws.BranchName, ws.Repos); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	// Create the worktree directory but WITHOUT a .git directory
	worktreePath := filepath.Join(deps.workspacesRoot, "ORPHAN-TEST-3", "invalid-git-repo")
	if err := os.MkdirAll(worktreePath, 0o750); err != nil {
		t.Fatalf("failed to create worktree directory: %v", err)
	}

	// Add a dummy file to prove it's a real directory
	if err := os.WriteFile(filepath.Join(worktreePath, "README.md"), []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	orphans, err := deps.svc.DetectOrphans()
	if err != nil {
		t.Fatalf("DetectOrphans failed: %v", err)
	}

	if len(orphans) != 1 {
		t.Fatalf("expected 1 orphan, got %d: %+v", len(orphans), orphans)
	}

	if orphans[0].WorkspaceID != "ORPHAN-TEST-3" {
		t.Errorf("expected workspace ID ORPHAN-TEST-3, got %s", orphans[0].WorkspaceID)
	}

	if orphans[0].RepoName != "invalid-git-repo" {
		t.Errorf("expected repo name invalid-git-repo, got %s", orphans[0].RepoName)
	}

	if orphans[0].Reason != domain.OrphanReasonInvalidGitDir {
		t.Errorf("expected reason invalid_git_dir, got %s", orphans[0].Reason)
	}
}

func TestDetectOrphans_NoOrphans(t *testing.T) {
	deps := newTestService(t)

	// Create a proper workspace with canonical repo and worktree
	sourceRepo := filepath.Join(deps.projectsRoot, "source-clean")
	createRepoWithCommit(t, sourceRepo)

	canonicalPath := filepath.Join(deps.projectsRoot, "clean-repo")
	runGit(t, "", "clone", "--bare", sourceRepo, canonicalPath)

	repoURL := "file://" + sourceRepo

	if _, err := deps.svc.CreateWorkspace("CLEAN-WS", "", []domain.Repo{{Name: "clean-repo", URL: repoURL}}); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	orphans, err := deps.svc.DetectOrphans()
	if err != nil {
		t.Fatalf("DetectOrphans failed: %v", err)
	}

	if len(orphans) != 0 {
		t.Fatalf("expected 0 orphans, got %d: %+v", len(orphans), orphans)
	}
}

func TestGetWorkspacesUsingRepo(t *testing.T) {
	deps := newTestService(t)

	// Create a bare canonical repo
	sourceRepo := filepath.Join(deps.projectsRoot, "source-shared")
	createRepoWithCommit(t, sourceRepo)

	canonicalPath := filepath.Join(deps.projectsRoot, "shared-repo")
	runGit(t, "", "clone", "--bare", sourceRepo, canonicalPath)

	repoURL := "file://" + sourceRepo

	// Create two workspaces using the same repo
	if _, err := deps.svc.CreateWorkspace("WS-1", "", []domain.Repo{{Name: "shared-repo", URL: repoURL}}); err != nil {
		t.Fatalf("failed to create workspace WS-1: %v", err)
	}

	if _, err := deps.svc.CreateWorkspace("WS-2", "", []domain.Repo{{Name: "shared-repo", URL: repoURL}}); err != nil {
		t.Fatalf("failed to create workspace WS-2: %v", err)
	}

	// Create a workspace that doesn't use the repo
	if _, err := deps.svc.CreateWorkspace("WS-3", "", []domain.Repo{}); err != nil {
		t.Fatalf("failed to create workspace WS-3: %v", err)
	}

	usedBy, err := deps.svc.GetWorkspacesUsingRepo("shared-repo")
	if err != nil {
		t.Fatalf("GetWorkspacesUsingRepo failed: %v", err)
	}

	if len(usedBy) != 2 {
		t.Fatalf("expected 2 workspaces using repo, got %d", len(usedBy))
	}

	// Check that both WS-1 and WS-2 are in the list
	foundWS1, foundWS2 := false, false

	for _, wsID := range usedBy {
		if wsID == "WS-1" {
			foundWS1 = true
		}

		if wsID == "WS-2" {
			foundWS2 = true
		}
	}

	if !foundWS1 || !foundWS2 {
		t.Errorf("expected both WS-1 and WS-2 in usedBy, got %v", usedBy)
	}
}

func TestPreviewCloseWorkspace(t *testing.T) {
	deps := newTestService(t)

	// Create a workspace
	if _, err := deps.svc.CreateWorkspace("TEST-PREVIEW", "", []domain.Repo{}); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	// Create a test file to have some disk usage
	testFile := filepath.Join(deps.workspacesRoot, "TEST-PREVIEW", "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test preview
	preview, err := deps.svc.PreviewCloseWorkspace("TEST-PREVIEW", true)
	if err != nil {
		t.Fatalf("PreviewCloseWorkspace failed: %v", err)
	}

	if preview.WorkspaceID != "TEST-PREVIEW" {
		t.Errorf("expected workspace ID TEST-PREVIEW, got %s", preview.WorkspaceID)
	}

	if preview.KeepMetadata != true {
		t.Errorf("expected KeepMetadata true, got false")
	}

	expectedPath := filepath.Join(deps.workspacesRoot, "TEST-PREVIEW")
	if preview.WorkspacePath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, preview.WorkspacePath)
	}

	if preview.DiskUsageBytes <= 0 {
		t.Errorf("expected positive disk usage, got %d", preview.DiskUsageBytes)
	}

	// Verify workspace still exists (dry run doesn't delete)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("workspace should still exist after preview")
	}
}

func TestPreviewCloseWorkspaceNonexistent(t *testing.T) {
	deps := newTestService(t)

	_, err := deps.svc.PreviewCloseWorkspace("NONEXISTENT", false)
	if err == nil {
		t.Fatalf("expected error when previewing nonexistent workspace")
	}
}

func TestPreviewRemoveCanonicalRepo(t *testing.T) {
	deps := newTestService(t)

	// Create a bare repository
	sourceRepo := filepath.Join(deps.projectsRoot, "source")
	createRepoWithCommit(t, sourceRepo)

	repoPath := filepath.Join(deps.projectsRoot, "test-repo")
	runGit(t, "", "clone", "--bare", sourceRepo, repoPath)

	// Create a workspace that uses this repo (using file:// URL for local repo)
	if _, err := deps.svc.CreateWorkspace("WS-USING-REPO", "", []domain.Repo{{Name: "test-repo", URL: "file://" + sourceRepo}}); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	// Test preview
	preview, err := deps.svc.PreviewRemoveCanonicalRepo("test-repo")
	if err != nil {
		t.Fatalf("PreviewRemoveCanonicalRepo failed: %v", err)
	}

	if preview.RepoName != "test-repo" {
		t.Errorf("expected repo name test-repo, got %s", preview.RepoName)
	}

	if preview.RepoPath != repoPath {
		t.Errorf("expected path %s, got %s", repoPath, preview.RepoPath)
	}

	if len(preview.WorkspacesAffected) != 1 || preview.WorkspacesAffected[0] != "WS-USING-REPO" {
		t.Errorf("expected workspaces affected [WS-USING-REPO], got %v", preview.WorkspacesAffected)
	}

	if preview.DiskUsageBytes <= 0 {
		t.Errorf("expected positive disk usage, got %d", preview.DiskUsageBytes)
	}

	// Verify repo still exists (dry run doesn't delete)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Errorf("repo should still exist after preview")
	}
}

func TestPreviewRemoveCanonicalRepoNonexistent(t *testing.T) {
	deps := newTestService(t)

	_, err := deps.svc.PreviewRemoveCanonicalRepo("nonexistent-repo")
	if err == nil {
		t.Fatalf("expected error when previewing nonexistent repo")
	}
}
