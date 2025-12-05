# Tasks: Add Dry Run Mode

## Implementation Checklist

### Phase 1: Workspace Close Dry Run
- [ ] Add `--dry-run` flag to `workspaceCloseCmd`
- [ ] When dry-run:
  - [ ] Load workspace metadata
  - [ ] Calculate what would be deleted (directories, files)
  - [ ] Show summary without deleting:
    ```
    [DRY RUN] Would close workspace: PROJ-123
      - Remove directory: ~/workspaces/PROJ-123
      - Repos affected: backend, frontend
      - Total size: 1.2 GB
    ```
- [ ] Support `--json` output for dry run

### Phase 2: Repo Remove Dry Run
- [ ] Add `--dry-run` flag to `repoRemoveCmd`
- [ ] When dry-run:
  - [ ] Check if repo exists
  - [ ] List workspaces using this repo
  - [ ] Show summary:
    ```
    [DRY RUN] Would remove repository: backend
      - Remove directory: ~/projects/backend
      - Used by workspaces: PROJ-123, PROJ-456 (will become orphaned)
      - Size: 500 MB
    ```

### Phase 3: Service Layer Support
- [ ] Option 1: Add `DryRun bool` parameter to methods
- [ ] Option 2: Add separate `Preview*()` methods
- [ ] Option 3: Return action plan that can be executed or displayed
- [ ] Choose approach and implement consistently

### Phase 4: Output Formatting
- [ ] Create consistent dry-run output format
- [ ] Use clear `[DRY RUN]` prefix
- [ ] Color output yellow/orange for warning
- [ ] Ensure `--json` works with dry-run

### Phase 5: Testing
- [ ] Test dry run shows correct preview
- [ ] Test dry run doesn't modify filesystem
- [ ] Test dry run with `--json`
- [ ] Test dry run with `--force` (should still be dry)
