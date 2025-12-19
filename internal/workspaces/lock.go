// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
)

const lockFileName = ".canopy.lock"

// LockManager manages workspace-level file locks.
type LockManager struct {
	root           string
	timeout        time.Duration
	staleThreshold time.Duration
	logger         *logging.Logger
	now            func() time.Time
	sleep          func(time.Duration)
}

// LockHandle represents an acquired lock.
type LockHandle struct {
	workspaceID string
	path        string
	file        *os.File
	logger      *logging.Logger
}

// NewLockManager creates a new LockManager.
func NewLockManager(root string, timeout, staleThreshold time.Duration, logger *logging.Logger) *LockManager {
	return &LockManager{
		root:           root,
		timeout:        timeout,
		staleThreshold: staleThreshold,
		logger:         logger,
		now:            time.Now,
		sleep:          time.Sleep,
	}
}

// Acquire obtains an exclusive lock for a workspace.
// If createDir is true, the workspace directory is created when missing.
func (m *LockManager) Acquire(ctx context.Context, workspaceID string, createDir bool) (*LockHandle, error) {
	lockPath, err := m.lockPath(workspaceID, createDir)
	if err != nil {
		return nil, err
	}

	deadline := m.now().Add(m.timeout)

	for {
		if ctx.Err() != nil {
			return nil, cerrors.NewContextError(ctx, "acquire lock", workspaceID)
		}

		file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600) //nolint:gosec // lockPath is derived from workspace root and ID
		if err == nil {
			handle := &LockHandle{
				workspaceID: workspaceID,
				path:        lockPath,
				file:        file,
				logger:      m.logger,
			}

			if m.logger != nil {
				m.logger.Debug("workspace lock acquired", "workspace_id", workspaceID, "path", lockPath)
			}

			return handle, nil
		}

		if !errors.Is(err, os.ErrExist) {
			return nil, cerrors.NewIOFailed(fmt.Sprintf("acquire lock %s", lockPath), err)
		}

		stale, staleErr := m.removeIfStale(lockPath)
		if staleErr != nil {
			return nil, staleErr
		}

		if stale {
			continue
		}

		if m.now().After(deadline) {
			return nil, cerrors.NewWorkspaceLocked(workspaceID)
		}

		m.sleep(100 * time.Millisecond)
	}
}

// IsLocked reports whether a non-stale lock exists for the workspace.
func (m *LockManager) IsLocked(workspaceID string) (bool, error) {
	lockPath, err := m.lockPath(workspaceID, false)
	if err != nil {
		return false, err
	}

	info, err := os.Stat(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, cerrors.NewIOFailed(fmt.Sprintf("stat lock %s", lockPath), err)
	}

	if m.staleThreshold > 0 && m.now().Sub(info.ModTime()) > m.staleThreshold {
		return false, nil
	}

	return true, nil
}

func (m *LockManager) lockPath(workspaceID string, createDir bool) (string, error) {
	workspacePath := filepath.Join(m.root, workspaceID)
	if createDir {
		if err := os.MkdirAll(workspacePath, 0o750); err != nil {
			return "", cerrors.NewIOFailed("create workspace directory", err)
		}
	} else {
		if _, err := os.Stat(workspacePath); err != nil {
			if os.IsNotExist(err) {
				return "", cerrors.NewWorkspaceNotFound(workspaceID)
			}

			return "", cerrors.NewIOFailed("stat workspace directory", err)
		}
	}

	return filepath.Join(workspacePath, lockFileName), nil
}

func (m *LockManager) removeIfStale(lockPath string) (bool, error) {
	if m.staleThreshold <= 0 {
		return false, nil
	}

	info, err := os.Stat(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, cerrors.NewIOFailed(fmt.Sprintf("stat lock %s", lockPath), err)
	}

	if m.now().Sub(info.ModTime()) <= m.staleThreshold {
		return false, nil
	}

	if err := os.Remove(lockPath); err != nil {
		return false, cerrors.NewIOFailed(fmt.Sprintf("remove stale lock %s", lockPath), err)
	}

	if m.logger != nil {
		m.logger.Debug("stale workspace lock removed", "path", lockPath)
	}

	return true, nil
}

// Release releases the lock and removes the lock file.
func (h *LockHandle) Release() error {
	if h.file != nil {
		_ = h.file.Close()
	}

	if err := os.Remove(h.path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return cerrors.NewIOFailed(fmt.Sprintf("release lock %s", h.path), err)
	}

	if h.logger != nil {
		h.logger.Debug("workspace lock released", "workspace_id", h.workspaceID, "path", h.path)
	}

	return nil
}
