# Change: Refactor WorkspaceStorage Interface to be ID-based

## Why
The current `WorkspaceStorage` interface leaks filesystem implementation details through `dirName` parameters. Methods like `Create(dirName, id, ...)`, `Save(dirName, ws)`, and `Load(dirName)` expose that workspaces are stored in directories, coupling callers to the filesystem implementation. This violates hexagonal architecture principles where ports should be implementation-agnostic.

Additionally, methods lack `context.Context` support, preventing proper cancellation and timeout propagation for I/O operations.

## What Changes
- **BREAKING** Add `context.Context` as first parameter to all `WorkspaceStorage` methods
- **BREAKING** Refactor `Create` to accept a `domain.Workspace` object instead of separate parameters
- **BREAKING** Refactor `Save`, `Load`, `Close` to use workspace ID instead of `dirName`
- **BREAKING** Refactor `Rename` to use old/new IDs instead of dirNames
- **BREAKING** Refactor `DeleteClosed` to use workspace ID and timestamp instead of path
- **BREAKING** Remove `LoadByID` in favor of unified `Load(ctx, id)`
- Storage implementation manages ID-to-directory mapping internally

## Impact
- Affected specs: core-architecture
- Affected code:
  - `internal/ports/storage.go` (interface definition)
  - `internal/storage/storage.go` (implementation - assumes rename from workspace package)
  - `internal/workspaces/service.go` (primary caller)
  - `internal/workspaces/git_service.go` (caller)
  - `internal/mocks/storage.go` (mock implementation)
  - `internal/workspaces/service_test.go` (test updates)
- **Dependency**: Should be implemented after `refactor-rename-workspace-package`
