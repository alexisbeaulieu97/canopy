```markdown
# Change: Add Workspace Rename Command

## Why
Users sometimes need to rename a workspace—for example, when a ticket ID changes, when reorganizing work, or when a workspace name was chosen poorly. Currently, there's no safe way to rename a workspace; users must manually rename directories and update metadata. A `canopy workspace rename` command would handle all the bookkeeping safely.

## What Changes
- Add `canopy workspace rename <OLD-ID> <NEW-ID>` command
- Rename workspace directory in workspaces_root
- Update workspace.yaml metadata with new ID
- Update branch names in all repos (optional, with `--rename-branches`)
- Validate new ID doesn't conflict with existing workspace
- Update any closed workspace references (optional)

## Impact
- Affected specs: `specs/workspace-management/spec.md`
- Affected code:
  - `cmd/canopy/workspace.go` - Add new `rename` subcommand
  - `internal/workspaces/service.go` - Add `RenameWorkspace()` method
  - `internal/workspace/workspace.go` - Add `Rename()` method to engine
  - `internal/gitx/git.go` - Add `RenameBranch()` if implementing branch rename
```
