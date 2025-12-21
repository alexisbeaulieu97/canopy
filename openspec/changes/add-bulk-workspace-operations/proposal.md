# Change: Add bulk workspace operations with pattern matching

## Why
Users managing many workspaces (20+) need efficient batch operations. Currently, operations are per-workspace only, requiring multiple commands for common tasks like syncing or closing multiple workspaces.

## What Changes
- Add `--pattern` flag to `workspace close`, `workspace sync`, and `workspace branch` commands
- Implement pattern-based workspace filtering using regex
- Add confirmation dialog for bulk destructive operations
- Add `--all` flag as shorthand for matching all workspaces

## Impact
- Affected specs: workspace-management, cli
- Affected code:
  - `internal/workspaces/service.go` - Add pattern-matching list methods
  - `cmd/canopy/workspace_close.go` - Add --pattern flag
  - `cmd/canopy/workspace_sync.go` - Add --pattern flag
  - `cmd/canopy/workspace_branch.go` - Add --pattern flag
