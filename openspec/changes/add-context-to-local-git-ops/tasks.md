## 1. Interface Updates
- [ ] 1.1 Update `ports.GitOperations.CreateWorktree` to accept context
- [ ] 1.2 Update `ports.GitOperations.Status` to accept context
- [ ] 1.3 Update `ports.GitOperations.Checkout` to accept context
- [ ] 1.4 Update `ports.GitOperations.List` to accept context

## 2. Implementation Updates
- [ ] 2.1 Update `gitx.GitEngine.CreateWorktree` implementation
- [ ] 2.2 Update `gitx.GitEngine.Status` implementation
- [ ] 2.3 Update `gitx.GitEngine.Checkout` implementation
- [ ] 2.4 Update `gitx.GitEngine.List` implementation
- [ ] 2.5 Add context deadline checks in long-running loops

## 3. Mock Updates
- [ ] 3.1 Update `mocks.GitOps` to match new interface

## 4. Caller Updates
- [ ] 4.1 Update calls in `internal/workspaces/service.go`
- [ ] 4.2 Update calls in `internal/workspaces/git_service.go`
- [ ] 4.3 Update calls in `internal/workspaces/orphan_service.go`
- [ ] 4.4 Verify CLI commands pass cmd.Context()

## 5. Testing
- [ ] 5.1 Add tests for context cancellation in git operations
- [ ] 5.2 Update existing tests to provide context

