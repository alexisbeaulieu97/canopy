## 1. Progress Infrastructure
- [ ] 1.1 Create `internal/output/progress.go` with progress bar implementation
- [ ] 1.2 Support TTY detection (disable for non-interactive)
- [ ] 1.3 Support `--no-progress` flag for scripting

## 2. Workspace Sync Progress
- [ ] 2.1 Add progress to single workspace sync
- [ ] 2.2 Add progress to bulk workspace sync (--pattern)
- [ ] 2.3 Show per-repo status during sync

## 3. Workspace Create Progress
- [ ] 3.1 Add progress for multi-repo workspace creation
- [ ] 3.2 Show clone/worktree creation status

## 4. Workspace Close Progress
- [ ] 4.1 Add progress for bulk close (--pattern)
- [ ] 4.2 Show per-workspace status during close

## 5. Testing
- [ ] 5.1 Add tests for progress output
- [ ] 5.2 Verify non-interactive fallback works
- [ ] 5.3 Test cancellation behavior
