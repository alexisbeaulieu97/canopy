````markdown
# Tasks: Add Orphan Detection

## Implementation Checklist

### 1. Define Orphan Types
- [x] 1.1 Define `OrphanedWorktree` struct:
  ```go
  type OrphanedWorktree struct {
      WorkspaceID   string
      RepoName      string
      WorktreePath  string
      Reason        string  // "canonical_missing", "directory_missing", etc.
  }
  ```

### 2. Detection Logic
- [x] 2.1 Add `DetectOrphans() ([]OrphanedWorktree, error)` to service
- [x] 2.2 For each workspace, check each repo in metadata exists in canonical repos
- [x] 2.3 Check worktree directory exists
- [x] 2.4 Check worktree is valid git directory
- [x] 2.5 Return list of orphans with reasons

### 3. Check Command Integration
- [x] 3.1 Add `--orphans` flag to `canopy check`
- [x] 3.2 Display orphan summary:
  ```text
  Found 2 orphaned worktrees:
    - PROJ-123/backend: canonical repo 'backend' not found
    - PROJ-456/frontend: worktree directory missing
  ```
- [x] 3.3 Suggest remediation commands

### 4. Repo Remove Warning
- [x] 4.1 Before removing canonical repo, check workspace usage
- [x] 4.2 Warn if repo is used by any workspace
- [x] 4.3 Require `--force` to proceed
- [x] 4.4 Enhance messaging (already partially implemented)

### 5. TUI Integration
- [x] 5.1 Add orphan status to workspace item
- [x] 5.2 Show warning icon for workspaces with orphans
- [x] 5.3 Show details in workspace detail view

### 6. Testing
- [x] 6.1 Add test for detecting missing canonical repo
- [x] 6.2 Add test for detecting missing worktree directory
- [x] 6.3 Add integration test for `canopy check --orphans`

````
