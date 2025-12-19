# Tasks: Add Workspace-Level Concurrency Control

## 1. Lock Manager Implementation
- [x] 1.1 Create `internal/workspaces/lock.go` with `LockManager` type
- [x] 1.2 Implement file-based locking using `os.OpenFile` with exclusive flag
- [x] 1.3 Implement lock acquisition with configurable timeout
- [x] 1.4 Implement stale lock detection (check file mtime)
- [x] 1.5 Implement lock release with cleanup

## 2. Service Integration
- [x] 2.1 Add lock acquisition to `CreateWorkspace`
- [x] 2.2 Add lock acquisition to `CloseWorkspace` and `CloseWorkspaceKeepMetadata`
- [x] 2.3 Add lock acquisition to `RenameWorkspace`
- [x] 2.4 Add lock acquisition to `RestoreWorkspace`
- [x] 2.5 Add lock acquisition to `AddRepoToWorkspace` and `RemoveRepoFromWorkspace`
- [x] 2.6 Add lock acquisition to `SyncWorkspace`

## 3. Configuration
- [x] 3.1 Add `lock_timeout` configuration option (default: 30s)
- [x] 3.2 Add `lock_stale_threshold` configuration option (default: 5m)

## 4. Observability
- [x] 4.1 Add lock status to workspace metadata (optional)
- [x] 4.2 Add `--show-locks` flag to `workspace list` command
- [x] 4.3 Log lock acquisition/release at debug level

## 5. Testing
- [x] 5.1 Add unit tests for lock manager
- [x] 5.2 Add integration tests for concurrent operations
- [x] 5.3 Test stale lock detection and cleanup
