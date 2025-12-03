```markdown
# Implementation Tasks

## 1. Service Layer
- [ ] 1.1 Add `SyncWorkspace(workspaceID string, opts SyncOpts) error` to service
- [ ] 1.2 Create SyncOpts struct (FetchOnly, Rebase, ContinueOnError)
- [ ] 1.3 For each repo: fetch then pull (unless fetch-only)
- [ ] 1.4 Collect and report per-repo results

## 2. Git Engine
- [ ] 2.1 Add `FetchWorktree(path string) error` method for worktree fetch
- [ ] 2.2 Ensure Pull respects rebase flag

## 3. CLI Command
- [ ] 3.1 Create `workspaceSyncCmd` cobra command
- [ ] 3.2 Add `--fetch-only` flag
- [ ] 3.3 Add `--rebase` flag
- [ ] 3.4 Add `--continue-on-error` flag
- [ ] 3.5 Display per-repo progress and status

## 4. TUI Integration
- [ ] 4.1 Add `r` key handler for refresh/sync
- [ ] 4.2 Show spinner during sync
- [ ] 4.3 Refresh workspace status after sync
- [ ] 4.4 Add to help keys

## 5. Testing
- [ ] 5.1 Unit test for SyncWorkspace
- [ ] 5.2 Manual test CLI command
- [ ] 5.3 Manual test TUI shortcut
```
