# Change: Use True Git Worktrees Instead of Full Clones

## Why
The current `CreateWorktree` implementation in `internal/gitx/git.go` performs a full `git clone` from the canonical bare repository instead of creating a true git worktree. This causes three critical issues:
1. **Disk usage explosion**: Each workspace repo gets a complete copy of all git objects (~2-10x expected disk usage)
2. **Broken remote**: The cloned repo's `origin` points to the canonical path instead of the real remote, so pushes never reach the upstream repository
3. **Spec violation**: The domain context defines worktrees as "a git worktree checked out to a branch" but the implementation creates full clones

## What Changes
- **BREAKING**: Replace `git.PlainClone` with proper worktree creation using `git worktree add` via CLI escape hatch (go-git's worktree API is limited per documented constraints)
- Configure worktree remotes to point to the original upstream URL instead of the canonical path
- Update canonical repository to properly track worktrees
- Add worktree pruning support for orphan cleanup
- Update workspace close to use `git worktree remove` for proper cleanup

**Implementation approach**:
1. Store original remote URL in canonical repo config during clone
2. Use `git worktree add -b <branch> <path>` for worktree creation
3. After creation, update worktree's `origin` remote to point to upstream URL
4. Push/pull operations will correctly use the real remote

## Impact
- Affected specs: `core-architecture`
- Affected code:
  - `internal/gitx/git.go:CreateWorktree` - Rewrite to use true worktrees
  - `internal/gitx/git.go:EnsureCanonical` - Store upstream URL in config
  - `internal/workspaces/service.go` - Update workspace deletion to use worktree remove
  - `internal/workspaces/orphan_service.go` - Add worktree prune integration
  - `cmd/canopy/workspace.go` - No changes needed (uses service layer)
- **Migration**: Existing workspaces will continue to work but won't benefit from disk savings until recreated

