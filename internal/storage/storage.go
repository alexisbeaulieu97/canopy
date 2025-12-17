// Package storage manages workspace metadata and directories.
package storage

import (
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

// Create creates a new workspace directory and metadata.
func (e *Engine) Create(dirName, id, branchName string, repos []domain.Repo) error {
	safeDir, err := sanitizeDirName(dirName)
	if err != nil {
		return cerrors.NewPathInvalid(dirName, err.Error())
	}

	path := filepath.Join(e.WorkspacesRoot, safeDir)

	if err := os.Mkdir(path, 0o750); err != nil {
		if os.IsExist(err) {
			return cerrors.NewWorkspaceExists(id)
		}

		return cerrors.NewIOFailed("create workspace directory", err)
	}

	workspace := domain.Workspace{
		ID:         id,
		BranchName: branchName,
		Repos:      repos,
	}

	metaPath := filepath.Join(path, "workspace.yaml")

	return e.saveMetadata(metaPath, workspace)
}

// Save updates the metadata for an existing workspace directory.
func (e *Engine) Save(dirName string, workspace domain.Workspace) error {
	safeDir, err := sanitizeDirName(dirName)
	if err != nil {
		return cerrors.NewPathInvalid(dirName, err.Error())
	}

	path := filepath.Join(e.WorkspacesRoot, safeDir)
	metaPath := filepath.Join(path, "workspace.yaml")

	return e.saveMetadata(metaPath, workspace)
}

// Close copies workspace metadata into the closed root and returns the closed entry.
func (e *Engine) Close(dirName string, workspace domain.Workspace, closedAt time.Time) (*domain.ClosedWorkspace, error) {
	if e.ClosedRoot == "" {
		return nil, cerrors.NewConfigInvalid("closed_root is not configured")
	}

	safeDir, err := sanitizeDirName(dirName)
	if err != nil {
		return nil, cerrors.NewPathInvalid(dirName, err.Error())
	}

	closedDir := filepath.Join(e.ClosedRoot, safeDir, closedAt.UTC().Format("20060102T150405Z"))

	if err := os.MkdirAll(closedDir, 0o750); err != nil {
		return nil, cerrors.NewIOFailed("create closed directory", err)
	}

	workspace.ClosedAt = &closedAt

	metaPath := filepath.Join(closedDir, "workspace.yaml")

	if err := e.saveMetadata(metaPath, workspace); err != nil {
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
func (e *Engine) List() (map[string]domain.Workspace, error) {
	entries, err := os.ReadDir(e.WorkspacesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, cerrors.NewIOFailed("read workspaces root", err)
	}

	workspaces := make(map[string]domain.Workspace)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if w, ok := e.tryLoadMetadata(filepath.Join(e.WorkspacesRoot, entry.Name())); ok {
			workspaces[entry.Name()] = w
		}
	}

	return workspaces, nil
}

// ListClosed returns closed workspaces stored on disk, sorted by newest first.
func (e *Engine) ListClosed() ([]domain.ClosedWorkspace, error) {
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

// Load reads the metadata for a specific workspace.
func (e *Engine) Load(dirName string) (*domain.Workspace, error) {
	safeDir, err := sanitizeDirName(dirName)
	if err != nil {
		return nil, cerrors.NewPathInvalid(dirName, err.Error())
	}

	path := filepath.Join(e.WorkspacesRoot, safeDir)
	metaPath := filepath.Join(path, "workspace.yaml")

	f, err := os.Open(metaPath) //nolint:gosec // path is derived from workspace directory
	if err != nil {
		return nil, cerrors.NewWorkspaceMetadataError(dirName, "read", err)
	}

	defer func() { _ = f.Close() }()

	var w domain.Workspace
	if err := yaml.NewDecoder(f).Decode(&w); err != nil {
		return nil, cerrors.NewWorkspaceMetadataError(dirName, "decode", err)
	}

	// Handle version: missing version defaults to 0 (legacy)
	// Version 0 workspaces are auto-migrated to version 1 on next save
	// Note: Future versions are loaded as-is - callers should validate if needed

	return &w, nil
}

// LoadByID looks up a workspace by its ID and returns the workspace metadata
// and directory name. It attempts direct path access first (assuming ID == dirName),
// then falls back to scanning all workspaces if the direct lookup fails.
func (e *Engine) LoadByID(id string) (*domain.Workspace, string, error) {
	// First, try direct path access assuming ID == dirName
	safeID, err := sanitizeDirName(id)
	if err == nil {
		ws, err := e.Load(safeID)
		if err == nil && ws.ID == id {
			return ws, safeID, nil
		}
	}

	// Fallback: scan all workspaces to find the one with matching ID
	workspaces, err := e.List()
	if err != nil {
		return nil, "", cerrors.NewIOFailed("list workspaces", err)
	}

	for dirName, ws := range workspaces {
		if ws.ID == id {
			return &ws, dirName, nil
		}
	}

	return nil, "", cerrors.NewWorkspaceNotFound(id)
}

// Delete removes a workspace.
func (e *Engine) Delete(workspaceID string) error {
	safeDir, err := sanitizeDirName(workspaceID)
	if err != nil {
		return cerrors.NewPathInvalid(workspaceID, err.Error())
	}

	path := filepath.Join(e.WorkspacesRoot, safeDir)

	return os.RemoveAll(path)
}

// Rename renames a workspace directory and updates its metadata.
func (e *Engine) Rename(oldDirName, newDirName, newID string) error {
	safeOldDir, err := sanitizeDirName(oldDirName)
	if err != nil {
		return cerrors.NewPathInvalid(oldDirName, err.Error())
	}

	safeNewDir, err := sanitizeDirName(newDirName)
	if err != nil {
		return cerrors.NewPathInvalid(newDirName, err.Error())
	}

	oldPath := filepath.Join(e.WorkspacesRoot, safeOldDir)
	newPath := filepath.Join(e.WorkspacesRoot, safeNewDir)

	// Check that old path exists
	if _, err := os.Stat(oldPath); err != nil {
		if os.IsNotExist(err) {
			return cerrors.NewWorkspaceNotFound(oldDirName)
		}

		return cerrors.NewIOFailed("stat old workspace", err)
	}

	// Check that new path doesn't exist
	if _, err := os.Stat(newPath); err == nil {
		return cerrors.NewWorkspaceExists(newID)
	} else if !os.IsNotExist(err) {
		return cerrors.NewIOFailed("stat new workspace path", err)
	}

	// Load existing metadata
	ws, err := e.Load(safeOldDir)
	if err != nil {
		return cerrors.NewWorkspaceMetadataError(oldDirName, "load", err)
	}

	// Rename the directory
	if err := os.Rename(oldPath, newPath); err != nil {
		return cerrors.NewIOFailed("rename workspace directory", err)
	}

	// Update metadata with new ID
	ws.ID = newID
	if err := e.Save(safeNewDir, *ws); err != nil {
		// Attempt rollback on metadata save failure
		_ = os.Rename(newPath, oldPath)
		return cerrors.NewWorkspaceMetadataError(newID, "update", err)
	}

	return nil
}

// LatestClosed returns the newest closed entry for the given workspace ID.
func (e *Engine) LatestClosed(workspaceID string) (*domain.ClosedWorkspace, error) { //nolint:gocyclo // handles filesystem traversal and selection
	if e.ClosedRoot == "" {
		return nil, cerrors.NewConfigInvalid("closed_root is not configured")
	}

	safeDir, err := sanitizeDirName(workspaceID)
	if err != nil {
		return nil, cerrors.NewPathInvalid(workspaceID, err.Error())
	}

	workspaceDir := filepath.Join(e.ClosedRoot, safeDir)

	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, cerrors.NewWorkspaceNotFound(workspaceID).WithContext("state", "closed")
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
		return nil, cerrors.NewWorkspaceNotFound(workspaceID).WithContext("state", "closed")
	}

	return latest, nil
}

// DeleteClosed removes a closed workspace entry.
func (e *Engine) DeleteClosed(path string) error {
	if path == "" {
		return cerrors.NewInvalidArgument("path", "closed path is required")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return cerrors.NewIOFailed("resolve closed path", err)
	}

	if e.ClosedRoot != "" {
		root := filepath.Clean(e.ClosedRoot)

		if !strings.HasPrefix(absPath, root+string(os.PathSeparator)) && absPath != root {
			return cerrors.NewPathInvalid(path, "closed path must be within closed root")
		}
	}

	// Use the validated absolute path to prevent TOCTOU race conditions
	return os.RemoveAll(absPath)
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
