// Package gitx wraps git operations used by canopy.
package gitx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Compile-time check that GitEngine implements ports.GitOperations.
var _ ports.GitOperations = (*GitEngine)(nil)

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

// CreateWorktree creates a worktree for a workspace branch
func (g *GitEngine) CreateWorktree(repoName, worktreePath, branchName string) error {
	canonicalPath := filepath.Join(g.ProjectsRoot, repoName)

	// Use git CLI for robustness
	// 1. Clone
	cmd := exec.Command("git", "clone", canonicalPath, worktreePath) //nolint:gosec // arguments are constructed internally
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %s: %w", string(output), err)
	}

	// 2. Checkout new branch
	cmd = exec.Command("git", "-C", worktreePath, "checkout", "-b", branchName) //nolint:gosec // arguments are constructed internally
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout -b failed: %s: %w", string(output), err)
	}

	return nil
}

// Status returns isDirty, unpushedCommits, behindRemote, branchName, error
func (g *GitEngine) Status(path string) (bool, int, int, string, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return false, 0, 0, "", fmt.Errorf("failed to open repo: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return false, 0, 0, "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return false, 0, 0, "", fmt.Errorf("failed to get status: %w", err)
	}

	isDirty := !status.IsClean()

	// Get current branch
	head, err := r.Head()
	if err != nil {
		return isDirty, 0, 0, "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	branchName := head.Name().Short()

	// Check unpushed commits
	unpushed := 0
	behindRemote := 0

	remoteName := "origin"
	remoteRefName := plumbing.NewRemoteReferenceName(remoteName, branchName)

	if ref, refErr := r.Reference(remoteRefName, true); refErr == nil {
		// Prefer git CLI for accurate ahead/behind counts; fallback to go-git walk if it fails.
		if ahead, behind, cliErr := g.aheadBehindCounts(path, branchName); cliErr == nil {
			unpushed = ahead
			behindRemote = behind
		} else {
			// Calculate log between remote and HEAD
			commits, err := r.Log(&git.LogOptions{
				From: head.Hash(),
			})
			if err == nil {
				_ = commits.ForEach(func(c *object.Commit) error {
					if c.Hash == ref.Hash() {
						return fmt.Errorf("found") // Stop iteration
					}

					unpushed++

					return nil
				})
			}
		}
	}

	return isDirty, unpushed, behindRemote, branchName, nil
}

// Clone clones a repository to the projects root (bare)
func (g *GitEngine) Clone(url, name string) error {
	path := filepath.Join(g.ProjectsRoot, name)

	// Check if exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("repository %s already exists", name)
	}

	// Use git CLI for robustness
	cmd := exec.Command("git", "clone", "--bare", url, path) //nolint:gosec // arguments are constructed internally
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
	cmd := exec.Command("git", "-C", path, "fetch", "--all") //nolint:gosec // arguments are constructed internally
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch failed: %s: %w", string(output), err)
	}

	return nil
}

// Pull pulls updates for a repository worktree
func (g *GitEngine) Pull(path string) error {
	// Use git CLI
	cmd := exec.Command("git", "-C", path, "pull") //nolint:gosec // arguments are constructed internally
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %s: %w", string(output), err)
	}

	return nil
}

// Push pushes the current branch to its upstream.
func (g *GitEngine) Push(path, branch string) error {
	args := []string{"-C", path, "push"}
	if branch != "" {
		args = append(args, "--set-upstream", "origin", branch)
	}

	cmd := exec.Command("git", args...) //nolint:gosec // arguments are constructed internally
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %s: %w", strings.TrimSpace(string(output)), err)
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

	cmd := exec.Command("git", args...) //nolint:gosec // arguments are constructed internally
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %s: %w", string(output), err)
	}

	return nil
}

func (g *GitEngine) aheadBehindCounts(path, branch string) (int, int, error) {
	if branch == "" {
		return 0, 0, fmt.Errorf("branch name is required")
	}

	cmd := exec.Command( //nolint:gosec // arguments are constructed internally
		"git",
		"-C", path,
		"rev-list",
		"--left-right",
		"--count",
		fmt.Sprintf("HEAD...origin/%s", branch),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("git rev-list failed: %s: %w", strings.TrimSpace(string(output)), err)
	}

	fields := strings.Fields(strings.TrimSpace(string(output)))
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output: %s", strings.TrimSpace(string(output)))
	}

	ahead, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse ahead count: %w", err)
	}

	behind, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse behind count: %w", err)
	}

	return ahead, behind, nil
}
