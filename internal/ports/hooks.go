// Package ports defines interfaces for external dependencies (hexagonal architecture).
package ports

import (
	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// HookExecutor defines the interface for executing lifecycle hooks.
type HookExecutor interface {
	// ExecuteHooks runs a list of hooks with the given context.
	// If ContinueOnError is true, it continues even if a hook fails.
	// If DryRun is true, it returns command previews without executing.
	ExecuteHooks(hooks []config.Hook, ctx domain.HookContext, opts HookExecuteOptions) ([]domain.HookCommandPreview, error)
}

// HookExecuteOptions controls hook execution behavior.
type HookExecuteOptions struct {
	ContinueOnError bool
	DryRun          bool
}
