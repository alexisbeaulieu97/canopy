# Tasks: Add Orphan Detection

## Implementation Checklist

### Phase 1: Define Orphan Types
- [ ] Define `OrphanedWorktree` struct:
  ```go
  type OrphanedWorktree struct {
      WorkspaceID   string
      RepoName      string
      WorktreePath  string
      Reason        string  // "canonical_missing", "directory_missing", etc.
  }
  ```

### Phase 2: Detection Logic
- [ ] Add `DetectOrphans() ([]OrphanedWorktree, error)` to service
- [ ] For each workspace:
  - [ ] Check each repo in metadata exists in canonical repos
  - [ ] Check worktree directory exists
  - [ ] Check worktree is valid git directory
- [ ] Return list of orphans with reasons

### Phase 3: Check Command Integration
- [ ] Add `--orphans` flag to `canopy check`
- [ ] Display orphan summary:
  ```
  Found 2 orphaned worktrees:
    - PROJ-123/backend: canonical repo 'backend' not found
    - PROJ-456/frontend: worktree directory missing
  ```
- [ ] Suggest remediation commands

### Phase 4: Repo Remove Warning
- [ ] Before removing canonical repo, check workspace usage
- [ ] Warn if repo is used by any workspace
- [ ] Require `--force` to proceed
- [ ] (Already partially implemented, enhance messaging)

### Phase 5: TUI Integration
- [ ] Add orphan status to workspace item
- [ ] Show warning icon for workspaces with orphans
- [ ] Show details in workspace detail view

### Phase 6: Testing
- [ ] Add test for detecting missing canonical repo
- [ ] Add test for detecting missing worktree directory
- [ ] Add integration test for `canopy check --orphans`
