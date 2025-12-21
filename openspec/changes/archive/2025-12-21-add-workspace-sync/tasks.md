## 1. Service Layer
- [x] 1.1 Add `SyncWorkspace(ctx, id string, opts SyncOptions) (*SyncResult, error)` to `internal/workspaces/service.go`
- [x] 1.2 Define `SyncOptions` struct with timeout configuration
- [x] 1.3 Define `SyncResult` and `RepoSyncStatus` structs in `internal/domain/domain.go`
- [x] 1.4 Implement parallel fetch/pull with per-repo timeout
- [x] 1.5 Add unit tests for sync logic

## 2. CLI Command
- [x] 2.1 Add `workspaceSyncCmd` in `cmd/canopy/workspace.go`
- [x] 2.2 Parse `--timeout` flag (default: 60s per repo)
- [x] 2.3 Format output as summary table
- [x] 2.4 Support `--json` output for scripting
- [x] 2.5 Add integration test for sync command

## 3. Documentation
- [x] 3.1 Update README.md command reference
- [x] 3.2 Update docs/usage.md with sync examples
