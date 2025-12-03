// Package workspace manages workspace metadata and directories.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
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

// ClosedWorkspace is an alias for domain.ClosedWorkspace for backward compatibility.
//
// Deprecated: Use domain.ClosedWorkspace directly.
type ClosedWorkspace = domain.ClosedWorkspace

// Create creates a new workspace directory and metadata.
func (e *Engine) Create(dirName, id, branchName string, repos []domain.Repo) error {
	safeDir, err := sanitizeDirName(dirName)
	if err != nil {
		return fmt.Errorf("invalid workspace directory: %w", err)
	}

	path := filepath.Join(e.WorkspacesRoot, safeDir)

	if err := os.Mkdir(path, 0o750); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("workspace already exists: %s", path)
		}

		return fmt.Errorf("failed to create workspace directory: %w", err)
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
		return fmt.Errorf("invalid workspace directory: %w", err)
	}

	path := filepath.Join(e.WorkspacesRoot, safeDir)
	metaPath := filepath.Join(path, "workspace.yaml")

	return e.saveMetadata(metaPath, workspace)
}

// Close copies workspace metadata into the closed root and returns the closed entry.
func (e *Engine) Close(dirName string, workspace domain.Workspace, closedAt time.Time) (*ClosedWorkspace, error) {
	if e.ClosedRoot == "" {
		return nil, fmt.Errorf("closed root is not configured")
	}

	safeDir, err := sanitizeDirName(dirName)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace directory: %w", err)
	}

	closedDir := filepath.Join(e.ClosedRoot, safeDir, closedAt.UTC().Format("20060102T150405Z"))

	if err := os.MkdirAll(closedDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create closed directory: %w", err)
	}

	workspace.ClosedAt = &closedAt

	metaPath := filepath.Join(closedDir, "workspace.yaml")

	if err := e.saveMetadata(metaPath, workspace); err != nil {
		return nil, fmt.Errorf("failed to write closed metadata: %w", err)
	}

	return &ClosedWorkspace{
		DirName:  safeDir,
		Path:     closedDir,
		Metadata: workspace,
	}, nil
}

func (e *Engine) saveMetadata(path string, workspace domain.Workspace) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o640) //nolint:gosec // path is constructed internally
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}

	defer func() { _ = f.Close() }()

	enc := yaml.NewEncoder(f)
	if err := enc.Encode(workspace); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	if err := enc.Close(); err != nil {
		return fmt.Errorf("failed to flush metadata: %w", err)
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

		return nil, fmt.Errorf("failed to read workspaces root: %w", err)
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
func (e *Engine) ListClosed() ([]ClosedWorkspace, error) {
	if e.ClosedRoot == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(e.ClosedRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read closed root: %w", err)
	}

	var closed []ClosedWorkspace

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workspaceDir := filepath.Join(e.ClosedRoot, entry.Name())

		versionDirs, err := os.ReadDir(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read closed directory %s: %w", workspaceDir, err)
		}

		for _, version := range versionDirs {
			if !version.IsDir() {
				continue
			}

			dirPath := filepath.Join(workspaceDir, version.Name())

			if w, ok := e.tryLoadMetadata(dirPath); ok {
				closed = append(closed, ClosedWorkspace{
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
		return nil, fmt.Errorf("invalid workspace directory: %w", err)
	}

	path := filepath.Join(e.WorkspacesRoot, safeDir)
	metaPath := filepath.Join(path, "workspace.yaml")

	f, err := os.Open(metaPath) //nolint:gosec // path is derived from workspace directory
	if err != nil {
		return nil, fmt.Errorf("failed to open workspace metadata: %w", err)
	}

	defer func() { _ = f.Close() }()

	var w domain.Workspace
	if err := yaml.NewDecoder(f).Decode(&w); err != nil {
		return nil, fmt.Errorf("failed to decode workspace metadata: %w", err)
	}

	return &w, nil
}

// Delete removes a workspace.
func (e *Engine) Delete(workspaceID string) error {
	safeDir, err := sanitizeDirName(workspaceID)
	if err != nil {
		return fmt.Errorf("invalid workspace directory: %w", err)
	}

	path := filepath.Join(e.WorkspacesRoot, safeDir)

	return os.RemoveAll(path)
}

// LatestClosed returns the newest closed entry for the given workspace ID.
func (e *Engine) LatestClosed(workspaceID string) (*ClosedWorkspace, error) { //nolint:gocyclo // handles filesystem traversal and selection
	if e.ClosedRoot == "" {
		return nil, fmt.Errorf("closed root is not configured")
	}

	safeDir, err := sanitizeDirName(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace id: %w", err)
	}

	workspaceDir := filepath.Join(e.ClosedRoot, safeDir)

	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("closed workspace %s not found", workspaceID)
		}

		return nil, fmt.Errorf("failed to read closed entries: %w", err)
	}

	var latest *ClosedWorkspace

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(workspaceDir, entry.Name())

		if w, ok := e.tryLoadMetadata(dirPath); ok {
			candidate := &ClosedWorkspace{
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
		return nil, fmt.Errorf("closed workspace %s not found", workspaceID)
	}

	return latest, nil
}

// DeleteClosed removes a closed workspace entry.
func (e *Engine) DeleteClosed(path string) error {
	if path == "" {
		return fmt.Errorf("closed path is required")
	}

	if e.ClosedRoot != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to resolve closed path: %w", err)
		}

		root := filepath.Clean(e.ClosedRoot)

		if !strings.HasPrefix(absPath, root+string(os.PathSeparator)) && absPath != root {
			return fmt.Errorf("closed path must be within closed root")
		}
	}

	return os.RemoveAll(path)
}

func sanitizeDirName(name string) (string, error) {
	cleaned := filepath.Clean(strings.TrimSpace(name))
	if cleaned == "" || cleaned == "." {
		return "", fmt.Errorf("workspace name cannot be empty")
	}

	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("workspace name must be relative")
	}

	if cleaned != filepath.Base(cleaned) || strings.Contains(cleaned, "..") || strings.ContainsRune(cleaned, filepath.Separator) {
		return "", fmt.Errorf("workspace name contains invalid path elements")
	}

	return cleaned, nil
}
