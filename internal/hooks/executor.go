// Package hooks provides lifecycle hook execution for workspaces.
package hooks

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
)

// DefaultTimeout is the default hook execution timeout.
const DefaultTimeout = 30 * time.Second

// HookContext provides context for hook execution.
type HookContext struct {
	WorkspaceID   string
	WorkspacePath string
	BranchName    string
	Repos         []domain.Repo
}

// Executor executes lifecycle hooks.
type Executor struct {
	logger *logging.Logger
}

// NewExecutor creates a new hook executor.
func NewExecutor(logger *logging.Logger) *Executor {
	return &Executor{logger: logger}
}

// ExecuteHooks runs a list of hooks with the given context.
// If continueOnError is true at the executor level, it continues even if a hook fails.
func (e *Executor) ExecuteHooks(hooks []config.Hook, ctx HookContext, continueOnError bool) error {
	for i, hook := range hooks {
		err := e.executeHook(hook, ctx, i)
		if err != nil {
			if hook.ContinueOnError || continueOnError {
				e.logger.Warn("Hook failed but continuing", "index", i, "command", hook.Command, "error", err)
				continue
			}

			return err
		}
	}

	return nil
}

// executeHook runs a single hook command.
func (e *Executor) executeHook(hook config.Hook, ctx HookContext, index int) error {
	// Determine repos to run against
	repos := ctx.Repos
	if len(hook.Repos) > 0 {
		repos = filterRepos(ctx.Repos, hook.Repos)
	}

	// If repos filter specified, run once per matching repo
	if len(hook.Repos) > 0 {
		for _, repo := range repos {
			repoPath := fmt.Sprintf("%s/%s", ctx.WorkspacePath, repo.Name)
			if err := e.runCommand(hook, ctx, repoPath, &repo, index); err != nil {
				return err
			}
		}

		return nil
	}

	// No repos filter - run once in workspace root
	return e.runCommand(hook, ctx, ctx.WorkspacePath, nil, index)
}

// runCommand executes the hook command in the specified directory.
func (e *Executor) runCommand(hook config.Hook, ctx HookContext, workDir string, repo *domain.Repo, index int) error {
	// Determine shell
	shell := hook.Shell
	if shell == "" {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
	}

	// Determine timeout
	timeout := DefaultTimeout
	if hook.Timeout > 0 {
		timeout = time.Duration(hook.Timeout) * time.Second
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Build command
	// The hook command comes from user-controlled config, which is the trust boundary.
	// See design.md threat model for security considerations.
	cmd := exec.CommandContext(execCtx, shell, "-c", hook.Command) //nolint:gosec // user-controlled config is trusted
	cmd.Dir = workDir

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("CANOPY_WORKSPACE_ID=%s", ctx.WorkspaceID),
		fmt.Sprintf("CANOPY_WORKSPACE_PATH=%s", ctx.WorkspacePath),
		fmt.Sprintf("CANOPY_BRANCH=%s", ctx.BranchName),
	)

	if repo != nil {
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("CANOPY_REPO_NAME=%s", repo.Name),
			fmt.Sprintf("CANOPY_REPO_PATH=%s/%s", ctx.WorkspacePath, repo.Name),
		)
	}

	// Capture output
	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	e.logger.Debug("Executing hook",
		"index", index,
		"command", hook.Command,
		"working_dir", workDir,
		"timeout", timeout,
	)

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		// Check if it was a timeout
		if execCtx.Err() == context.DeadlineExceeded {
			return cerrors.NewHookTimeout(index, hook.Command, timeout)
		}

		// Get exit code if available
		exitCode := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}

		repoName := ""
		if repo != nil {
			repoName = repo.Name
		}

		return cerrors.NewHookFailed(index, hook.Command, exitCode, repoName, stderr.String())
	}

	e.logger.Info("Hook completed", "index", index, "exit_code", 0, "duration", duration.Round(time.Millisecond))

	// Log stderr as warning if present
	if stderr.Len() > 0 {
		e.logger.Warn("Hook stderr output", "index", index, "stderr", stderr.String())
	}

	return nil
}

// filterRepos returns only repos whose names match the filter list.
func filterRepos(repos []domain.Repo, filter []string) []domain.Repo {
	filterSet := make(map[string]bool)
	for _, name := range filter {
		filterSet[name] = true
	}

	var result []domain.Repo

	for _, repo := range repos {
		if filterSet[repo.Name] {
			result = append(result, repo)
		}
	}

	return result
}
