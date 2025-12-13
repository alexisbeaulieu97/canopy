# Change: Add Missing Port Interfaces for Testability

## Why
Several concrete types used by the service layer lack port interfaces, reducing testability and violating the hexagonal architecture principle of depending on abstractions. The `hooks.Executor`, `DiskUsageCalculator`, and `WorkspaceCache` are currently injected as concrete types, making it difficult to mock them in unit tests.

## What Changes
- Add `HookExecutor` interface to `internal/ports/`
- Add `DiskUsage` interface to `internal/ports/`
- Add `WorkspaceCache` interface to `internal/ports/`
- Update `Service` to depend on interfaces instead of concrete types
- Add mock implementations to `internal/mocks/`
- Update functional options in `internal/app/app.go`

## Impact
- Affected specs: core-architecture
- Affected code:
  - `internal/ports/hooks.go` (new file)
  - `internal/ports/diskusage.go` (new file)
  - `internal/ports/cache.go` (new file)
  - `internal/mocks/hooks.go` (new file)
  - `internal/mocks/diskusage.go` (new file)
  - `internal/mocks/cache.go` (new file)
  - `internal/workspaces/service.go` (type changes)
  - `internal/app/app.go` (functional options)

