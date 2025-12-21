## 1. Add Dependency
- [x] 1.1 Run `go get golang.org/x/sync/errgroup`
- [x] 1.2 Verify dependency in `go.mod`

## 2. Refactor Parallel Implementation
- [x] 2.1 Import `golang.org/x/sync/errgroup` in `parallel.go`
- [x] 2.2 Rewrite `runParallelCanonical` using `errgroup.Group`
- [x] 2.3 Use `g.SetLimit(opts.workers)` for bounded concurrency
- [x] 2.4 Use `g.Go()` for spawning goroutines
- [x] 2.5 Remove `canonicalResult` struct
- [x] 2.6 Remove `canonicalWorker` function
- [x] 2.7 Remove `collectCanonicalResults` function
- [x] 2.8 Keep `runSequentialCanonical` unchanged (single-item optimization)

## 3. Update Tests
- [x] 3.1 Review `parallel_test.go` for any test updates needed
- [x] 3.2 Ensure bounded concurrency tests still pass
- [x] 3.3 Ensure fail-fast behavior tests still pass
- [x] 3.4 Ensure context cancellation tests still pass

## 4. Verification
- [x] 4.1 Run `go build ./...` to verify compilation
- [x] 4.2 Run `go test ./internal/workspaces/...` to verify tests pass
- [x] 4.3 Run `golangci-lint run` to check for issues
- [x] 4.4 Remove `nolint:gocyclo` directive (no longer needed)
