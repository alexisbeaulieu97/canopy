# Tasks: Refactor Dependency Injection

## Implementation Checklist

- [ ] Define `AppOption` type as functional option
- [ ] Create option functions:
  - [ ] `WithGitOperations(ports.GitOperations)`
  - [ ] `WithWorkspaceStorage(ports.WorkspaceStorage)`
  - [ ] `WithConfigProvider(ports.ConfigProvider)`
  - [ ] `WithLogger(*logging.Logger)`
- [ ] Refactor `App` struct to store interfaces instead of concrete types
- [ ] Update `app.New()` to:
  - [ ] Accept variadic `AppOption` parameters
  - [ ] Use defaults when options not provided
  - [ ] Maintain backward compatibility with zero options
- [ ] Add `app_test.go` tests demonstrating mock injection
- [ ] Verify all CLI commands still work with `go test ./...`
- [ ] Run integration tests to confirm no regressions
