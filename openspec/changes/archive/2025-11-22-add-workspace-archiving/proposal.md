# Change: Add Workspace Closing

## Why
Currently, `canopy workspace close` permanently deletes workspace directories and metadata. Users who want to reference old work or restore a workspace later have no option. Closing provides a middle ground: remove active worktrees (freeing disk space) while preserving metadata and history for future reference or restoration.

## What Changes
- Add `canopy workspace close <ID> --archive` command option
- Create `~/.canopy/closed/` directory (configurable `closed_root`) for closed workspace metadata
- Move workspace metadata to closed directory (preserve workspace.yaml)
- Remove worktrees but keep metadata for potential restoration
- Add `canopy workspace reopen <ID>` command to recreate from closed entry
- Add `canopy workspace list --closed` flag to view closed workspaces
- Update `canopy workspace close` to prompt "Keep workspace instead? [Y/n]"

## Impact
- Affected specs: `specs/workspace-management/spec.md`
- Affected code:
  - `internal/workspaces/service.go` - Add closing/reopen methods
  - `internal/workspace/workspace.go` - Add closed directory handling
  - `cmd/canopy/workspace.go` - Add close, reopen, and list --closed commands
  - `internal/config/config.go` - Add closed_root configuration option
