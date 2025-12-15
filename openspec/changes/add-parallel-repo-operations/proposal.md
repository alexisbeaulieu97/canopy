# Change: Add Parallel Repository Operations During Workspace Creation

## Why
Currently, repository cloning/setup during workspace creation is sequential in `service.go`. For workspaces with many repositories (3+), this can be slow:
- Each clone/fetch operation is network-bound
- Operations are independent and could run concurrently
- Users wait unnecessarily for sequential network I/O

## What Changes
- Add parallel execution for `EnsureCanonical` calls during workspace creation
- Use bounded concurrency (default 4 workers) to avoid overwhelming the network
- Add progress reporting for parallel operations
- Maintain atomic rollback on failure (if any repo fails, clean up all)
- Make concurrency configurable via config option

## Impact
- Affected specs: `core-architecture`
- Affected code:
  - `internal/workspaces/service.go:CreateWorkspace` - Parallelize repo setup
  - `internal/config/config.go` - Add `parallel_workers` config option
  - Error handling - Collect errors from all workers, fail-fast optional
- **Performance**: Significant speedup for workspaces with multiple repos (up to 4x for 4+ repos)
- **Risk**: Medium - Concurrency adds complexity, needs careful error handling

