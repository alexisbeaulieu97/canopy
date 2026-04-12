package workspaces

import (
	"context"
	"fmt"

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
func (s *Service) RunHooks(ctx context.Context, workspaceID string, phase HookPhase, continueOnError bool) error {
	selected, hookCtx, err := s.resolveHookExecution(ctx, workspaceID, phase)
	if err != nil {
		return err
	}

	if len(selected) == 0 {
		return nil
	}

	if _, err := s.hookExecutor.ExecuteHooks(ctx, selected, hookCtx, ports.HookExecuteOptions{
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
func (s *Service) PreviewHooks(ctx context.Context, workspaceID string, phase HookPhase) ([]domain.HookCommandPreview, error) {
	selected, hookCtx, err := s.resolveHookExecution(ctx, workspaceID, phase)
	if err != nil {
		return nil, err
	}

	if len(selected) == 0 {
		return nil, nil
	}

	previews, err := s.hookExecutor.ExecuteHooks(ctx, selected, hookCtx, ports.HookExecuteOptions{
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

func (s *Service) resolveHookExecution(ctx context.Context, workspaceID string, phase HookPhase) ([]ports.HookSpec, domain.HookContext, error) {
	workspace, dirName, err := s.findWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, domain.HookContext{}, err
	}

	selected, err := selectHooksForPhase(s.config.GetHooks(), phase)
	if err != nil {
		return nil, domain.HookContext{}, err
	}

	return selected, workspaceHookContext(s.config.GetWorkspacesRoot(), workspaceID, dirName, workspace.BranchName, workspace.Repos), nil
}

func selectHooksForPhase(hooksConfig ports.HooksConfig, phase HookPhase) ([]ports.HookSpec, error) {
	switch phase {
	case HookPhasePostCreate:
		return hooksConfig.PostCreate, nil
	case HookPhasePreClose:
		return hooksConfig.PreClose, nil
	default:
		return nil, cerrors.NewInvalidArgument("hook_phase", fmt.Sprintf("unsupported hook phase %q", phase))
	}
}
