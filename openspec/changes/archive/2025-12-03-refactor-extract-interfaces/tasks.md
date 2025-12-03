# Implementation Tasks

## 1. Define Interfaces
- [x] 1.1 Create `internal/ports/` directory
- [x] 1.2 Define `GitOperations` interface in `ports/git.go`
  - EnsureCanonical, CreateWorktree, Status, Push, Pull, Fetch, Checkout, List
- [x] 1.3 Define `WorkspaceStorage` interface in `ports/storage.go`
  - Create, Save, Delete, List, Close, ListClosed, LatestClosed
- [x] 1.4 Define `ConfigProvider` interface in `ports/config.go`
  - GetReposForWorkspace, Validate, registry access

## 2. Update Implementations
- [x] 2.1 Ensure `gitx.GitEngine` implements `GitOperations`
- [x] 2.2 Ensure `workspace.Engine` implements `WorkspaceStorage`
- [x] 2.3 Add compile-time interface checks (`var _ GitOperations = (*GitEngine)(nil)`)

## 3. Update Service Layer
- [x] 3.1 Change Service fields from concrete to interface types
- [x] 3.2 Update NewService constructor signature
- [x] 3.3 Update all call sites (app.go, tests) - backward compatible, no changes needed

## 4. Create Mocks
- [x] 4.1 Create `internal/mocks/git.go` with MockGitOperations
- [x] 4.2 Create `internal/mocks/storage.go` with MockWorkspaceStorage
- [x] 4.3 Create `internal/mocks/config.go` with MockConfigProvider
- [x] 4.4 Add helper constructors for test setup

## 5. Update Tests
- [x] 5.1 Add service_mock_test.go demonstrating mock usage
- [x] 5.2 Add tests for error scenarios using mocks
- [x] 5.3 Keep integration tests with real implementations (unchanged)

## 6. Documentation
- [x] 6.1 Document interface contracts in godoc (inline comments in ports/*.go)
- [x] 6.2 Update architecture docs to reflect hexagonal pattern (core-architecture/spec.md)
