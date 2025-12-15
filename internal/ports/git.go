// Package ports defines interfaces for external dependencies (hexagonal architecture).
package ports

import (
	"context"

	"github.com/go-git/go-git/v5"
)

// CommandResult holds the output and exit code from a git command execution.
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// GitOperations defines the interface for git operations.
type GitOperations interface {
	// EnsureCanonical ensures the repo is cloned in ProjectsRoot (bare).
	EnsureCanonical(ctx context.Context, repoURL, repoName string) (*git.Repository, error)

	// CreateWorktree creates a worktree for a workspace branch.
	CreateWorktree(ctx context.Context, repoName, worktreePath, branchName string) error

	// Status returns isDirty, unpushedCommits, behindRemote, branchName, error.
	Status(ctx context.Context, path string) (isDirty bool, unpushed, behind int, branch string, err error)

	// Clone clones a repository to the projects root (bare).
	Clone(ctx context.Context, url, name string) error

	// Fetch fetches updates for a canonical repository.
	Fetch(ctx context.Context, name string) error

	// Pull pulls updates for a repository worktree.
	Pull(ctx context.Context, path string) error

	// Push pushes the current branch to its upstream.
	Push(ctx context.Context, path, branch string) error

	// List returns a list of repository names in the projects root.
	List(ctx context.Context) ([]string, error)

	// Checkout checks out a branch in the given path, optionally creating it.
	Checkout(ctx context.Context, path, branchName string, create bool) error

	// RenameBranch renames a branch in the given repository.
	RenameBranch(ctx context.Context, repoPath, oldName, newName string) error

	// RunCommand executes an arbitrary git command in the specified repository path.
	RunCommand(ctx context.Context, repoPath string, args ...string) (*CommandResult, error)

	// GetUpstreamURL retrieves the upstream URL from a canonical repository's config.
	GetUpstreamURL(repoName string) (string, error)

	// RemoveWorktree removes a git worktree from the canonical repository.
	RemoveWorktree(ctx context.Context, repoName, worktreePath string) error

	// PruneWorktrees cleans up stale worktree references from a canonical repository.
	PruneWorktrees(ctx context.Context, repoName string) error
}
