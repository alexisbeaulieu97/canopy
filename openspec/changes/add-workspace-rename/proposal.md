# Change: Add Workspace Rename Command

## Why
Users sometimes need to rename workspaces:
- Typo in original name
- Ticket ID changed
- Better naming convention adopted
- Reusing workspace for different work

Currently, renaming requires:
1. Create new workspace with desired name
2. Manually copy/move repos
3. Delete old workspace

This is error-prone and tedious. A rename command simplifies the operation.

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
