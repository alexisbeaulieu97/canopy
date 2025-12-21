## 1. Implementation

- [x] 1.1 Update `ensureWorkspaceClean` to check unpushed commits in addition to dirty state
- [x] 1.2 Add new error type `ErrRepoHasUnpushedCommits` if distinct from `ErrRepoNotClean`
- [x] 1.3 Update `PreviewCloseWorkspace` output to show unpushed commit counts
- [x] 1.4 Update CLI error messages to clearly distinguish uncommitted vs unpushed

## 2. Testing

- [x] 2.1 Add unit test: close blocked when repo has unpushed commits
- [x] 2.2 Add unit test: close allowed when repo is clean (no dirty, no unpushed)
- [x] 2.3 Add unit test: --force bypasses unpushed check
- [x] 2.4 Add unit test: preview shows unpushed commit warning
- [x] 2.5 Update existing close tests to ensure they still pass

## 3. Documentation

- [x] 3.1 Update usage.md to document unpushed commit safety check
- [x] 3.2 Document --force flag behavior for bypassing safety checks
