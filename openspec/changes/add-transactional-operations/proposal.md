# Change: Add Transactional Operation Hardening

## Why
Several service operations can leave the system in an inconsistent state on partial failure:
- `CreateWorkspaceWithOptions` writes metadata before cloning; failures leave orphan metadata
- `AddRepoToWorkspace` creates worktrees before updating metadata; failures leave orphan worktrees
- `RestoreWorkspace` deletes closed entries after recreation without compensating for failures

A staged approach with cleanup-on-failure would prevent these inconsistencies without adding full transaction complexity.

## What Changes
- Reorder operations to minimize inconsistent states (create filesystem artifacts before metadata)
- Add deferred cleanup functions that run on failure
- Implement a lightweight `Operation` helper that tracks steps and provides rollback
- Add explicit validation before mutations (fail fast)
- Log compensation actions for debugging

## Impact
- Affected specs: `specs/core-architecture/spec.md`
- Affected code:
  - `internal/workspaces/operation.go` (new) - Operation helper type
  - `internal/workspaces/service.go` - Refactor CreateWorkspace, AddRepoToWorkspace, RestoreWorkspace
  - `internal/workspaces/canonical.go` - Refactor Add operation
- Breaking changes: None (internal refactoring)
