# Tasks: Unify Error Handling with Typed Errors

## Implementation Checklist

### 1. Add New Error Types
- [ ] 1.1 Add `ErrConfigValidation` error code to `internal/errors/errors.go`
- [ ] 1.2 Add `ErrPathInvalid` error code for path-related errors
- [ ] 1.3 Add `ErrPathNotDirectory` error code
- [ ] 1.4 Add `NewConfigValidation(field, detail string)` constructor
- [ ] 1.5 Add `NewPathInvalid(path, reason string)` constructor
- [ ] 1.6 Add `NewPathNotDirectory(path string)` constructor
- [ ] 1.7 Add sentinel errors for new types
- [ ] 1.8 Add tests for new error constructors

### 2. Convert workspace/workspace.go (27 instances)
- [ ] 2.1 Convert path existence check errors to `NewPathInvalid`
- [ ] 2.2 Convert directory creation errors to `NewIOFailed`
- [ ] 2.3 Convert metadata read/write errors to `NewWorkspaceMetadataError`
- [ ] 2.4 Convert YAML marshal/unmarshal errors to `NewIOFailed`
- [ ] 2.5 Convert directory removal errors to `NewIOFailed`
- [ ] 2.6 Update tests to use `errors.Is()` where appropriate

### 3. Convert config/config.go (26 instances)
- [ ] 3.1 Convert config file read errors to `NewIOFailed`
- [ ] 3.2 Convert validation errors to `NewConfigValidation`
- [ ] 3.3 Convert path validation errors to `NewPathInvalid` or `NewPathNotDirectory`
- [ ] 3.4 Convert regex compile errors to `NewConfigValidation`
- [ ] 3.5 Update tests to use `errors.Is()` where appropriate

### 4. Convert config/repo_registry.go (10 instances)
- [ ] 4.1 Convert registry file read/write errors to `NewRegistryError`
- [ ] 4.2 Convert alias validation errors to `NewInvalidArgument`
- [ ] 4.3 Convert duplicate alias errors to `NewRegistryError`
- [ ] 4.4 Update tests to use `errors.Is()` where appropriate

### 5. Convert tui/commands.go (3 instances)
- [ ] 5.1 Identify and convert remaining `fmt.Errorf` calls
- [ ] 5.2 Use appropriate existing or new error types

### 6. Convert workspaces/service.go (2 instances)
- [ ] 6.1 Identify and convert remaining `fmt.Errorf` calls
- [ ] 6.2 Use appropriate existing error types

### 7. Final Validation
- [ ] 7.1 Run `grep -r "fmt.Errorf" internal/` to verify zero instances (excluding tests)
- [ ] 7.2 Run full test suite
- [ ] 7.3 Verify error messages are user-friendly
- [ ] 7.4 Verify `errors.Is()` works for all new error paths

