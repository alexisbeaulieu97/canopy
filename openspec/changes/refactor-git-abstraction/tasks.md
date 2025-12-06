# Tasks: Consolidate Git Abstraction

## Implementation Checklist

### 1. Audit Current CLI Usage
- [x] 1.1 Identify all `exec.Command("git", ...)` calls in `gitx/git.go`
- [x] 1.2 Document each operation's current behavior and edge cases
- [x] 1.3 Research go-git equivalents for each CLI operation

### 2. Migrate CreateWorktree
- [x] 2.1 Implement pure go-git worktree creation
- [x] 2.2 Handle branch creation within go-git
- [x] 2.3 Remove CLI fallback code
- [x] 2.4 Add unit tests for worktree creation

### 3. Migrate Clone Operation
- [x] 3.1 Verify existing `EnsureCanonical()` uses go-git correctly
- [x] 3.2 Migrate `Clone()` from CLI to `git.PlainClone()`
- [x] 3.3 Handle bare clone options properly
- [x] 3.4 Add unit tests for clone

### 4. Migrate Fetch/Pull/Push
- [x] 4.1 Implement `Fetch()` using `repo.Fetch()` from go-git
- [x] 4.2 Implement `Pull()` using worktree `Pull()` method
- [x] 4.3 Implement `Push()` using `repo.Push()` with refspecs
- [x] 4.4 Handle authentication for remote operations
- [x] 4.5 Add unit tests for each operation

### 5. Migrate Checkout and Utilities
- [x] 5.1 Implement `Checkout()` using go-git worktree checkout
- [x] 5.2 Migrate `aheadBehindCounts()` to pure go-git rev walking
- [x] 5.3 Evaluate `RunCommand()` - may need to keep for user escape hatch
- [x] 5.4 Add unit tests

### 6. Consistent Error Handling
- [x] 6.1 Ensure all operations wrap errors with `cerrors.WrapGitError()`
- [x] 6.2 Map go-git errors to appropriate CanopyError types
- [x] 6.3 Add tests verifying error types are correct

### 7. Integration Testing
- [x] 7.1 Run full test suite to verify no regressions
- [x] 7.2 Test with real repositories for edge cases
- [x] 7.3 Document any go-git limitations discovered

## Notes

### go-git Limitations Discovered
- Authentication: go-git supports SSH keys and basic auth, but for local file:// URLs no auth is needed
- `RunCommand()` kept as escape hatch for user-initiated git commands that aren't covered by go-git API
- Push with `--set-upstream` emulated by setting branch config after push
