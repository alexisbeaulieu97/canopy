package gitx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitEngine wraps git operations
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
		return nil, fmt.Errorf("failed to clone %s: %w", repoURL, err)
	}

	return r, nil
}

// CreateWorktree creates a worktree for a ticket branch
func (g *GitEngine) CreateWorktree(repoName, worktreePath, branchName string) error {
	canonicalPath := filepath.Join(g.ProjectsRoot, repoName)

	// Use git CLI for robustness
	// 1. Clone
	cmd := exec.Command("git", "clone", canonicalPath, worktreePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %s: %w", string(output), err)
	}

	// 2. Checkout new branch
	cmd = exec.Command("git", "-C", worktreePath, "checkout", "-b", branchName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout -b failed: %s: %w", string(output), err)
	}

	return nil
}

// Status returns isDirty, unpushedCommits, branchName, error
func (g *GitEngine) Status(path string) (bool, int, string, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to open repo: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to get status: %w", err)
	}

	isDirty := !status.IsClean()

	// Get current branch
	head, err := r.Head()
	if err != nil {
		return isDirty, 0, "", fmt.Errorf("failed to get HEAD: %w", err)
	}
	branchName := head.Name().Short()

	// Check unpushed commits
	// This is a simplified check: count commits in HEAD that are not in origin/<branch>
	// Note: This assumes 'origin' is the remote and we have fetched recently.
	// For MVP, we won't auto-fetch to avoid network latency on 'status'.

	unpushed := 0

	// Try to resolve remote reference
	remoteName := "origin"
	remoteRefName := plumbing.NewRemoteReferenceName(remoteName, branchName)

	remoteRef, err := r.Reference(remoteRefName, true)
	if err == nil {
		// Calculate log between remote and HEAD
		commits, err := r.Log(&git.LogOptions{
			From: head.Hash(),
		})
		if err == nil {
			err = commits.ForEach(func(c *object.Commit) error {
				if c.Hash == remoteRef.Hash() {
					return fmt.Errorf("found") // Stop iteration
				}
				unpushed++
				return nil
			})
			if err != nil && err.Error() != "found" {
				// Real error, but maybe just ignore for status?
				// Let's just return what we have
			}
			// If we iterated everything and didn't find remote hash, it means we are ahead or diverged.
			// Or remote ref is not in history (e.g. rebase).
		}
	}

	return isDirty, unpushed, branchName, nil
}

// Clone clones a repository to the projects root (bare)
func (g *GitEngine) Clone(url, name string) error {
	path := filepath.Join(g.ProjectsRoot, name)

	// Check if exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("repository %s already exists", name)
	}

	// Use git CLI for robustness
	cmd := exec.Command("git", "clone", "--bare", url, path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %s: %w", string(output), err)
	}
	return nil
}

// Fetch fetches updates for a canonical repository
func (g *GitEngine) Fetch(name string) error {
	path := filepath.Join(g.ProjectsRoot, name)

	// Check if exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("repository %s does not exist", name)
	}

	// Use git CLI
	cmd := exec.Command("git", "-C", path, "fetch", "--all")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch failed: %s: %w", string(output), err)
	}
	return nil
}

// Pull pulls updates for a repository worktree
func (g *GitEngine) Pull(path string) error {
	// Use git CLI
	cmd := exec.Command("git", "-C", path, "pull")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %s: %w", string(output), err)
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
		return nil, fmt.Errorf("failed to read projects root: %w", err)
	}

	var repos []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Verify it's a git repo? For now, just assume directories are repos.
		// Or maybe check for HEAD/config if bare?
		// Let's keep it simple for MVP.
		repos = append(repos, entry.Name())
	}
	return repos, nil
}

// Checkout checks out a branch in the given path, optionally creating it
func (g *GitEngine) Checkout(path, branchName string, create bool) error {
	args := []string{"-C", path, "checkout"}
	if create {
		args = append(args, "-b")
	}
	args = append(args, branchName)

	cmd := exec.Command("git", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %s: %w", string(output), err)
	}
	return nil
}
