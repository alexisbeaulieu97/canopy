## 1. Canonical Repository Enhancement
- [ ] 1.1 Store upstream URL in canonical repo config during EnsureCanonical
- [ ] 1.2 Add helper to retrieve upstream URL from canonical config
- [ ] 1.3 Add tests for URL storage and retrieval

## 2. Worktree Creation Rewrite
- [ ] 2.1 Implement CreateWorktree using `git worktree add -b <branch> <path>`
- [ ] 2.2 Configure worktree remote to point to upstream URL after creation
- [ ] 2.3 Set up branch tracking for proper push/pull behavior
- [ ] 2.4 Add tests for worktree creation with remote verification
- [ ] 2.5 Update Status() to work correctly with worktrees

## 3. Worktree Cleanup
- [ ] 3.1 Add RemoveWorktree method using `git worktree remove`
- [ ] 3.2 Add PruneWorktrees method using `git worktree prune`
- [ ] 3.3 Update CloseWorkspace to use RemoveWorktree
- [ ] 3.4 Integrate worktree pruning into orphan remediation

## 4. Documentation and Migration
- [ ] 4.1 Update code documentation to reflect worktree behavior
- [ ] 4.2 Add migration notes for existing workspaces
- [ ] 4.3 Update architecture.md with worktree details

