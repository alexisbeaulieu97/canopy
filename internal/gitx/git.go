// Package gitx wraps git operations used by canopy.
//
// # go-git Implementation Notes
//
// This package uses go-git (github.com/go-git/go-git/v5) for pure Go git operations,
// eliminating the need for exec.Command("git", ...) calls in most cases.
//
// ## Known Limitations
//
//   - Authentication: go-git relies on SSH agents for authentication. Users must have
//     their SSH keys properly configured. HTTPS authentication with credentials is
//     not directly supported without additional configuration.
//
//   - Worktree creation: go-git's Worktree.Add() does not support detached HEAD or
//     creating a worktree for a non-existent branch. We use git CLI via [GitEngine.RunCommand]
//     as a fallback for this operation.
//
//   - Sparse checkout: Not natively supported by go-git. Would require CLI fallback.
//
//   - Interactive operations: Rebase, merge conflict resolution, and other interactive
//     git operations are not available in go-git.
//
// ## Escape Hatch
//
// The [GitEngine.RunCommand] method provides an escape hatch for operations that cannot
// be performed with go-git. It executes git commands directly via exec.Command and
// should be used sparingly.
package gitx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// DefaultNetworkTimeout is the default timeout for network operations (clone, fetch, push, pull).
const DefaultNetworkTimeout = 5 * time.Minute

// errStopIteration is a sentinel error used to break out of commit iteration loops.
var errStopIteration = errors.New("stop iteration")

// Compile-time check that GitEngine implements ports.GitOperations.
var _ ports.GitOperations = (*GitEngine)(nil)

// GitEngine wraps git operations using go-git for pure Go implementations.
// See package documentation for known limitations.
type GitEngine struct {
	ProjectsRoot string
	RetryConfig  RetryConfig
}

// New creates a new GitEngine with default retry configuration.
func New(projectsRoot string) *GitEngine {
	return &GitEngine{
		ProjectsRoot: projectsRoot,
		RetryConfig:  DefaultRetryConfig(),
	}
}

// NewWithRetry creates a new GitEngine with custom retry configuration.
func NewWithRetry(projectsRoot string, retryCfg RetryConfig) *GitEngine {
	return &GitEngine{
		ProjectsRoot: projectsRoot,
		RetryConfig:  retryCfg,
	}
}

// EnsureCanonical ensures the repo is cloned in ProjectsRoot (bare)
func (g *GitEngine) EnsureCanonical(ctx context.Context, repoURL, repoName string) (*git.Repository, error) {
	path := filepath.Join(g.ProjectsRoot, repoName)

	// Check if exists
	r, err := git.PlainOpen(path)
	if err == nil {
		return r, nil
	}

	// Apply default timeout if context has no deadline
	ctx, cancel := g.withDefaultTimeout(ctx)
	defer cancel()

	// Clone if not exists, with retry for transient failures
	r, err = WithRetry(ctx, g.RetryConfig, func() (*git.Repository, error) {
		repo, cloneErr := git.PlainCloneContext(ctx, path, true, &git.CloneOptions{
			URL: repoURL,
		})
		if cloneErr != nil {
			// Clean up partial clone on error before retry
			if cleanupErr := os.RemoveAll(path); cleanupErr != nil {
				log.Warn("failed to cleanup partial clone", "path", path, "error", cleanupErr)
			}
		}

		return repo, cloneErr
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, cerrors.NewOperationCanceledWithTarget("clone", repoURL)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, cerrors.NewOperationTimeout("clone", repoURL)
		}

		return nil, cerrors.WrapGitError(err, fmt.Sprintf("clone %s", repoURL))
	}

	return r, nil
}

// CreateWorktree creates a worktree for a workspace branch
func (g *GitEngine) CreateWorktree(repoName, worktreePath, branchName string) error {
	canonicalPath := filepath.Join(g.ProjectsRoot, repoName)

	// Open the canonical (bare) repository
	canonicalRepo, err := git.PlainOpen(canonicalPath)
	if err != nil {
		return cerrors.WrapGitError(err, "open canonical repo")
	}

	// Clone from the canonical repo to the worktree path (non-bare)
	repo, err := git.PlainClone(worktreePath, false, &git.CloneOptions{
		URL: canonicalPath,
	})
	if err != nil {
		return cerrors.WrapGitError(err, "clone")
	}

	// Get the worktree
	wt, err := repo.Worktree()
	if err != nil {
		return cerrors.WrapGitError(err, "get worktree")
	}

	// Get the HEAD reference to use as the starting point for the new branch
	head, err := canonicalRepo.Head()
	if err != nil {
		return cerrors.WrapGitError(err, "get HEAD")
	}

	// Create and checkout a new branch
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Hash:   head.Hash(),
		Create: true,
	})
	if err != nil {
		return cerrors.WrapGitError(err, "checkout -b")
	}

	return nil
}

// Status returns isDirty, unpushedCommits, behindRemote, branchName, error
func (g *GitEngine) Status(path string) (bool, int, int, string, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return false, 0, 0, "", cerrors.WrapGitError(err, "open repo")
	}

	w, err := r.Worktree()
	if err != nil {
		return false, 0, 0, "", cerrors.WrapGitError(err, "get worktree")
	}

	status, err := w.Status()
	if err != nil {
		return false, 0, 0, "", cerrors.WrapGitError(err, "get status")
	}

	isDirty := !status.IsClean()

	// Get current branch
	head, err := r.Head()
	if err != nil {
		return isDirty, 0, 0, "", cerrors.WrapGitError(err, "get HEAD")
	}

	branchName := head.Name().Short()

	// Check unpushed commits using pure go-git
	unpushed := 0
	behindRemote := 0

	remoteName := "origin"
	remoteRefName := plumbing.NewRemoteReferenceName(remoteName, branchName)

	if ref, refErr := r.Reference(remoteRefName, true); refErr == nil {
		// Calculate ahead/behind using go-git rev walking
		ahead, behind, countErr := g.countAheadBehind(r, head.Hash(), ref.Hash())
		if countErr == nil {
			unpushed = ahead
			behindRemote = behind
		}
	}

	return isDirty, unpushed, behindRemote, branchName, nil
}

// Clone clones a repository to the projects root (bare)
func (g *GitEngine) Clone(ctx context.Context, url, name string) error {
	path := filepath.Join(g.ProjectsRoot, name)

	// Check if exists
	_, err := os.Stat(path)
	if err == nil {
		// Path exists
		return cerrors.NewRepoAlreadyExists(name, "projects root")
	}

	if !os.IsNotExist(err) {
		// Some other error (permission, I/O, etc.)
		return cerrors.NewIOFailed(fmt.Sprintf("check path %s", path), err)
	}

	// Path does not exist - proceed with clone

	// Apply default timeout if context has no deadline
	ctx, cancel := g.withDefaultTimeout(ctx)
	defer cancel()

	// Clone as bare using go-git, with retry for transient failures
	err = WithRetryNoResult(ctx, g.RetryConfig, func() error {
		_, cloneErr := git.PlainCloneContext(ctx, path, true, &git.CloneOptions{
			URL: url,
		})
		if cloneErr != nil {
			// Clean up partial clone on error before retry
			if cleanupErr := os.RemoveAll(path); cleanupErr != nil {
				log.Warn("failed to cleanup partial clone", "path", path, "error", cleanupErr)
			}
		}

		return cloneErr
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return cerrors.NewOperationCanceledWithTarget("clone", url)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return cerrors.NewOperationTimeout("clone", url)
		}

		return cerrors.WrapGitError(err, "clone")
	}

	return nil
}

// Fetch fetches updates for a canonical repository
func (g *GitEngine) Fetch(ctx context.Context, name string) error {
	path := filepath.Join(g.ProjectsRoot, name)

	// Check if exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return cerrors.NewRepoNotFound(name)
	}

	if err != nil {
		// Some other error (permission, I/O, etc.)
		return cerrors.NewIOFailed(fmt.Sprintf("check path %s", path), err)
	}

	// Open the repository
	r, err := git.PlainOpen(path)
	if err != nil {
		return cerrors.WrapGitError(err, "open repo")
	}

	// Apply default timeout if context has no deadline
	ctx, cancel := g.withDefaultTimeout(ctx)
	defer cancel()

	// Fetch from all remotes
	remotes, err := r.Remotes()
	if err != nil {
		return cerrors.WrapGitError(err, "list remotes")
	}

	for _, remote := range remotes {
		// Fetch into refs/remotes/<remote>/* to properly track remote branches
		// For bare repos used as canonical storage, we fetch directly into refs/heads/*
		remoteName := remote.Config().Name
		refSpec := config.RefSpec(fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", remoteName))

		// Wrap fetch with retry for transient failures
		fetchErr := WithRetryNoResult(ctx, g.RetryConfig, func() error {
			return remote.FetchContext(ctx, &git.FetchOptions{
				RefSpecs: []config.RefSpec{refSpec},
			})
		})
		if fetchErr != nil && !errors.Is(fetchErr, git.NoErrAlreadyUpToDate) {
			if errors.Is(fetchErr, context.Canceled) {
				return cerrors.NewOperationCanceledWithTarget("fetch", name)
			}

			if errors.Is(fetchErr, context.DeadlineExceeded) {
				return cerrors.NewOperationTimeout("fetch", name)
			}

			return cerrors.WrapGitError(fetchErr, "fetch")
		}
	}

	return nil
}

// Pull pulls updates for a repository worktree
func (g *GitEngine) Pull(ctx context.Context, path string) error {
	// Open the repository
	r, err := git.PlainOpen(path)
	if err != nil {
		return cerrors.WrapGitError(err, "open repo")
	}

	// Get the worktree
	w, err := r.Worktree()
	if err != nil {
		return cerrors.WrapGitError(err, "get worktree")
	}

	// Apply default timeout if context has no deadline
	ctx, cancel := g.withDefaultTimeout(ctx)
	defer cancel()

	// Pull changes, with retry for transient failures
	err = WithRetryNoResult(ctx, g.RetryConfig, func() error {
		return w.PullContext(ctx, &git.PullOptions{
			RemoteName: "origin",
		})
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		if errors.Is(err, context.Canceled) {
			return cerrors.NewOperationCanceledWithTarget("pull", path)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return cerrors.NewOperationTimeout("pull", path)
		}

		return cerrors.WrapGitError(err, "pull")
	}

	return nil
}

// Push pushes the current branch to its upstream.
func (g *GitEngine) Push(ctx context.Context, path, branch string) error {
	// Open the repository
	r, err := git.PlainOpen(path)
	if err != nil {
		return cerrors.WrapGitError(err, "open repo")
	}

	// Build push options
	pushOpts := &git.PushOptions{
		RemoteName: "origin",
	}

	// If branch is specified, set up the refspec for pushing and tracking
	if branch != "" {
		// Push the branch and set upstream tracking
		refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))
		pushOpts.RefSpecs = []config.RefSpec{refSpec}

		// Set the branch to track the remote
		cfg, err := r.Config()
		if err == nil {
			// Initialize Branches map if nil (can happen on freshly cloned repos)
			if cfg.Branches == nil {
				cfg.Branches = make(map[string]*config.Branch)
			}

			cfg.Branches[branch] = &config.Branch{
				Name:   branch,
				Remote: "origin",
				Merge:  plumbing.NewBranchReferenceName(branch),
			}

			if setErr := r.SetConfig(cfg); setErr != nil {
				// Log but don't fail - tracking is nice-to-have, push is the priority
				log.Warn("failed to set branch tracking config", "branch", branch, "error", setErr)
			}
		}
	}

	// Apply default timeout if context has no deadline
	ctx, cancel := g.withDefaultTimeout(ctx)
	defer cancel()

	// Push changes, with retry for transient failures
	err = WithRetryNoResult(ctx, g.RetryConfig, func() error {
		return r.PushContext(ctx, pushOpts)
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		if errors.Is(err, context.Canceled) {
			return cerrors.NewOperationCanceledWithTarget("push", path)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return cerrors.NewOperationTimeout("push", path)
		}

		return cerrors.WrapGitError(err, "push")
	}

	return nil
}

// List returns a list of repository names in the projects root
func (g *GitEngine) List() ([]string, error) {
	entries, err := os.ReadDir(g.ProjectsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, cerrors.NewIOFailed("read projects root", err)
	}

	var repos []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Verify it's a bare git repo by checking for HEAD file.
		// The projects root contains only canonical bare repos (not regular .git repos),
		// so we only check for HEAD at the root level, not .git/HEAD.
		headPath := filepath.Join(g.ProjectsRoot, entry.Name(), "HEAD")

		_, err := os.Stat(headPath)
		if err != nil {
			if os.IsNotExist(err) {
				// Not a bare git repo, skip silently
				continue
			}
			// Other stat errors (permission denied, etc.) - skip but could log in debug mode
			continue
		}

		repos = append(repos, entry.Name())
	}

	return repos, nil
}

// Checkout checks out a branch in the given path, optionally creating it
func (g *GitEngine) Checkout(path, branchName string, create bool) error {
	// Open the repository
	r, err := git.PlainOpen(path)
	if err != nil {
		return cerrors.WrapGitError(err, "open repo")
	}

	// Get the worktree
	w, err := r.Worktree()
	if err != nil {
		return cerrors.WrapGitError(err, "get worktree")
	}

	// Build checkout options
	checkoutOpts := &git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: create,
	}

	// If creating a new branch, use HEAD as the starting point
	if create {
		head, err := r.Head()
		if err != nil {
			return cerrors.WrapGitError(err, "get HEAD")
		}

		checkoutOpts.Hash = head.Hash()
	}

	// Checkout the branch
	err = w.Checkout(checkoutOpts)
	if err != nil {
		return cerrors.WrapGitError(err, "checkout")
	}

	return nil
}

// RenameBranch renames a branch in the given repository.
// This uses git CLI via RunCommand as go-git does not support branch renaming directly.
func (g *GitEngine) RenameBranch(ctx context.Context, repoPath, oldName, newName string) error {
	result, err := g.RunCommand(ctx, repoPath, "branch", "-m", oldName, newName)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return cerrors.NewCommandFailed(fmt.Sprintf("git branch -m %s %s", oldName, newName),
			fmt.Errorf("exit code %d: %s", result.ExitCode, result.Stderr))
	}

	return nil
}

// countAheadBehind calculates how many commits local is ahead and behind remote.
// This is a pure go-git implementation of git rev-list --left-right --count.
func (g *GitEngine) countAheadBehind(r *git.Repository, localHash, remoteHash plumbing.Hash) (int, int, error) {
	// If both hashes are the same, we're not ahead or behind
	if localHash == remoteHash {
		return 0, 0, nil
	}

	// Find the merge base (common ancestor)
	localCommit, err := r.CommitObject(localHash)
	if err != nil {
		return 0, 0, cerrors.WrapGitError(err, "get local commit")
	}

	remoteCommit, err := r.CommitObject(remoteHash)
	if err != nil {
		return 0, 0, cerrors.WrapGitError(err, "get remote commit")
	}

	// Find merge bases
	bases, err := localCommit.MergeBase(remoteCommit)
	if err != nil {
		return 0, 0, cerrors.WrapGitError(err, "find merge base")
	}

	if len(bases) == 0 {
		// No common ancestor - count all commits
		ahead, err := g.countCommitsTo(r, localHash, plumbing.ZeroHash)
		if err != nil {
			return 0, 0, err
		}

		behind, err := g.countCommitsTo(r, remoteHash, plumbing.ZeroHash)
		if err != nil {
			return 0, 0, err
		}

		return ahead, behind, nil
	}

	// Count commits from merge base to local (ahead)
	ahead, err := g.countCommitsTo(r, localHash, bases[0].Hash)
	if err != nil {
		return 0, 0, err
	}

	// Count commits from merge base to remote (behind)
	behind, err := g.countCommitsTo(r, remoteHash, bases[0].Hash)
	if err != nil {
		return 0, 0, err
	}

	return ahead, behind, nil
}

// countCommitsTo counts the number of commits from 'from' to 'to' (exclusive).
// If 'to' is ZeroHash, counts all commits reachable from 'from'.
func (g *GitEngine) countCommitsTo(r *git.Repository, from, to plumbing.Hash) (int, error) {
	commits, err := r.Log(&git.LogOptions{
		From: from,
	})
	if err != nil {
		return 0, cerrors.WrapGitError(err, "get log")
	}

	count := 0
	err = commits.ForEach(func(c *object.Commit) error {
		if c.Hash == to {
			return errStopIteration
		}

		count++

		return nil
	})

	// Ignore the sentinel error - it's our way of breaking iteration
	if err != nil && !errors.Is(err, errStopIteration) {
		return 0, cerrors.WrapGitError(err, "iterate commits")
	}

	return count, nil
}

// RunCommand executes an arbitrary git command in the specified repository path.
// This is an escape hatch for operations that cannot be performed with go-git,
// such as worktree creation with specific options. Use sparingly.
//
// Security note: The git binary path is hardcoded and arguments are passed
// as separate parameters to prevent shell injection.
func (g *GitEngine) RunCommand(ctx context.Context, repoPath string, args ...string) (*ports.CommandResult, error) {
	if len(args) == 0 {
		return nil, cerrors.NewInvalidArgument("args", "git command requires at least one argument")
	}

	cmdArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.CommandContext(ctx, "git", cmdArgs...) //nolint:gosec // git binary is hardcoded, args passed safely as separate parameters

	var stdout, stderr strings.Builder

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &ports.CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, cerrors.NewOperationCanceledWithTarget("git command", strings.Join(args, " "))
		}

		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, cerrors.NewOperationTimeout("git command", strings.Join(args, " "))
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, cerrors.NewCommandFailed("git", err)
		}
	}

	return result, nil
}

// withDefaultTimeout returns a context with the default network timeout applied
// if the provided context has no deadline set. The returned cancel function
// must be called to release resources (use defer cancel()).
func (g *GitEngine) withDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		// Context already has a deadline, return no-op cancel
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, DefaultNetworkTimeout)
}
