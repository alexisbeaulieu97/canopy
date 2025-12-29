## 1. Progress Infrastructure
- [x] 1.1 Create `internal/output/progress.go` with progress bar implementation
- [x] 1.2 Support TTY detection (disable for non-interactive)
- [x] 1.3 Support `--no-progress` flag for scripting

## 2. Workspace Sync Progress
- [ ] 2.1 Add progress to single workspace sync
- [x] 2.2 Add progress to bulk workspace sync (--pattern)
- [ ] 2.3 Show per-repo status during sync

## 3. Workspace Create Progress
- [ ] 3.1 Add progress for multi-repo workspace creation
- [ ] 3.2 Show clone/worktree creation status

## 4. Workspace Close Progress
- [x] 4.1 Add progress for bulk close (--pattern)
- [x] 4.2 Show per-workspace status during close

## 5. Testing
- [x] 5.1 Add tests for progress output
- [x] 5.2 Verify non-interactive fallback works
- [x] 5.3 Test cancellation behavior
