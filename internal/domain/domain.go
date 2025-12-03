// Package domain contains core domain models.
package domain

import "time"

// Repo represents a git repository
type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// Workspace represents a work item
type Workspace struct {
	ID             string     `yaml:"id"`
	BranchName     string     `yaml:"branch_name,omitempty"`
	Repos          []Repo     `yaml:"repos"`
	ClosedAt       *time.Time `yaml:"closed_at,omitempty"`
	LastModified   time.Time  `yaml:"-"`
	DiskUsageBytes int64      `yaml:"-"`
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
}

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
