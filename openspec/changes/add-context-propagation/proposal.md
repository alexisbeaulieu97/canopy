# Change: Add Context Propagation Through Interfaces

## Why
Several internal interfaces (WorkspaceFinder, ConfigProvider) don't accept context.Context parameters, preventing proper cancellation and timeout propagation. This means operations can't be cancelled when a parent context is cancelled (e.g., Ctrl+C), leading to poor user experience during long-running operations.

## What Changes
- Add context.Context parameter to WorkspaceFinder interface methods
- Update callers (GitService, OrphanService, ExportService) to pass context through
- Ensure context cancellation propagates to all sub-operations
- Add context-aware timeout handling for workspace operations

## Impact
- Affected specs: core-architecture
- Affected code:
  - `internal/ports/git.go` (interface changes)
  - `internal/workspaces/service.go` (implementation)
  - `internal/workspaces/git_service.go`
  - `internal/workspaces/orphan_service.go`
  - `internal/workspaces/export_service.go`
- **BREAKING**: Interface signature changes (internal only, no public API impact)
