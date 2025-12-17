# Change: Extract Registry Save/Rollback Pattern

## Why

The registry save-with-rollback pattern is duplicated 3 times in `cmd/canopy/repo.go`:
- `repoAddCmd` (lines 103-114)
- `repoRegisterCmd` (lines 215-225)
- `repoUnregisterCmd` (lines 250-261)

Each instance follows the same pattern: save, handle error with rollback, log rollback failures. This duplication makes maintenance harder and risks inconsistent error handling if one instance is updated but others are not.

## What Changes

- Extract the save-with-rollback pattern into a reusable helper function
- Update all three command handlers to use the shared helper
- Improve consistency of error messages and logging

## Impact

- Affected specs: core-architecture (DRY principle)
- Affected code:
  - `cmd/canopy/repo.go` (refactor 3 locations)
  - Optionally `internal/config/registry.go` if helper belongs there
- Risk: Low - behavior-preserving refactoring
