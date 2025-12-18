# Change: Add Workspace-Level Concurrency Control

## Why
Currently, concurrent operations on the same workspace can race (e.g., close/sync/rename running in parallel from different terminals). This can lead to corrupted metadata or inconsistent filesystem state. A lightweight lock mechanism would prevent these races without adding significant complexity.

## What Changes
- Implement file-based locking per workspace (handles multiple CLI instances)
- Lock acquired at service boundary before mutating operations
- Lock file stored in workspace directory (`.canopy.lock`)
- Timeout-based lock acquisition with configurable wait time
- Stale lock detection and cleanup (locks older than configurable threshold)
- Read operations do not require locks
- Lock status visible in `workspace list` output (optional `--show-locks` flag)

## Impact
- Affected specs: `specs/core-architecture/spec.md`
- Affected code:
  - `internal/workspaces/lock.go` (new) - Lock manager implementation
  - `internal/workspaces/service.go` - Add lock acquisition to mutating methods
  - `internal/domain/workspace.go` - Add lock status to domain model
- Breaking changes: None (additive only)
