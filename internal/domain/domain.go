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
	ID         string     `yaml:"id"`
	BranchName string     `yaml:"branch_name,omitempty"`
	Repos      []Repo     `yaml:"repos"`
	ArchivedAt *time.Time `yaml:"archived_at,omitempty"`
}

// RepoStatus represents the git status of a repo
type RepoStatus struct {
	Name            string
	IsDirty         bool
	UnpushedCommits int
	Branch          string
}

// WorkspaceStatus represents the aggregate status of a workspace
type WorkspaceStatus struct {
	ID         string
	BranchName string
	Repos      []RepoStatus
}
