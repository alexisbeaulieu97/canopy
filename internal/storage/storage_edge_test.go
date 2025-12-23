package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

func TestListWorkspaces_RootMissing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "missing")
	engine := New(workspacesRoot, filepath.Join(tmpDir, "closed"))

	workspaces, err := engine.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(workspaces) != 0 {
		t.Fatalf("expected no workspaces, got %d", len(workspaces))
	}
}

func TestListWorkspaces_SkipsInvalidEntries(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "workspaces")
	closedRoot := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	engine := New(workspacesRoot, closedRoot)

	if err := os.WriteFile(filepath.Join(workspacesRoot, "not-a-dir"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	badDir := filepath.Join(workspacesRoot, "bad")
	if err := os.MkdirAll(badDir, 0o750); err != nil {
		t.Fatalf("failed to create bad dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(badDir, "workspace.yaml"), []byte("invalid: ["), 0o644); err != nil {
		t.Fatalf("failed to write bad metadata: %v", err)
	}

	if err := engine.Create(context.Background(), domain.Workspace{ID: "good"}); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	workspaces, err := engine.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(workspaces))
	}

	if workspaces[0].ID != "good" {
		t.Errorf("expected workspace ID good, got %q", workspaces[0].ID)
	}
}

func TestClose_RequiresClosedRoot(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "workspaces")

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	engine := New(workspacesRoot, "")
	if err := engine.Create(context.Background(), domain.Workspace{ID: "ws-1"}); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	if _, err := engine.Close(context.Background(), "ws-1", time.Now()); err == nil {
		t.Fatal("expected error when closed root is not configured")
	}
}

func TestClose_CreatesClosedEntry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "workspaces")
	closedRoot := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	engine := New(workspacesRoot, closedRoot)
	if err := engine.Create(context.Background(), domain.Workspace{ID: "ws-1"}); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	closedAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	closed, err := engine.Close(context.Background(), "ws-1", closedAt)
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if closed == nil || closed.Metadata.ClosedAt == nil {
		t.Fatal("expected closed workspace metadata to include closed time")
	}

	if _, err := os.Stat(filepath.Join(closed.Path, "workspace.yaml")); err != nil {
		t.Fatalf("expected closed metadata to exist: %v", err)
	}
}
