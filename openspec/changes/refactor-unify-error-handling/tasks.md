# Tasks: Unify Error Handling with Typed Errors

## Implementation Checklist

### 1. Add New Error Types
- [x] 1.1 Add `ErrConfigValidation` error code to `internal/errors/errors.go`
- [x] 1.2 Add `ErrPathInvalid` error code for path-related errors
- [x] 1.3 Add `ErrPathNotDirectory` error code
- [x] 1.4 Add `NewConfigValidation(field, detail string)` constructor
- [x] 1.5 Add `NewPathInvalid(path, reason string)` constructor
- [x] 1.6 Add `NewPathNotDirectory(path string)` constructor
- [x] 1.7 Add sentinel errors for new types
- [x] 1.8 Add tests for new error constructors

### 2. Convert workspace/workspace.go (27 instances)
- [x] 2.1 Convert path existence check errors to `NewPathInvalid`
- [x] 2.2 Convert directory creation errors to `NewIOFailed`
- [x] 2.3 Convert metadata read/write errors to `NewWorkspaceMetadataError`
- [x] 2.4 Convert YAML marshal/unmarshal errors to `NewIOFailed`
- [x] 2.5 Convert directory removal errors to `NewIOFailed`
- [x] 2.6 Update tests to use `errors.Is()` where appropriate

### 3. Convert config/config.go (26 instances)
- [x] 3.1 Convert config file read errors to `NewIOFailed`
- [x] 3.2 Convert validation errors to `NewConfigValidation`
- [x] 3.3 Convert path validation errors to `NewPathInvalid` or `NewPathNotDirectory`
- [x] 3.4 Convert regex compile errors to `NewConfigValidation`
- [x] 3.5 Update tests to use `errors.Is()` where appropriate

### 4. Convert config/repo_registry.go (10 instances)
- [x] 4.1 Convert registry file read/write errors to `NewRegistryError`
- [x] 4.2 Convert alias validation errors to `NewInvalidArgument`
- [x] 4.3 Convert duplicate alias errors to `NewRegistryError`
- [x] 4.4 Update tests to use `errors.Is()` where appropriate

### 5. Convert tui/commands.go (3 instances)
- [x] 5.1 Identify and convert remaining `fmt.Errorf` calls
- [x] 5.2 Use appropriate existing or new error types

### 6. Convert workspaces/service.go (2 instances)
- [x] 6.1 Identify and convert remaining `fmt.Errorf` calls
- [x] 6.2 Use appropriate existing error types

### 7. Final Validation
- [x] 7.1 Run `grep -r "fmt.Errorf" internal/` to verify zero instances (excluding tests)
- [x] 7.2 Run full test suite
- [x] 7.3 Verify error messages are user-friendly
- [x] 7.4 Verify `errors.Is()` works for all new error paths
