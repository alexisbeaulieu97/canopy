# Change: Add Workspace Rename Command

## Why
Users need to rename workspaces (due to typos, changed requirements, or naming conventions) but the current workflow—create new workspace, move repos, delete old—is tedious and error-prone.

## What Changes
- Add `canopy workspace rename <OLD> <NEW>` command
- Rename workspace directory
- Update metadata file with new ID
- Handle branch name updates if branch matches old workspace ID
- Validate new name doesn't conflict with existing workspace

## Impact
- **Affected specs**: `specs/workspace-management/spec.md`
- **Affected code**:
  - `cmd/canopy/workspace.go` - Add rename subcommand
  - `internal/workspaces/service.go` - Add `RenameWorkspace()` method
  - `internal/workspace/workspace.go` - Add `Rename()` method
- **Risk**: Low - Filesystem rename with metadata update
