# Tasks: Add Workspace Rename Command

## Implementation Checklist

### Phase 1: Service Layer
- [ ] Add `RenameWorkspace(oldID, newID string, renameBranch bool) error` to service
- [ ] Validate oldID exists
- [ ] Validate newID doesn't exist
- [ ] Validate newID is a valid workspace name

### Phase 2: Storage Layer
- [ ] Add `Rename(oldDir, newDir string) error` to `WorkspaceStorage` interface
- [ ] Implement in `workspace/workspace.go`:
  - [ ] Rename directory
  - [ ] Update metadata with new ID
  - [ ] Handle errors and rollback

### Phase 3: Branch Handling
- [ ] If workspace branch matches old ID, optionally rename branch
- [ ] Add `--rename-branch` flag (default true if branch == oldID)
- [ ] Use `git branch -m` via gitEngine

### Phase 4: CLI Command
- [ ] Add `workspaceRenameCmd` to `cmd/canopy/workspace.go`
- [ ] Parse arguments: `canopy workspace rename <OLD> <NEW>`
- [ ] Add `--rename-branch` flag
- [ ] Add `--force` flag to overwrite if new exists
- [ ] Print success message

### Phase 5: Closed Workspace Handling
- [ ] Decide: should rename also work on closed workspaces?
- [ ] If yes, update closed workspace directory and metadata
- [ ] If no, return error for closed workspaces

### Phase 6: Testing
- [ ] Add unit test for `RenameWorkspace()`
- [ ] Test rename with branch rename
- [ ] Test rename conflict detection
- [ ] Add integration test
