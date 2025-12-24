## 1. Interface Updates
- [x] 1.1 Update WorkspaceFinder interface in `internal/workspaces/git_service.go` to accept context
- [x] 1.2 Update FindWorkspace implementation in service.go
- [x] 1.3 Update all callers in git_service.go to pass context

## 2. Sub-Service Updates
- [x] 2.1 Update OrphanService to pass context through
- [x] 2.2 Update ExportService to pass context through
- [x] 2.3 Update any other services using WorkspaceFinder

## 3. Testing
- [x] 3.1 Add tests for context cancellation behavior
- [x] 3.2 Verify existing tests still pass
- [x] 3.3 Test timeout propagation

## 4. Validation
- [x] 4.1 Run full test suite
- [ ] 4.2 Manual verification of Ctrl+C handling
