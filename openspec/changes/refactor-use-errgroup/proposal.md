# Change: Replace Custom Parallel Logic with errgroup

## Why
The current `parallel.go` is complex due to manual channel/worker pool and cancellation logic (~137 lines). Using `golang.org/x/sync/errgroup` simplifies this to ~30 lines while providing bounded concurrency and automatic fail-fast cancellation.

## What Changes
- Add `golang.org/x/sync/errgroup` dependency
- Refactor `runParallelCanonical` to use `errgroup.Group`
- Remove custom `canonicalWorker` and `collectCanonicalResults` functions
- Remove `canonicalResult` struct (no longer needed)
- Simplify parallel execution to ~30 lines

## Impact
- Affected specs: core-architecture (implementation detail, no behavior change)
- Affected code:
  - `internal/workspaces/parallel.go` (major simplification)
  - `internal/workspaces/parallel_test.go` (test updates if needed)
  - `go.mod` (new dependency)
- No breaking changes - same fail-fast behavior preserved
