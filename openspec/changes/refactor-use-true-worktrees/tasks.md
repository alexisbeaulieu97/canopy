## 1. Canonical Repository Enhancement
- [x] 1.1 Store upstream URL in canonical repo config during EnsureCanonical
- [x] 1.2 Add helper to retrieve upstream URL from canonical config
- [x] 1.3 Add tests for URL storage and retrieval

## 2. Worktree Creation Rewrite
- [x] 2.1 Implement CreateWorktree using `git worktree add -b <branch> <path>`
- [x] 2.2 Configure worktree remote to point to upstream URL after creation
- [x] 2.3 Set up branch tracking for proper push/pull behavior
- [x] 2.4 Add tests for worktree creation with remote verification
- [x] 2.5 Update Status() to work correctly with worktrees

## 3. Worktree Cleanup
- [x] 3.1 Add RemoveWorktree method using `git worktree remove`
- [x] 3.2 Add PruneWorktrees method using `git worktree prune`
- [x] 3.3 Update CloseWorkspace to use RemoveWorktree
- [x] 3.4 Integrate worktree pruning into orphan remediation

## 4. Documentation and Migration
- [x] 4.1 Update code documentation to reflect worktree behavior
- [x] 4.2 Add migration notes for existing workspaces
- [x] 4.3 Update architecture.md with worktree details

