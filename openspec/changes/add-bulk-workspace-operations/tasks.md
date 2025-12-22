## 1. Service Layer
- [x] 1.1 Add `ListWorkspacesMatching(pattern string)` to service
- [x] 1.2 Add `CloseWorkspacesMatching(ctx, pattern string, opts)` method
- [x] 1.3 Add `SyncWorkspacesMatching(ctx, pattern string, opts)` method
- [x] 1.4 Implement regex pattern validation

## 2. CLI Commands
- [x] 2.1 Add `--pattern` flag to `workspace close`
- [x] 2.2 Add `--pattern` flag to `workspace sync`
- [x] 2.3 Add `--pattern` flag to `workspace branch`
- [x] 2.4 Add `--all` flag as shorthand for `--pattern ".*"`

## 3. User Experience
- [x] 3.1 Add confirmation dialog for bulk close operations
- [x] 3.2 Add preview mode showing affected workspaces before execution
- [x] 3.3 Add progress output for long-running bulk operations
- [x] 3.4 Add summary output (success/failed counts)

## 4. Testing
- [x] 4.1 Add unit tests for pattern matching
- [x] 4.2 Add integration tests for bulk close
- [x] 4.3 Add integration tests for bulk sync

## 5. Documentation
- [x] 5.1 Update docs/usage.md with bulk operation examples
- [x] 5.2 Update README command reference
