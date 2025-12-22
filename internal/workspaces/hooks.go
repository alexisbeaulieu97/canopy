package workspaces

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// HookPhase identifies which lifecycle hook set to execute.
type HookPhase string

const (
	// HookPhasePostCreate executes post_create hooks.
	HookPhasePostCreate HookPhase = "post_create"
	// HookPhasePreClose executes pre_close hooks.
	HookPhasePreClose HookPhase = "pre_close"
)

// RunHooks executes lifecycle hooks for an existing workspace without performing other actions.
//
//nolint:contextcheck // Hooks manage their own timeout context per-hook
func (s *Service) RunHooks(ctx context.Context, workspaceID string, phase HookPhase, continueOnError bool) error {
	workspace, dirName, err := s.findWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}

	hooksConfig := s.config.GetHooks()

	var selected []config.Hook

	switch phase {
	case HookPhasePostCreate:
		selected = hooksConfig.PostCreate
	case HookPhasePreClose:
		selected = hooksConfig.PreClose
	default:
		return cerrors.NewInvalidArgument("hook_phase", fmt.Sprintf("unsupported hook phase %q", phase))
	}

	if len(selected) == 0 {
		return nil
	}

	hookCtx := domain.HookContext{
		WorkspaceID:   workspaceID,
		WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), dirName),
		BranchName:    workspace.BranchName,
		Repos:         workspace.Repos,
	}

	if _, err := s.hookExecutor.ExecuteHooks(selected, hookCtx, ports.HookExecuteOptions{
		ContinueOnError: continueOnError,
	}); err != nil {
		if s.logger != nil {
			s.logger.Error(fmt.Sprintf("%s hooks failed", phase), "error", err)
		}

		if !continueOnError {
			return err
		}
	}

	return nil
}

// PreviewHooks returns a dry-run preview of lifecycle hooks for an existing workspace.
//
//nolint:contextcheck // Hooks manage their own timeout context per-hook
func (s *Service) PreviewHooks(ctx context.Context, workspaceID string, phase HookPhase) ([]domain.HookCommandPreview, error) {
	workspace, dirName, err := s.findWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	hooksConfig := s.config.GetHooks()

	var selected []config.Hook

	switch phase {
	case HookPhasePostCreate:
		selected = hooksConfig.PostCreate
	case HookPhasePreClose:
		selected = hooksConfig.PreClose
	default:
		return nil, cerrors.NewInvalidArgument("hook_phase", fmt.Sprintf("unsupported hook phase %q", phase))
	}

	if len(selected) == 0 {
		return nil, nil
	}

	hookCtx := domain.HookContext{
		WorkspaceID:   workspaceID,
		WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), dirName),
		BranchName:    workspace.BranchName,
		Repos:         workspace.Repos,
	}

	previews, err := s.hookExecutor.ExecuteHooks(selected, hookCtx, ports.HookExecuteOptions{
		DryRun: true,
	})
	if err != nil {
		if s.logger != nil {
			s.logger.Error(fmt.Sprintf("%s hook dry-run failed", phase), "error", err)
		}

		return nil, err
	}

	return previews, nil
}
