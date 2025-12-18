# Tasks: Add Transactional Operation Hardening

## 1. Operation Helper
- [ ] 1.1 Create `internal/workspaces/operation.go` with `Operation` type
- [ ] 1.2 Implement `AddStep(action, rollback)` for tracking steps
- [ ] 1.3 Implement `Execute()` that runs rollbacks on failure
- [ ] 1.4 Add logging for rollback actions

## 2. CreateWorkspace Hardening
- [ ] 2.1 Reorder: create worktrees before saving metadata
- [ ] 2.2 Add rollback: remove worktrees if metadata save fails
- [ ] 2.3 Add rollback: cleanup workspace directory on any failure
- [ ] 2.4 Add validation: check workspace ID doesn't exist before any work

## 3. AddRepoToWorkspace Hardening
- [ ] 3.1 Add rollback: remove worktree if metadata update fails
- [ ] 3.2 Add validation: check repo doesn't already exist in workspace

## 4. RestoreWorkspace Hardening
- [ ] 4.1 Reorder: verify closed entry before any restoration work
- [ ] 4.2 Add rollback: restore closed entry if recreation fails
- [ ] 4.3 Only delete closed entry after successful restoration

## 5. Canonical Repo Add Hardening
- [ ] 5.1 Add rollback: remove cloned repo if registry update fails
- [ ] 5.2 Add validation: check repo doesn't already exist before clone

## 6. Testing
- [ ] 6.1 Add unit tests for Operation helper
- [ ] 6.2 Add integration tests simulating failures at each step
- [ ] 6.3 Verify rollback behavior cleans up correctly
