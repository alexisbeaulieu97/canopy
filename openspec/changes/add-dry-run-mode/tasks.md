# Tasks: Add Dry Run Mode

## Implementation Checklist

### 1. Workspace Close Dry Run
- [ ] 1.1 Add `--dry-run` flag to `workspaceCloseCmd`
- [ ] 1.2 When dry-run, load workspace metadata
- [ ] 1.3 Calculate what would be deleted (directories, files)
- [ ] 1.4 Show summary without deleting:
  ```text
  [DRY RUN] Would close workspace: PROJ-123
    - Remove directory: ~/workspaces/PROJ-123
    - Repos affected: backend, frontend
    - Total size: 1.2 GB
  ```
- [ ] 1.5 Support `--json` output for dry run

### 2. Repo Remove Dry Run
- [ ] 2.1 Add `--dry-run` flag to `repoRemoveCmd`
- [ ] 2.2 When dry-run, check if repo exists
- [ ] 2.3 List workspaces using this repo
- [ ] 2.4 Show summary:
  ```text
  [DRY RUN] Would remove repository: backend
    - Remove directory: ~/projects/backend
    - Used by workspaces: PROJ-123, PROJ-456 (will become orphaned)
    - Size: 500 MB
  ```

### 3. Service Layer Support
- [ ] 3.1 Choose approach: `DryRun bool` parameter, `Preview*()` methods, or action plan return
- [ ] 3.2 Implement consistently across affected methods
- [ ] 3.3 Document chosen approach in code comments

### 4. Output Formatting
- [ ] 4.1 Create consistent dry-run output format
- [ ] 4.2 Use clear `[DRY RUN]` prefix
- [ ] 4.3 Color output yellow/orange for warning
- [ ] 4.4 Ensure `--json` works with dry-run

### 5. Testing
- [ ] 5.1 Test dry run shows correct preview
- [ ] 5.2 Test dry run doesn't modify filesystem
- [ ] 5.3 Test dry run with `--json`
- [ ] 5.4 Test dry run with `--force` (should still be dry)
