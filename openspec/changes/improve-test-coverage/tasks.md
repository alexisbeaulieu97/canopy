## 1. TUI Testing
- [ ] 1.1 Add tests for `update.go` message handling
- [ ] 1.2 Add tests for `model.go` state management
- [ ] 1.3 Add tests for `commands.go` workspace loading
- [ ] 1.4 Refactor `handleListKeyWithState` to reduce complexity

## 2. Workspace Operations Testing
- [ ] 2.1 Add unit tests for `create.go`
- [ ] 2.2 Add unit tests for `close.go`
- [ ] 2.3 Add unit tests for `sync.go`
- [ ] 2.4 Add unit tests for `rename.go`
- [ ] 2.5 Add unit tests for `restore.go`

## 3. Storage Layer Testing
- [ ] 3.1 Add tests for ListWorkspaces edge cases
- [ ] 3.2 Add tests for CloseWorkspace with different options
- [ ] 3.3 Add tests for migration scenarios

## 4. Test Infrastructure
- [ ] 4.1 Create shared test fixtures for common scenarios
- [ ] 4.2 Add test helpers for workspace creation
- [ ] 4.3 Document testing patterns in CONTRIBUTING.md

## 5. Code Quality
- [ ] 5.1 Address `//nolint:gocyclo` in update.go
- [ ] 5.2 Extract message handlers to reduce complexity
- [ ] 5.3 Verify test coverage improvements
