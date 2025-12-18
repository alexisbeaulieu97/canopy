# Tasks: Add Workspace-Level Concurrency Control

## 1. Lock Manager Implementation
- [ ] 1.1 Create `internal/workspaces/lock.go` with `LockManager` type
- [ ] 1.2 Implement file-based locking using `os.OpenFile` with exclusive flag
- [ ] 1.3 Implement lock acquisition with configurable timeout
- [ ] 1.4 Implement stale lock detection (check file mtime)
- [ ] 1.5 Implement lock release with cleanup

## 2. Service Integration
- [ ] 2.1 Add lock acquisition to `CreateWorkspace`
- [ ] 2.2 Add lock acquisition to `CloseWorkspace` and `CloseWorkspaceKeepMetadata`
- [ ] 2.3 Add lock acquisition to `RenameWorkspace`
- [ ] 2.4 Add lock acquisition to `RestoreWorkspace`
- [ ] 2.5 Add lock acquisition to `AddRepoToWorkspace` and `RemoveRepoFromWorkspace`
- [ ] 2.6 Add lock acquisition to `SyncWorkspace`

## 3. Configuration
- [ ] 3.1 Add `lock_timeout` configuration option (default: 30s)
- [ ] 3.2 Add `lock_stale_threshold` configuration option (default: 5m)

## 4. Observability
- [ ] 4.1 Add lock status to workspace metadata (optional)
- [ ] 4.2 Add `--show-locks` flag to `workspace list` command
- [ ] 4.3 Log lock acquisition/release at debug level

## 5. Testing
- [ ] 5.1 Add unit tests for lock manager
- [ ] 5.2 Add integration tests for concurrent operations
- [ ] 5.3 Test stale lock detection and cleanup
