# Tasks: Add Orphan Detection

## Implementation Checklist

### 1. Define Orphan Types
- [ ] 1.1 Define `OrphanedWorktree` struct:
  ```go
  type OrphanedWorktree struct {
      WorkspaceID   string
      RepoName      string
      WorktreePath  string
      Reason        string  // "canonical_missing", "directory_missing", etc.
  }
  ```

### 2. Detection Logic
- [ ] 2.1 Add `DetectOrphans() ([]OrphanedWorktree, error)` to service
- [ ] 2.2 For each workspace, check each repo in metadata exists in canonical repos
- [ ] 2.3 Check worktree directory exists
- [ ] 2.4 Check worktree is valid git directory
- [ ] 2.5 Return list of orphans with reasons

### 3. Check Command Integration
- [ ] 3.1 Add `--orphans` flag to `canopy check`
- [ ] 3.2 Display orphan summary:
  ```text
  Found 2 orphaned worktrees:
    - PROJ-123/backend: canonical repo 'backend' not found
    - PROJ-456/frontend: worktree directory missing
  ```
- [ ] 3.3 Suggest remediation commands

### 4. Repo Remove Warning
- [ ] 4.1 Before removing canonical repo, check workspace usage
- [ ] 4.2 Warn if repo is used by any workspace
- [ ] 4.3 Require `--force` to proceed
- [ ] 4.4 Enhance messaging (already partially implemented)

### 5. TUI Integration
- [ ] 5.1 Add orphan status to workspace item
- [ ] 5.2 Show warning icon for workspaces with orphans
- [ ] 5.3 Show details in workspace detail view

### 6. Testing
- [ ] 6.1 Add test for detecting missing canonical repo
- [ ] 6.2 Add test for detecting missing worktree directory
- [ ] 6.3 Add integration test for `canopy check --orphans`
