```markdown
# Implementation Tasks

## 1. Define Error Types
- [ ] 1.1 Create `internal/errors/errors.go`
- [ ] 1.2 Define `ErrWorkspaceNotFound` with workspace ID field
- [ ] 1.3 Define `ErrWorkspaceExists` with workspace ID field
- [ ] 1.4 Define `ErrRepoNotFound` with repo name field
- [ ] 1.5 Define `ErrUncleanWorkspace` with workspace ID and dirty repos
- [ ] 1.6 Define `ErrInvalidConfig` with validation details
- [ ] 1.7 Implement `Error()` method for each type
- [ ] 1.8 Implement `Is()` method for sentinel error matching

## 2. Update Workspaces Service
- [ ] 2.1 Replace `fmt.Errorf("workspace %s not found")` with ErrWorkspaceNotFound
- [ ] 2.2 Replace repo not found errors with ErrRepoNotFound
- [ ] 2.3 Replace unclean workspace errors with ErrUncleanWorkspace
- [ ] 2.4 Ensure all errors are properly wrapped with context

## 3. Update Workspace Engine
- [ ] 3.1 Return ErrWorkspaceExists from Create when directory exists
- [ ] 3.2 Return typed errors from List, Save, Delete

## 4. Update CLI Error Handling
- [ ] 4.1 Create error handler helper in cmd/canopy/
- [ ] 4.2 Map error types to user-friendly messages
- [ ] 4.3 Map error types to exit codes
- [ ] 4.4 Include error codes in JSON output

## 5. Testing
- [ ] 5.1 Unit tests for error type Is() matching
- [ ] 5.2 Test error messages are user-friendly
- [ ] 5.3 Test CLI exit codes for different errors
```
