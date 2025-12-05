# Change: Add Orphan Detection

## Why
When a canonical repository is removed with `canopy repo remove`, existing worktrees in workspaces become orphaned:
- The worktree directory still exists
- The workspace metadata still references the repo
- Git operations may fail with confusing errors

Similarly, manual filesystem operations can leave orphaned state. Detecting and warning about orphans improves reliability.

## What Changes
- Add `canopy check --orphans` to detect orphaned worktrees
- Add `canopy repo remove` warning when repo is in use
- Add TUI indicator for workspaces with orphaned repos
- Provide remediation suggestions (add repo back or remove from workspace)

## Impact
- **Affected specs**: `specs/repository-management/spec.md`, `specs/workspace-management/spec.md`
- **Affected code**:
  - `cmd/canopy/check.go` - Add orphan detection
  - `cmd/canopy/repo.go` - Add warning on remove
  - `internal/workspaces/service.go` - Add `DetectOrphans()` method
  - `internal/tui/` - Add orphan indicator
- **Risk**: Low - Diagnostic feature, non-destructive
