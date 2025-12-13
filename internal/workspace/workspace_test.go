package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

func TestLoadByID_DirectPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "workspaces")
	closedRoot := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	engine := New(workspacesRoot, closedRoot)

	// Create a workspace where ID == dirName
	wsID := "test-workspace"
	repos := []domain.Repo{{Name: "test-repo", URL: "https://github.com/org/repo.git"}}

	if err := engine.Create(wsID, wsID, "main", repos); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	// Test LoadByID with direct path access
	ws, dirName, err := engine.LoadByID(wsID)
	if err != nil {
		t.Fatalf("LoadByID failed: %v", err)
	}

	if ws.ID != wsID {
		t.Errorf("expected ID %q, got %q", wsID, ws.ID)
	}

	if dirName != wsID {
		t.Errorf("expected dirName %q, got %q", wsID, dirName)
	}

	if ws.BranchName != "main" {
		t.Errorf("expected BranchName %q, got %q", "main", ws.BranchName)
	}

	if len(ws.Repos) != 1 || ws.Repos[0].Name != "test-repo" {
		t.Errorf("expected 1 repo named test-repo, got %v", ws.Repos)
	}
}

func TestLoadByID_FallbackScan(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "workspaces")
	closedRoot := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	engine := New(workspacesRoot, closedRoot)

	// Create a workspace where dirName differs from ID
	dirName := "custom-dir"
	wsID := "workspace-id"

	if err := engine.Create(dirName, wsID, "feature", nil); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	// Test LoadByID falls back to scanning when direct path fails
	ws, foundDir, err := engine.LoadByID(wsID)
	if err != nil {
		t.Fatalf("LoadByID fallback failed: %v", err)
	}

	if ws.ID != wsID {
		t.Errorf("expected ID %q, got %q", wsID, ws.ID)
	}

	if foundDir != dirName {
		t.Errorf("expected dirName %q, got %q", dirName, foundDir)
	}
}

func TestLoadByID_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "workspaces")
	closedRoot := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	engine := New(workspacesRoot, closedRoot)

	// Test LoadByID with non-existent workspace
	_, _, err := engine.LoadByID("non-existent")
	if err == nil {
		t.Fatal("expected error for non-existent workspace")
	}
}

func TestLoadByID_DirectPathMismatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "workspaces")
	closedRoot := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	engine := New(workspacesRoot, closedRoot)

	// Create workspace where dirName == searchID but ID is different
	// This tests that we correctly verify the ID matches, not just the directory
	if err := engine.Create("some-dir", "different-id", "main", nil); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	// Search for "some-dir" as an ID - it exists as a directory but ID is "different-id"
	_, _, err := engine.LoadByID("some-dir")
	if err == nil {
		t.Fatal("expected error when directory exists but ID doesn't match")
	}
}
