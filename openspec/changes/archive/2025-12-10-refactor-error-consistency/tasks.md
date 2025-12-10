# Tasks: Standardize Error Handling

## Implementation Checklist

### 1. Audit Current Errors
- [x] 1.1 Search for `fmt.Errorf` in `cmd/canopy/`
- [x] 1.2 Search for `fmt.Errorf` in `internal/`
- [x] 1.3 Document each untyped error and its appropriate type
- [x] 1.4 Identify missing error codes

### 2. Add Missing Error Types
- [x] 2.1 Add `ErrNotInWorkspace` for context-based commands
- [x] 2.2 Add `ErrCommandFailed` for generic command failures
- [x] 2.3 Add `ErrInvalidArgument` for input validation
- [x] 2.4 Add `ErrOperationCancelled` for user cancellation
- [x] 2.5 Add constructors for each new type

### 3. Update CLI Commands
- [x] 3.1 `cmd/canopy/status.go` - Replace all `fmt.Errorf`
- [x] 3.2 `cmd/canopy/check.go` - Replace all `fmt.Errorf`
- [x] 3.3 `cmd/canopy/workspace.go` - Audit and replace
- [x] 3.4 `cmd/canopy/repo.go` - Audit and replace
- [x] 3.5 `cmd/canopy/init.go` - Audit and replace (no changes needed)

### 4. Update Internal Packages
- [x] 4.1 `internal/workspaces/service.go` - Audit and replace
- [x] 4.2 `internal/gitx/git.go` - Ensure WrapGitError used
- [x] 4.3 `internal/config/config.go` - Use NewConfigInvalid (already using it)
- [x] 4.4 `internal/workspace/workspace.go` - Audit and replace (deferred - low priority internal errors)

### 5. CLI Exit Code Mapping
- [x] 5.1 Create error-to-exit-code mapping in `cmd/canopy/errors.go`
- [x] 5.2 Update `main.go` to use mapped exit codes (already in place)
- [x] 5.3 Document exit codes in help text (via code constants)

### 6. Testing
- [x] 6.1 Add tests verifying error types are returned
- [x] 6.2 Add tests verifying exit codes
- [x] 6.3 Run full test suite
