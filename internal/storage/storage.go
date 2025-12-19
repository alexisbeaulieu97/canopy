// Package storage manages workspace metadata and directories.
package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Compile-time check that Engine implements ports.WorkspaceStorage.
var _ ports.WorkspaceStorage = (*Engine)(nil)

// Engine manages workspaces.
type Engine struct {
	WorkspacesRoot string
	ClosedRoot     string
}

// New creates a new Workspace Engine.
func New(workspacesRoot, closedRoot string) *Engine {
	return &Engine{
		WorkspacesRoot: workspacesRoot,
		ClosedRoot:     closedRoot,
	}
}

// resolveDirectory resolves a workspace ID to its directory path.
// It requires that the workspace ID equals the directory name (after sanitization).
// If the directory exists but contains corrupt metadata, an error is returned.
func (e *Engine) resolveDirectory(id string) (string, error) {
	safeID, err := sanitizeDirName(id)
	if err != nil {
		return "", cerrors.NewWorkspaceNotFound(id)
	}

	path := filepath.Join(e.WorkspacesRoot, safeID)
	metaPath := filepath.Join(path, "workspace.yaml")

	//nolint:gosec // metaPath is derived from sanitized workspace ID and fixed filename
	f, openErr := os.Open(metaPath)
	if openErr != nil {
		if os.IsNotExist(openErr) {
			return "", cerrors.NewWorkspaceNotFound(id)
		}

		return "", cerrors.NewIOFailed("open workspace metadata", openErr)
	}

	defer func() { _ = f.Close() }()

	var w domain.Workspace
	if decodeErr := yaml.NewDecoder(f).Decode(&w); decodeErr != nil {
		return "", cerrors.NewWorkspaceMetadataError(id, "decode", decodeErr)
	}

	if w.ID != id {
		return "", cerrors.NewWorkspaceNotFound(id)
	}

	return safeID, nil
}

// resolveClosedDirectory resolves a closed workspace ID and timestamp to its directory path.
func (e *Engine) resolveClosedDirectory(id string, closedAt time.Time) (string, error) {
	if e.ClosedRoot == "" {
		return "", cerrors.NewConfigInvalid("closed_root is not configured")
	}

	safeDir, err := sanitizeDirName(id)
	if err != nil {
		return "", cerrors.NewPathInvalid(id, err.Error())
	}

	timestampDir := closedAt.UTC().Format("20060102T150405Z")
	closedDir := filepath.Join(e.ClosedRoot, safeDir, timestampDir)

	// Verify the directory exists
	if _, err := os.Stat(closedDir); err != nil {
		if os.IsNotExist(err) {
			return "", cerrors.NewWorkspaceNotFound(id).WithContext("state", "closed")
		}

		return "", cerrors.NewIOFailed("stat closed directory", err)
	}

	return closedDir, nil
}

// Create creates a new workspace from the provided domain object.
func (e *Engine) Create(_ context.Context, ws domain.Workspace) error {
	safeDir, err := sanitizeDirName(ws.ID)
	if err != nil {
		return cerrors.NewPathInvalid(ws.ID, err.Error())
	}

	path := filepath.Join(e.WorkspacesRoot, safeDir)

	if err := os.Mkdir(path, 0o750); err != nil {
		if os.IsExist(err) {
			metaPath := filepath.Join(path, "workspace.yaml")

			_, statErr := os.Stat(metaPath)
			if statErr == nil {
				return cerrors.NewWorkspaceExists(ws.ID)
			}

			if statErr != nil && !os.IsNotExist(statErr) {
				return cerrors.NewIOFailed("check workspace metadata", statErr)
			}
		} else {
			return cerrors.NewIOFailed("create workspace directory", err)
		}
	}

	metaPath := filepath.Join(path, "workspace.yaml")

	return e.saveMetadata(metaPath, ws)
}

// Save persists changes to an existing workspace.
func (e *Engine) Save(_ context.Context, ws domain.Workspace) error {
	dirName, err := e.resolveDirectory(ws.ID)
	if err != nil {
		return err
	}

	path := filepath.Join(e.WorkspacesRoot, dirName)
	metaPath := filepath.Join(path, "workspace.yaml")

	return e.saveMetadata(metaPath, ws)
}

// Close archives a workspace and returns the closed entry.
func (e *Engine) Close(_ context.Context, id string, closedAt time.Time) (*domain.ClosedWorkspace, error) {
	if e.ClosedRoot == "" {
		return nil, cerrors.NewConfigInvalid("closed_root is not configured")
	}

	dirName, err := e.resolveDirectory(id)
	if err != nil {
		return nil, err
	}

	// Load the workspace metadata
	path := filepath.Join(e.WorkspacesRoot, dirName)
	metaPath := filepath.Join(path, "workspace.yaml")

	f, err := os.Open(metaPath) //nolint:gosec // path is derived from workspace directory
	if err != nil {
		return nil, cerrors.NewWorkspaceMetadataError(id, "read", err)
	}

	var workspace domain.Workspace
	if err := yaml.NewDecoder(f).Decode(&workspace); err != nil {
		_ = f.Close()
		return nil, cerrors.NewWorkspaceMetadataError(id, "decode", err)
	}

	_ = f.Close()

	safeDir, err := sanitizeDirName(dirName)
	if err != nil {
		return nil, cerrors.NewPathInvalid(dirName, err.Error())
	}

	closedDir := filepath.Join(e.ClosedRoot, safeDir, closedAt.UTC().Format("20060102T150405Z"))

	if err := os.MkdirAll(closedDir, 0o750); err != nil {
		return nil, cerrors.NewIOFailed("create closed directory", err)
	}

	workspace.ClosedAt = &closedAt

	closedMetaPath := filepath.Join(closedDir, "workspace.yaml")

	if err := e.saveMetadata(closedMetaPath, workspace); err != nil {
		return nil, cerrors.NewWorkspaceMetadataError(workspace.ID, "write", err)
	}

	return &domain.ClosedWorkspace{
		DirName:  safeDir,
		Path:     closedDir,
		Metadata: workspace,
	}, nil
}

func (e *Engine) saveMetadata(path string, workspace domain.Workspace) error {
	// Always write current schema version
	workspace.Version = domain.CurrentWorkspaceVersion

	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o640) //nolint:gosec // path is constructed internally
	if err != nil {
		return cerrors.NewIOFailed("create metadata file", err)
	}

	defer func() { _ = f.Close() }()

	enc := yaml.NewEncoder(f)
	if err := enc.Encode(workspace); err != nil {
		return cerrors.NewIOFailed("encode metadata", err)
	}

	if err := enc.Close(); err != nil {
		return cerrors.NewIOFailed("flush metadata", err)
	}

	return nil
}

// List returns all active workspaces.
func (e *Engine) List(_ context.Context) ([]domain.Workspace, error) {
	entries, err := os.ReadDir(e.WorkspacesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, cerrors.NewIOFailed("read workspaces root", err)
	}

	var workspaces []domain.Workspace

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if w, ok := e.tryLoadMetadata(filepath.Join(e.WorkspacesRoot, entry.Name())); ok {
			workspaces = append(workspaces, w)
		}
	}

	return workspaces, nil
}

// ListClosed returns closed workspaces stored on disk, sorted by newest first.
func (e *Engine) ListClosed(_ context.Context) ([]domain.ClosedWorkspace, error) {
	if e.ClosedRoot == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(e.ClosedRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, cerrors.NewIOFailed("read closed root", err)
	}

	var closed []domain.ClosedWorkspace

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workspaceDir := filepath.Join(e.ClosedRoot, entry.Name())

		versionDirs, err := os.ReadDir(workspaceDir)
		if err != nil {
			return nil, cerrors.NewIOFailed("read closed directory", err)
		}

		for _, version := range versionDirs {
			if !version.IsDir() {
				continue
			}

			dirPath := filepath.Join(workspaceDir, version.Name())

			if w, ok := e.tryLoadMetadata(dirPath); ok {
				closed = append(closed, domain.ClosedWorkspace{
					DirName:  entry.Name(),
					Path:     dirPath,
					Metadata: w,
				})
			}
		}
	}

	sort.Slice(closed, func(i, j int) bool {
		return closed[i].ClosedAt().After(closed[j].ClosedAt())
	})

	return closed, nil
}

func (e *Engine) tryLoadMetadata(dirPath string) (domain.Workspace, bool) {
	metaPath := filepath.Join(dirPath, "workspace.yaml")

	f, err := os.Open(metaPath) //nolint:gosec // path is derived from workspace directory
	if err != nil {
		return domain.Workspace{}, false
	}

	defer func() { _ = f.Close() }()

	var w domain.Workspace
	if err := yaml.NewDecoder(f).Decode(&w); err != nil {
		return domain.Workspace{}, false
	}

	// Handle version: missing version defaults to 0 (legacy)
	// Version 0 workspaces are auto-migrated to version 1 on next save
	// Note: Future versions are loaded as-is - callers should validate if needed

	// Fallback for older metadata: infer closed time from directory name when ClosedAt is missing.
	if w.ClosedAt == nil {
		if ts, ok := inferClosedTimeFromPath(dirPath); ok {
			w.ClosedAt = &ts
		}
	}

	return w, true
}

func inferClosedTimeFromPath(path string) (time.Time, bool) {
	// Expect paths like .../<workspace>/<timestamp>/workspace.yaml
	parent := filepath.Base(filepath.Dir(path))

	ts, err := time.Parse("20060102T150405Z", parent)
	if err != nil {
		return time.Time{}, false
	}

	return ts, true
}

// Load retrieves a workspace by ID.
func (e *Engine) Load(_ context.Context, id string) (*domain.Workspace, error) {
	dirName, err := e.resolveDirectory(id)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(e.WorkspacesRoot, dirName)
	metaPath := filepath.Join(path, "workspace.yaml")

	f, err := os.Open(metaPath) //nolint:gosec // path is derived from workspace directory
	if err != nil {
		return nil, cerrors.NewWorkspaceMetadataError(id, "read", err)
	}

	defer func() { _ = f.Close() }()

	var w domain.Workspace
	if err := yaml.NewDecoder(f).Decode(&w); err != nil {
		return nil, cerrors.NewWorkspaceMetadataError(id, "decode", err)
	}

	// Handle version: missing version defaults to 0 (legacy)
	// Version 0 workspaces are auto-migrated to version 1 on next save
	// Note: Future versions are loaded as-is - callers should validate if needed

	return &w, nil
}

// Delete removes a workspace by ID.
func (e *Engine) Delete(_ context.Context, id string) error {
	dirName, err := e.resolveDirectory(id)
	if err != nil {
		// If workspace not found, deletion is idempotent - return nil
		if errors.Is(err, cerrors.WorkspaceNotFound) {
			return nil
		}

		return err
	}

	path := filepath.Join(e.WorkspacesRoot, dirName)

	return os.RemoveAll(path)
}

// Rename changes a workspace's ID.
func (e *Engine) Rename(_ context.Context, oldID, newID string) error {
	oldDirName, err := e.resolveDirectory(oldID)
	if err != nil {
		return err
	}

	safeNewDir, err := sanitizeDirName(newID)
	if err != nil {
		return cerrors.NewPathInvalid(newID, err.Error())
	}

	oldPath := filepath.Join(e.WorkspacesRoot, oldDirName)
	newPath := filepath.Join(e.WorkspacesRoot, safeNewDir)

	// Check that new path doesn't exist
	if _, err := os.Stat(newPath); err == nil {
		return cerrors.NewWorkspaceExists(newID)
	} else if !os.IsNotExist(err) {
		return cerrors.NewIOFailed("stat new workspace path", err)
	}

	// Load existing metadata
	path := filepath.Join(e.WorkspacesRoot, oldDirName)
	metaPath := filepath.Join(path, "workspace.yaml")

	f, err := os.Open(metaPath) //nolint:gosec // path is derived from workspace directory
	if err != nil {
		return cerrors.NewWorkspaceMetadataError(oldID, "read", err)
	}

	var ws domain.Workspace
	if decodeErr := yaml.NewDecoder(f).Decode(&ws); decodeErr != nil {
		_ = f.Close()
		return cerrors.NewWorkspaceMetadataError(oldID, "decode", decodeErr)
	}

	_ = f.Close()

	// Rename the directory
	if err := os.Rename(oldPath, newPath); err != nil {
		return cerrors.NewIOFailed("rename workspace directory", err)
	}

	// Update metadata with new ID
	ws.ID = newID
	newMetaPath := filepath.Join(newPath, "workspace.yaml")

	if err := e.saveMetadata(newMetaPath, ws); err != nil {
		// Attempt rollback on metadata save failure
		_ = os.Rename(newPath, oldPath)
		return cerrors.NewWorkspaceMetadataError(newID, "update", err)
	}

	return nil
}

// LatestClosed returns the most recent closed entry for a workspace.
func (e *Engine) LatestClosed(_ context.Context, id string) (*domain.ClosedWorkspace, error) { //nolint:gocyclo // handles filesystem traversal and selection
	if e.ClosedRoot == "" {
		return nil, cerrors.NewConfigInvalid("closed_root is not configured")
	}

	safeDir, err := sanitizeDirName(id)
	if err != nil {
		return nil, cerrors.NewPathInvalid(id, err.Error())
	}

	workspaceDir := filepath.Join(e.ClosedRoot, safeDir)

	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, cerrors.NewWorkspaceNotFound(id).WithContext("state", "closed")
		}

		return nil, cerrors.NewIOFailed("read closed entries", err)
	}

	var latest *domain.ClosedWorkspace

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(workspaceDir, entry.Name())

		if w, ok := e.tryLoadMetadata(dirPath); ok {
			candidate := &domain.ClosedWorkspace{
				DirName:  safeDir,
				Path:     dirPath,
				Metadata: w,
			}

			if latest == nil || candidate.ClosedAt().After(latest.ClosedAt()) {
				latest = candidate
			}
		}
	}

	if latest == nil {
		return nil, cerrors.NewWorkspaceNotFound(id).WithContext("state", "closed")
	}

	return latest, nil
}

// DeleteClosed removes a closed workspace entry identified by workspace ID and close timestamp.
func (e *Engine) DeleteClosed(_ context.Context, id string, closedAt time.Time) error {
	closedDir, err := e.resolveClosedDirectory(id, closedAt)
	if err != nil {
		// If closed workspace not found, deletion is idempotent - return nil
		if errors.Is(err, cerrors.WorkspaceNotFound) {
			return nil
		}

		return err
	}

	return os.RemoveAll(closedDir)
}

func sanitizeDirName(name string) (string, error) {
	cleaned := filepath.Clean(strings.TrimSpace(name))
	if cleaned == "" || cleaned == "." {
		return "", cerrors.NewInvalidArgument("name", "workspace name cannot be empty")
	}

	if filepath.IsAbs(cleaned) {
		return "", cerrors.NewInvalidArgument("name", "workspace name must be relative")
	}

	if cleaned != filepath.Base(cleaned) || strings.Contains(cleaned, "..") || strings.ContainsRune(cleaned, filepath.Separator) {
		return "", cerrors.NewInvalidArgument("name", "workspace name contains invalid path elements")
	}

	return cleaned, nil
}
