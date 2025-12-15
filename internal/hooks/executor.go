// Package hooks provides lifecycle hook execution for workspaces.
//
// Hooks are user-defined commands that execute at specific points in the workspace
// lifecycle. They run sequentially in the order defined in configuration, and each
// hook can be filtered to run only in specific repositories.
//
// Security: Hooks execute arbitrary commands from user-controlled configuration.
// The trust boundary is the user's config file - no sandboxing is applied.
package hooks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Compile-time check that Executor implements ports.HookExecutor.
var _ ports.HookExecutor = (*Executor)(nil)

// DefaultTimeout is the default hook execution timeout.
const DefaultTimeout = 30 * time.Second

// HookContext is an alias for domain.HookContext for backwards compatibility.
//
// Deprecated: Use domain.HookContext directly.
type HookContext = domain.HookContext

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
func (e *Executor) ExecuteHooks(hks []config.Hook, ctx domain.HookContext, continueOnError bool) error {
	for i, hook := range hks {
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
func (e *Executor) executeHook(hook config.Hook, ctx domain.HookContext, index int) error {
	// Determine repos to run against
	repos := ctx.Repos
	if len(hook.Repos) > 0 {
		repos = filterRepos(ctx.Repos, hook.Repos)
	}

	// If repos filter specified, run once per matching repo
	if len(hook.Repos) > 0 {
		for _, repo := range repos {
			repoPath := filepath.Join(ctx.WorkspacePath, repo.Name)
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
func (e *Executor) runCommand(hook config.Hook, ctx domain.HookContext, workDir string, repo *domain.Repo, index int) error {
	shell := resolveShell(hook.Shell)
	timeout := resolveTimeout(hook.Timeout)

	execCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := e.buildCommand(execCtx, shell, hook.Command, workDir, ctx, repo)

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
		return e.handleCommandError(execCtx, err, hook, index, repo, timeout, stderr.String())
	}

	e.logCommandSuccess(index, duration, stdout.String(), stderr.String())

	return nil
}

// resolveShell determines the shell to use for executing the hook.
func resolveShell(hookShell string) string {
	if hookShell != "" {
		return hookShell
	}

	if envShell := os.Getenv("SHELL"); envShell != "" {
		return envShell
	}

	return "/bin/sh"
}

// resolveTimeout determines the timeout duration for the hook.
func resolveTimeout(hookTimeout int) time.Duration {
	if hookTimeout > 0 {
		return time.Duration(hookTimeout) * time.Second
	}

	return DefaultTimeout
}

// buildCommand creates the exec.Cmd with proper environment variables.
func (e *Executor) buildCommand(
	ctx context.Context,
	shell, command, workDir string,
	hookCtx HookContext,
	repo *domain.Repo,
) *exec.Cmd {
	// The hook command comes from user-controlled config, which is the trust boundary.
	// See design.md threat model for security considerations.
	cmd := exec.CommandContext(ctx, shell, "-c", command) //nolint:gosec // user-controlled config is trusted
	cmd.Dir = workDir
	cmd.Env = e.buildEnvVars(hookCtx, repo)

	return cmd
}

// buildEnvVars creates the environment variables for the hook command.
func (e *Executor) buildEnvVars(ctx HookContext, repo *domain.Repo) []string {
	env := append(os.Environ(),
		fmt.Sprintf("CANOPY_WORKSPACE_ID=%s", ctx.WorkspaceID),
		fmt.Sprintf("CANOPY_WORKSPACE_PATH=%s", ctx.WorkspacePath),
		fmt.Sprintf("CANOPY_BRANCH=%s", ctx.BranchName),
	)

	if repo != nil {
		env = append(env,
			fmt.Sprintf("CANOPY_REPO_NAME=%s", repo.Name),
			fmt.Sprintf("CANOPY_REPO_PATH=%s", filepath.Join(ctx.WorkspacePath, repo.Name)),
		)
	}

	return env
}

// handleCommandError processes errors from hook command execution.
func (e *Executor) handleCommandError(
	execCtx context.Context,
	err error,
	hook config.Hook,
	index int,
	repo *domain.Repo,
	timeout time.Duration,
	stderrOutput string,
) error {
	if execCtx.Err() == context.DeadlineExceeded {
		return cerrors.NewHookTimeout(index, hook.Command, timeout)
	}

	exitCode := -1

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}

	repoName := ""
	if repo != nil {
		repoName = repo.Name
	}

	return cerrors.NewHookFailed(index, hook.Command, exitCode, repoName, stderrOutput)
}

// logCommandSuccess logs successful hook completion and any output.
func (e *Executor) logCommandSuccess(index int, duration time.Duration, stdoutOutput, stderrOutput string) {
	e.logger.Info("Hook completed", "index", index, "exit_code", 0, "duration", duration.Round(time.Millisecond))

	if stdoutOutput != "" {
		e.logger.Debug("Hook stdout output", "index", index, "stdout", stdoutOutput)
	}

	if stderrOutput != "" {
		e.logger.Warn("Hook stderr output", "index", index, "stderr", stderrOutput)
	}
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
