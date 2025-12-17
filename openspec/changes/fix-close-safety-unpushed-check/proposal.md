# Change: Fix Close Safety to Check Unpushed Commits

## Why
The `project.md` explicitly states: "Safe Deletion: `workspace close` must verify no unpushed/uncommitted changes before deletion."

However, the current implementation in `ensureWorkspaceClean` only checks for dirty (uncommitted) changes, ignoring the unpushed commits return value from `gitEngine.Status()`. This violates the documented constraint and risks data loss when users close workspaces with committed but unpushed work.

Current code at `internal/workspaces/service.go:1116`:
```go
isDirty, _, _, _, err := s.gitEngine.Status(context.Background(), worktreePath)
// unpushed is ignored with _
if isDirty {
    return cerrors.NewRepoNotClean(repo.Name, action)
}
```

## What Changes
- Update `ensureWorkspaceClean` to also check for unpushed commits
- Add new error type for unpushed commits (distinct from dirty/uncommitted)
- Update `PreviewCloseWorkspace` to show unpushed status in preview
- Add `--force` flag behavior documentation (bypasses all safety checks)

## Impact
- Affected specs: `workspace-management`
- Affected code:
  - `internal/workspaces/service.go` - ensureWorkspaceClean function
  - `internal/errors/errors.go` - new error type if needed
  - `cmd/canopy/workspace.go` - error message handling
- **Breaking change**: Users who previously could close workspaces with unpushed commits will now be blocked (safety improvement)
