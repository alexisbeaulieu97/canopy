## 1. Output Infrastructure

- [ ] 1.1 Create `internal/output/table.go` with box-drawn table renderer
- [ ] 1.2 Create `internal/output/progress.go` with spinner and progress bar
- [ ] 1.3 Create `internal/output/errors.go` with styled error box renderer
- [ ] 1.4 Update `internal/output/colors.go` with expanded semantic palette
- [ ] 1.5 Create `internal/output/icons.go` with status icons (with NO_COLOR support)

## 2. Workspace List Command

- [ ] 2.1 Refactor `workspace list` to use table output
- [ ] 2.2 Add status column with colored indicators
- [ ] 2.3 Add summary footer (total count, disk usage)
- [ ] 2.4 Ensure `--json` output remains unchanged

## 3. Workspace View Command

- [ ] 3.1 Create sectioned output with headers
- [ ] 3.2 Add metadata section with key-value formatting
- [ ] 3.3 Add repository table with status columns
- [ ] 3.4 Add warning section for orphaned worktrees

## 4. Sync/Push Commands

- [ ] 4.1 Add progress spinner during operation
- [ ] 4.2 Create per-repo result lines with status icons
- [ ] 4.3 Add summary section with success/failure counts
- [ ] 4.4 Format errors with context and suggestions

## 5. Error Handling

- [ ] 5.1 Create error box component with border
- [ ] 5.2 Add "Did you mean?" suggestions for typos
- [ ] 5.3 Add actionable hints for common errors
- [ ] 5.4 Ensure errors go to stderr with proper exit codes

## 6. Testing & Polish

- [ ] 6.1 Test output with NO_COLOR=1 (no colors)
- [ ] 6.2 Test output piped to file (no colors, no special chars)
- [ ] 6.3 Verify table alignment with long workspace names
- [ ] 6.4 Update integration tests for new output format
