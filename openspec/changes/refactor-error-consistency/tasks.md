# Tasks: Standardize Error Handling

## Implementation Checklist

### Phase 1: Audit Current Errors
- [ ] Search for `fmt.Errorf` in `cmd/canopy/`
- [ ] Search for `fmt.Errorf` in `internal/`
- [ ] Document each untyped error and its appropriate type
- [ ] Identify missing error codes

### Phase 2: Add Missing Error Types
- [ ] Add `ErrNotInWorkspace` for context-based commands
- [ ] Add `ErrCommandFailed` for generic command failures
- [ ] Add `ErrInvalidArgument` for input validation
- [ ] Add `ErrOperationCancelled` for user cancellation
- [ ] Add constructors for each new type

### Phase 3: Update CLI Commands
- [ ] `cmd/canopy/status.go` - Replace all `fmt.Errorf`
- [ ] `cmd/canopy/check.go` - Replace all `fmt.Errorf`
- [ ] `cmd/canopy/workspace.go` - Audit and replace
- [ ] `cmd/canopy/repo.go` - Audit and replace
- [ ] `cmd/canopy/init.go` - Audit and replace

### Phase 4: Update Internal Packages
- [ ] `internal/workspaces/service.go` - Audit and replace
- [ ] `internal/gitx/git.go` - Ensure WrapGitError used
- [ ] `internal/config/config.go` - Use NewConfigInvalid
- [ ] `internal/workspace/workspace.go` - Audit and replace

### Phase 5: CLI Exit Code Mapping
- [ ] Create error-to-exit-code mapping in `cmd/canopy/errors.go`
- [ ] Update `main.go` to use mapped exit codes
- [ ] Document exit codes in help text

### Phase 6: Testing
- [ ] Add tests verifying error types are returned
- [ ] Add tests verifying exit codes
- [ ] Run full test suite
