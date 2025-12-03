# Change: Refactor to Extract Interfaces for Testability

## Why
Current code uses concrete types directly, making unit tests slow and brittle due to required filesystem/git I/O. Extracting interfaces enables mock implementations for fast, reliable unit tests and better hexagonal architecture separation.

## What Changes
- Define `GitOperations` interface for git operations
- Define `WorkspaceStorage` interface for workspace persistence
- Define `ConfigProvider` interface for configuration
- **BREAKING** - Update `Service` to depend on interfaces, not concrete types (constructor signature change)
- Create mock implementations for testing
- Update existing tests to use mocks where appropriate

## Impact
- Affected specs: `specs/core-architecture/spec.md`
- Affected code:
  - `internal/ports/git.go` (new) - GitOperations interface
  - `internal/ports/storage.go` (new) - WorkspaceStorage interface
  - `internal/workspaces/service.go` - Accept interfaces in constructor
  - `internal/workspaces/service_test.go` - Use mocks
  - `internal/mocks/` (new) - Mock implementations
