# Change: Consolidate Path Building

## Why
Path construction is scattered throughout the codebase using `fmt.Sprintf`:
```go
// In service.go
worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)
path := fmt.Sprintf("%s/%s", s.config.GetProjectsRoot(), name)

// In workspace.go  
path := filepath.Join(e.WorkspacesRoot, safeDir)
```

This inconsistency leads to:
- Mixed use of `fmt.Sprintf` with `/` and `filepath.Join`
- Potential bugs on Windows (though Go handles `/` well)
- No type safety for path components
- Repeated path patterns

## What Changes
- Create `internal/paths/paths.go` with type-safe path builders
- Define path types: `WorkspacePath`, `CanonicalRepoPath`, `WorktreePath`
- Provide constructors that validate and build paths correctly
- Replace scattered `fmt.Sprintf` and `filepath.Join` calls

## Impact
- **Affected specs**: None (internal utility)
- **Affected code**:
  - `internal/paths/paths.go` - New file
  - `internal/workspaces/service.go` - Use path builders
  - `internal/workspace/workspace.go` - Use path builders
  - `internal/gitx/git.go` - Use path builders
- **Risk**: Low - Utility refactoring
