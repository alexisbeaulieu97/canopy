package workspaces

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexisbeaulieu97/yard/internal/config"
	"github.com/alexisbeaulieu97/yard/internal/domain"
	"github.com/alexisbeaulieu97/yard/internal/gitx"
	"github.com/alexisbeaulieu97/yard/internal/workspace"
)

func TestResolveRepos(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.Defaults{
			WorkspacePatterns: []config.WorkspacePattern{
				{Pattern: "^TEST-", Repos: []string{"repo-a"}},
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

	if len(repos) != 1 || repos[0].Name != "repo-a" {
		t.Errorf("expected [repo-a], got %v", repos)
	}

	// Test case 2: Explicit repos
	repos, err = svc.ResolveRepos("OTHER-123", []string{"myorg/repo-b", "https://github.com/org/repo-c.git"})
	if err != nil {
		t.Fatalf("ResolveRepos failed: %v", err)
	}

	if len(repos) != 2 {
		t.Errorf("expected 2 repos, got %d", len(repos))
	}

	if repos[0].Name != "repo-b" {
		t.Errorf("expected repo-b, got %s", repos[0].Name)
	}

	if repos[1].Name != "repo-c" {
		t.Errorf("expected repo-c, got %s", repos[1].Name)
	}
}

func TestCreateWorkspace(t *testing.T) {
	// Setup temp dirs
	tmpDir, err := os.MkdirTemp("", "yard-service-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	projectsRoot := filepath.Join(tmpDir, "projects")
	workspacesRoot := filepath.Join(tmpDir, "workspaces")

	if err := os.MkdirAll(projectsRoot, 0o750); err != nil {
		t.Fatalf("failed to create projects root: %v", err)
	}

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	cfg := &config.Config{
		ProjectsRoot:    projectsRoot,
		WorkspacesRoot:  workspacesRoot,
		WorkspaceNaming: "{{.ID}}",
	}

	gitEngine := gitx.New(projectsRoot)
	wsEngine := workspace.New(workspacesRoot)
	svc := NewService(cfg, gitEngine, wsEngine, nil)

	// We can't easily test full CreateWorkspace because it calls git commands.
	// But we can test the directory creation part if we mock git or use bare repos.
	// For now, let's test a "bare" workspace creation (no repos) if allowed.
	// CreateWorkspace requires repos? No, it iterates over them.

	// Test creating a workspace with NO repos
	dirName, err := svc.CreateWorkspace("TEST-EMPTY", "", "", []domain.Repo{})
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
