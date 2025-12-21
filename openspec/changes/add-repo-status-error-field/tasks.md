## 1. Domain Model Update
- [ ] 1.1 Add `Error` field to `domain.RepoStatus` struct
- [ ] 1.2 Add `StatusError` type for typed status errors (timeout, fetch failed, etc.)

## 2. Status Retrieval
- [ ] 2.1 Update `GetStatus` to populate Error field instead of Branch for errors
- [ ] 2.2 Update `GetWorkspaceStatusBatch` for parallel status with error handling
- [ ] 2.3 Remove "timeout" and "error" encoding from Branch field

## 3. CLI Output
- [ ] 3.1 Update `cmd/canopy/status.go` to display Error field
- [ ] 3.2 Update `cmd/canopy/workspace_list.go` status display
- [ ] 3.3 Update `formatRepoStatusIndicator` helper

## 4. TUI Update
- [ ] 4.1 Update `internal/tui/view.go` to render error states
- [ ] 4.2 Update status badges for error display

## 5. Testing
- [ ] 5.1 Add unit tests for RepoStatus with error field
- [ ] 5.2 Update integration tests for status output
- [ ] 5.3 Verify JSON output format

## 6. Documentation
- [ ] 6.1 Update docs/usage.md with new status output format
- [ ] 6.2 Update docs/error-codes.md if needed
