# Change: Add Parallel Repository Operations During Workspace Creation

## Why
Currently, repository cloning/setup during workspace creation is sequential. For workspaces with many repositories (3+), this can be slow due to network-bound operations that are independent and could run concurrently.

## What Changes
- Add parallel execution for `EnsureCanonical` calls during workspace creation
- Use bounded concurrency (default 4 workers) to avoid overwhelming the network
- Add progress reporting for parallel operations
- Default to fail-fast behavior (cancel remaining on first failure, clean up)
- Add optional `continue_on_error` config to aggregate errors instead
- Make concurrency configurable via `parallel_workers` config option

## Impact
- Affected specs: `core-architecture`
- Affected code:
  - `internal/workspaces/service.go:CreateWorkspace` - Parallelize repo setup
  - `internal/config/config.go` - Add `parallel_workers` and `continue_on_error` config options
  - Error handling - Fail-fast by default, continue-on-error optional
- **Performance**: Significant speedup for workspaces with multiple repos (up to 4x for 4+ repos)
- **Risk**: Medium - Concurrency adds complexity, needs careful error handling

