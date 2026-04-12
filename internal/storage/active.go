package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// Create creates a new workspace from the provided domain object.
func (e *Engine) Create(_ context.Context, ws domain.Workspace) error {
	dirName, err := e.workspaceDirName(ws)
	if err != nil {
		return cerrors.NewPathInvalid(ws.ID, err.Error())
	}

	path := filepath.Join(e.WorkspacesRoot, dirName)

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

	return e.saveMetadata(filepath.Join(path, "workspace.yaml"), ws)
}

// Save persists changes to an existing workspace.
func (e *Engine) Save(_ context.Context, ws domain.Workspace) error {
	dirName, err := e.resolveDirectory(ws.ID)
	if err != nil {
		return err
	}

	return e.saveMetadata(filepath.Join(e.WorkspacesRoot, dirName, "workspace.yaml"), ws)
}

func (e *Engine) saveMetadata(path string, workspace domain.Workspace) error {
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

		if workspace, ok := e.tryLoadMetadata(filepath.Join(e.WorkspacesRoot, entry.Name())); ok {
			workspaces = append(workspaces, workspace)
		}
	}

	return workspaces, nil
}

func (e *Engine) tryLoadMetadata(dirPath string) (domain.Workspace, bool) {
	metaPath := filepath.Join(dirPath, "workspace.yaml")

	f, err := os.Open(metaPath) //nolint:gosec // path is derived from workspace directory
	if err != nil {
		return domain.Workspace{}, false
	}

	defer func() { _ = f.Close() }()

	var workspace domain.Workspace
	if err := yaml.NewDecoder(f).Decode(&workspace); err != nil {
		return domain.Workspace{}, false
	}

	workspace.DirName = filepath.Base(dirPath)

	if workspace.ClosedAt == nil {
		if ts, ok := inferClosedTimeFromPath(dirPath); ok {
			workspace.ClosedAt = &ts
		}
	}

	return workspace, true
}

func inferClosedTimeFromPath(path string) (time.Time, bool) {
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

	metaPath := filepath.Join(e.WorkspacesRoot, dirName, "workspace.yaml")

	f, err := os.Open(metaPath) //nolint:gosec // path is derived from workspace directory
	if err != nil {
		return nil, cerrors.NewWorkspaceMetadataError(id, "read", err)
	}

	defer func() { _ = f.Close() }()

	var workspace domain.Workspace
	if err := yaml.NewDecoder(f).Decode(&workspace); err != nil {
		return nil, cerrors.NewWorkspaceMetadataError(id, "decode", err)
	}

	workspace.DirName = dirName

	return &workspace, nil
}

// Delete removes a workspace by ID.
func (e *Engine) Delete(_ context.Context, id string) error {
	dirName, err := e.resolveDirectory(id)
	if err != nil {
		if errors.Is(err, cerrors.WorkspaceNotFound) {
			return nil
		}

		return err
	}

	return os.RemoveAll(filepath.Join(e.WorkspacesRoot, dirName))
}

// Rename changes a workspace's ID.
func (e *Engine) Rename(_ context.Context, oldID, newID string) error {
	oldDirName, err := e.resolveDirectory(oldID)
	if err != nil {
		return err
	}

	safeNewDir, err := e.dirNameForID(newID)
	if err != nil {
		return cerrors.NewPathInvalid(newID, err.Error())
	}

	oldPath := filepath.Join(e.WorkspacesRoot, oldDirName)
	newPath := filepath.Join(e.WorkspacesRoot, safeNewDir)

	if _, err := os.Stat(newPath); err == nil {
		return cerrors.NewWorkspaceExists(newID)
	} else if !os.IsNotExist(err) {
		return cerrors.NewIOFailed("stat new workspace path", err)
	}

	metaPath := filepath.Join(e.WorkspacesRoot, oldDirName, "workspace.yaml")

	f, err := os.Open(metaPath) //nolint:gosec // path is derived from workspace directory
	if err != nil {
		return cerrors.NewWorkspaceMetadataError(oldID, "read", err)
	}

	var workspace domain.Workspace
	if decodeErr := yaml.NewDecoder(f).Decode(&workspace); decodeErr != nil {
		_ = f.Close()
		return cerrors.NewWorkspaceMetadataError(oldID, "decode", decodeErr)
	}

	_ = f.Close()

	if err := os.Rename(oldPath, newPath); err != nil {
		return cerrors.NewIOFailed("rename workspace directory", err)
	}

	workspace.ID = newID

	if err := e.saveMetadata(filepath.Join(newPath, "workspace.yaml"), workspace); err != nil {
		_ = os.Rename(newPath, oldPath)
		return cerrors.NewWorkspaceMetadataError(newID, "update", err)
	}

	return nil
}
