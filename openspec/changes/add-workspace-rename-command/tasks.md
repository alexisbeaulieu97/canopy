# Implementation Tasks

## 1. Validation
- [ ] 1.1 Check old workspace exists
- [ ] 1.2 Check new ID doesn't conflict with existing workspace
- [ ] 1.3 Validate new ID format (no path traversal, valid characters)

## 2. Workspace Engine
- [ ] 2.1 Add `Rename(oldDir, newDir string) error` to workspace.Engine
- [ ] 2.2 Rename directory atomically
- [ ] 2.3 Update workspace.yaml with new ID

## 3. Service Layer
- [ ] 3.1 Add `RenameWorkspace(oldID, newID string, opts RenameOpts) error`
- [ ] 3.2 Create RenameOpts struct (RenameBranches bool)
- [ ] 3.3 Optionally rename branches in all repos

## 4. Git Engine (Optional)
- [ ] 4.1 Add `RenameBranch(path, oldBranch, newBranch string) error`
- [ ] 4.2 Handle remote tracking branch updates

## 5. CLI Command
- [ ] 5.1 Create `workspaceRenameCmd` cobra command
- [ ] 5.2 Require exactly 2 args (old, new)
- [ ] 5.3 Add `--rename-branches` flag (default: false)
- [ ] 5.4 Add `--force` flag to skip safety checks
- [ ] 5.5 Show confirmation before renaming

## 6. Testing
- [ ] 6.1 Unit test for RenameWorkspace
- [ ] 6.2 Test conflict detection
- [ ] 6.3 Test with branch rename
- [ ] 6.4 Test rollback on failure
