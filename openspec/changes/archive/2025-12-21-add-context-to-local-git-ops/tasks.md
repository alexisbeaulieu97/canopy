## 1. Interface Updates
- [x] 1.1 Update `ports.GitOperations.CreateWorktree` to accept context
- [x] 1.2 Update `ports.GitOperations.Status` to accept context
- [x] 1.3 Update `ports.GitOperations.Checkout` to accept context
- [x] 1.4 Update `ports.GitOperations.List` to accept context

## 2. Implementation Updates
- [x] 2.1 Update `gitx.GitEngine.CreateWorktree` implementation
- [x] 2.2 Update `gitx.GitEngine.Status` implementation
- [x] 2.3 Update `gitx.GitEngine.Checkout` implementation
- [x] 2.4 Update `gitx.GitEngine.List` implementation
- [x] 2.5 Add context deadline checks in long-running loops

## 3. Mock Updates
- [x] 3.1 Update `mocks.GitOps` to match new interface

## 4. Caller Updates
- [x] 4.1 Update calls in `internal/workspaces/service.go`
- [x] 4.2 Update calls in `internal/workspaces/git_service.go`
- [x] 4.3 Update calls in `internal/workspaces/orphan_service.go`
- [x] 4.4 Verify CLI commands pass cmd.Context()

## 5. Testing
- [x] 5.1 Add tests for context cancellation in git operations
- [x] 5.2 Update existing tests to provide context

