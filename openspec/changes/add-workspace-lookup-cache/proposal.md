# Change: Add Efficient Workspace Lookup with Caching

## Why
The current `findWorkspace()` method performs O(n) lookup by listing all workspaces and iterating to find a match. This is called multiple times per CLI command (e.g., `GetStatus`, `CloseWorkspace`, `AddRepoToWorkspace` all call `findWorkspace`). With many workspaces, this becomes inefficient and causes unnecessary filesystem I/O.

## What Changes
- Add direct `LoadByID(id string)` method to `WorkspaceStorage` interface
- Implement efficient lookup in workspace engine using direct path access
- Add optional in-memory cache for workspace metadata with TTL
- Cache invalidation on write operations (Create, Save, Delete, Close)

## Impact
- Affected specs: `core-architecture`, `workspace-management`
- Affected code:
  - `internal/ports/storage.go` - Add `LoadByID` method to interface
  - `internal/workspace/workspace.go` - Implement direct lookup
  - `internal/workspaces/service.go` - Use direct lookup, add caching
