# Change: Replace Custom Parallel Logic with errgroup

## Why
The current `parallel.go` implementation uses custom channel-based work distribution with manual worker pools and context cancellation. This is ~137 lines of complex concurrent code that can be significantly simplified using the standard `golang.org/x/sync/errgroup` package, which provides:
- Built-in bounded concurrency via `SetLimit()`
- Automatic context cancellation on first error (fail-fast)
- Cleaner, more idiomatic Go concurrent patterns
- Well-tested standard library code

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
