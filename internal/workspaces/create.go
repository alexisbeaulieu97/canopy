package workspaces

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
)

// CreateOptions configures workspace creation behavior.
type CreateOptions struct {
	SkipHooks         bool // Skip post_create hooks
	ContinueOnHookErr bool // Continue if hooks fail
	Template          *config.Template
}

// CreateWorkspace creates a new workspace directory and returns the directory name
func (s *Service) CreateWorkspace(ctx context.Context, id, branchName string, repos []domain.Repo) (string, error) {
	return s.CreateWorkspaceWithOptions(ctx, id, branchName, repos, CreateOptions{})
}

// CreateWorkspaceWithOptions creates a new workspace with configurable options.
func (s *Service) CreateWorkspaceWithOptions(ctx context.Context, id, branchName string, repos []domain.Repo, opts CreateOptions) (string, error) {
	if opts.Template != nil && branchName == "" && opts.Template.DefaultBranch != "" {
		branchName = opts.Template.DefaultBranch
	}

	branchName, err := s.resolveCreateBranchName(id, branchName)
	if err != nil {
		return "", err
	}

	dirName, err := s.config.ComputeWorkspaceDir(id)
	if err != nil {
		return "", err
	}

	if err := s.withWorkspaceLock(ctx, id, true, func() error {
		return s.createWorkspaceWithOptionsUnlocked(ctx, id, dirName, branchName, repos, opts)
	}); err != nil {
		return dirName, err
	}

	return dirName, nil
}

func (s *Service) createWorkspaceWithOptionsUnlocked(ctx context.Context, id, dirName, branchName string, repos []domain.Repo, opts CreateOptions) error {
	if err := s.ensureWorkspaceAvailable(id, dirName); err != nil {
		return err
	}

	ws := domain.Workspace{
		ID:         id,
		BranchName: branchName,
		Repos:      repos,
		DirName:    dirName,
	}

	if err := s.executeWorkspaceCreate(ctx, ws, repos, dirName); err != nil {
		return err
	}

	// Invalidate cache for this workspace ID
	s.cache.Invalidate(id)

	if opts.Template != nil && len(opts.Template.SetupCommands) > 0 {
		setupFailed := s.runTemplateSetupCommands(ctx, id, dirName, opts.Template.SetupCommands)
		if setupFailed {
			ws.SetupIncomplete = true
			if err := s.wsEngine.Save(ctx, ws); err != nil {
				if s.logger != nil {
					s.logger.Warn("Failed to mark workspace as partially initialized", "workspace_id", id, "error", err)
				}
			}
		}
	}

	// Run post_create hooks
	//nolint:contextcheck // Hooks manage their own timeout context per-hook
	if err := s.runPostCreateHooks(id, dirName, branchName, repos, opts); err != nil {
		// Hook failures don't rollback the workspace (per design.md)
		// But we return the error if not continuing on hook errors
		return err
	}

	return nil
}

func (s *Service) runTemplateSetupCommands(ctx context.Context, workspaceID, dirName string, commands []string) bool {
	if len(commands) == 0 {
		return false
	}

	workspacePath := filepath.Join(s.config.GetWorkspacesRoot(), dirName)
	failed := false

	for i, command := range commands {
		trimmed := strings.TrimSpace(command)
		if trimmed == "" {
			if s.logger != nil {
				s.logger.Warn("Skipping empty template setup command", "index", i, "workspace_id", workspaceID)
			}

			continue
		}

		if s.logger != nil {
			s.logger.Info("Running template setup command", "index", i, "workspace_id", workspaceID, "command", trimmed)
		}

		cmd := shellCommand(ctx, trimmed)
		cmd.Dir = workspacePath

		outputBytes, err := cmd.CombinedOutput()

		outputText := strings.TrimSpace(string(outputBytes))
		if outputText != "" && s.logger != nil {
			s.logger.Info("Template setup output", "index", i, "workspace_id", workspaceID, "output", outputText)
		}

		if err != nil {
			failed = true

			if s.logger != nil {
				s.logger.Warn("Template setup command failed", "index", i, "workspace_id", workspaceID, "command", trimmed, "error", err)
			}
		}
	}

	return failed
}

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		// #nosec G204 -- template setup commands are user-configured and expected.
		return exec.CommandContext(ctx, "cmd.exe", "/c", command)
	}

	// #nosec G204 -- template setup commands are user-configured and expected.
	return exec.CommandContext(ctx, "sh", "-c", command)
}

func (s *Service) resolveCreateBranchName(id, branchName string) (string, error) {
	// Validate inputs
	if err := validation.ValidateWorkspaceID(id); err != nil {
		return "", err
	}

	if err := validation.ValidateBranchName(branchName); err != nil {
		return "", err
	}

	// Default branch name is the workspace ID
	if branchName == "" {
		branchName = id
		// Validate the derived branch name (workspace IDs may contain chars invalid for git refs)
		if err := validation.ValidateBranchName(branchName); err != nil {
			return "", cerrors.NewInvalidArgument("workspace-id", "cannot be used as default branch name: "+err.Error())
		}
	}

	return branchName, nil
}

func (s *Service) executeWorkspaceCreate(ctx context.Context, ws domain.Workspace, repos []domain.Repo, dirName string) error {
	op := NewOperation(s.logger)
	op.AddStep(func() error {
		return s.wsEngine.Create(ctx, ws)
	}, func() error {
		return s.wsEngine.Delete(ctx, ws.ID)
	})
	op.AddStep(func() error {
		return s.cloneWorkspaceRepos(ctx, repos, dirName, ws.BranchName)
	}, func() error {
		return s.removeWorkspaceRepoWorktrees(ctx, dirName, repos)
	})

	return op.Execute()
}

func (s *Service) ensureWorkspaceAvailable(workspaceID, dirName string) error {
	metaPath := filepath.Join(s.config.GetWorkspacesRoot(), dirName, "workspace.yaml")
	if _, err := os.Stat(metaPath); err == nil {
		return cerrors.NewWorkspaceExists(workspaceID)
	} else if err != nil && !os.IsNotExist(err) {
		return cerrors.NewIOFailed("check workspace metadata", err)
	}

	return nil
}

func (s *Service) removeWorkspaceRepoWorktrees(ctx context.Context, dirName string, repos []domain.Repo) error {
	if s.gitEngine == nil {
		return cerrors.NewInternalError("git engine not initialized", nil)
	}

	var errs []error

	for _, repo := range repos {
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), dirName, repo.Name)
		if err := s.gitEngine.RemoveWorktree(ctx, repo.Name, worktreePath); err != nil {
			errs = append(errs, err)
		}
	}

	return joinErrors(errs...)
}

// cloneWorkspaceRepos clones all repositories for a workspace.
// It runs EnsureCanonical operations in parallel (bounded by config.parallel_workers)
// for performance, then creates worktrees sequentially (as they depend on the canonical).
func (s *Service) cloneWorkspaceRepos(ctx context.Context, repos []domain.Repo, dirName, branchName string) error {
	if len(repos) == 0 {
		return nil
	}

	if s.gitEngine == nil {
		return cerrors.NewInternalError("git engine not initialized", nil)
	}

	// Check for context cancellation before starting
	if ctx.Err() != nil {
		return cerrors.NewContextError(ctx, "create workspace", dirName)
	}

	executor := NewParallelExecutor(s.config.GetParallelWorkers())

	err := executor.Run(ctx, len(repos), func(runCtx context.Context, index int) error {
		repo := repos[index]

		_, err := s.gitEngine.EnsureCanonical(runCtx, repo.URL, repo.Name)
		if err != nil {
			return cerrors.WrapGitError(err, "ensure canonical for "+repo.Name)
		}

		return nil
	}, ParallelOptions{ContinueOnError: false})
	if err != nil {
		return err
	}

	// Create worktrees sequentially (they depend on the canonical being ready)
	for _, repo := range repos {
		// Check for context cancellation before each operation
		if ctx.Err() != nil {
			return cerrors.NewContextError(ctx, "create workspace", dirName)
		}

		// Create worktree
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), dirName, repo.Name)
		if err := s.gitEngine.CreateWorktree(ctx, repo.Name, worktreePath, branchName); err != nil {
			return cerrors.WrapGitError(err, fmt.Sprintf("create worktree for %s", repo.Name))
		}
	}

	return nil
}

// runPostCreateHooks runs post_create hooks if configured and not skipped.
// Returns nil if hooks are skipped or succeed, error otherwise.
func (s *Service) runPostCreateHooks(id, dirName, branchName string, repos []domain.Repo, opts CreateOptions) error {
	if opts.SkipHooks {
		return nil
	}

	hooksConfig := s.config.GetHooks()
	if len(hooksConfig.PostCreate) == 0 {
		return nil
	}

	hookCtx := domain.HookContext{
		WorkspaceID:   id,
		WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), dirName),
		BranchName:    branchName,
		Repos:         repos,
	}

	//nolint:contextcheck // Hooks manage their own timeout context per-hook
	if _, err := s.hookExecutor.ExecuteHooks(hooksConfig.PostCreate, hookCtx, ports.HookExecuteOptions{
		ContinueOnError: opts.ContinueOnHookErr,
	}); err != nil {
		s.logger.Error("post_create hooks failed", "error", err)

		if !opts.ContinueOnHookErr {
			return err
		}
	}

	return nil
}
