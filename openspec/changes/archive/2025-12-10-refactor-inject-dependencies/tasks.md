# Tasks: Refactor Dependency Injection

## Implementation Checklist

### 1. Define Functional Options
- [x] 1.1 Define `AppOption` type as functional option
- [x] 1.2 Create `WithGitOperations(ports.GitOperations)` option
- [x] 1.3 Create `WithWorkspaceStorage(ports.WorkspaceStorage)` option
- [x] 1.4 Create `WithConfigProvider(ports.ConfigProvider)` option
- [x] 1.5 Create `WithLogger(*logging.Logger)` option

### 2. Refactor App Struct
- [x] 2.1 Refactor `App` struct to store interfaces instead of concrete types
- [x] 2.2 Update `app.New()` to accept variadic `AppOption` parameters
- [x] 2.3 Implement default instantiation when options not provided
- [x] 2.4 Ensure backward compatibility with zero options

### 3. Testing
- [x] 3.1 Add `app_test.go` tests demonstrating mock injection
- [x] 3.2 Verify all CLI commands still work with `go test ./...`
- [x] 3.3 Run integration tests to confirm no regressions
