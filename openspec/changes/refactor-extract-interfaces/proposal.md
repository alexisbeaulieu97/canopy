```markdown
# Change: Refactor to Extract Interfaces for Testability

## Why
Current code uses concrete types (`*gitx.GitEngine`, `*workspace.Engine`) directly, making unit testing difficult. Tests require real filesystem and git operations, which are slow and brittle. Extracting interfaces allows:
- Mock implementations for fast unit tests
- Test business logic without I/O
- Dependency injection for different environments
- Better adherence to hexagonal architecture principles

## What Changes
- Define `GitOperations` interface for git operations
- Define `WorkspaceStorage` interface for workspace persistence
- Define `ConfigProvider` interface for configuration
- Update `Service` to depend on interfaces, not concrete types
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
```
