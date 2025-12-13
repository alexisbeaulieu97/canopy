# Tasks: Add Workspace Rename Command

## Implementation Checklist

### 1. Service Layer
- [x] 1.1 Add `RenameWorkspace(oldID, newID string, renameBranch bool) error` to service
- [x] 1.2 Validate oldID exists
- [x] 1.3 Validate newID doesn't exist
- [ ] 1.4 Validate newID is a valid workspace name

### 2. Storage Layer
- [x] 2.1 Add `Rename(oldDir, newDir string) error` to `WorkspaceStorage` interface
- [x] 2.2 Implement in `workspace/workspace.go`: rename directory
- [x] 2.3 Update metadata with new ID
- [x] 2.4 Handle errors and rollback

### 3. Branch Handling
- [x] 3.1 If workspace branch matches old ID, optionally rename branch
- [x] 3.2 Add `--rename-branch` flag (default true if branch == oldID)
- [x] 3.3 Use `git branch -m` via gitEngine

### 4. CLI Command
- [x] 4.1 Add `workspaceRenameCmd` to `cmd/canopy/workspace.go`
- [x] 4.2 Parse arguments: `canopy workspace rename <OLD> <NEW>`
- [x] 4.3 Add `--rename-branch` flag
- [ ] 4.4 Add `--force` flag to overwrite if new exists
- [x] 4.5 Print success message

### 5. Closed Workspace Handling
- [ ] 5.1 Return error when attempting to rename closed workspaces
- [ ] 5.2 Error message: "cannot rename closed workspace; reopen first with 'workspace open'"
- [ ] 5.3 Add test for closed workspace rename rejection

### 6. Testing
- [ ] 6.1 Add unit test for `RenameWorkspace()`
- [ ] 6.2 Test rename with branch rename
- [ ] 6.3 Test rename conflict detection
- [ ] 6.4 Add integration test
