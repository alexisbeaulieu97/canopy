## 1. Add Dependency
- [ ] 1.1 Run `go get golang.org/x/sync/errgroup`
- [ ] 1.2 Verify dependency in `go.mod`

## 2. Refactor Parallel Implementation
- [ ] 2.1 Import `golang.org/x/sync/errgroup` in `parallel.go`
- [ ] 2.2 Rewrite `runParallelCanonical` using `errgroup.Group`
- [ ] 2.3 Use `g.SetLimit(opts.workers)` for bounded concurrency
- [ ] 2.4 Use `g.Go()` for spawning goroutines
- [ ] 2.5 Remove `canonicalResult` struct
- [ ] 2.6 Remove `canonicalWorker` function
- [ ] 2.7 Remove `collectCanonicalResults` function
- [ ] 2.8 Keep `runSequentialCanonical` unchanged (single-item optimization)

## 3. Update Tests
- [ ] 3.1 Review `parallel_test.go` for any test updates needed
- [ ] 3.2 Ensure bounded concurrency tests still pass
- [ ] 3.3 Ensure fail-fast behavior tests still pass
- [ ] 3.4 Ensure context cancellation tests still pass

## 4. Verification
- [ ] 4.1 Run `go build ./...` to verify compilation
- [ ] 4.2 Run `go test ./internal/workspaces/...` to verify tests pass
- [ ] 4.3 Run `golangci-lint run` to check for issues
- [ ] 4.4 Remove `nolint:gocyclo` directive (no longer needed)
