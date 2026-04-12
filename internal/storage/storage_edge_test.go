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

func TestListWorkspaces_ReturnsErrorForInvalidMetadata(t *testing.T) {
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
	if err == nil {
		t.Fatal("expected List to fail for invalid active workspace metadata")
	}

	if workspaces != nil {
		t.Fatalf("expected no workspaces on error, got %d", len(workspaces))
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

func TestClose_CreatesDistinctClosedEntriesWithinSameSecond(t *testing.T) {
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

	firstClosedAt := time.Date(2024, 1, 2, 3, 4, 5, 100, time.UTC)
	secondClosedAt := time.Date(2024, 1, 2, 3, 4, 5, 200, time.UTC)

	first, err := engine.Close(context.Background(), "ws-1", firstClosedAt)
	if err != nil {
		t.Fatalf("first Close failed: %v", err)
	}

	second, err := engine.Close(context.Background(), "ws-1", secondClosedAt)
	if err != nil {
		t.Fatalf("second Close failed: %v", err)
	}

	if first.Path == second.Path {
		t.Fatalf("expected distinct closed paths, got %q", first.Path)
	}

	resolvedFirst, err := engine.resolveClosedDirectory("ws-1", firstClosedAt)
	if err != nil {
		t.Fatalf("failed to resolve first closed directory: %v", err)
	}

	resolvedSecond, err := engine.resolveClosedDirectory("ws-1", secondClosedAt)
	if err != nil {
		t.Fatalf("failed to resolve second closed directory: %v", err)
	}

	if resolvedFirst != first.Path {
		t.Fatalf("expected first resolved path %q, got %q", first.Path, resolvedFirst)
	}

	if resolvedSecond != second.Path {
		t.Fatalf("expected second resolved path %q, got %q", second.Path, resolvedSecond)
	}
}

func TestResolveClosedDirectory_AcceptsLegacyTimestampPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "workspaces")
	closedRoot := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("failed to create workspaces root: %v", err)
	}

	engine := New(workspacesRoot, closedRoot)
	closedAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	legacyDir := filepath.Join(closedRoot, "ws-1", closedAt.UTC().Format(legacyClosedEntryTimestampLayout))

	if err := os.MkdirAll(legacyDir, 0o750); err != nil {
		t.Fatalf("failed to create legacy closed dir: %v", err)
	}

	if err := engine.saveMetadata(filepath.Join(legacyDir, "workspace.yaml"), domain.Workspace{ID: "ws-1"}); err != nil {
		t.Fatalf("failed to save legacy metadata: %v", err)
	}

	resolved, err := engine.resolveClosedDirectory("ws-1", closedAt)
	if err != nil {
		t.Fatalf("resolveClosedDirectory failed: %v", err)
	}

	if resolved != legacyDir {
		t.Fatalf("expected legacy path %q, got %q", legacyDir, resolved)
	}

	closedEntries, err := engine.ListClosed(context.Background())
	if err != nil {
		t.Fatalf("ListClosed failed: %v", err)
	}

	if len(closedEntries) != 1 {
		t.Fatalf("expected 1 closed entry, got %d", len(closedEntries))
	}

	if !closedEntries[0].ClosedAt().Equal(closedAt) {
		t.Fatalf("expected closed time %v, got %v", closedAt, closedEntries[0].ClosedAt())
	}
}
