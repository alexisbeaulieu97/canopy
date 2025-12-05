# Tasks: Consolidate Git Abstraction

## Implementation Checklist

### Phase 1: Document and Plan
- [ ] Add comment block in `gitx/git.go` explaining CLI vs library choice
- [ ] Update `openspec/project.md` to clarify the constraint exception
- [ ] Identify which operations use go-git vs CLI

### Phase 2: Extract GoGit Adapter
- [ ] Create `internal/gitx/gogit.go`
- [ ] Move `EnsureCanonical()` implementation
- [ ] Move `Status()` go-git parts
- [ ] Define `GoGitAdapter` struct

### Phase 3: Extract CLI Adapter
- [ ] Create `internal/gitx/cli.go`
- [ ] Move `CreateWorktree()`, `Clone()`, `Fetch()`, `Pull()`, `Push()`
- [ ] Move `Checkout()`, `RunCommand()`, `aheadBehindCounts()`
- [ ] Define `GitCLIAdapter` struct

### Phase 4: Create Facade
- [ ] Create `internal/gitx/facade.go`
- [ ] Define `GitEngine` as facade that delegates to appropriate adapter
- [ ] Add constructor that accepts adapter preference option
- [ ] Ensure backward compatibility with existing `New()` function

### Phase 5: Consistent Error Handling
- [ ] Ensure all CLI operations wrap errors with `cerrors.WrapGitError()`
- [ ] Ensure all go-git operations wrap errors consistently
- [ ] Add tests verifying error types

### Phase 6: Testing
- [ ] Add unit tests for `GoGitAdapter` (can mock go-git)
- [ ] Add integration tests for `GitCLIAdapter`
- [ ] Run full test suite to verify no regressions
