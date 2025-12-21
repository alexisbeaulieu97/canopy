# Change: Add Context Support to Local Git Operations

## Why
Several `GitOperations` methods don't accept `context.Context`, making them impossible to cancel or time out. This is inconsistent with network operations (Clone, Fetch, Push, Pull) which already support context. A hung filesystem or git operation can stall the CLI/TUI indefinitely.

Affected methods without context:
- `CreateWorktree(repoName, worktreePath, branchName string)`
- `Status(path string)`
- `Checkout(path, branchName string, create bool)`
- `List() ([]string, error)`

## What Changes
- Add `context.Context` as first parameter to `CreateWorktree`, `Status`, `Checkout`, and `List`
- Update all callers in service layer and CLI to pass context
- Apply reasonable default timeouts for local operations (30s)
- Ensure context cancellation is checked at appropriate points

## Impact
- Affected specs: `core-architecture`
- Affected code:
  - `internal/ports/git.go` - Update interface signatures
  - `internal/gitx/git.go` - Update implementations
  - `internal/mocks/git.go` - Update mock
  - `internal/workspaces/service.go` - Pass context to git operations
  - `internal/workspaces/git_service.go` - Pass context to git operations
  - `internal/workspaces/orphan_service.go` - Pass context to git operations
  - `cmd/canopy/*.go` - Ensure cmd.Context() is propagated
- **No breaking changes** for end users (internal interface change only)

