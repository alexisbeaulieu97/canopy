## 1. Progress Infrastructure
- [x] 1.1 Create `internal/output/progress.go` with progress bar implementation
- [x] 1.2 Support TTY detection (disable for non-interactive)
- [x] 1.3 Support `--no-progress` flag for scripting

## 2. Workspace Sync Progress
- [x] 2.1 Add progress to bulk workspace sync (--pattern/--all)
- [x] 2.2 Show current workspace name during bulk sync

## 3. Workspace Close Progress
- [x] 3.1 Add progress for bulk close (--pattern/--all)
- [x] 3.2 Show per-workspace status during close

## 4. Testing
- [x] 4.1 Add tests for progress output
- [x] 4.2 Verify non-interactive fallback works
- [x] 4.3 Test cancellation behavior
