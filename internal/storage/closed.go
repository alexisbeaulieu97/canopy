package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
)

// Close archives a workspace and returns the closed entry.
func (e *Engine) Close(_ context.Context, id string, closedAt time.Time) (*domain.ClosedWorkspace, error) {
	if e.ClosedRoot == "" {
		return nil, cerrors.NewConfigInvalid("closed_root is not configured")
	}

	dirName, err := e.resolveDirectory(id)
	if err != nil {
		return nil, err
	}

	metaPath := filepath.Join(e.WorkspacesRoot, dirName, "workspace.yaml")

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

	safeDir, err := validation.NormalizeWorkspaceDirName(workspace.ID)
	if err != nil {
		return nil, cerrors.NewPathInvalid(dirName, err.Error())
	}

	closedDir := filepath.Join(e.ClosedRoot, safeDir, closedAt.UTC().Format("20060102T150405Z"))
	if err := os.MkdirAll(closedDir, 0o750); err != nil {
		return nil, cerrors.NewIOFailed("create closed directory", err)
	}

	workspace.ClosedAt = &closedAt

	if err := e.saveMetadata(filepath.Join(closedDir, "workspace.yaml"), workspace); err != nil {
		return nil, cerrors.NewWorkspaceMetadataError(workspace.ID, "write", err)
	}

	return &domain.ClosedWorkspace{
		DirName:  safeDir,
		Path:     closedDir,
		Metadata: workspace,
	}, nil
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
			if workspace, ok := e.tryLoadMetadata(dirPath); ok {
				closed = append(closed, domain.ClosedWorkspace{
					DirName:  entry.Name(),
					Path:     dirPath,
					Metadata: workspace,
				})
			}
		}
	}

	sort.Slice(closed, func(i, j int) bool {
		return closed[i].ClosedAt().After(closed[j].ClosedAt())
	})

	return closed, nil
}

// LatestClosed returns the most recent closed entry for a workspace.
func (e *Engine) LatestClosed(_ context.Context, id string) (*domain.ClosedWorkspace, error) { //nolint:gocyclo // handles filesystem traversal and selection
	if e.ClosedRoot == "" {
		return nil, cerrors.NewConfigInvalid("closed_root is not configured")
	}

	safeDir, err := validation.NormalizeWorkspaceDirName(id)
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
		if workspace, ok := e.tryLoadMetadata(dirPath); ok {
			candidate := &domain.ClosedWorkspace{
				DirName:  safeDir,
				Path:     dirPath,
				Metadata: workspace,
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
		if errors.Is(err, cerrors.WorkspaceNotFound) {
			return nil
		}

		return err
	}

	return os.RemoveAll(closedDir)
}
