## 1. Test Infrastructure

- [ ] 1.1 Create `test/integration/helpers_test.go` with common test utilities
- [ ] 1.2 Add workspace creation helper that handles full setup
- [ ] 1.3 Add repo creation helper with configurable initial state
- [ ] 1.4 Add cleanup utilities for reliable test isolation

## 2. Workspace Lifecycle Tests

- [ ] 2.1 Add `TestWorkspaceRestoreFlow` - close with metadata, then restore
- [ ] 2.2 Add `TestWorkspaceRestoreForceOverwrite` - restore over existing workspace
- [ ] 2.3 Add `TestWorkspaceRename` - rename workspace and verify branches
- [ ] 2.4 Add `TestWorkspaceRenameWithBranch` - rename including branch rename

## 3. Repository Management Tests

- [ ] 3.1 Add `TestAddRepoToExistingWorkspace` - add repo after creation
- [ ] 3.2 Add `TestRemoveRepoFromWorkspace` - remove repo from workspace
- [ ] 3.3 Add `TestRepoStatusInWorkspace` - verify status reporting

## 4. Branch Operation Tests

- [ ] 4.1 Add `TestBranchSwitchInWorkspace` - switch all repos to existing branch
- [ ] 4.2 Add `TestBranchCreateInWorkspace` - create and switch to new branch
- [ ] 4.3 Add `TestBranchSwitchPartialFailure` - handle mixed success/failure

## 5. Orphan Detection Tests

- [ ] 5.1 Add `TestOrphanDetection` - detect orphaned worktrees
- [ ] 5.2 Add `TestOrphanCleanup` - prune stale worktree references

## 6. Error Scenario Tests

- [ ] 6.1 Add `TestCreateWorkspaceDirtyRepo` - dirty repo blocks close
- [ ] 6.2 Add `TestCreateWorkspaceInvalidConfig` - invalid config handling
- [ ] 6.3 Add `TestWorkspaceNotFound` - error messages for missing workspaces
