package workspaces

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

func TestAddRepoBlockedByWorkspaceLock(t *testing.T) {
	t.Parallel()

	mockConfig := mocks.NewMockConfigProvider()
	mockConfig.WorkspacesRoot = t.TempDir()
	mockConfig.LockTimeout = 150 * time.Millisecond
	mockConfig.LockStaleThreshold = time.Minute

	lockPath := filepath.Join(mockConfig.WorkspacesRoot, "LOCKED", lockFileName)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o750); err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	if err := os.WriteFile(lockPath, []byte("lock"), 0o600); err != nil {
		t.Fatalf("failed to create lock: %v", err)
	}

	svc := NewService(mockConfig, mocks.NewMockGitOperations(), mocks.NewMockWorkspaceStorage(), nil)

	err := svc.AddRepoToWorkspace(context.Background(), "LOCKED", "repo")
	if err == nil {
		t.Fatalf("expected lock error")
	}

	var canopyErr *cerrors.CanopyError
	if !errors.As(err, &canopyErr) || canopyErr.Code != cerrors.ErrWorkspaceLocked {
		t.Fatalf("expected ErrWorkspaceLocked, got %v", err)
	}
}
