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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// errStopIteration is a sentinel error used to break out of commit iteration loops.
var errStopIteration = errors.New("stop iteration")

// Compile-time check that GitEngine implements ports.GitOperations.
var _ ports.GitOperations = (*GitEngine)(nil)

// GitEngine wraps git operations using go-git for pure Go implementations.
// See package documentation for known limitations.
type GitEngine struct {
	ProjectsRoot string
}

// New creates a new GitEngine
func New(projectsRoot string) *GitEngine {
	return &GitEngine{ProjectsRoot: projectsRoot}
}

// EnsureCanonical ensures the repo is cloned in ProjectsRoot (bare)
func (g *GitEngine) EnsureCanonical(repoURL, repoName string) (*git.Repository, error) {
	path := filepath.Join(g.ProjectsRoot, repoName)

	// Check if exists
	r, err := git.PlainOpen(path)
	if err == nil {
		return r, nil
	}

	// Clone if not exists
	r, err = git.PlainClone(path, true, &git.CloneOptions{
		URL: repoURL,
	})
	if err != nil {
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
func (g *GitEngine) Clone(url, name string) error {
	path := filepath.Join(g.ProjectsRoot, name)

	// Check if exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return cerrors.NewRepoAlreadyExists(name, "projects root")
	}

	// Clone as bare using go-git
	_, err := git.PlainClone(path, true, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return cerrors.WrapGitError(err, "clone")
	}

	return nil
}

// Fetch fetches updates for a canonical repository
func (g *GitEngine) Fetch(name string) error {
	path := filepath.Join(g.ProjectsRoot, name)

	// Check if exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cerrors.NewRepoNotFound(name)
	}

	// Open the repository
	r, err := git.PlainOpen(path)
	if err != nil {
		return cerrors.WrapGitError(err, "open repo")
	}

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

		err := remote.Fetch(&git.FetchOptions{
			RefSpecs: []config.RefSpec{refSpec},
		})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return cerrors.WrapGitError(err, "fetch")
		}
	}

	return nil
}

// Pull pulls updates for a repository worktree
func (g *GitEngine) Pull(path string) error {
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

	// Pull changes
	err = w.Pull(&git.PullOptions{
		RemoteName: "origin",
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return cerrors.WrapGitError(err, "pull")
	}

	return nil
}

// Push pushes the current branch to its upstream.
func (g *GitEngine) Push(path, branch string) error {
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

	// Push changes
	err = r.Push(pushOpts)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
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

		// Verify it's a bare git repo by checking for HEAD file
		headPath := filepath.Join(g.ProjectsRoot, entry.Name(), "HEAD")
		if _, err := os.Stat(headPath); os.IsNotExist(err) {
			// Not a git repo, skip
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
func (g *GitEngine) RunCommand(repoPath string, args ...string) (*ports.CommandResult, error) {
	if len(args) == 0 {
		return nil, cerrors.NewInvalidArgument("args", "git command requires at least one argument")
	}

	cmdArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.Command("git", cmdArgs...) //nolint:gosec // git binary is hardcoded, args passed safely as separate parameters

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
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, cerrors.NewCommandFailed("git", err)
		}
	}

	return result, nil
}
