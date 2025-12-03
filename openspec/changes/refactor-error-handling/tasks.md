# Implementation Tasks

## 1. Define Error Types
- [x] 1.1 Create `internal/errors/errors.go`
- [x] 1.2 Define `ErrWorkspaceNotFound` with workspace ID field
- [x] 1.3 Define `ErrWorkspaceExists` with workspace ID field
- [x] 1.4 Define `ErrRepoNotFound` with repo name field
- [x] 1.5 Define `ErrUncleanWorkspace` with workspace ID and dirty repos
- [x] 1.6 Define `ErrInvalidConfig` with validation details
- [x] 1.7 Implement `Error()` method for each type
- [x] 1.8 Implement `Is()` method for sentinel error matching

## 2. Update Workspaces Service
- [x] 2.1 Replace `fmt.Errorf("workspace %s not found")` with ErrWorkspaceNotFound
- [x] 2.2 Replace repo not found errors with ErrRepoNotFound
- [x] 2.3 Replace unclean workspace errors with ErrUncleanWorkspace
- [x] 2.4 Ensure all errors are properly wrapped with context

## 3. Update Workspace Engine
- [x] 3.1 Return ErrWorkspaceExists from Create when directory exists
- [x] 3.2 Return typed errors from List, Save, Delete

## 4. Update CLI Error Handling
- [x] 4.1 Create error handler helper in cmd/canopy/
- [x] 4.2 Map error types to user-friendly messages
- [x] 4.3 Map error types to exit codes
- [x] 4.4 Include error codes in JSON output

## 5. Testing
- [x] 5.1 Unit tests for error type Is() matching
- [x] 5.2 Test error messages are user-friendly
- [x] 5.3 Test CLI exit codes for different errors
