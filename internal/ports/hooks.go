// Package ports defines interfaces for external dependencies (hexagonal architecture).
package ports

import (
	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// HookExecutor defines the interface for executing lifecycle hooks.
type HookExecutor interface {
	// ExecuteHooks runs a list of hooks with the given context.
	// If continueOnError is true, it continues even if a hook fails.
	ExecuteHooks(hooks []config.Hook, ctx domain.HookContext, continueOnError bool) error
}
