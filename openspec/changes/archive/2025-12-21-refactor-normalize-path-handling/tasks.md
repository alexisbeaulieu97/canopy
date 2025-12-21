## 1. Service Layer Path Normalization
- [x] 1.1 Replace sprintf paths in `internal/workspaces/service.go`
- [x] 1.2 Replace sprintf paths in `internal/workspaces/git_service.go`

## 2. Hooks Path Normalization
- [x] 2.1 Replace sprintf paths in `internal/hooks/executor.go`

## 3. Verification
- [x] 3.1 Run tests to verify no regressions
- [x] 3.2 Verify no remaining `%s/%s` patterns for path construction

