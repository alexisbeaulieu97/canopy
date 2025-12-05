# Tasks: Consolidate Path Building

## Implementation Checklist

### Phase 1: Create Paths Package
- [ ] Create `internal/paths/paths.go`
- [ ] Define `Roots` struct holding base paths:
  ```go
  type Roots struct {
      Projects   string
      Workspaces string
      Closed     string
  }
  ```

### Phase 2: Define Path Builders
- [ ] Add `WorkspacePath(roots Roots, workspaceDir string) string`
- [ ] Add `WorkspaceMetadataPath(roots Roots, workspaceDir string) string`
- [ ] Add `CanonicalRepoPath(roots Roots, repoName string) string`
- [ ] Add `WorktreePath(roots Roots, workspaceDir, repoName string) string`
- [ ] Add `ClosedWorkspacePath(roots Roots, workspaceDir, timestamp string) string`

### Phase 3: Add Validation
- [ ] Validate path components don't contain path separators
- [ ] Validate paths are within expected roots (prevent traversal)
- [ ] Return errors for invalid inputs

### Phase 4: Update Consumers
- [ ] Update `internal/workspaces/service.go`:
  - [ ] Replace `fmt.Sprintf("%s/%s/%s", ...)` with `paths.WorktreePath()`
  - [ ] Replace other path constructions
- [ ] Update `internal/workspace/workspace.go`:
  - [ ] Use `paths.WorkspacePath()` and `paths.WorkspaceMetadataPath()`
- [ ] Update `internal/gitx/git.go`:
  - [ ] Use `paths.CanonicalRepoPath()`

### Phase 5: Testing
- [ ] Add unit tests for path builders
- [ ] Test path traversal prevention
- [ ] Run full test suite to verify no regressions
