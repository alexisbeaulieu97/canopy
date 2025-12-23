# Change: Add Workspace Health Check Command

## Why
Users need comprehensive health checks for their workspaces beyond orphan detection. Issues like corrupted worktrees, stale branches, large repos, or metadata inconsistencies should be proactively detected and optionally fixed.

## What Changes
- Add `canopy doctor workspace [ID]` subcommand
- Check worktree integrity, git config validity, remote connectivity
- Provide health scores and remediation suggestions
- Support `--fix` flag for auto-fixable issues

## Impact
- Affected specs: cli, workspace-management
- Affected code:
  - `cmd/canopy/doctor.go` (new subcommand)
  - `internal/workspaces/` (health check logic)
  - `internal/domain/domain.go` (health check types)
- No breaking changes
