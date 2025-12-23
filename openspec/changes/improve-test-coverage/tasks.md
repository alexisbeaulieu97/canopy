## 1. TUI Testing
- [x] 1.1 Add tests for `update.go` message handling
- [x] 1.2 Add tests for `model.go` state management
- [x] 1.3 Add tests for `commands.go` workspace loading
- [x] 1.4 Refactor `handleListKeyWithState` to reduce complexity

## 2. Workspace Operations Testing
- [x] 2.1 Add unit tests for `create.go`
- [x] 2.2 Add unit tests for `close.go`
- [x] 2.3 Add unit tests for `sync.go`
- [x] 2.4 Add unit tests for `rename.go`
- [x] 2.5 Add unit tests for `restore.go`

## 3. Storage Layer Testing
- [x] 3.1 Add tests for ListWorkspaces edge cases
- [x] 3.2 Add tests for CloseWorkspace with different options
- [x] 3.3 Add tests for migration scenarios

## 4. Test Infrastructure
- [x] 4.1 Create shared test fixtures for common scenarios
- [x] 4.2 Add test helpers for workspace creation
- [x] 4.3 Document testing patterns in CONTRIBUTING.md

## 5. Code Quality
- [x] 5.1 Address `//nolint:gocyclo` in update.go
- [x] 5.2 Extract message handlers to reduce complexity
- [x] 5.3 Verify test coverage improvements
