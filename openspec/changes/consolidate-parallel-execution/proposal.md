# Change: Consolidate parallel execution patterns

## Why
Parallel execution logic is duplicated across sync, status, and canonical operations with inconsistent patterns for errgroup usage, limits, and cancellation. This creates maintenance burden and behavioral drift.

## What Changes
- Create unified `ParallelExecutor` abstraction
- Consolidate errgroup + limits + cancellation patterns
- Add configurable worker pool limits
- Ensure consistent context propagation

## Impact
- Affected specs: core-architecture
- Affected code:
  - `internal/workspaces/parallel.go` - Create unified executor
  - `internal/workspaces/sync.go` - Use unified executor
  - `internal/workspaces/status_batch.go` - Use unified executor
  - `internal/workspaces/canonical.go` - Use unified executor
  - `internal/workspaces/git_service.go` - Use unified executor
