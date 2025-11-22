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
)

// Engine manages workspaces
type Engine struct {
	WorkspacesRoot string
	ArchivesRoot   string
}

// New creates a new Workspace Engine
func New(workspacesRoot, archivesRoot string) *Engine {
	return &Engine{
		WorkspacesRoot: workspacesRoot,
		ArchivesRoot:   archivesRoot,
	}
}

// ArchivedWorkspace describes a stored archived workspace entry.
type ArchivedWorkspace struct {
	DirName  string
	Path     string
	Metadata domain.Workspace
}

// ArchivedAt returns the time the workspace was archived, if recorded.
func (a ArchivedWorkspace) ArchivedAt() time.Time {
	if a.Metadata.ArchivedAt != nil {
		return *a.Metadata.ArchivedAt
	}

	return time.Time{}
}

// Create creates a new workspace directory and metadata
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

	// Create metadata file
	workspace := domain.Workspace{
		ID:         id,
		BranchName: branchName,
		Repos:      repos,
	}

	metaPath := filepath.Join(path, "workspace.yaml")

	return e.saveMetadata(metaPath, workspace)
}

// Save updates the metadata for an existing workspace directory
func (e *Engine) Save(dirName string, workspace domain.Workspace) error {
	safeDir, err := sanitizeDirName(dirName)
	if err != nil {
		return fmt.Errorf("invalid workspace directory: %w", err)
	}

	path := filepath.Join(e.WorkspacesRoot, safeDir)
	metaPath := filepath.Join(path, "workspace.yaml")

	return e.saveMetadata(metaPath, workspace)
}

// Archive copies workspace metadata into the archives root and returns the archive entry.
func (e *Engine) Archive(dirName string, workspace domain.Workspace, archivedAt time.Time) (*ArchivedWorkspace, error) {
	if e.ArchivesRoot == "" {
		return nil, fmt.Errorf("archives root is not configured")
	}

	safeDir, err := sanitizeDirName(dirName)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace directory: %w", err)
	}

	archiveDir := filepath.Join(e.ArchivesRoot, safeDir, archivedAt.UTC().Format("20060102T150405Z"))

	if err := os.MkdirAll(archiveDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create archive directory: %w", err)
	}

	workspace.ArchivedAt = &archivedAt

	metaPath := filepath.Join(archiveDir, "workspace.yaml")

	if err := e.saveMetadata(metaPath, workspace); err != nil {
		return nil, fmt.Errorf("failed to write archive metadata: %w", err)
	}

	return &ArchivedWorkspace{
		DirName:  safeDir,
		Path:     archiveDir,
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

// List returns all active workspaces
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

// ListArchived returns archived workspaces stored on disk, sorted by newest first.
func (e *Engine) ListArchived() ([]ArchivedWorkspace, error) {
	if e.ArchivesRoot == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(e.ArchivesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read archives root: %w", err)
	}

	var archives []ArchivedWorkspace

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workspaceDir := filepath.Join(e.ArchivesRoot, entry.Name())

		versionDirs, err := os.ReadDir(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read archive directory %s: %w", workspaceDir, err)
		}

		for _, version := range versionDirs {
			if !version.IsDir() {
				continue
			}

			dirPath := filepath.Join(workspaceDir, version.Name())

			if w, ok := e.tryLoadMetadata(dirPath); ok {
				archives = append(archives, ArchivedWorkspace{
					DirName:  entry.Name(),
					Path:     dirPath,
					Metadata: w,
				})
			}
		}
	}

	sort.Slice(archives, func(i, j int) bool {
		return archives[i].ArchivedAt().After(archives[j].ArchivedAt())
	})

	return archives, nil
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

	return w, true
}

// Load reads the metadata for a specific workspace
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

// Delete removes a workspace
func (e *Engine) Delete(workspaceID string) error {
	safeDir, err := sanitizeDirName(workspaceID)
	if err != nil {
		return fmt.Errorf("invalid workspace directory: %w", err)
	}

	path := filepath.Join(e.WorkspacesRoot, safeDir)

	return os.RemoveAll(path)
}

// LatestArchive returns the newest archived entry for the given workspace ID.
func (e *Engine) LatestArchive(workspaceID string) (*ArchivedWorkspace, error) { //nolint:gocyclo // handles filesystem traversal and selection
	if e.ArchivesRoot == "" {
		return nil, fmt.Errorf("archives root is not configured")
	}

	safeDir, err := sanitizeDirName(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace id: %w", err)
	}

	workspaceDir := filepath.Join(e.ArchivesRoot, safeDir)

	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("archived workspace %s not found", workspaceID)
		}

		return nil, fmt.Errorf("failed to read archives: %w", err)
	}

	var latest *ArchivedWorkspace

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(workspaceDir, entry.Name())

		if w, ok := e.tryLoadMetadata(dirPath); ok {
			candidate := &ArchivedWorkspace{
				DirName:  safeDir,
				Path:     dirPath,
				Metadata: w,
			}

			if latest == nil || candidate.ArchivedAt().After(latest.ArchivedAt()) {
				latest = candidate
			}
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("archived workspace %s not found", workspaceID)
	}

	return latest, nil
}

// DeleteArchive removes an archived workspace entry.
func (e *Engine) DeleteArchive(path string) error {
	if path == "" {
		return fmt.Errorf("archive path is required")
	}

	if e.ArchivesRoot != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to resolve archive path: %w", err)
		}

		root := filepath.Clean(e.ArchivesRoot)

		if !strings.HasPrefix(absPath, root+string(os.PathSeparator)) && absPath != root {
			return fmt.Errorf("archive path must be within archives root")
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
