// Package domain contains core domain models.
package domain

// Repo represents a git repository
type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// Workspace represents a work item
type Workspace struct {
	ID         string `yaml:"id"`
	Slug       string `yaml:"slug,omitempty"`
	BranchName string `yaml:"branch_name,omitempty"`
	Repos      []Repo `yaml:"repos"`
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
