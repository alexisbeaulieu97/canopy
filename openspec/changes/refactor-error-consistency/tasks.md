# Tasks: Standardize Error Handling

## Implementation Checklist

### 1. Audit Current Errors
- [ ] 1.1 Search for `fmt.Errorf` in `cmd/canopy/`
- [ ] 1.2 Search for `fmt.Errorf` in `internal/`
- [ ] 1.3 Document each untyped error and its appropriate type
- [ ] 1.4 Identify missing error codes

### 2. Add Missing Error Types
- [ ] 2.1 Add `ErrNotInWorkspace` for context-based commands
- [ ] 2.2 Add `ErrCommandFailed` for generic command failures
- [ ] 2.3 Add `ErrInvalidArgument` for input validation
- [ ] 2.4 Add `ErrOperationCancelled` for user cancellation
- [ ] 2.5 Add constructors for each new type

### 3. Update CLI Commands
- [ ] 3.1 `cmd/canopy/status.go` - Replace all `fmt.Errorf`
- [ ] 3.2 `cmd/canopy/check.go` - Replace all `fmt.Errorf`
- [ ] 3.3 `cmd/canopy/workspace.go` - Audit and replace
- [ ] 3.4 `cmd/canopy/repo.go` - Audit and replace
- [ ] 3.5 `cmd/canopy/init.go` - Audit and replace

### 4. Update Internal Packages
- [ ] 4.1 `internal/workspaces/service.go` - Audit and replace
- [ ] 4.2 `internal/gitx/git.go` - Ensure WrapGitError used
- [ ] 4.3 `internal/config/config.go` - Use NewConfigInvalid
- [ ] 4.4 `internal/workspace/workspace.go` - Audit and replace

### 5. CLI Exit Code Mapping
- [ ] 5.1 Create error-to-exit-code mapping in `cmd/canopy/errors.go`
- [ ] 5.2 Update `main.go` to use mapped exit codes
- [ ] 5.3 Document exit codes in help text

### 6. Testing
- [ ] 6.1 Add tests verifying error types are returned
- [ ] 6.2 Add tests verifying exit codes
- [ ] 6.3 Run full test suite
