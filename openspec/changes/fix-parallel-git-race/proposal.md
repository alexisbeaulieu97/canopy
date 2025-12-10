# Change: Fix Race Condition in Parallel Git Operations

## Why
The `runGitParallel` function in `service.go:604-660` has a race condition: when `continueOnError=false`, the first error should stop other goroutines, but currently all goroutines run to completion. This wastes resources and can cause confusing error messages.

## What Changes
- Add early termination support to `runGitParallel` using context cancellation
- Ensure goroutines check for cancellation before starting work
- Properly synchronize error collection
- Add tests for race condition scenarios

## Impact
- **Affected specs**: `specs/core-architecture/spec.md`
- **Affected code**:
  - `internal/workspaces/service.go:604-660` - `runGitParallel` function
- **Risk**: Low - Bug fix with clear scope, improves existing behavior

