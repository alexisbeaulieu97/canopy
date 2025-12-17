# Change: Fix Close Safety to Check Unpushed Commits

## Why
The `project.md` constraint "Safe Deletion" requires verifying no unpushed/uncommitted changes before deletion, but `ensureWorkspaceClean` only checks dirty state, ignoring unpushed commits and risking data loss.

## What Changes
- **BREAKING** Update `ensureWorkspaceClean` to also check for unpushed commits
- Add new error type for unpushed commits (distinct from dirty/uncommitted)
- Update `PreviewCloseWorkspace` to show unpushed status in preview
- Add `--force` flag behavior documentation (bypasses all safety checks)

## Impact
- Affected specs: `workspace-management`
- Affected code:
  - `internal/workspaces/service.go` - ensureWorkspaceClean function
  - `internal/errors/errors.go` - new error type if needed
  - `cmd/canopy/workspace.go` - error message handling
