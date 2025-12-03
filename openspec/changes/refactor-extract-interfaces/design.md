# Design Document: Extract Interfaces for Testability

## Context

The current `Service` struct in `internal/workspaces/service.go` directly depends on concrete types:
- `*gitx.Engine` for git operations
- `*workspace.Engine` for workspace filesystem operations
- `*config.Config` for configuration

This tight coupling requires real filesystem and git I/O in tests, making unit tests slow and brittle. Tests cannot run in isolation and depend on external state.

## Goals

- Enable fast, isolated unit tests without filesystem/git I/O
- Adopt hexagonal architecture with clear ports and adapters
- Reduce coupling between service layer and infrastructure
- Allow mock implementations for deterministic testing

## Non-Goals

- Changing the public API of existing commands
- Introducing a DI framework or container
- Refactoring the TUI layer (separate change)
- Modifying git operation semantics

## Decisions

### Interface Extraction Strategy

Extract three core interfaces to decouple the service layer:

1. **GitOperations** - Git clone, fetch, pull, push, status
2. **WorkspaceStorage** - CRUD for workspaces on filesystem
3. **ConfigProvider** - Configuration loading and validation

### Rationale

Extracting interfaces allows:
- Mock implementations for unit tests
- Clear boundaries between business logic and infrastructure
- Easier swapping of implementations (e.g., in-memory storage for tests)
- Better adherence to dependency inversion principle

### Constructor Signature

```go
// Interface-driven signature
func NewService(cfg ConfigProvider, git GitOperations, storage WorkspaceStorage) *Service
```

### Dependency Injection Patterns (Decision Details)

- Constructor injection (preferred for required dependencies)
- Functional options for optional/configurable behavior
- No DI framework needed; manual wiring is sufficient for this codebase size

## Risks and Trade-offs

### Migration Complexity
- All call sites to `NewService` must be updated
- Existing integration tests continue to work with real implementations
- Risk: missed call site causes compile error (safe failure)

### Performance Implications
- Interface indirection has negligible overhead
- Mock implementations enable faster test suites overall
- No runtime performance impact for production code

### Testing Impacts
- Unit tests become faster and more reliable
- Integration tests still use real implementations
- Need to maintain both mock and real implementations

## Migration Plan

### Phase 1: Define Interfaces
1. Create `internal/ports/git.go` with `GitOperations` interface
2. Create `internal/ports/storage.go` with `WorkspaceStorage` interface
3. Create `internal/ports/config.go` with `ConfigProvider` interface

### Phase 2: Update Service
1. Modify `Service` struct to accept interfaces
2. Update `NewService` constructor signature
3. Ensure existing adapters implement interfaces

### Phase 3: Create Mocks
1. Create `internal/mocks/git.go` implementing `GitOperations`
2. Create `internal/mocks/storage.go` implementing `WorkspaceStorage`
3. Create `internal/mocks/config.go` implementing `ConfigProvider`

### Phase 4: Update Tests
1. Update unit tests to use mocks
2. Verify integration tests still work
3. Add new unit tests for previously untestable code paths

### Deprecation Timeline
- Old constructor deprecated in v0.x
- Removed in next minor version

### Rollback Criteria
- Tests fail after migration
- Unexpected runtime errors in production
- Performance regression > 5%

## Open Questions

### Scope of Injectable Components
- Should `logging.Logger` also be an interface?
- Should TUI model accept interfaces for service calls?

### Phased Rollout
- Can interfaces be introduced incrementally (one at a time)?
- Should we use build tags to support both old and new constructors?

### Mock Generation
- Manual mocks vs generated (mockgen/mockery)?
- Trade-off: manual mocks are simpler but require maintenance
