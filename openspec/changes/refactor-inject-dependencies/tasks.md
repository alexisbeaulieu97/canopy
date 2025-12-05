# Tasks: Refactor Dependency Injection

## Implementation Checklist

### 1. Define Functional Options
- [ ] 1.1 Define `AppOption` type as functional option
- [ ] 1.2 Create `WithGitOperations(ports.GitOperations)` option
- [ ] 1.3 Create `WithWorkspaceStorage(ports.WorkspaceStorage)` option
- [ ] 1.4 Create `WithConfigProvider(ports.ConfigProvider)` option
- [ ] 1.5 Create `WithLogger(*logging.Logger)` option

### 2. Refactor App Struct
- [ ] 2.1 Refactor `App` struct to store interfaces instead of concrete types
- [ ] 2.2 Update `app.New()` to accept variadic `AppOption` parameters
- [ ] 2.3 Implement default instantiation when options not provided
- [ ] 2.4 Ensure backward compatibility with zero options

### 3. Testing
- [ ] 3.1 Add `app_test.go` tests demonstrating mock injection
- [ ] 3.2 Verify all CLI commands still work with `go test ./...`
- [ ] 3.3 Run integration tests to confirm no regressions
