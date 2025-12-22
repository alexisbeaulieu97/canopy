package workspaces

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

func TestLockManagerAcquireRelease(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	lm := NewLockManager(root, 500*time.Millisecond, time.Minute, nil, nil)

	handle, err := lm.Acquire(context.Background(), "WS-1", true)
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}

	lockPath := filepath.Join(root, "WS-1", lockFileName)
	if _, statErr := os.Stat(lockPath); statErr != nil {
		t.Fatalf("expected lock file to exist: %v", statErr)
	}

	if err := handle.Release(); err != nil {
		t.Fatalf("release failed: %v", err)
	}

	if _, statErr := os.Stat(lockPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected lock file to be removed, got %v", statErr)
	}
}

func TestLockManagerTimeout(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	lm := NewLockManager(root, 200*time.Millisecond, time.Minute, nil, nil)

	handle, err := lm.Acquire(context.Background(), "WS-1", true)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	defer func() { _ = handle.Release() }()

	_, err = lm.Acquire(context.Background(), "WS-1", false)
	if err == nil {
		t.Fatalf("expected lock timeout error")
	}

	var canopyErr *cerrors.CanopyError
	if !errors.As(err, &canopyErr) || canopyErr.Code != cerrors.ErrWorkspaceLocked {
		t.Fatalf("expected ErrWorkspaceLocked, got %v", err)
	}
}

func TestLockManagerStaleCleanup(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	staleThreshold := 50 * time.Millisecond
	lm := NewLockManager(root, 200*time.Millisecond, staleThreshold, nil, nil)

	lockPath := filepath.Join(root, "WS-1", lockFileName)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o750); err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	if err := os.WriteFile(lockPath, []byte("stale"), 0o600); err != nil {
		t.Fatalf("failed to create stale lock: %v", err)
	}

	staleTime := time.Now().Add(-1 * time.Second)
	if err := os.Chtimes(lockPath, staleTime, staleTime); err != nil {
		t.Fatalf("failed to backdate lock: %v", err)
	}

	handle, err := lm.Acquire(context.Background(), "WS-1", false)
	if err != nil {
		t.Fatalf("acquire after stale cleanup failed: %v", err)
	}

	if err := handle.Release(); err != nil {
		t.Fatalf("release failed: %v", err)
	}
}
