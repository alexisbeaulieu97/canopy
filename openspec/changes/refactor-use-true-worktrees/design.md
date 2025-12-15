## Context
The current implementation clones from canonical repos, creating full copies. This violates the documented design where worktrees should share objects with the canonical repository.

## Goals / Non-Goals
- **Goals**:
  - Reduce disk usage by sharing git objects between canonical and worktrees
  - Enable push/pull to work with real remotes
  - Maintain backward compatibility with existing workspace metadata
- **Non-Goals**:
  - Automatic migration of existing workspaces
  - Supporting multiple remotes per worktree

## Decisions

### Decision 1: Use git CLI for worktree operations
go-git's worktree API is documented as limited (cannot create worktree for non-existent branch). We will use the existing `RunCommand` escape hatch for worktree operations:
- `git worktree add -b <branch> <path>`
- `git worktree remove <path>`
- `git worktree prune`

**Alternatives considered**:
- Pure go-git: Rejected due to API limitations
- git2go (libgit2 bindings): Rejected due to CGO dependency

### Decision 2: Store upstream URL during canonical clone
During `EnsureCanonical`, store the original URL in the canonical repo's git config:
```
[canopy]
    upstreamUrl = https://github.com/owner/repo.git
```
This allows worktrees to be configured with the correct remote.

**Alternatives considered**:
- Separate metadata file: Rejected; adds another file to manage and risks desync with repo state
- Rely on existing git remotes: Rejected; bare repos don't have a working origin remote by default
- Embed URL in branch names: Rejected; branch names have character/length limits and would break existing workflows

### Decision 3: Configure worktree remotes post-creation
After `git worktree add`, update the worktree's origin remote:
```bash
git -C <worktree> remote set-url origin <upstream-url>
```

**Alternatives considered**:
- Pre-configure remotes at worktree creation: Rejected; `git worktree add` doesn't support custom remote configuration
- Centralized remote mapping service: Rejected; over-engineered for this use case, adds external dependency
- Keep origin pointing to canonical, fetch from upstream: Rejected; breaks push workflows and confuses users

## Risks / Trade-offs
- **Risk**: Shelling out to git CLI reduces portability
  - **Mitigation**: git is required for worktree functionality anyway; document as system requirement
- **Risk**: Existing workspaces won't automatically benefit
  - **Mitigation**: Document that users should close and recreate workspaces; existing ones continue to work

## Migration Plan
1. Deploy new version
2. Existing workspaces continue working (using full clones)
3. New workspaces use true worktrees
4. Users can optionally close and recreate existing workspaces for disk savings
5. No automatic migration to avoid data loss risk

## Open Questions
- Should we add a `canopy workspace migrate` command to convert existing clones to worktrees?
- Should we warn users about the disk usage difference in existing vs new workspaces?

