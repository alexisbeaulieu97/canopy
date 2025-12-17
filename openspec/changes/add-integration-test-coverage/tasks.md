## 1. Test Infrastructure

- [x] 1.1 Create `test/integration/helpers_test.go` with common test utilities
- [x] 1.2 Add workspace creation helper that handles full setup
- [x] 1.3 Add repo creation helper with configurable initial state
- [x] 1.4 Add cleanup utilities for reliable test isolation

## 2. Workspace Lifecycle Tests

- [x] 2.1 Add `TestWorkspaceRestoreFlow` - close with metadata, then restore
- [x] 2.2 Add `TestWorkspaceRestoreForceOverwrite` - restore over existing workspace
- [x] 2.3 Add `TestWorkspaceRename` - rename workspace and verify branches
- [x] 2.4 Add `TestWorkspaceRenameWithBranch` - rename including branch rename

## 3. Repository Management Tests

- [x] 3.1 Add `TestAddRepoToExistingWorkspace` - add repo after creation
- [x] 3.2 Add `TestRemoveRepoFromWorkspace` - remove repo from workspace
- [x] 3.3 Add `TestRepoStatusInWorkspace` - verify status reporting

## 4. Branch Operation Tests

- [x] 4.1 Add `TestBranchSwitchInWorkspace` - switch all repos to existing branch
- [x] 4.2 Add `TestBranchCreateInWorkspace` - create and switch to new branch
- [x] 4.3 Add `TestBranchSwitchPartialFailure` - handle mixed success/failure

## 5. Orphan Detection Tests

- [x] 5.1 Add `TestOrphanDetection` - detect orphaned worktrees
- [x] 5.2 Add `TestOrphanCleanup` - prune stale worktree references

## 6. Error Scenario Tests

- [x] 6.1 Add `TestCreateWorkspaceDirtyRepo` - dirty repo blocks close
- [x] 6.2 Add `TestCreateWorkspaceInvalidConfig` - invalid config handling
- [x] 6.3 Add `TestWorkspaceNotFound` - error messages for missing workspaces
