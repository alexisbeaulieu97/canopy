# Change: Normalize Path Handling with filepath.Join

## Why
Many paths are built using `fmt.Sprintf("%s/%s", ...)` string concatenation instead of `filepath.Join`. This is:
1. **Brittle on Windows**: Hardcoded forward slashes don't work correctly on Windows
2. **Risk of duplicate slashes**: If a path already ends with `/`, concatenation creates `//`
3. **Inconsistent**: Some code uses `filepath.Join`, some uses sprintf

Found instances in:
- `internal/workspaces/service.go` (~6 instances)
- `internal/workspaces/git_service.go` (~4 instances)
- `internal/hooks/executor.go` (~2 instances)

## What Changes
- Replace all `fmt.Sprintf("%s/%s", ...)` path constructions with `filepath.Join`
- Optionally extract common path patterns into helper functions for DRY
- No functional changes - purely a refactor for correctness and consistency

## Impact
- Affected specs: `core-architecture` (code quality)
- Affected code:
  - `internal/workspaces/service.go`
  - `internal/workspaces/git_service.go`
  - `internal/hooks/executor.go`
- **Risk**: Very Low - No functional changes, just using proper path APIs

