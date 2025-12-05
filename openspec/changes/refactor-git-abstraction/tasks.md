# Tasks: Consolidate Git Abstraction

## Implementation Checklist

### 1. Audit Current CLI Usage
- [ ] 1.1 Identify all `exec.Command("git", ...)` calls in `gitx/git.go`
- [ ] 1.2 Document each operation's current behavior and edge cases
- [ ] 1.3 Research go-git equivalents for each CLI operation

### 2. Migrate CreateWorktree
- [ ] 2.1 Implement pure go-git worktree creation
- [ ] 2.2 Handle branch creation within go-git
- [ ] 2.3 Remove CLI fallback code
- [ ] 2.4 Add unit tests for worktree creation

### 3. Migrate Clone Operation
- [ ] 3.1 Verify existing `EnsureCanonical()` uses go-git correctly
- [ ] 3.2 Migrate `Clone()` from CLI to `git.PlainClone()`
- [ ] 3.3 Handle bare clone options properly
- [ ] 3.4 Add unit tests for clone

### 4. Migrate Fetch/Pull/Push
- [ ] 4.1 Implement `Fetch()` using `repo.Fetch()` from go-git
- [ ] 4.2 Implement `Pull()` using worktree `Pull()` method
- [ ] 4.3 Implement `Push()` using `repo.Push()` with refspecs
- [ ] 4.4 Handle authentication for remote operations
- [ ] 4.5 Add unit tests for each operation

### 5. Migrate Checkout and Utilities
- [ ] 5.1 Implement `Checkout()` using go-git worktree checkout
- [ ] 5.2 Migrate `aheadBehindCounts()` to pure go-git rev walking
- [ ] 5.3 Evaluate `RunCommand()` - may need to keep for user escape hatch
- [ ] 5.4 Add unit tests

### 6. Consistent Error Handling
- [ ] 6.1 Ensure all operations wrap errors with `cerrors.WrapGitError()`
- [ ] 6.2 Map go-git errors to appropriate CanopyError types
- [ ] 6.3 Add tests verifying error types are correct

### 7. Integration Testing
- [ ] 7.1 Run full test suite to verify no regressions
- [ ] 7.2 Test with real repositories for edge cases
- [ ] 7.3 Document any go-git limitations discovered
