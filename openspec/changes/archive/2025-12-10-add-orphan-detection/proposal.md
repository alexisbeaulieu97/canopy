# Change: Add Orphan Detection

## Why
Removing a canonical repository (or manual filesystem changes) can leave worktrees orphaned—stale directories and metadata causing confusing Git failures. Detecting and warning about orphans improves reliability and user experience.

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
