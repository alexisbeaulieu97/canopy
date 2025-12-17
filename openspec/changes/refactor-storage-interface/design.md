## Context
The `WorkspaceStorage` interface currently exposes filesystem implementation details through `dirName` parameters. This creates tight coupling between the service layer and the storage implementation.

**Current Interface (problematic methods):**
```go
type WorkspaceStorage interface {
    Create(dirName, id, branchName string, repos []domain.Repo) error
    Save(dirName string, ws domain.Workspace) error
    Load(dirName string) (*domain.Workspace, error)
    Close(dirName string, ws domain.Workspace, closedAt time.Time) (*domain.ClosedWorkspace, error)
    Rename(oldDirName, newDirName, newID string) error
    // ... other methods
}
```

The service layer must track both workspace IDs and directory names, passing them through multiple method calls.

## Goals / Non-Goals

**Goals:**
- Make the interface implementation-agnostic (ID-based)
- Add context support for cancellation/timeout
- Simplify the service layer by removing dirName tracking
- Maintain backward compatibility during migration

**Non-Goals:**
- Changing the underlying storage format (YAML files)
- Adding new storage backends (database, etc.)
- Changing domain types

## Decisions

### Decision 1: ID-based interface with internal directory mapping
The storage implementation will manage the mapping from workspace ID to directory path internally. Callers only provide IDs.

**Proposed Interface:**
```go
type WorkspaceStorage interface {
    // Create creates a new workspace from the provided domain object.
    Create(ctx context.Context, ws domain.Workspace) error

    // Save persists changes to an existing workspace.
    Save(ctx context.Context, ws domain.Workspace) error

    // Load retrieves a workspace by ID.
    Load(ctx context.Context, id string) (*domain.Workspace, error)

    // Close archives a workspace and returns the closed entry.
    Close(ctx context.Context, id string, closedAt time.Time) (*domain.ClosedWorkspace, error)

    // List returns all active workspaces.
    List(ctx context.Context) ([]domain.Workspace, error)

    // Delete removes a workspace by ID.
    Delete(ctx context.Context, id string) error

    // Rename changes a workspace's ID.
    Rename(ctx context.Context, oldID, newID string) error

    // ListClosed returns archived workspaces.
    ListClosed(ctx context.Context) ([]domain.ClosedWorkspace, error)

    // LatestClosed returns the most recent closed entry for a workspace.
    LatestClosed(ctx context.Context, id string) (*domain.ClosedWorkspace, error)

    // DeleteClosed removes a closed workspace entry.
    DeleteClosed(ctx context.Context, path string) error
}
```

**Rationale:** This is the cleanest approach that fully abstracts the storage implementation. The storage layer decides how to map IDs to paths.

### Decision 2: List returns slice instead of map
Change `List()` return type from `map[string]domain.Workspace` to `[]domain.Workspace`.

**Rationale:** The map key was `dirName`, which is an implementation detail. Since IDs are now in the domain object, a slice is sufficient and cleaner.

### Decision 3: ID-to-directory mapping strategy
The storage implementation uses a simple strategy: `directory = sanitize(id)`. For workspaces where ID differs from directory (legacy or renamed), the implementation falls back to scanning.

**Rationale:** This maintains backward compatibility with existing workspaces while providing fast-path lookup for the common case.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Breaking change to interface | Implement incrementally; update callers file-by-file |
| Performance regression from scanning | Keep direct ID-to-path lookup as fast path; scanning only for edge cases |
| Test updates required | Update mocks and test helpers as part of implementation |

## Migration Plan

1. Update interface definition with new signatures
2. Update mock implementation
3. Update storage implementation (Engine)
4. Update service layer callers one method at a time
5. Update tests
6. Remove deprecated LoadByID (now redundant with new Load)

## Open Questions

None - design is straightforward refactoring.
