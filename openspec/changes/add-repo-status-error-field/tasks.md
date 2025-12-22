## 1. Domain Model Update
- [x] 1.1 Add `Error` field to `domain.RepoStatus` struct
- [x] 1.2 Add `StatusError` type for typed status errors (timeout, fetch failed, etc.)

## 2. Status Retrieval
- [x] 2.1 Update `GetStatus` to populate Error field instead of Branch for errors
- [x] 2.2 Update `GetWorkspaceStatusBatch` for parallel status with error handling
- [x] 2.3 Remove "timeout" and "error" encoding from Branch field

## 3. CLI Output
- [x] 3.1 Update `cmd/canopy/status.go` to display Error field
- [x] 3.2 Update `cmd/canopy/workspace_list.go` status display
- [x] 3.3 Update `formatRepoStatusIndicator` helper

## 4. TUI Update
- [x] 4.1 Update `internal/tui/view.go` to render error states
- [x] 4.2 Update status badges for error display

## 5. Testing
- [x] 5.1 Add unit tests for RepoStatus with error field
- [x] 5.2 Update integration tests for status output
- [x] 5.3 Verify JSON output format

## 6. Documentation
- [x] 6.1 Update docs/usage.md with new status output format
- [x] 6.2 Update docs/error-codes.md if needed (no changes required)
