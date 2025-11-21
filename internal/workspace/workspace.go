// Package workspace manages workspace metadata and directories.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alexisbeaulieu97/yard/internal/domain"
)

// Engine manages workspaces
type Engine struct {
	WorkspacesRoot string
}

// New creates a new Workspace Engine
func New(workspacesRoot string) *Engine {
	return &Engine{WorkspacesRoot: workspacesRoot}
}

// Create creates a new workspace directory and metadata
func (e *Engine) Create(dirName, id, slug, branchName string, repos []domain.Repo) error {
	safeDir, err := sanitizeDirName(dirName)
	if err != nil {
		return fmt.Errorf("invalid workspace directory: %w", err)
	}

	path := filepath.Join(e.WorkspacesRoot, safeDir)

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("workspace already exists: %s", path)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat workspace path: %w", err)
	}

	if err := os.MkdirAll(path, 0o750); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Create metadata file
	workspace := domain.Workspace{
		ID:         id,
		Slug:       slug,
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

func (e *Engine) tryLoadMetadata(dirPath string) (domain.Workspace, bool) {
	metaPath := filepath.Join(dirPath, "workspace.yaml")

	f, err := os.Open(metaPath) //nolint:gosec // path is derived from workspace directory
	if err != nil {
		metaPath = filepath.Join(dirPath, "ticket.yaml")

		f, err = os.Open(metaPath) //nolint:gosec // path is derived from workspace directory
		if err != nil {
			return domain.Workspace{}, false
		}
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
		// Try legacy ticket.yaml
		metaPath = filepath.Join(path, "ticket.yaml")

		f, err = os.Open(metaPath) //nolint:gosec // path is derived from workspace directory
		if err != nil {
			return nil, fmt.Errorf("failed to open workspace metadata: %w", err)
		}
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
