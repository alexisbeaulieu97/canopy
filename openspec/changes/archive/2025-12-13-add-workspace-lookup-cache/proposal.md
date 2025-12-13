# Change: Add Efficient Workspace Lookup with Caching

## Why
The current `findWorkspace()` performs O(n) lookups by listing all workspaces, causing inefficient filesystem I/O when called repeatedly across CLI commands.

## What Changes
- **BREAKING**: Add `LoadByID(id string) (*Workspace, string, error)` method to `WorkspaceStorage` interface
  - All `WorkspaceStorage` implementers (including mocks) must add this method
  - Migration: Implementers can delegate to `List()` + filter as a temporary implementation
- Implement efficient lookup in workspace engine using direct path access
- Add optional in-memory cache for workspace metadata with TTL
- Cache invalidation on write operations (Create, Save, Delete, Close)

## Impact
- Affected specs: `core-architecture`, `workspace-management`
- Affected code:
  - `internal/ports/storage.go` - Add `LoadByID` method to interface
  - `internal/workspace/workspace.go` - Implement direct lookup
  - `internal/workspaces/service.go` - Use direct lookup, add caching
