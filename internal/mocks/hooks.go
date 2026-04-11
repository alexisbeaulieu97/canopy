// Package mocks provides mock implementations for testing.
package mocks

import (
	"context"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Compile-time check that MockHookExecutor implements ports.HookExecutor.
var _ ports.HookExecutor = (*MockHookExecutor)(nil)

// MockHookExecutor is a mock implementation of ports.HookExecutor for testing.
type MockHookExecutor struct {
	// ExecuteHooksFunc is called when ExecuteHooks is invoked.
	ExecuteHooksFunc func(ctx context.Context, hks []ports.HookSpec, hookCtx domain.HookContext, opts ports.HookExecuteOptions) ([]domain.HookCommandPreview, error)

	// ExecuteHooksCalls records all calls to ExecuteHooks for verification.
	ExecuteHooksCalls []ExecuteHooksCall

	// ExecuteHooksErr is the error to return if ExecuteHooksFunc is not set.
	ExecuteHooksErr error

	// ExecuteHooksPreviews is returned when ExecuteHooksFunc is not set.
	ExecuteHooksPreviews []domain.HookCommandPreview
}

// ExecuteHooksCall records a single call to ExecuteHooks.
type ExecuteHooksCall struct {
	BaseCtx context.Context
	Hooks   []ports.HookSpec
	Ctx     domain.HookContext
	Options ports.HookExecuteOptions
}

// NewMockHookExecutor creates a new MockHookExecutor with sensible defaults.
func NewMockHookExecutor() *MockHookExecutor {
	return &MockHookExecutor{
		ExecuteHooksCalls: make([]ExecuteHooksCall, 0),
	}
}

// ExecuteHooks calls the mock function if set, otherwise returns ExecuteHooksErr.
func (m *MockHookExecutor) ExecuteHooks(
	ctx context.Context,
	hks []ports.HookSpec,
	hookCtx domain.HookContext,
	opts ports.HookExecuteOptions,
) ([]domain.HookCommandPreview, error) {
	m.ExecuteHooksCalls = append(m.ExecuteHooksCalls, ExecuteHooksCall{
		BaseCtx: ctx,
		Hooks:   hks,
		Ctx:     hookCtx,
		Options: opts,
	})

	if m.ExecuteHooksFunc != nil {
		return m.ExecuteHooksFunc(ctx, hks, hookCtx, opts)
	}

	return m.ExecuteHooksPreviews, m.ExecuteHooksErr
}

// ResetCalls clears the recorded calls.
func (m *MockHookExecutor) ResetCalls() {
	m.ExecuteHooksCalls = make([]ExecuteHooksCall, 0)
}

// CallCount returns the number of times ExecuteHooks was called.
func (m *MockHookExecutor) CallCount() int {
	return len(m.ExecuteHooksCalls)
}
