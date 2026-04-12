package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
)

// resolveDirectory resolves a workspace ID to its directory path.
// It requires that the workspace ID equals the directory name (after sanitization).
// If the directory exists but contains corrupt metadata, an error is returned.
func (e *Engine) resolveDirectory(id string) (string, error) {
	if dirName, ok, err := e.resolveDirectoryFromTemplate(id); err != nil {
		return "", err
	} else if ok {
		return dirName, nil
	}

	return e.resolveDirectoryByScan(id)
}

func (e *Engine) resolveDirectoryFromTemplate(id string) (string, bool, error) {
	dirName, err := e.dirNameForID(id)
	if err != nil {
		if isInvalidWorkspaceID(err) {
			return "", false, nil
		}

		return "", false, err
	}

	path := filepath.Join(e.WorkspacesRoot, dirName)

	match, err := e.matchWorkspaceID(path, id)
	if err != nil {
		return "", false, err
	}

	if match {
		return dirName, true, nil
	}

	return "", false, nil
}

func isInvalidWorkspaceID(err error) bool {
	var canopyErr *cerrors.CanopyError
	if errors.As(err, &canopyErr) {
		return canopyErr.Code == cerrors.ErrInvalidArgument
	}

	return false
}

func (e *Engine) resolveDirectoryByScan(id string) (string, error) {
	entries, err := os.ReadDir(e.WorkspacesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return "", cerrors.NewWorkspaceNotFound(id)
		}

		return "", cerrors.NewIOFailed("read workspaces root", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(e.WorkspacesRoot, entry.Name())

		match, err := e.matchWorkspaceID(path, id)
		if err != nil {
			return "", err
		}

		if match {
			return entry.Name(), nil
		}
	}

	return "", cerrors.NewWorkspaceNotFound(id)
}

func (e *Engine) matchWorkspaceID(dirPath, id string) (bool, error) {
	metaPath := filepath.Join(dirPath, "workspace.yaml")

	//nolint:gosec // metaPath is derived from workspace directory and fixed filename
	f, openErr := os.Open(metaPath)
	if openErr != nil {
		if os.IsNotExist(openErr) {
			return false, nil
		}

		return false, cerrors.NewIOFailed("open workspace metadata", openErr)
	}

	defer func() { _ = f.Close() }()

	var workspace domain.Workspace
	if decodeErr := yaml.NewDecoder(f).Decode(&workspace); decodeErr != nil {
		return false, cerrors.NewWorkspaceMetadataError(id, "decode", decodeErr)
	}

	return workspace.ID == id, nil
}

// resolveClosedDirectory resolves a closed workspace ID and timestamp to its directory path.
func (e *Engine) resolveClosedDirectory(id string, closedAt time.Time) (string, error) {
	if e.ClosedRoot == "" {
		return "", cerrors.NewConfigInvalid("closed_root is not configured")
	}

	safeDir, err := validation.NormalizeWorkspaceDirName(id)
	if err != nil {
		return "", cerrors.NewPathInvalid(id, err.Error())
	}

	closedDir := filepath.Join(e.ClosedRoot, safeDir, closedAt.UTC().Format("20060102T150405Z"))

	if _, err := os.Stat(closedDir); err != nil {
		if os.IsNotExist(err) {
			return "", cerrors.NewWorkspaceNotFound(id).WithContext("state", "closed")
		}

		return "", cerrors.NewIOFailed("stat closed directory", err)
	}

	return closedDir, nil
}

func (e *Engine) workspaceDirName(ws domain.Workspace) (string, error) {
	if strings.TrimSpace(ws.DirName) != "" {
		return validation.NormalizeWorkspaceDirName(ws.DirName)
	}

	return e.dirNameForID(ws.ID)
}
