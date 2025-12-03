```markdown
# Implementation Tasks

## 1. Service Layer
- [ ] 1.1 Add `PullWorkspace(workspaceID string, rebase bool) error` to service
- [ ] 1.2 Iterate through all repos and call gitEngine.Pull()
- [ ] 1.3 Add `--rebase` support to gitx.Pull()
- [ ] 1.4 Return aggregated errors or first error based on continue-on-error

## 2. CLI Command
- [ ] 2.1 Create `workspacePullCmd` cobra command
- [ ] 2.2 Add `--rebase` flag (default: false)
- [ ] 2.3 Add `--continue-on-error` flag (default: false)
- [ ] 2.4 Show per-repo success/failure output
- [ ] 2.5 Exit with non-zero if any repo fails

## 3. TUI Integration
- [ ] 3.1 Add `l` key handler in handleListKey()
- [ ] 3.2 Add confirmation prompt similar to push
- [ ] 3.3 Show spinner during pull operation
- [ ] 3.4 Add `l` to help keys

## 4. Testing
- [ ] 4.1 Unit test for PullWorkspace service method
- [ ] 4.2 Manual test CLI command
- [ ] 4.3 Manual test TUI shortcut
```
