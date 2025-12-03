```markdown
# Implementation Tasks

## 1. Define Interfaces
- [ ] 1.1 Create `internal/ports/` directory
- [ ] 1.2 Define `GitOperations` interface in `ports/git.go`
  - EnsureCanonical, CreateWorktree, Status, Push, Pull, Fetch, Checkout, List
- [ ] 1.3 Define `WorkspaceStorage` interface in `ports/storage.go`
  - Create, Save, Delete, List, Close, ListClosed, LatestClosed
- [ ] 1.4 Define `ConfigProvider` interface in `ports/config.go`
  - GetReposForWorkspace, Validate, registry access

## 2. Update Implementations
- [ ] 2.1 Ensure `gitx.GitEngine` implements `GitOperations`
- [ ] 2.2 Ensure `workspace.Engine` implements `WorkspaceStorage`
- [ ] 2.3 Add compile-time interface checks (`var _ GitOperations = (*GitEngine)(nil)`)

## 3. Update Service Layer
- [ ] 3.1 Change Service fields from concrete to interface types
- [ ] 3.2 Update NewService constructor signature
- [ ] 3.3 Update all call sites (app.go, tests)

## 4. Create Mocks
- [ ] 4.1 Create `internal/mocks/git.go` with MockGitOperations
- [ ] 4.2 Create `internal/mocks/storage.go` with MockWorkspaceStorage
- [ ] 4.3 Add helper methods for test setup

## 5. Update Tests
- [ ] 5.1 Refactor service_test.go to use mocks
- [ ] 5.2 Add tests for error scenarios using mocks
- [ ] 5.3 Keep integration tests with real implementations

## 6. Documentation
- [ ] 6.1 Document interface contracts in godoc
- [ ] 6.2 Update architecture docs to reflect hexagonal pattern
```
