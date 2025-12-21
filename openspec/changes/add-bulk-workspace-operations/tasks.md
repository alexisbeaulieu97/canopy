## 1. Service Layer
- [ ] 1.1 Add `ListWorkspacesMatching(pattern string)` to service
- [ ] 1.2 Add `CloseWorkspacesMatching(ctx, pattern string, opts)` method
- [ ] 1.3 Add `SyncWorkspacesMatching(ctx, pattern string, opts)` method
- [ ] 1.4 Implement regex pattern validation

## 2. CLI Commands
- [ ] 2.1 Add `--pattern` flag to `workspace close`
- [ ] 2.2 Add `--pattern` flag to `workspace sync`
- [ ] 2.3 Add `--pattern` flag to `workspace branch`
- [ ] 2.4 Add `--all` flag as shorthand for `--pattern ".*"`

## 3. User Experience
- [ ] 3.1 Add confirmation dialog for bulk close operations
- [ ] 3.2 Add preview mode showing affected workspaces before execution
- [ ] 3.3 Add progress output for long-running bulk operations
- [ ] 3.4 Add summary output (success/failed counts)

## 4. Testing
- [ ] 4.1 Add unit tests for pattern matching
- [ ] 4.2 Add integration tests for bulk close
- [ ] 4.3 Add integration tests for bulk sync

## 5. Documentation
- [ ] 5.1 Update docs/usage.md with bulk operation examples
- [ ] 5.2 Update README command reference
