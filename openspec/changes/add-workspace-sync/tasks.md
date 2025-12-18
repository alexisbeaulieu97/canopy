## 1. Service Layer
- [ ] 1.1 Add `SyncWorkspace(ctx, id string, opts SyncOptions) (*SyncResult, error)` to `internal/workspaces/service.go`
- [ ] 1.2 Define `SyncOptions` struct with timeout configuration
- [ ] 1.3 Define `SyncResult` and `RepoSyncStatus` structs in `internal/domain/domain.go`
- [ ] 1.4 Implement parallel fetch/pull with per-repo timeout
- [ ] 1.5 Add unit tests for sync logic

## 2. CLI Command
- [ ] 2.1 Add `workspaceSyncCmd` in `cmd/canopy/workspace.go`
- [ ] 2.2 Parse `--timeout` flag (default: 60s per repo)
- [ ] 2.3 Format output as summary table
- [ ] 2.4 Support `--json` output for scripting
- [ ] 2.5 Add integration test for sync command

## 3. Documentation
- [ ] 3.1 Update README.md command reference
- [ ] 3.2 Update docs/usage.md with sync examples
