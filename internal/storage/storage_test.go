package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

func TestLoad_ByID(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "workspaces")
	closedRoot := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	engine := New(workspacesRoot, closedRoot)
	ctx := context.Background()

	// Create a workspace
	wsID := "test-workspace"
	repos := []domain.Repo{{Name: "test-repo", URL: "https://github.com/org/repo.git"}}
	ws := domain.Workspace{
		ID:         wsID,
		BranchName: "main",
		Repos:      repos,
	}

	if err := engine.Create(ctx, ws); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	// Test Load by ID
	loaded, err := engine.Load(ctx, wsID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.ID != wsID {
		t.Errorf("expected ID %q, got %q", wsID, loaded.ID)
	}

	if loaded.BranchName != "main" {
		t.Errorf("expected BranchName %q, got %q", "main", loaded.BranchName)
	}

	if len(loaded.Repos) != 1 || loaded.Repos[0].Name != "test-repo" {
		t.Errorf("expected 1 repo named test-repo, got %v", loaded.Repos)
	}
}

func TestLoad_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "workspaces")
	closedRoot := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	engine := New(workspacesRoot, closedRoot)
	ctx := context.Background()

	// Test Load with non-existent workspace
	_, err := engine.Load(ctx, "non-existent")
	if err == nil {
		t.Fatal("expected error for non-existent workspace")
	}
}

func TestCreateAndList(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "workspaces")
	closedRoot := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	engine := New(workspacesRoot, closedRoot)
	ctx := context.Background()

	// Create two workspaces
	ws1 := domain.Workspace{ID: "workspace-1", BranchName: "main"}
	ws2 := domain.Workspace{ID: "workspace-2", BranchName: "feature"}

	if err := engine.Create(ctx, ws1); err != nil {
		t.Fatalf("failed to create workspace 1: %v", err)
	}

	if err := engine.Create(ctx, ws2); err != nil {
		t.Fatalf("failed to create workspace 2: %v", err)
	}

	// List workspaces
	workspaces, err := engine.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(workspaces) != 2 {
		t.Errorf("expected 2 workspaces, got %d", len(workspaces))
	}

	// Verify workspace IDs are in the list
	ids := make(map[string]bool)
	for _, w := range workspaces {
		ids[w.ID] = true
	}

	if !ids["workspace-1"] {
		t.Error("workspace-1 not found in list")
	}

	if !ids["workspace-2"] {
		t.Error("workspace-2 not found in list")
	}
}
