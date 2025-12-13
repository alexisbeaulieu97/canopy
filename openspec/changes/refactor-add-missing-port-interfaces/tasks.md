# Tasks: Add Missing Port Interfaces for Testability

## Implementation Checklist

### 1. Add HookExecutor Interface
- [ ] 1.1 Create `internal/ports/hooks.go` with `HookExecutor` interface:
  ```go
  type HookExecutor interface {
      ExecuteHooks(hooks []config.Hook, ctx hooks.HookContext, continueOnError bool) error
  }
  ```
- [ ] 1.2 Verify `hooks.Executor` satisfies the interface
- [ ] 1.3 Create `internal/mocks/hooks.go` with mock implementation
- [ ] 1.4 Write tests for mock implementation

### 2. Add DiskUsage Interface
- [ ] 2.1 Create `internal/ports/diskusage.go` with `DiskUsage` interface:
  ```go
  type DiskUsage interface {
      CachedUsage(root string) (int64, time.Time, error)
      Calculate(root string) (int64, time.Time, error)
      InvalidateCache(root string)
      ClearCache()
  }
  ```
- [ ] 2.2 Verify `DiskUsageCalculator` satisfies the interface
- [ ] 2.3 Create `internal/mocks/diskusage.go` with mock implementation
- [ ] 2.4 Write tests for mock implementation

### 3. Add WorkspaceCache Interface
- [ ] 3.1 Create `internal/ports/cache.go` with `WorkspaceCache` interface:
  ```go
  type WorkspaceCache interface {
      Get(id string) (*domain.Workspace, string, bool)
      Set(id string, ws *domain.Workspace, dirName string)
      Invalidate(id string)
      InvalidateAll()
      Size() int
  }
  ```
- [ ] 3.2 Verify `workspaces.WorkspaceCache` satisfies the interface
- [ ] 3.3 Create `internal/mocks/cache.go` with mock implementation
- [ ] 3.4 Write tests for mock implementation

### 4. Update Service Dependencies
- [ ] 4.1 Change `Service.hookExecutor` from `*hooks.Executor` to `ports.HookExecutor`
- [ ] 4.2 Change `Service.diskUsage` from `*DiskUsageCalculator` to `ports.DiskUsage`
- [ ] 4.3 Change `Service.cache` from `*WorkspaceCache` to `ports.WorkspaceCache`
- [ ] 4.4 Update `NewService` constructor parameter types
- [ ] 4.5 Update functional options in `app.go`

### 5. Update Tests
- [ ] 5.1 Update existing tests to use mock implementations where appropriate
- [ ] 5.2 Verify all tests pass
- [ ] 5.3 Add test demonstrating mockability of new interfaces

### 6. Verification
- [ ] 6.1 Run `go build ./...` to verify compilation
- [ ] 6.2 Run full test suite with race detector
- [ ] 6.3 Verify interface satisfaction with compile-time checks

