// Package domain contains core domain models for Canopy.
//
// This package defines the fundamental data structures used throughout the application.
// Domain types are pure data with no external dependencies, making them safe to use
// across all layers of the architecture.
//
// # Key Types
//
// Workspace-related types:
//   - Workspace: Represents an active workspace with its repositories
//   - ClosedWorkspace: Represents an archived workspace
//   - WorkspaceStatus: Aggregate git status for a workspace
//   - WorkspaceClosePreview: Preview of what closing a workspace would do
//   - WorkspaceExport: Portable format for workspace import/export
//
// Repository-related types:
//   - Repo: A git repository with name and URL
//   - RepoStatus: Git status of a single repository
//   - RepoRemovePreview: Preview of what removing a repo would do
//
// Hook-related types:
//   - HookContext: Context provided to lifecycle hooks
//
// Orphan detection:
//   - OrphanedWorktree: A worktree with missing or invalid references
//   - OrphanReason: Why a worktree is considered orphaned
//
// Canonical repository status:
//   - CanonicalRepoStatus: Detailed status of a bare repository
package domain

import "time"

// CurrentWorkspaceVersion is the current workspace metadata schema version.
// Version history:
//   - 0: Legacy workspaces without version field (implicit)
//   - 1: First versioned schema (adds version field)
const CurrentWorkspaceVersion = 1

// Repo represents a git repository
type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// Workspace represents a work item
type Workspace struct {
	Version         int        `yaml:"version"`
	ID              string     `yaml:"id"`
	BranchName      string     `yaml:"branch_name,omitempty"`
	Repos           []Repo     `yaml:"repos"`
	ClosedAt        *time.Time `yaml:"closed_at,omitempty"`
	SetupIncomplete bool       `yaml:"setup_incomplete,omitempty"`
	Locked          bool       `yaml:"-" json:"locked,omitempty"`
	DirName         string     `yaml:"-" json:"-"`
	LastModified    time.Time  `yaml:"-"`
	DiskUsageBytes  int64      `yaml:"-"`
}

// ClosedWorkspace describes a stored closed workspace entry.
type ClosedWorkspace struct {
	DirName  string
	Path     string
	Metadata Workspace
}

// ClosedAt returns the time the workspace was closed, if recorded.
func (a ClosedWorkspace) ClosedAt() time.Time {
	if a.Metadata.ClosedAt != nil {
		return *a.Metadata.ClosedAt
	}

	return time.Time{}
}

// RepoStatus represents the git status of a repo
type RepoStatus struct {
	Name            string
	IsDirty         bool
	UnpushedCommits int
	BehindRemote    int
	Branch          string
	Error           StatusError
}

// StatusError represents an error state for repository status checks.
type StatusError string

// Known status error values.
const (
	StatusErrorTimeout StatusError = "timeout"
)

// WorkspaceStatus represents the aggregate status of a workspace
type WorkspaceStatus struct {
	ID         string
	BranchName string
	Repos      []RepoStatus
}

// IsStale reports whether the workspace is older than the provided threshold.
func (w Workspace) IsStale(thresholdDays int) bool {
	if thresholdDays <= 0 || w.LastModified.IsZero() {
		return false
	}

	cutoff := time.Now().AddDate(0, 0, -thresholdDays)

	return w.LastModified.Before(cutoff)
}

// OrphanReason identifies why a worktree is orphaned.
type OrphanReason string

// Orphan reasons.
const (
	OrphanReasonCanonicalMissing OrphanReason = "canonical_missing"
	OrphanReasonDirectoryMissing OrphanReason = "directory_missing"
	OrphanReasonInvalidGitDir    OrphanReason = "invalid_git_dir"
)

// OrphanedWorktree represents a worktree that references missing or invalid resources.
type OrphanedWorktree struct {
	WorkspaceID  string       `json:"workspace_id"`
	RepoName     string       `json:"repo_name"`
	WorktreePath string       `json:"worktree_path"`
	Reason       OrphanReason `json:"reason"`
}

// ReasonDescription returns a human-readable description of the orphan reason.
func (o OrphanedWorktree) ReasonDescription() string {
	switch o.Reason {
	case OrphanReasonCanonicalMissing:
		return "canonical repo not found"
	case OrphanReasonDirectoryMissing:
		return "worktree directory missing"
	case OrphanReasonInvalidGitDir:
		return "invalid git directory"
	default:
		return string(o.Reason)
	}
}

// RepoCloseStatus describes the status of a repository when closing a workspace.
type RepoCloseStatus struct {
	Name          string `json:"name"`
	IsDirty       bool   `json:"is_dirty"`
	UnpushedCount int    `json:"unpushed_count"`
}

// WorkspaceClosePreview describes what would happen when closing a workspace.
type WorkspaceClosePreview struct {
	WorkspaceID    string            `json:"workspace_id"`
	WorkspacePath  string            `json:"workspace_path"`
	BranchName     string            `json:"branch_name"`
	ReposAffected  []string          `json:"repos_affected"`
	RepoStatuses   []RepoCloseStatus `json:"repo_statuses,omitempty"`
	DiskUsageBytes int64             `json:"disk_usage_bytes"`
	KeepMetadata   bool              `json:"keep_metadata"`
}

// RepoRemovePreview describes what would happen when removing a canonical repo.
type RepoRemovePreview struct {
	RepoName           string   `json:"repo_name"`
	RepoPath           string   `json:"repo_path"`
	DiskUsageBytes     int64    `json:"disk_usage_bytes"`
	WorkspacesAffected []string `json:"workspaces_affected"`
}

// CanonicalRepoStatus describes the health and usage of a canonical repository.
type CanonicalRepoStatus struct {
	Name           string     `json:"name"`
	Path           string     `json:"path"`
	DiskUsageBytes int64      `json:"disk_usage_bytes"`
	LastFetchTime  *time.Time `json:"last_fetch_time"`
	UsedByCount    int        `json:"used_by_count"`
	UsedBy         []string   `json:"used_by"`
}

// SyncStatus identifies the outcome of a repository sync operation.
type SyncStatus string

const (
	// SyncStatusUpdated means the repository was pulled and updated with new commits.
	SyncStatusUpdated SyncStatus = "updated"
	// SyncStatusUpToDate means the repository was already up-to-date.
	SyncStatusUpToDate SyncStatus = "up-to-date"
	// SyncStatusConflict means the pull operation failed due to merge conflicts.
	SyncStatusConflict SyncStatus = "conflict"
	// SyncStatusTimeout means the sync operation timed out for this repository.
	SyncStatusTimeout SyncStatus = "timeout"
	// SyncStatusError means an unexpected error occurred during sync.
	SyncStatusError SyncStatus = "error"
)

// RepoSyncStatus describes the sync result for a single repository.
type RepoSyncStatus struct {
	Name    string     `json:"name"`
	Status  SyncStatus `json:"status"`
	Updated int        `json:"updated"` // Number of commits pulled
	Error   string     `json:"error,omitempty"`
}

// SyncResult aggregates sync results for an entire workspace.
type SyncResult struct {
	WorkspaceID  string           `json:"workspace_id"`
	Repos        []RepoSyncStatus `json:"repos"`
	TotalUpdated int              `json:"total_updated"`
	TotalErrors  int              `json:"total_errors"`
}

// WorkspaceExport is the portable format for exporting/importing workspaces.
type WorkspaceExport struct {
	Version          string       `yaml:"version" json:"version"`
	WorkspaceVersion int          `yaml:"workspace_version" json:"workspace_version"`
	ID               string       `yaml:"id" json:"id"`
	Branch           string       `yaml:"branch" json:"branch"`
	Repos            []RepoExport `yaml:"repos" json:"repos"`
	ExportedAt       time.Time    `yaml:"exported_at" json:"exported_at"`
}

// RepoExport is the portable format for a repository in an export.
type RepoExport struct {
	Name  string `yaml:"name" json:"name"`
	URL   string `yaml:"url" json:"url"`
	Alias string `yaml:"alias,omitempty" json:"alias,omitempty"`
}

// HookContext provides context for hook execution.
type HookContext struct {
	WorkspaceID   string
	WorkspacePath string
	BranchName    string
	Repos         []Repo
}

// HookCommandPreview describes a resolved hook command in dry-run mode.
type HookCommandPreview struct {
	Index         int    `json:"index"`
	Command       string `json:"command"`
	WorkingDir    string `json:"working_dir"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspacePath string `json:"workspace_path"`
	BranchName    string `json:"branch_name"`
	RepoName      string `json:"repo_name,omitempty"`
	RepoPath      string `json:"repo_path,omitempty"`
}
