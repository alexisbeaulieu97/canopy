# Tasks: Add Transactional Operation Hardening

## 1. Operation Helper
- [x] 1.1 Create `internal/workspaces/operation.go` with `Operation` type
- [x] 1.2 Implement `AddStep(action, rollback)` for tracking steps
- [x] 1.3 Implement `Execute()` that runs rollbacks on failure
- [x] 1.4 Add logging for rollback actions

## 2. CreateWorkspace Hardening
- [x] 2.1 Reorder: create worktrees before saving metadata
- [x] 2.2 Add rollback: remove worktrees if metadata save fails
- [x] 2.3 Add rollback: cleanup workspace directory on any failure
- [x] 2.4 Add validation: check workspace ID doesn't exist before any work

## 3. AddRepoToWorkspace Hardening
- [x] 3.1 Add rollback: remove worktree if metadata update fails
- [x] 3.2 Add validation: check repo doesn't already exist in workspace

## 4. RestoreWorkspace Hardening
- [x] 4.1 Reorder: verify closed entry before any restoration work
- [x] 4.2 Add rollback: restore closed entry if recreation fails
- [x] 4.3 Only delete closed entry after successful restoration

## 5. Canonical Repo Add Hardening
- [x] 5.1 Add rollback: remove cloned repo if registry update fails
- [x] 5.2 Add validation: check repo doesn't already exist before clone

## 6. Testing
- [x] 6.1 Add unit tests for Operation helper
- [x] 6.2 Add integration tests simulating failures at each step
- [x] 6.3 Verify rollback behavior cleans up correctly
