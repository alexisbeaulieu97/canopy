# Tasks: Add Missing Port Interfaces for Testability

## Implementation Checklist

### 1. Add HookExecutor Interface
- [x] 1.1 Create `internal/ports/hooks.go` with `HookExecutor` interface:
  ```go
  type HookExecutor interface {
      ExecuteHooks(hooks []config.Hook, ctx domain.HookContext, continueOnError bool) error
  }
  ```
- [x] 1.2 Verify `hooks.Executor` satisfies the interface
- [x] 1.3 Create `internal/mocks/hooks.go` with mock implementation
- [x] 1.4 Write tests for mock implementation

### 2. Add DiskUsage Interface
- [x] 2.1 Create `internal/ports/diskusage.go` with `DiskUsage` interface:
  ```go
  type DiskUsage interface {
      CachedUsage(root string) (int64, time.Time, error)
      Calculate(root string) (int64, time.Time, error)
      InvalidateCache(root string)
      ClearCache()
  }
  ```
- [x] 2.2 Verify `DiskUsageCalculator` satisfies the interface
- [x] 2.3 Create `internal/mocks/diskusage.go` with mock implementation
- [x] 2.4 Write tests for mock implementation

### 3. Add WorkspaceCache Interface
- [x] 3.1 Create `internal/ports/cache.go` with `WorkspaceCache` interface:
  ```go
  type WorkspaceCache interface {
      Get(id string) (*domain.Workspace, string, bool)
      Set(id string, ws *domain.Workspace, dirName string)
      Invalidate(id string)
      InvalidateAll()
      Size() int
  }
  ```
- [x] 3.2 Verify `workspaces.WorkspaceCache` satisfies the interface
- [x] 3.3 Create `internal/mocks/cache.go` with mock implementation
- [x] 3.4 Write tests for mock implementation

### 4. Update Service Dependencies
- [x] 4.1 Change `Service.hookExecutor` from `*hooks.Executor` to `ports.HookExecutor`
- [x] 4.2 Change `Service.diskUsage` from `*DiskUsageCalculator` to `ports.DiskUsage`
- [x] 4.3 Change `Service.cache` from `*WorkspaceCache` to `ports.WorkspaceCache`
- [x] 4.4 Update `NewService` constructor with `ServiceOption` variadic parameter
- [x] 4.5 Add functional options (`WithHookExecutor`, `WithDiskUsage`, `WithCache`) in `workspaces/service.go`

### 5. Update Tests
- [x] 5.1 Update existing tests to use mock implementations where appropriate
- [x] 5.2 Verify all tests pass
- [x] 5.3 Add test demonstrating mockability of new interfaces

### 6. Verification
- [x] 6.1 Run `go build ./...` to verify compilation
- [x] 6.2 Run full test suite with race detector
- [x] 6.3 Verify interface satisfaction with compile-time checks

