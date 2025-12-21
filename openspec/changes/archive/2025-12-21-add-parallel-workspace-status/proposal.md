# Change: Add Parallel Workspace Status Fetching

## Why
The `canopy workspace list --status` command currently fetches status for each workspace sequentially. For users with many workspaces (10+), this results in a slow "bird's-eye view" experience. The status check involves git operations that are largely I/O bound and can be parallelized for significant speedup.

## What Changes
- Implement parallel worker pool for workspace status fetching in `workspace list --status`
- Use bounded concurrency (respecting `parallel_workers` config setting)
- Maintain deterministic output ordering (workspaces appear in consistent order regardless of fetch completion order)
- Add `--parallel-status` flag to explicitly control behavior (default: parallel)
- Add `--sequential-status` flag for debugging or when parallelism causes issues

## Impact
- Affected specs: `specs/cli/spec.md`
- Affected code:
  - `cmd/canopy/workspace.go:200-350` - Refactor list command status fetching
  - `internal/workspaces/service.go` - Add `GetWorkspaceStatusBatch` method
- No breaking changes - existing behavior preserved with new default
- Performance improvement: ~5-10x faster for users with many workspaces
